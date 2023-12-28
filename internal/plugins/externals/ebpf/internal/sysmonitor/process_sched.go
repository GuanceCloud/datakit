//go:build linux
// +build linux

package sysmonitor

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"debug/elf"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/cilium/ebpf"
	"github.com/golang/groupcache/lru"
	pr "github.com/shirou/gopsutil/v3/process"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/tracing"

	"golang.org/x/sys/unix"
)

// #include "../c/process_sched/process_sched.h"
import "C"

type ProcessSchedC C.struct_rec_process_sched_status
type ProcInjectC C.struct_proc_inject

type ProcessSchedWithFNameC C.struct_rec_process_sched_status_with_filename

type perfHandler func(cpu int, data []byte, perfmap *manager.PerfMap, manager *manager.Manager)

const (
	SchedFork = 0b1 << iota
	SchedExec
	SchedExit
)

const (
	bmapProcInject = "bmap_procinject"
)

type ProcessInfo struct {
	Pid   int32
	PName string

	// (LWP) thread info
	// ProcessInfo map[int32]*ProcessInfo

	ENV             map[string]string
	CMD             string
	ExePath         string
	ExeResolvedPath string
	AttachUProbe    bool
}

type ProcessAttachInfo struct {
	cannotAttach *lru.Cache
	attach       *lru.Cache
	sync.RWMutex
}

func (procAttach *ProcessAttachInfo) AddAtach(name string, tn time.Time) {
	procAttach.Lock()
	defer procAttach.Unlock()
	if procAttach.attach == nil {
		procAttach.attach = lru.New(2048)
	}
	if procAttach.cannotAttach == nil {
		procAttach.cannotAttach = lru.New(2048)
	}

	procAttach.cannotAttach.Remove(name)
	procAttach.attach.Add(name, tn)
}

func (procAttach *ProcessAttachInfo) AddCannotAttach(name string, tn time.Time) {
	procAttach.Lock()
	defer procAttach.Unlock()
	if procAttach.attach == nil {
		procAttach.attach = lru.New(2048)
	}
	if procAttach.cannotAttach == nil {
		procAttach.cannotAttach = lru.New(2048)
	}
	procAttach.cannotAttach.Add(name, tn)
}

func (procAttach *ProcessAttachInfo) GetAttachInfo(name string) (time.Time, bool) {
	procAttach.RLock()
	defer procAttach.RUnlock()
	if procAttach.attach == nil {
		procAttach.attach = lru.New(2048)
	}
	if procAttach.cannotAttach == nil {
		procAttach.cannotAttach = lru.New(2048)
	}
	if v, ok := procAttach.attach.Get(name); ok {
		if v, ok := v.(time.Time); ok {
			return v, true
		}
	}

	return time.Time{}, false
}

func (procAttach *ProcessAttachInfo) GetCannotAndAttachInfo(name string) (time.Time, bool) {
	procAttach.RLock()
	defer procAttach.RUnlock()
	if procAttach.attach == nil {
		procAttach.attach = lru.New(2048)
	}
	if procAttach.cannotAttach == nil {
		procAttach.cannotAttach = lru.New(2048)
	}
	if v, ok := procAttach.attach.Get(name); ok {
		if v, ok := v.(time.Time); ok {
			return v, true
		}
	}

	if v, ok := procAttach.cannotAttach.Get(name); ok {
		if v, ok := v.(time.Time); ok {
			return v, true
		}
	}
	return time.Time{}, false
}

var execGoFnName = []string{
	"uprobe__go_runtime_execute",
}

func NewProcessSchedTracer(filter *tracing.ProcessFilter) (*SchedTracer, error) {
	tracer := SchedTracer{
		processFilter: filter,
	}

	var err error
	tracer.Manager, err = NewSchedManger(tracer.ProcessSchedHandler)
	if err != nil {
		return nil, err
	}

	return &tracer, nil
}

type SchedTracer struct {
	Manager *manager.Manager

	processFilter *tracing.ProcessFilter
	attachInfo    ProcessAttachInfo

	sync.Mutex
}

func (tracer *SchedTracer) GetGOSchedMap() (map[string]*ebpf.Map, bool) {
	if tracer.Manager == nil {
		return nil, false
	}

	bmaps := map[string]*ebpf.Map{}

	// if m, ok, err := tracer.Manager.GetMap("bmap_goid2tid"); !ok || err != nil {
	// 	return nil, false
	// } else {
	// 	bmaps["bmap_goid2tid"] = m
	// }
	if m, ok, err := tracer.Manager.GetMap("bmap_tid2goid"); !ok || err != nil {
		return nil, false
	} else {
		bmaps["bmap_tid2goid"] = m
	}
	return bmaps, true
}

func (tracer *SchedTracer) Start(ctx context.Context) error {
	err := tracer.Manager.Start()
	if err != nil {
		return err
	}

	pses, err := pr.Processes()
	if err != nil {
		return nil
	}

	for _, p := range pses {
		if err := tracer.goProbeRegister(p); err != nil {
			l.Warn(err)
		}
	}

	return nil
}

func (tracer *SchedTracer) Stop() error {
	if err := tracer.Manager.Stop(manager.CleanAll); err != nil {
		return err
	}

	return nil
}

func NewSchedManger(handler perfHandler) (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sched_process_fork",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sched_process_exec",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sched_process_exit",
				},
			},
		},
		PerfMaps: []*manager.PerfMap{
			{
				Map: manager.Map{
					Name: "process_sched_event",
				},
				PerfMapOptions: manager.PerfMapOptions{
					PerfRingBufferSize: 32 * os.Getpagesize(),
					DataHandler:        handler,
				},
			},
		},
	}
	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}

	buf, err := dkebpf.ProcessSchedBin()
	if err != nil {
		return nil, fmt.Errorf("conntrack.o: %w", err)
	}

	if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, fmt.Errorf("init process sched tracer: %w", err)
	}

	return m, nil
}

func (tracer *SchedTracer) ProcessSchedHandler(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager) {
	evetC := (*ProcessSchedC)(unsafe.Pointer(&data[0]))

	switch evetC.status {
	case SchedFork:
	case SchedExec:
		// eventC := (*ProcessSchedWithFNameC)(unsafe.Pointer(&data[0]))
		// pid := eventC.sched_status.nxt_pid

		p, err := pr.NewProcess(int32(evetC.nxt_pid))
		if err != nil {
			break
		}

		if err := tracer.goProbeRegister(p); err != nil {
			l.Debug(err)
		}
	case SchedExit:
		if tracer.processFilter != nil {
			tracer.processFilter.Delete(int(evetC.nxt_pid))
		}
	default:
		return
	}
}

func (tracer *SchedTracer) goProbeRegister(p *pr.Process) error {
	pname, err := p.Name()
	if err != nil {
		return err
	}
	env, err := p.Environ()
	if err != nil {
		l.Debug(err)
		return nil
	}

	envMap := map[string]string{}
	for _, v := range env {
		s := strings.Index(v, "=")
		if s > 0 {
			envMap[v[:s]] = v[s+1:]
		}
	}
	exePath, err := p.Exe()
	if err != nil {
		l.Debug(err)
		return nil
	}

	pid := p.Pid
	exeResolvePath := resolveBinPath(int(pid), exePath)
	if exeResolvePath == "" {
		return nil
	}

	exeResolvePath = HostRoot(exeResolvePath)

	if tracer.processFilter != nil {
		if !tracer.processFilter.Filter(int(pid), pname, exePath, envMap) {
			return nil
		}
	}

	// check file modified

	exeFstat, err := os.Stat(exeResolvePath)
	if err != nil {
		return err
	}
	exeModTime := exeFstat.ModTime()

	if tmod, ok := tracer.attachInfo.GetCannotAndAttachInfo(exeResolvePath); ok {
		if tmod.Equal(exeFstat.ModTime()) {
			return nil
		}
	}

	var goVer = [2]int{}
	inf, err := buildinfo.ReadFile(exeResolvePath)
	if err != nil {
		l.Debug(err)
		// if the go version is greater than 1.13+, this function can get the go version

		// do not return, if we can find the symbol, just attach
		// we tested go1.5+(amd64)
	} else {
		goVer, _ = parseGoVersion(inf.GoVersion)
	}

	var symbolAddr uint64 = 0

	elfFile, err := elf.Open(exeResolvePath)
	if err != nil {
		return fmt.Errorf("nnable to open elf file %s: %w", exeResolvePath, err)
	}

	if syms, err := FindSymbol(elfFile, "runtime.execute"); err == nil {
		if len(syms) != 1 {
			l.Debugf("find symbol runtime.execute, exe %s, count %d", exeResolvePath, len(syms))
			return nil
		}
		symbolAddr = syms[0].Value
	} else {
		sym, err := getGoUprobeSymbolFromPCLN(elfFile, goVer[1] >= 20, "runtime.execute")
		if err != nil {
			l.Debug(err)
			tracer.attachInfo.AddCannotAttach(exeResolvePath, exeModTime)
			return nil
		}
		symbolAddr = sym.Start
	}

	if tracer.Manager == nil {
		return nil
	}

	emap, ok, err := tracer.Manager.GetMap(bmapProcInject)
	if err != nil {
		return fmt.Errorf("get bpf map bmap_proc_inject failed: %w", err)
	}
	if !ok {
		l.Warn("get bpf map bmap_proc_inject failed")
	}

	// offset, err := FindMemberOffsetFromFile(fpath, "runtime.g", "goid")
	// if err != nil {
	// 	// go1.10 ~ 1.21: 152
	// 	offset = 152
	// }

	val := ProcInjectC{
		// go1.10(arm64, amd64) ~ 1.21; go1.5+(amd64): 152
		offset_go_runtime_g_goid: C.__u64(152),
		go_use_register:          0,
	}

	switch runtime.GOARCH {
	case "arm64":
		if goVer[1] >= 18 {
			val.go_use_register = 1
		}
	case "amd64":
		if goVer[1] >= 17 {
			val.go_use_register = 1
		}
	}

	pidU32 := (uint32)(pid)
	if err := emap.Update(unsafe.Pointer(&pidU32), unsafe.Pointer(&val), ebpf.UpdateAny); err != nil {
		return err
	}

	var uid string
	if tmod, ok := tracer.attachInfo.GetAttachInfo(exeResolvePath); ok {
		if tmod.Equal(exeModTime) {
			return nil
		}

		uid = ShortID(exeResolvePath)

		l.Info("DetachHook: file modfied: ", exeResolvePath, " ShortID: ", uid)
		for _, fnName := range execGoFnName {
			p, ok := tracer.Manager.GetProbe(manager.ProbeIdentificationPair{
				UID:          uid,
				EBPFFuncName: fnName,
			})
			if !ok {
				continue
			}
			if err := tracer.Manager.DetachHook(manager.ProbeIdentificationPair{
				UID:          uid,
				EBPFFuncName: fnName,
			}); err != nil {
				l.Error(err)
			}
			pp := p.Program()
			if pp != nil {
				if err := pp.Close(); err != nil {
					l.Warn(err)
				}
			}
		}
	}

	tracer.attachInfo.AddAtach(exeResolvePath, exeModTime)

	if uid == "" {
		uid = ShortID(exeResolvePath)
	}

	l.Info("AddHook: ", exeResolvePath, " ShortID: ", uid)
	for _, fnName := range execGoFnName {
		if err := tracer.Manager.AddHook("", &manager.Probe{
			ProbeIdentificationPair: manager.ProbeIdentificationPair{
				UID:          uid,
				EBPFFuncName: fnName,
			},
			UprobeOffset: symbolAddr,
			BinaryPath:   exeResolvePath,
		}); err != nil {
			l.Warn(err)
		}
	}

	return nil
}
