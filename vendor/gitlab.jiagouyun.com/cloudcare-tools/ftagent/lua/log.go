package lua

import (
	"bytes"
	"io"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type Log struct {
	io.Writer
}

func NewLog(w io.Writer) *Log {
	return &Log{w}
}

func (lg *Log) logInfo(L *lua.LState) int {
	str := L.ToString(1)
	var b bytes.Buffer
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString(" [info] ")
	b.WriteString(str)
	b.WriteString("\n")
	lg.Write(b.Bytes())
	return 0
}

func (lg *Log) logDebug(L *lua.LState) int {
	str := L.ToString(1)
	var b bytes.Buffer
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString(" [debug] ")
	b.WriteString(str)
	b.WriteString("\n")
	lg.Write(b.Bytes())
	return 0
}

func (lg *Log) logWarn(L *lua.LState) int {
	str := L.ToString(1)
	var b bytes.Buffer
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString(" [warn] ")
	b.WriteString(str)
	b.WriteString("\n")
	lg.Write(b.Bytes())
	return 0
}

func (lg *Log) logError(L *lua.LState) int {
	str := L.ToString(1)
	var b bytes.Buffer
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString(" [error] ")
	b.WriteString(str)
	b.WriteString("\n")
	lg.Write(b.Bytes())
	return 0
}
