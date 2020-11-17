package luascript

import (
	"errors"
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

func ToBasic(value lua.LValue) interface{} {
	switch value.(type) {
	case lua.LString:
		return value.String()
	case lua.LNumber:
		return value.(lua.LNumber)
	case lua.LBool:
		return value.(lua.LBool)
	case *lua.LTable:
		val := value.(*lua.LTable)
		m := make(map[interface{}]interface{})
		val.ForEach(func(key, element lua.LValue) {
			m[ToBasic(key)] = ToBasic(element)
		})
		return m
	default:
		return nil
	}
}

func ToLValue(l *lua.LState, value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}
	if lval, ok := value.(lua.LValue); ok {
		return lval
	}

	switch val := reflect.ValueOf(value); val.Kind() {
	case reflect.Bool:
		return lua.LBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(val.Uint())
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(val.Float())
	case reflect.String:
		return lua.LString(val.String())
	case reflect.Map:
		if val.IsNil() {
			return lua.LNil
		}
		tb := l.NewTable()
		valMap := val.MapRange()
		for valMap.Next() {
			tb.RawSet(ToLValue(l, valMap.Key().Interface()), ToLValue(l, valMap.Value().Interface()))
		}
		return tb
	case reflect.Slice:
		if val.IsNil() {
			return lua.LNil
		}
		fallthrough
	case reflect.Array:
		tb := l.NewTable()
		for i := 0; i < val.Len(); i++ {
			tb.Append(ToLValue(l, val.Index(i).Interface()))
		}
		return tb
	case reflect.Ptr:
		if val.IsNil() {
			return lua.LNil
		}
		return ToLValue(l, val.Elem().Interface())
	case reflect.Struct:
		fallthrough
	default:
		ud := l.NewUserData()
		ud.Value = val.Interface()
		return ud
	}
}

// SeSendToLua 将数据以实参方式，发送到 lua 指定的函数中
// callbakTypeName 目前看来是多余的
// TODO: 支持可变数量的实参，支持动态修改函数名
func SendToLua(l *lua.LState, val lua.LValue, callbackFnName, callbackTypeName string) (lua.LValue, error) {
	l.SetMetatable(val, l.GetTypeMetatable(callbackTypeName))
	lv := l.GetGlobal(callbackFnName)

	switch lv.(type) {
	case *lua.LFunction:
	default:
		return nil, errors.New("invalid lua function: " + callbackFnName)
	}

	gf := lv.(*lua.LFunction)
	if err := l.CallByParam(lua.P{Fn: gf, NRet: 1, Protect: true}, val); err != nil {
		return nil, err
	}

	result := l.Get(-1)
	l.Pop(1)
	return result, nil
}
