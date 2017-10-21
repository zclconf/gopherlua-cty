package luacty

import (
	"fmt"
	"testing"

	lua "github.com/yuin/gopher-lua"
	"github.com/zclconf/go-cty/cty"
)

func TestConverterWrapCtyValue(t *testing.T) {
	tests := map[string]struct {
		Vals   map[string]cty.Value
		Assert string
	}{
		"equal": {
			map[string]cty.Value{
				"a": cty.StringVal("hello"),
				"b": cty.StringVal("hello"),
			},
			`
				assert(a == b)
			`,
		},
		"not equal": {
			map[string]cty.Value{
				"a": cty.StringVal("hello"),
				"b": cty.StringVal("world"),
			},
			`
				assert(a ~= b)
			`,
		},
		"not equal with native lua value": {
			map[string]cty.Value{
				"a": cty.StringVal("hello"),
			},
			`
				assert(a ~= "hello")
			`,
		},

		"less than": {
			map[string]cty.Value{
				"a": cty.NumberIntVal(0),
				"b": cty.NumberIntVal(2),
			},
			`
				assert(a < b)
			`,
		},
		"less than (false)": {
			map[string]cty.Value{
				"a": cty.NumberIntVal(2),
				"b": cty.NumberIntVal(0),
			},
			`
				assert(not (a < b))
			`,
		},
		"less than or equal": {
			map[string]cty.Value{
				"a": cty.NumberIntVal(0),
				"b": cty.NumberIntVal(0),
			},
			`
				assert(a <= b)
			`,
		},
		"greater than": {
			map[string]cty.Value{
				"a": cty.NumberIntVal(2),
				"b": cty.NumberIntVal(0),
			},
			`
				assert(a > b)
			`,
		},

		"add": {
			map[string]cty.Value{
				"a":    cty.NumberIntVal(2),
				"b":    cty.NumberIntVal(3),
				"want": cty.NumberIntVal(5),
			},
			`
				result = a + b
				assert(result == want)
			`,
		},
		"add with string": {
			map[string]cty.Value{
				"a":    cty.NumberIntVal(2),
				"b":    cty.StringVal("3"),
				"want": cty.NumberIntVal(5),
			},
			`
				result = a + b
				assert(result == want)
			`,
		},
		"add with lua number": {
			map[string]cty.Value{
				"a":    cty.NumberIntVal(2),
				"want": cty.NumberIntVal(5),
			},
			`
				result = a + 3
				assert(result == want)
			`,
		},
		"subtract": {
			map[string]cty.Value{
				"a":    cty.NumberIntVal(2),
				"b":    cty.NumberIntVal(3),
				"want": cty.NumberIntVal(-1),
			},
			`
				result = a - b
				assert(result == want)
			`,
		},
		"multiply": {
			map[string]cty.Value{
				"a":    cty.NumberIntVal(2),
				"b":    cty.NumberIntVal(3),
				"want": cty.NumberIntVal(6),
			},
			`
				result = a * b
				assert(result == want)
			`,
		},
		"divide": {
			map[string]cty.Value{
				"a":    cty.NumberIntVal(6),
				"b":    cty.NumberIntVal(2),
				"want": cty.NumberIntVal(3),
			},
			`
				result = a / b
				assert(result == want)
			`,
		},
		"modulo": {
			map[string]cty.Value{
				"a":    cty.NumberIntVal(7),
				"b":    cty.NumberIntVal(2),
				"want": cty.NumberIntVal(1),
			},
			`
				result = a % b
				assert(result == want)
			`,
		},
		"negate": {
			map[string]cty.Value{
				"a":    cty.NumberIntVal(7),
				"want": cty.NumberIntVal(-7),
			},
			`
				result = -a
				assert(result == want)
			`,
		},

		"concat": {
			map[string]cty.Value{
				"a":    cty.StringVal("hey"),
				"b":    cty.StringVal("world"),
				"want": cty.StringVal("heyworld"),
			},
			`
				result = a .. b
				assert(result == want)
			`,
		},
		"concat with lua string": {
			map[string]cty.Value{
				"a":    cty.StringVal("hey"),
				"want": cty.StringVal("hey world"),
			},
			`
				result = a .. " world"
				assert(result == want)
			`,
		},
		"concat with lua number": {
			map[string]cty.Value{
				"a":    cty.StringVal("hey"),
				"want": cty.StringVal("hey6"),
			},
			`
				result = a .. 6
				assert(result == want)
			`,
		},
		"concat with number": {
			map[string]cty.Value{
				"a":    cty.StringVal("hey"),
				"b":    cty.NumberIntVal(2),
				"want": cty.StringVal("hey2"),
			},
			`
				result = a .. b
				assert(result == want)
			`,
		},
		"concat with bool": {
			map[string]cty.Value{
				"a":    cty.StringVal("hey"),
				"b":    cty.True,
				"want": cty.StringVal("heytrue"),
			},
			`
				result = a .. b
				assert(result == want)
			`,
		},

		"len with string": {
			map[string]cty.Value{
				"a":    cty.StringVal("hey"),
				"want": cty.NumberIntVal(3),
			},
			`
				result = #a
				assert(result == want)
			`,
		},
		"len with tuple": {
			map[string]cty.Value{
				"a":    cty.EmptyTupleVal,
				"want": cty.NumberIntVal(0),
			},
			`
				result = #a
				assert(result == want)
			`,
		},

		"index into map": {
			map[string]cty.Value{
				"a":    cty.MapVal(map[string]cty.Value{"greeting": cty.StringVal("hello")}),
				"want": cty.StringVal("hello"),
			},
			`
				result = a.greeting
				assert(result == want)
			`,
		},
		"index into map (absent key)": {
			map[string]cty.Value{
				"a": cty.MapValEmpty(cty.String),
			},
			`
				result = a.greeting
				assert(result == nil)
			`,
		},
		"index into list": {
			map[string]cty.Value{
				"a":    cty.ListVal([]cty.Value{cty.StringVal("hello")}),
				"want": cty.StringVal("hello"),
			},
			`
				result = a[0]
				assert(result == want)
			`,
		},
		"index into list (absent index)": {
			map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{cty.StringVal("hello")}),
			},
			`
				result = a[1]
				assert(result == nil)
			`,
		},
		"index into object": {
			map[string]cty.Value{
				"a":    cty.ObjectVal(map[string]cty.Value{"greeting": cty.StringVal("hello")}),
				"want": cty.StringVal("hello"),
			},
			`
				result = a.greeting
				assert(result == want)
			`,
		},
		"index into object (absent attr)": {
			map[string]cty.Value{
				"a": cty.EmptyObjectVal,
			},
			`
				result = a.greeting
				assert(result == nil)
			`,
		},
		"index into tuple": {
			map[string]cty.Value{
				"a":    cty.TupleVal([]cty.Value{cty.StringVal("hello")}),
				"want": cty.StringVal("hello"),
			},
			`
				result = a[0]
				assert(result == want)
			`,
		},
		"index into tuple (absent index)": {
			map[string]cty.Value{
				"a": cty.TupleVal([]cty.Value{cty.StringVal("hello")}),
			},
			`
				result = a[1]
				assert(result == nil)
			`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			L := lua.NewState(lua.Options{
				SkipOpenLibs: true,
			})
			addTestFuncs(L, t)

			conv := NewConverter(L)

			for n, v := range test.Vals {
				L.SetGlobal(n, conv.WrapCtyValue(v))
			}

			err := L.DoString(test.Assert)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func addTestFuncs(L *lua.LState, t *testing.T) {
	print := L.NewFunction(func(L *lua.LState) int {
		val := L.CheckString(1)
		caller, _ := L.GetStack(1)
		L.GetInfo("l", caller, nil)
		t.Logf("lua line %d: %s", caller.CurrentLine, val)
		return 0
	})
	assert := L.NewFunction(func(L *lua.LState) int {
		val := L.CheckBool(1)
		if !val {
			caller, _ := L.GetStack(1)
			L.GetInfo("l", caller, nil)
			t.Errorf("assertion failed at lua line %d", caller.CurrentLine)
		}
		return 0
	})
	require := L.NewFunction(func(L *lua.LState) int {
		val := L.CheckBool(1)
		if !val {
			caller, _ := L.GetStack(1)
			L.GetInfo("l", caller, nil)
			t.Fatalf("assertion failed at lua line %d", caller.CurrentLine)
		}
		return 0
	})
	dump := L.NewFunction(func(L *lua.LState) int {
		val := L.CheckAny(1)
		L.Push(lua.LString(fmt.Sprintf("%#v", val)))
		return 1
	})

	L.SetGlobal("print", print)
	L.SetGlobal("assert", assert)
	L.SetGlobal("require", require)
	L.SetGlobal("dump", dump)
}
