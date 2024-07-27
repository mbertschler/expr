package expr_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/internal/testify/assert"
	"github.com/expr-lang/expr/internal/testify/require"
)

func TestExpr_named_types(t *testing.T) {
	type namedIntType int16
	type namedFloatType float32
	type namedStringType string
	env := map[string]any{
		"namedInt":    func() namedIntType { return 42 },
		"basicInt":    func() int { return 42 },
		"namedFloat":  func() namedFloatType { return 3.0 },
		"basicFloat":  func() float32 { return 3.0 },
		"namedString": func() namedStringType { return "abc" },
		"basicString": func() string { return "abc" },
	}

	cases := []struct {
		code     string
		expected bool
	}{
		{`namedFloat() == 3.0`, true},
		{`namedInt() == 42`, true},
		{`basicInt() == 42`, true},
		{`basicFloat() == 3.0`, true},
		{`namedString() == "abc"`, true},
		{`basicString() == "abc"`, true},
		{`namedInt() == 42 && namedFloat() == 3.0 && namedString() == "abc"`, true},
	}
	for _, c := range cases {
		program, err := expr.Compile(c.code,
			expr.Env(env),
			expr.Function("unpack", unpack),
			// expr.Patch(namedTypesPatcher{}),
			expr.Patch(unpackPatcher{}),
			// expr.Patch(printPatcher{}),
		)
		require.NoError(t, err)

		val, err := expr.Run(program, env)
		require.NoError(t, err)
		assert.Equal(t, c.expected, val.(bool), c.code)
	}
}

type printPatcher struct{}

func (p printPatcher) Visit(node *ast.Node) {
	n := *node
	// fmt.Printf("print node T %T: %s type %s kind %s nature %v | %#v\n", n, n.String(), n.Type().Name(), n.Type().Kind(), n.Nature(), n)
	fmt.Printf("node T %T: %s type: %s kind: %s\n", n, n.String(), n.Type().Name(), n.Type().Kind())
}

type unpackPatcher struct{}

func (p unpackPatcher) Visit(node *ast.Node) {
	n := *node

	switch n.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:

		if n.Type().Kind().String() == n.Type().Name() {
			// type is basic and not named
			return
		}

		wrapper := ast.CallNode{
			Callee:    &ast.IdentifierNode{Value: "unpack"},
			Arguments: []ast.Node{*node},
		}
		ast.Patch(node, &wrapper)
	}
}

type namedTypesPatcher struct{}

func (p namedTypesPatcher) Visit(node *ast.Node) {
	n := *node

	if n.Type().Kind().String() == n.Type().Name() {
		// type is basic and not named
		return
	}

	switch n.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p.wrapInt(node)
	case reflect.Float32, reflect.Float64:
		p.wrapFloat(node)
	case reflect.String:
		p.wrapString(node)
	}
}

func unpack(params ...any) (any, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("no params")
	}
	return unpackBasicType(params[0]), nil
}

func unpackBasicType(in any) any {
	v := reflect.ValueOf(in)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		return v.String()
	}
	// can't unpack
	return in
}

func (p namedTypesPatcher) wrapInt(node *ast.Node) {
	wrapper := ast.BuiltinNode{
		Name:      "int",
		Arguments: []ast.Node{*node},
	}
	ast.Patch(node, &wrapper)
}

func (p namedTypesPatcher) wrapFloat(node *ast.Node) {
	wrapper := ast.BuiltinNode{
		Name:      "float",
		Arguments: []ast.Node{*node},
	}
	ast.Patch(node, &wrapper)
}

func (p namedTypesPatcher) wrapString(node *ast.Node) {
	wrapper := ast.BuiltinNode{
		Name:      "string",
		Arguments: []ast.Node{*node},
	}
	ast.Patch(node, &wrapper)
}
