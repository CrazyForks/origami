package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/php-any/origami/data"
	"github.com/php-any/origami/parser"
	"github.com/php-any/origami/runtime"
	"github.com/php-any/origami/std"
	netdata "github.com/php-any/origami/std/net/data"
)

func TestResolvePathParamInt(t *testing.T) {
	p := parser.NewParser()
	vm := runtime.NewVM(p).(*runtime.VM)
	std.Load(vm)
	ctx := vm.CreateContext(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/product/42", nil)
	req.SetPathValue("id", "42")
	_, requestClass := beginRequest(req)
	reqProxy := data.NewProxyValue(requestClass, ctx)

	spec := netdata.HandlerSpec{
		Params: []netdata.ParamBinding{
			{Name: "id", TypeFQN: "int", Source: netdata.SourcePath, Index: 0, PathKey: "id"},
		},
	}

	args, acl := resolveHandlerArgs(vm, ctx, spec, reqProxy, nil)
	if acl != nil {
		t.Fatalf("resolve failed: %v", acl)
	}
	iv, ok := args[0].(*data.IntValue)
	if !ok || iv.Value != 42 {
		t.Fatalf("want int 42, got %#v", args[0])
	}
}

func TestResolveQueryParamString(t *testing.T) {
	p := parser.NewParser()
	vm := runtime.NewVM(p).(*runtime.VM)
	std.Load(vm)
	ctx := vm.CreateContext(nil)

	req := httptest.NewRequest(http.MethodGet, "/hello?name=Origami", nil)
	_, requestClass := beginRequest(req)
	reqProxy := data.NewProxyValue(requestClass, ctx)

	spec := netdata.HandlerSpec{
		Params: []netdata.ParamBinding{
			{Name: "name", TypeFQN: "string", Source: netdata.SourceQuery, Index: 0, QueryKey: "name"},
		},
	}

	args, acl := resolveHandlerArgs(vm, ctx, spec, reqProxy, nil)
	if acl != nil {
		t.Fatalf("resolve failed: %v", acl)
	}
	sv, ok := args[0].(*data.StringValue)
	if !ok || sv.AsString() != "Origami" {
		t.Fatalf("want name Origami, got %#v", args[0])
	}
}
