package module

import (
	"encoding/json"
	"strings"

	xj "github.com/basgys/goxml2json"
	lua "github.com/yuin/gopher-lua"
)

func xmlDecode(L *lua.LState) int {
	str := L.CheckString(1)

	// 转成json buffer，类型判断在转LTable时进行
	// Convert() 无需传入xj.WithTypeConverter(xj.Float)
	jsonBuff, err := xj.Convert(strings.NewReader(str))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Encode returns the xml encoding of value.
	value, err := func() (lua.LValue, error) {
		var value interface{}
		err := json.Unmarshal(jsonBuff.Bytes(), &value)
		if err != nil {
			return nil, err
		}

		return jsonDecodeValue(L, value), nil
	}()

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(value)
	return 1

}
