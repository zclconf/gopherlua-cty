package luacty

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// ToCtyValue attempts to convert the given Lua value to a cty Value of the
// given type.
//
// If the given type is cty.DynamicPseudoType then this method will select
// a cty type automatically based on the Lua value type, which is an obvious
// mapping for most types but note that Lua tables are always converted to
// object types unless specifically typed other wise.
//
// If the requested conversion is not possible -- because the given Lua value
// is not of a suitable type for the target type -- the result is cty.DynamicVal
// and an error is returned.
//
// Not all Lua types have corresponding cty types; those that don't will
// produce an error regardless of the target type.
//
// Error messages are written with a Lua developer as the audience, and so
// will not include Go-specific implementation details. Where possible, the
// result is a cty.PathError describing the location of the error within
// the given data structure.
func (c *Converter) ToCtyValue(val lua.LValue, ty cty.Type) (cty.Value, error) {
	// 'path' starts off as empty but will grow for each level of recursive
	// call we make, so by the time toCtyValue returns it is likely to have
	// unused capacity on the end of it, depending on how deeply-recursive
	// the given Type is.
	path := make(cty.Path, 0)
	return c.toCtyValue(val, ty, path)
}

func (c *Converter) toCtyValue(val lua.LValue, ty cty.Type, path cty.Path) (cty.Value, error) {
	if val.Type() == lua.LTNil {
		return cty.NullVal(ty), nil
	}

	if ty == cty.DynamicPseudoType {
		// Choose a type automatically
		var err error
		ty, err = c.impliedCtyType(val, path)
		if err != nil {
			return cty.DynamicVal, err
		}
	}

	// If the value is a userdata produced by this package then we will
	// unwrap it and attempt conversion using the standard cty conversion
	// logic.
	if val.Type() == lua.LTUserData {
		ud := val.(*lua.LUserData)
		if ctyV, isCty := ud.Value.(cty.Value); isCty {
			ret, err := convert.Convert(ctyV, ty)
			if err != nil {
				return cty.DynamicVal, path.NewError(err)
			}
			return ret, nil
		}
	}

	// If we have a native Lua value, our conversion strategy depends on our
	// target type, now that we've picked one.
	switch {
	case ty == cty.Bool:
		nv := lua.LVAsBool(val)
		return cty.BoolVal(nv), nil
	case ty == cty.Number:
		switch val.Type() {
		case lua.LTNumber:
			nv := float64(val.(lua.LNumber))
			return cty.NumberFloatVal(nv), nil
		default:
			dyVal, err := c.toCtyValue(val, cty.DynamicPseudoType, path)
			if err != nil {
				return cty.DynamicVal, err
			}
			numV, err := convert.Convert(dyVal, cty.Number)
			if err != nil {
				return cty.DynamicVal, path.NewError(err)
			}
			return numV, nil
		}
	case ty == cty.String:
		switch val.Type() {
		case lua.LTString:
			nv := string(val.(lua.LString))
			return cty.StringVal(nv), nil
		default:
			dyVal, err := c.toCtyValue(val, cty.DynamicPseudoType, path)
			if err != nil {
				return cty.DynamicVal, err
			}
			strV, err := convert.Convert(dyVal, cty.String)
			if err != nil {
				return cty.DynamicVal, path.NewError(err)
			}
			return strV, nil
		}
	default:
		return cty.DynamicVal, path.NewErrorf("%s values are not allowed", val.Type().String())
	}
}

// ImpliedCtyType attempts to produce a cty Type that is suitable to recieve
// the given Lua value, or returns an error if no mapping is possible.
//
// Error messages are written with a Lua developer as the audience, and so
// will not include Go-specific implementation details. Where possible, the
// result is a cty.PathError describing the location of the error within
// the given data structure.
func (c *Converter) ImpliedCtyType(val lua.LValue) (cty.Type, error) {
	path := make(cty.Path, 0)
	return c.impliedCtyType(val, path)
}

func (c *Converter) impliedCtyType(val lua.LValue, path cty.Path) (cty.Type, error) {
	switch val.Type() {

	case lua.LTNil:
		return cty.DynamicPseudoType, nil

	case lua.LTBool:
		return cty.Bool, nil

	case lua.LTNumber:
		return cty.Number, nil

	case lua.LTString:
		return cty.String, nil

	case lua.LTUserData:
		ud := val.(*lua.LUserData)
		if ctyV, isCty := ud.Value.(cty.Value); isCty {
			return ctyV.Type(), nil
		}

		// Other userdata types (presumably created by other packages) are not allowed
		return cty.DynamicPseudoType, path.NewErrorf("userdata values are not allowed")

	case lua.LTTable:
		table := val.(*lua.LTable)
		var err error

		// Make sure we have capacity in our path array for our key step
		path = append(path, cty.PathStep(nil))
		path = path[:len(path)-1]

		atys := make(map[string]cty.Type)

		table.ForEach(func(key lua.LValue, val lua.LValue) {
			if err != nil {
				return
			}
			keyCty, keyErr := c.ToCtyValue(key, cty.String)
			if keyErr != nil {
				err = path.NewErrorf("all table keys must be strings")
				return
			}
			attrName := keyCty.AsString()
			keyPath := append(path, cty.GetAttrStep{
				Name: attrName,
			})
			aty, valErr := c.impliedCtyType(val, keyPath)
			if valErr != nil {
				err = valErr
				return
			}
			atys[attrName] = aty
		})
		if err != nil {
			return cty.DynamicPseudoType, err
		}

		return cty.Object(atys), nil

	default:
		return cty.DynamicPseudoType, path.NewErrorf("%s values are not allowed", val.Type().String())

	}
}
