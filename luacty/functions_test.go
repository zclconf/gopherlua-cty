package luacty

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestConverterWrapCtyFunction(t *testing.T) {
	tests := map[string]struct {
		Funcs  map[string]function.Function
		Vals   map[string]cty.Value
		Assert string
	}{
		"simple": {
			map[string]function.Function{
				"upper": stdlib.UpperFunc,
			},
			map[string]cty.Value{
				"want": cty.StringVal("HELLO"),
			},
			`
				result = upper("hello")
				assert(result == want)
			`,
		},
		"variadic": {
			map[string]function.Function{
				"max": stdlib.MaxFunc,
			},
			map[string]cty.Value{
				"want": cty.NumberIntVal(10),
			},
			`
				result = max(1, 2, 10, 6)
				assert(result == want)
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

			for n, f := range test.Funcs {
				L.SetGlobal(n, conv.WrapCtyFunction(f))
			}
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
