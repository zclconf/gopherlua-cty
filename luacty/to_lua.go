package luacty

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// WrapCtyValue takes a cty Value and returns a Lua value (of type UserData)
// that represents the same value, with its metatable configured such that
// many Lua operations will delegate to the cty API.
//
// WrapCtyValue produces a result that stays as close as possible to cty
// semantics when used with other such wrapped values, but the result may
// not integrate well with native Lua values. For example, a wrapped cty.String
// value will not compare equal to any native Lua string.
func (c *Converter) WrapCtyValue(val cty.Value) lua.LValue {
	ret := c.lstate.NewUserData()
	ret.Value = val
	ret.Metatable = c.metatable
	return ret
}

func (c *Converter) ctyMetatable() *lua.LTable {
	L := c.lstate
	table := L.NewTable()

	table.RawSet(lua.LString("__eq"), c.lstate.NewFunction(c.ctyEq))
	table.RawSet(lua.LString("__add"), c.lstate.NewFunction(c.ctyArithmetic(stdlib.Add)))
	table.RawSet(lua.LString("__sub"), c.lstate.NewFunction(c.ctyArithmetic(stdlib.Subtract)))
	table.RawSet(lua.LString("__mul"), c.lstate.NewFunction(c.ctyArithmetic(stdlib.Multiply)))
	table.RawSet(lua.LString("__div"), c.lstate.NewFunction(c.ctyArithmetic(stdlib.Divide)))
	table.RawSet(lua.LString("__mod"), c.lstate.NewFunction(c.ctyArithmetic(stdlib.Modulo)))
	table.RawSet(lua.LString("__unm"), c.lstate.NewFunction(c.ctyNegate))

	return table
}

func (c *Converter) ctyEq(L *lua.LState) int {
	// On the stack we should have two LUserData values, because Lua
	// only calls __eq if both operands have the same type. However, we
	// don't know if both userdatas will be our own (other packages can
	// create UserData values too) and the user may call __eq directly,
	// so we will be defensive.
	a := L.CheckUserData(1)
	b := L.CheckUserData(2)
	L.Pop(2)

	if a == nil || b == nil {
		L.Push(lua.LBool(false))
		return 1
	}
	if _, isOurs := a.Value.(cty.Value); !isOurs {
		L.Push(lua.LBool(false))
		return 1
	}
	if _, isOurs := b.Value.(cty.Value); !isOurs {
		L.Push(lua.LBool(false))
		return 1
	}

	result := a.Value.(cty.Value).Equals(b.Value.(cty.Value))
	if result.IsKnown() {
		L.Push(lua.LBool(result.True()))
	} else {
		// Lua doesn't have the concept of an unknown bool, so we just
		// treat unknown result as false. (The result of eq is forced to
		// be a native lua bool, so we can't do better here.)
		L.Push(lua.LBool(false))
	}
	return 1
}

func (c *Converter) ctyArithmetic(op func(a, b cty.Value) (cty.Value, error)) lua.LGFunction {
	return func(L *lua.LState) int {
		aL := L.CheckAny(1)
		bL := L.CheckAny(2)

		a, err := c.ToCtyValue(aL, cty.Number)
		if err != nil {
			L.Error(lua.LString(err.Error()), 1)
		}
		b, err := c.ToCtyValue(bL, cty.Number)
		if err != nil {
			L.Error(lua.LString(err.Error()), 1)
		}

		result, err := op(a, b)
		if err != nil {
			L.Error(lua.LString(err.Error()), 1)
		}

		L.Push(c.WrapCtyValue(result))
		return 1
	}
}

func (c *Converter) ctyNegate(L *lua.LState) int {
	vL := L.CheckAny(1)

	v, err := c.ToCtyValue(vL, cty.Number)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}

	result, err := stdlib.Negate(v)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}

	L.Push(c.WrapCtyValue(result))
	return 1
}
