// Package tracing parse http header
package tracing

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
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

	ThrTraceid spanid.ID64

	ESpanType string

	TraceID      spanid.ID128
	ParentSpanID spanid.ID64
	HexEncode    bool
	HaveTracID   bool

	PidTid uint64

	ASpanSampled bool

	// TraceProvider string

	Headers map[string]string

	Param       string
	TaskComm    string
	ProcessName string
	Service     string
	AllowTrace  bool

	TS int64
}

func ParseHTTP1xHeader(payload []byte, ts int64, conv2dd bool) (*TraceInfo, bool) {
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
	uriAndParam := strings.Split(req[1], "?")

	uri := uriAndParam[0]

	switch {
	case len(uri) > 8 && (uri[:8] == "https://"):
		off := strings.Index(uri[8:], "/")
		if off == -1 {
			return nil, false
		} else {
			uri = uri[off+8:]
		}
	case len(uri) > 7 && (uri[:7] == "http://"):
		off := strings.Index(uri[7:], "/")
		if off == -1 {
			return nil, false
		}
		uri = uri[off+7:]
	case (len(uri) > 0) && (uri[:1] == "/"):
	default:
		return nil, false
	}

	tInfo := &TraceInfo{
		Method:       req[0],
		Path:         uri,
		Version:      req[2],
		TS:           ts,
		ASpanSampled: true,
		Headers:      map[string]string{},
	}

	if len(uriAndParam) == 2 {
		tInfo.Param = uriAndParam[1]
	}

	payload = payload[idx+2:]
	hdr := strings.Split(string(payload), "\r\n")
	for _, v := range hdr {
		kv := strings.SplitN(v, ":", 2)
		if len(kv) != 2 {
			break
		}
		if _, ok := tInfo.Headers[kv[0]]; !ok {
			tInfo.Headers[kv[0]] = strings.TrimSpace(kv[1])
		}
	}

	if v, ok := tInfo.Headers["traceparent"]; ok {
		traceParent := strings.Split(v, "-")
		if len(traceParent) == 4 {
			tInfo.ASpanSampled = sampledW3C(traceParent[3])
			trcidStr := traceParent[1]
			tInfo.TraceID = HexTraceid2ID128(trcidStr)
			tInfo.HaveTracID = true
			if conv2dd {
				tInfo.TraceID.High = 0
				tInfo.HexEncode = false
			} else {
				tInfo.HexEncode = true
			}
			prtidstr := traceParent[2]
			tInfo.ParentSpanID = HexSpanid2ID64(prtidstr)
		}
	} else if tid, ok := tInfo.Headers["x-datadog-trace-id"]; ok {
		tInfo.TraceID.Low = uint64(DecTraceOrSpanid2ID64(tid))
		if psid, ok := tInfo.Headers["x-datadog-parent-id"]; ok {
			tInfo.ParentSpanID = DecTraceOrSpanid2ID64(psid)
		}
		if v, ok := tInfo.Headers["x-datadog-sampling-priority"]; ok {
			tInfo.ASpanSampled = sampledDataDog(v)
		}
		tInfo.HexEncode = false
		tInfo.HaveTracID = true
	}

	return tInfo, true
}

func FormatSpanID(i uint64, base16 bool) string {
	if base16 {
		_id := make([]byte, 8)
		binary.BigEndian.PutUint64(_id, i)
		return hex.EncodeToString(_id)
	} else {
		return strconv.FormatUint(i, 10)
	}
}

func HexTraceid2ID128(s string) spanid.ID128 {
	if b, _ := hex.DecodeString(s); len(b) == 16 {
		return spanid.ID128{
			Low:  binary.BigEndian.Uint64(b[8:]),
			High: binary.BigEndian.Uint64(b[:8]),
		}
	}
	return spanid.ID128{}
}

func DecTraceOrSpanid2ID64(s string) spanid.ID64 {
	if strings.HasPrefix(s, "-") {
		id, _ := strconv.ParseInt(s, 10, 64)
		return spanid.ID64(id)
	} else {
		id, _ := strconv.ParseUint(s, 10, 64)
		return spanid.ID64(id)
	}
}

func HexSpanid2ID64(s string) spanid.ID64 {
	if b, _ := hex.DecodeString(s); len(b) == 8 {
		return spanid.ID64(binary.BigEndian.Uint64(b))
	} else {
		return 0
	}
}

type ProcessFilter struct {
	SvcAssignEnv []string
	RuleEnv      map[string]bool

	RuleProcessName map[string]bool
	RulePath        map[string]bool

	Pid2ProcInfo map[int]*ProcSvcInfo
	PidDeleted   *lru.Cache
	AnyProcess   bool
	Disable      bool
	sync.RWMutex
}

type ProcSvcInfo struct {
	Name       string
	Service    string
	AllowTrace bool
}

func NewProcessFilter(svcAssignEnv []string, ruleEnv map[string]bool, ruleProcessName map[string]bool,
	anyProcess, disable bool,
) *ProcessFilter {
	return &ProcessFilter{
		SvcAssignEnv:    svcAssignEnv,
		RuleEnv:         ruleEnv,
		RuleProcessName: ruleProcessName,

		RulePath: map[string]bool{},

		Pid2ProcInfo: map[int]*ProcSvcInfo{},
		PidDeleted:   lru.New(1024),

		AnyProcess: anyProcess,
		Disable:    disable,
	}
}

func (p *ProcessFilter) Filter(pid int, name, path string, env map[string]string) bool {
	p.Lock()
	defer p.Unlock()

	var filtered bool
	for i := 0; i < 1; i++ {
		for k, allow := range p.RuleEnv {
			if _, ok := env[k]; ok {
				if allow {
					filtered = true
				}
				break
			}
		}

		if allow, ok := p.RuleProcessName[name]; ok && allow {
			filtered = true
			break
		}

		if _, ok := p.RulePath[path]; ok {
			filtered = true
			break
		}

		if p.AnyProcess {
			filtered = true
			break
		}
	}

	if allow, ok := p.RuleProcessName[name]; ok && !allow {
		filtered = false
	}

	for k, allow := range p.RuleEnv {
		if _, ok := env[k]; ok && !allow {
			filtered = false
		}
	}

	if p.Disable {
		filtered = false
	}

	pinfo := &ProcSvcInfo{
		Name:       name,
		Service:    name,
		AllowTrace: filtered,
	}

	if len(env) > 0 && len(p.SvcAssignEnv) > 0 {
		for _, k := range p.SvcAssignEnv {
			if v, ok := env[k]; ok {
				pinfo.Service = v
				break
			}
		}
	}

	p.Pid2ProcInfo[pid] = pinfo

	return filtered
}

func (p *ProcessFilter) Delete(pid int) {
	p.Lock()
	defer p.Unlock()

	if v, ok := p.Pid2ProcInfo[pid]; ok {
		delete(p.Pid2ProcInfo, pid)
		p.PidDeleted.Add(pid, v)
	}
}

func (p *ProcessFilter) GetProcInfo(pid int) (*ProcSvcInfo, bool) {
	p.RLock()
	if len(p.Pid2ProcInfo) > 0 {
		if v, ok := p.Pid2ProcInfo[pid]; ok && v != nil {
			p.RUnlock()
			return v, true
		}
	}
	p.RUnlock()

	// search deleted proc info from lru map
	p.Lock()
	defer p.Unlock()
	if p.PidDeleted.Len() > 0 {
		if v, ok := p.PidDeleted.Get(pid); ok {
			if v, ok := v.(*ProcSvcInfo); ok && v != nil {
				return v, ok
			}
		}
	}

	return nil, false
}
