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

		"table to dynamic": {
			func(L *lua.LState) lua.LValue {
				return L.NewTable()
			},
			cty.DynamicPseudoType,
			cty.DynamicVal,
			true, // tables are not yet supported
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
