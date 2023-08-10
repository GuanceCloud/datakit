// Package tracing parse http header
package tracing

import (
	"bytes"
	"strings"
)

const (
	OpProbeUnknown     = "ebpf.unknown"
	OpSyscallRead      = "syscall.read"
	OpSyscallWrite     = "syscall.write"
	OpSycallSendto     = "syscall.sendto"
	OpSyscallRecvfrom  = "syscall.recvfrom"
	OpSyscallWritev    = "syscall.writev"
	OpSyscallReadv     = "syscall.readv"
	OpSyscallSendfile  = "syscall.sendfile"
	OpSyscallClose     = "syscall.close"
	OpUserFuncSSLRead  = "user.ssl_read"
	OpUserFuncSSLWrite = "user.ssl_write"

	SvcSyscall = "syscall"
)

const (
	PUnknown = iota
	PSyscallWrite
	PSyscallRead
	PSyscallSendto
	PSyscallRecvfrom
	PSyscallWritev
	PSyscallReadv
	PSyscallSendfile

	PSyacallClose

	PUsrSSLRead
	PUsrSSLWrite
)

var SysOP = func() func(uint16) string {
	mapping := map[uint16]string{
		PUnknown:         OpProbeUnknown,
		PSyscallWrite:    OpSyscallWrite,
		PSyscallRead:     OpSyscallRead,
		PSyscallSendto:   OpSycallSendto,
		PSyscallRecvfrom: OpSyscallRecvfrom,
		PSyscallWritev:   OpSyscallWritev,
		PSyscallReadv:    OpSyscallReadv,
		PUsrSSLRead:      OpUserFuncSSLRead,
		PUsrSSLWrite:     OpUserFuncSSLWrite,
		PSyscallSendfile: OpSyscallSendfile,
		PSyacallClose:    OpSyscallClose,
	}

	return func(id uint16) string {
		return mapping[id]
	}
}()

type TraceInfo struct {
	Host string

	Method  string
	Path    string
	Version string

	// param string

	TraceID      string
	ParentSpanID string

	Sample bool

	TraceProvider string

	TS int64
}

func ParseHTTP1xHeader(payload []byte, ts int64) (*TraceInfo, bool) {
	if payload[0] == '0' {
		return nil, false
	}

	idx := bytes.LastIndex(payload, []byte("\r\n\r\n"))
	if idx > 0 {
		payload = payload[:idx]
	} else if idx == 0 {
		return nil, false
	}

	// line 1
	idx = bytes.Index(payload, []byte("\r\n"))
	if idx < 0 {
		return nil, false
	}

	ln := payload[0:idx]

	req := strings.Split(string(ln), " ")
	if len(req) != 3 {
		return nil, false
	}
	uri := strings.Split(req[1], "?")

	tInfo := &TraceInfo{
		Method:  req[0],
		Path:    uri[0],
		Version: req[2],
		TS:      ts,
	}

	return tInfo, true
}
