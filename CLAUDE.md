# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**语言偏好：所有对话使用中文回答。**

## Build & Run

入口文件为根目录 `zy.go`，CLI 基于 `cobra`（`cmd/` 包）。

```bash
go build -o zy .              # 构建解释器（CI 产物名为 zy；也可命名为 origami）
./zy <script.php>             # 直接运行脚本（.zy 或 .php）
go run ./zy.go <script.php>   # 开发时推荐，无需先 build

# 子命令
./zy gen-std [-o .zy/std]     # 从 Go 标准库实现反射生成 PHP 伪代码（IDE 提示）
./zy compile [directory]      # 将 vendor 目录预编译为 Go 源码
./zy phpt [file-or-dir]       # 执行并验收 .phpt 测试文件
```

LSP 服务器为独立 Go module（`tools/lsp/`）：

```bash
cd tools/lsp && go build -o zy-lsp . && ./zy-lsp
# 或使用 ./build.sh
```

## Architecture

Origami-lang（折言）是 Go 实现的 PHP 类脚本语言解释器：词法分析 → 语法解析为 AST → 在 runtime VM 上执行。

### Pipeline

```
Source → lexer/ → tokens → parser/ → node/ (AST) → runtime/ (VM) executes nodes
```

### Core packages

| Package      | Role |
| ------------ | ---- |
| `token/`     | Token 类型枚举与关键字/运算符配置（`token.go`, `type_config.go`） |
| `lexer/`     | 词法分析：将 `.zy`/`.php` 源码 tokenize。入口 `lexer.go`；支持 HTML 块（`html_lexer.go`）、PHP 模式（`php_lexer.go`）、heredoc |
| `parser/`    | 递归下降解析器。`Parser`（`parser.go`）持有 token 流、scope manager、expression parser、class-path manager。语法规则通过 `all_parser.go` 的 `parserRouter` 注册 |
| `node/`      | AST 节点。所有语言结构实现 `data.GetValue`，同时承担结构与执行 |
| `data/`      | 核心接口与值类型：`VM`, `Context`, `GetValue`, `Types`, `ClassStmt`, `FuncStmt`，以及 `ZVal`, `ArrayValue`, `ObjectValue` 等 |
| `runtime/`   | `VM`（`vm.go`）为全局运行时容器；`TempVM`（`vm_temp.go`）提供请求级隔离（类似 php-fpm 每请求模型） |
| `cmd/`       | CLI 入口与子命令（`gen-std`, `compile`, `phpt`）；`cmd/runtime.go` 通过 `SetRuntimeLoader` 注入标准库加载 |
| `internal/`  | 内部工具，如 `pseudocode/`（`gen-std` 反射生成伪代码） |
| `std/`       | 标准库，通过 `std.Load` 及子包 `Load` 注册到 VM（见下方） |

### Standard library (`std/`)

`std.Load(vm)` 注册核心函数与基础类，并加载子模块：

| 子包 | 说明 |
| ---- | ---- |
| `php/` | PHP 内置函数（`empty`, `isset`, 字符串/数组、反射等）及子包 `pdo/`, `preg/`, `intl/`, `iconv/`, `attribute/` |
| `net/http/` | HTTP 服务器、路由、控制器 |
| `net/websocket/` | WebSocket 支持 |
| `net/annotation/` | HTTP 注解（路由、中间件等宏） |
| `database/` | 数据库访问；子包 `sql/`（`database/sql` 兼容层）、`migration/`（Schema 迁移）、`annotation/`（实体注解） |
| `container/` | 依赖注入容器及注解 |
| `validation/` | 数据校验 |
| `cli/` | CLI 运行时与注解 |
| `system/` | 系统相关（含 `os/`） |
| `channel/`, `context/`, `signal/`, `loop/` | 并发与信号 |
| `protowire/` | 协议解析 |
| `reflect/` | Go 反射桥接 |
| `exception/`, `log/` | 异常体系与日志 |

`zy.go` 的 `init` 额外加载：`php`, `http`, `websocket`, `net/annotation`, `system`（在 `std.Load` 之外）。

### Key interfaces (`data/` package)

- **`GetValue`** — `GetValue(ctx Context) (GetValue, Control)`。每个 AST 节点实现此接口
- **`Context`** — 变量作用域链，通过 `VM.CreateContext(vars)` 创建
- **`VM`** — 全局运行时：`AddClass`, `AddFunc`, `GetOrLoadClass`, `CreateContext`, `LoadAndRun`
- **`Types`** — 类型检查：`Is(value Value) bool`；实现包括 `BaseType`, `UnionType`, `NullableType`, `ClassType`

### Execution model

1. `zy.go` 通过 `cmd.SetRuntimeLoader` 注册标准库加载闭包，创建 `Parser` 与 `VM`
2. `VM.LoadAndRun(path)` 克隆 parser 解析文件，创建 `Context`，对顶层 AST 调用 `program.GetValue(ctx)`
3. `GetValue` 的第二个返回值 `Control` 表示非局部控制流：`ReturnControl`, `ThrowControl`, `BreakControl`, `ContinueControl`；`nil` 表示正常执行

### Class autoloading

类从文件系统自动发现。`ClassPathManager` 将类名映射到文件路径。类 `Foo\Bar` 须位于与命名空间对应的目录结构中，文件名为 `Bar.zy` 或 `Bar.php`（PSR-4 约定）。

### TempVM (request-level isolation)

`runtime.NewTempVM(baseVM)` 创建临时 VM：读取委托给 base VM，新的 class/interface/function 注册保存在本地。用于模拟 php-fpm 每请求模型；LSP 也用它解析文件而不污染全局 VM。

## Tests

### 脚本测试（`.zy` / `.php`）

`tests/` 下按主题分子目录（`basic/`, `func/`, `obj/`, `php/`, `operator/`, `net/`, `cli/`, `signal/`, `validation/` 等）。测试文件为 `.zy` 或 `.php`，通过打印输出人工验收，无断言框架；控制台红色输出表示失败。

```bash
go run ./zy.go tests/run_tests.zy   # 批量运行 tests/ 子目录中的 .zy 文件
go run ./zy.go tests/php/xxx.php    # 单独运行某个 .php 测试
```

注意：`tests/run_tests.zy` 目前仅扫描并 `include()` `.zy` 文件，不自动运行 `.php` 测试。

### Go 单元测试

```bash
go test ./...    # CI 使用此命令
```

Go 测试分布在 `lexer/`, `parser/`, `runtime/`, `cmd/compile/`, `std/`（如 `net/http/`, `database/migration/`, `validation/`）, `tools/lsp/` 等包中。

### Agent Skills

项目内 Cursor Agent Skills（`.cursor/skills/`）：

- `function-docs` — 函数表示、注册与执行
- `interpreter-docs` — VM、TempVM、子解析器
- `param-types-docs` — 参数类型语法与 `data.Types`
- `php-runtime-func-test` — 实现/修改 `std/php` 函数时须编写 `tests/php/` 测试并用 `go run ./zy.go` 验证
