package luacty

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zclconf/go-cty/cty"
)

// ToLuaValue takes a cty Value and returns a Lua value (of type UserData)
// that represents the same value, with its metatable configured such that
// many Lua operations will delegate to the cty API.
func (c Converter) ToLuaValue(val cty.Value) lua.LValue {
	return &lua.LUserData{
		Value:     val,
		Metatable: c.metatable,
	}
}

func (c Converter) ctyMetatable() lua.LValue {
	L := c.lstate
	table := L.NewTable()

	table.RawSet(lua.LString("__eq"), &lua.LFunction{
		Proto: &lua.FunctionProto{
			SourceName:    "__eq",
			NumParameters: 2,
		},
		GFunction: c.ctyEq,
	})

	return table
}

func (c Converter) ctyEq(L *lua.LState) int {
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
	L.Push(c.ToLuaValue(result))
	return 1
}
