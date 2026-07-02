package netdata

// BindingSource 描述 HTTP 控制器形参值的来源。
type BindingSource int

const (
	SourceRequest BindingSource = iota
	SourceResponse
	SourceDTO
	SourcePath
	SourceQuery
)

// ParamBinding 单形参绑定计划。
type ParamBinding struct {
	Name     string
	TypeFQN  string
	Source   BindingSource
	Index    int
	Validate bool
	PathKey  string // SourcePath 时对应路由模板中的 {key}
	QueryKey string // SourceQuery 时对应 ?key= 查询参数名
	Nullable bool
}

// HandlerSpec HTTP 控制器方法的参数解析计划。
type HandlerSpec struct {
	Params []ParamBinding
}

// ArgCount 返回应传入控制器方法的参数个数。
func (s HandlerSpec) ArgCount() int {
	return len(s.Params)
}
