package module

import (
	"bytes"
	"fmt"
	"io"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type lualog struct {
	io.Writer
}

func (lg *lualog) logInfo(L *lua.LState) int {
	lg.Write(logformat(L, " [info ] "))
	return 0
}

func (lg *lualog) logDebug(L *lua.LState) int {
	lg.Write(logformat(L, " [debug] "))
	return 0
}

func (lg *lualog) logWarn(L *lua.LState) int {
	lg.Write(logformat(L, " [warn ] "))
	return 0
}

func (lg *lualog) logError(L *lua.LState) int {
	lg.Write(logformat(L, " [error] "))
	return 0
}

func logformat(L *lua.LState, lv string) []byte {
	top := L.GetTop()
	args := make([]interface{}, 0, top)

	for i := 1; i <= top; i++ {
		args = append(args, sqlGetValue(L, i))
	}

	var b bytes.Buffer
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString(lv)
	for i, v := range args {
		b.WriteString(fmt.Sprintf("%v", v))
		if i != len(args)-1 {
			b.WriteString(" ")
		}
	}

	b.WriteString("\n")
	return b.Bytes()
}
