package module

import (
	"encoding/csv"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func csvDecode(L *lua.LState) int {
	str := L.CheckString(1)

	value, err := csvDecodeValue(L, str)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(value)
	return 1
}

func csvDecodeValue(L *lua.LState, value string) (*lua.LTable, error) {
	r := csv.NewReader(strings.NewReader(value))

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	tb := L.NewTable()

	// records[0] is header
	for i := 1; i < len(records); i++ {
		tc := L.NewTable()

		for k, v := range records[i] {
			if n, err := strconv.Atoi(v); err == nil {
				tc.RawSetString(records[0][k], lua.LNumber(n))
				continue
			}
			if n, err := strconv.ParseFloat(v, 64); err == nil {
				tc.RawSetString(records[0][k], lua.LNumber(n))
				continue
			}
			tc.RawSetString(records[0][k], lua.LString(v))
		}

		tb.Append(tc)
	}

	return tb, nil
}
