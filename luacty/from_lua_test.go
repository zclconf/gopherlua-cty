package luacty

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
	"github.com/zclconf/go-cty/cty"
)

func TestConverterToCtyValue(t *testing.T) {
	tests := map[string]struct {
		Cons func(*lua.LState) lua.LValue
		Type cty.Type
		Want cty.Value
		Err  bool
	}{
		"nil to string": {
			func(L *lua.LState) lua.LValue {
				return lua.LNil
			},
			cty.String,
			cty.NullVal(cty.String),
			false,
		},
		"nil to bool": {
			func(L *lua.LState) lua.LValue {
				return lua.LNil
			},
			cty.Bool,
			cty.NullVal(cty.Bool),
			false,
		},
		"nil to dynamic": {
			func(L *lua.LState) lua.LValue {
				return lua.LNil
			},
			cty.DynamicPseudoType,
			cty.NullVal(cty.DynamicPseudoType),
			false,
		},

		"string to string": {
			func(L *lua.LState) lua.LValue {
				return lua.LString("hello")
			},
			cty.String,
			cty.StringVal("hello"),
			false,
		},
		"string to number": {
			func(L *lua.LState) lua.LValue {
				return lua.LString("12")
			},
			cty.Number,
			cty.NumberIntVal(12),
			false,
		},
		"string to number (invalid)": {
			func(L *lua.LState) lua.LValue {
				return lua.LString("not a number")
			},
			cty.Number,
			cty.DynamicVal,
			true, // a number is required
		},
		"string to dynamic": {
			func(L *lua.LState) lua.LValue {
				return lua.LString("hello")
			},
			cty.DynamicPseudoType,
			cty.StringVal("hello"),
			false,
		},

		"number to number": {
			func(L *lua.LState) lua.LValue {
				return lua.LNumber(12)
			},
			cty.Number,
			cty.NumberIntVal(12),
			false,
		},
		"number to string": {
			func(L *lua.LState) lua.LValue {
				return lua.LNumber(12)
			},
			cty.String,
			cty.StringVal("12"),
			false,
		},
		"number to dynamic": {
			func(L *lua.LState) lua.LValue {
				return lua.LNumber(12)
			},
			cty.DynamicPseudoType,
			cty.NumberIntVal(12),
			false,
		},

		"bool to bool": {
			func(L *lua.LState) lua.LValue {
				return lua.LBool(true)
			},
			cty.Bool,
			cty.True,
			false,
		},
		"bool to dynamic": {
			func(L *lua.LState) lua.LValue {
				return lua.LBool(true)
			},
			cty.DynamicPseudoType,
			cty.True,
			false,
		},

		"table to object": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.RawSet(lua.LString("greeting"), lua.LString("hello"))
				return table
			},
			cty.Object(map[string]cty.Type{
				"greeting": cty.String,
			}),
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.StringVal("hello"),
			}),
			false,
		},
		"table to object (extra keys)": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.RawSet(lua.LString("greeting"), lua.LString("hello"))
				return table
			},
			cty.EmptyObject,
			cty.DynamicVal,
			true, // unexpected key "greeting"
		},
		"table to object (missing keys)": {
			func(L *lua.LState) lua.LValue {
				return L.NewTable()
			},
			cty.Object(map[string]cty.Type{
				"greeting": cty.String,
			}),
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.NullVal(cty.String),
			}),
			false,
		},
		"table to object (indices present)": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.Append(lua.LString("hello"))
				return table
			},
			cty.EmptyObject,
			cty.DynamicVal,
			true, // unexpected key "1"
		},
		"table to tuple": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.Append(lua.LString("hello"))
				table.Append(lua.LString("12"))
				return table
			},
			cty.Tuple([]cty.Type{
				cty.String,
				cty.Number,
			}),
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
				cty.NumberIntVal(12),
			}),
			false,
		},
		"table to tuple (extra indices present)": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.Append(lua.LString("hello"))
				return table
			},
			cty.EmptyTuple,
			cty.DynamicVal,
			true, // index 1 out of range
		},
		"table to tuple (insufficient indices present)": {
			func(L *lua.LState) lua.LValue {
				return L.NewTable()
			},
			cty.Tuple([]cty.Type{
				cty.String,
				cty.Number,
			}),
			cty.TupleVal([]cty.Value{
				cty.NullVal(cty.String),
				cty.NullVal(cty.Number),
			}),
			false,
		},
		"table to tuple (keys present)": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.RawSet(lua.LString("greeting"), lua.LString("hello"))
				return table
			},
			cty.Tuple([]cty.Type{
				cty.String,
				cty.Number,
			}),
			cty.DynamicVal,
			true, // unexpected key "greeting"
		},
		"table to map of string": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.RawSet(lua.LString("greeting"), lua.LString("hello"))
				table.Append(lua.LNumber(10))
				return table
			},
			cty.Map(cty.String),
			cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("hello"),
				"1":        cty.StringVal("10"),
			}),
			false,
		},
		"table to map of string (empty)": {
			func(L *lua.LState) lua.LValue {
				return L.NewTable()
			},
			cty.Map(cty.String),
			cty.MapValEmpty(cty.String),
			false,
		},
		"table to map of dynamic": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.RawSet(lua.LString("greeting"), lua.LString("hello"))
				table.Append(lua.LBool(true))
				return table
			},
			cty.Map(cty.DynamicPseudoType),
			cty.MapVal(map[string]cty.Value{
				"greeting": cty.StringVal("hello"),
				"1":        cty.StringVal("true"),
			}),
			false,
		},
		"table to map of dynamic (incompatible types)": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.RawSet(lua.LString("num"), lua.LNumber(10))
				table.RawSet(lua.LString("bool"), lua.LBool(true))
				return table
			},
			cty.Map(cty.DynamicPseudoType),
			cty.DynamicVal,
			true, // all values must be of the same type
		},
		"table to dynamic": {
			func(L *lua.LState) lua.LValue {
				table := L.NewTable()
				table.RawSet(lua.LString("greeting"), lua.LString("hello"))
				return table
			},
			cty.DynamicPseudoType,
			cty.ObjectVal(map[string]cty.Value{
				"greeting": cty.StringVal("hello"),
			}),
			false,
		},
		"table to bool": {
			func(L *lua.LState) lua.LValue {
				return L.NewTable()
			},
			cty.Bool,
			cty.True, // due to Lua bool conversion semantics
			false,
		},
		"table to string": {
			func(L *lua.LState) lua.LValue {
				return L.NewTable()
			},
			cty.String,
			cty.DynamicVal,
			true, // a string is required
		},

		"function to dynamic": {
			func(L *lua.LState) lua.LValue {
				return L.NewFunction(func(*lua.LState) int { return 0 })
			},
			cty.DynamicPseudoType,
			cty.DynamicVal,
			true, // function values are not allowed
		},

		"userdata (ours) to dynamic": {
			func(L *lua.LState) lua.LValue {
				ret := L.NewUserData()
				ret.Value = cty.StringVal("howdy")
				return ret
			},
			cty.DynamicPseudoType,
			cty.StringVal("howdy"),
			false,
		},
		"userdata (other than ours) to dynamic": {
			func(L *lua.LState) lua.LValue {
				return L.NewUserData()
			},
			cty.DynamicPseudoType,
			cty.DynamicVal,
			true, // userdata values are not allowed
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			L := lua.NewState()
			conv := NewConverter(L)

			vL := test.Cons(L)
			got, err := conv.ToCtyValue(vL, test.Type)
			if (err != nil) != test.Err {
				if test.Err {
					t.Errorf("conversion succeeded; want error")
				} else {
					t.Errorf("unexpected error: %s", err)
				}
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v", vL, got, test.Want)
			}
		})
	}
}
