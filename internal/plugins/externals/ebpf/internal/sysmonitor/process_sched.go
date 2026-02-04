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
	"sync"
	"unsafe"

	ebpfmanager "github.com/DataDog/ebpf-manager"
	"github.com/cilium/ebpf"
	"github.com/josharian/intern"
	pr "github.com/shirou/gopsutil/v3/process"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"

	"golang.org/x/sys/unix"
)

// #include "../c/process_sched/process_sched.h"
import "C"

type ProcessSchedC C.struct_rec_process_sched_status

type ProcFilterInfoC C.struct_proc_filter_info

func (sc *ProcessSchedC) Comm() string {
	return fmt.Sprint(*(*[16]byte)(unsafe.Pointer(&sc.comm[0])))
}

func (sc *ProcessSchedC) String() string {
	comm := *(*[16]byte)(unsafe.Pointer(&sc.comm[0]))
	return fmt.Sprintf("st %d, prv %d, nxt %d, name `%s`", sc.status,
		sc.prv_pid, sc.nxt_pid,
		string(bytes.TrimRight(comm[:], "\x00")))
}

type ProcInjectC C.struct_proc_inject

type ProcessSchedWithFNameC C.struct_rec_process_sched_status_with_filename

type perfHandler func(cpu int, data []byte, perfmap *ebpfmanager.PerfMap, manager *ebpfmanager.Manager)

const (
	SchedFork = 0b1 << iota
	SchedExec
	SchedExit
)

const (
	bmapProcInject = "bmap_procinject"
	bmapProcFilter = "bmap_proc_filter"
	bmapTid2Goid   = "bmap_tid2goid"
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
	fileUpdater PassiveFileUpdater
	sync.RWMutex
}

func NewProcessAttachInfo() *ProcessAttachInfo {
	return &ProcessAttachInfo{
		fileUpdater: *NewPassiveFileUpdater(),
	}
}

var execGoFnName = []string{
	"uprobe__go_runtime_execute",
}

func NewProcessSchedTracer(filter *ProcessFilter) (*SchedTracer, error) {
	tracer := SchedTracer{
		processFilter: filter,
		attachInfo:    NewProcessAttachInfo(),
	}

	var err error
	tracer.Manager, err = NewSchedManger(tracer.ProcessSchedHandler)
	if err != nil {
		return nil, err
	}

	return &tracer, nil
}

type SchedTracer struct {
	Manager *ebpfmanager.Manager

	processFilter *ProcessFilter
	attachInfo    *ProcessAttachInfo

	selfPid int

	sync.Mutex
}

func (tracer *SchedTracer) GetSchedMap() (map[string]*ebpf.Map, bool) {
	if tracer.Manager == nil {
		return nil, false
	}

	bmaps := map[string]*ebpf.Map{}

	if m, ok, err := tracer.Manager.GetMap(bmapTid2Goid); !ok || err != nil {
		return nil, false
	} else {
		bmaps[bmapTid2Goid] = m
	}

	if m, ok, err := tracer.Manager.GetMap(bmapProcFilter); !ok || err != nil {
		return nil, false
	} else {
		bmaps[bmapProcFilter] = m
	}

	return bmaps, true
}

type kernelProcFilter func(int)

func (tracer *SchedTracer) Start(ctx context.Context) error {
	tracer.selfPid = os.Getpid()
	err := tracer.Manager.Start()
	if err != nil {
		return err
	}

	fn := func() kernelProcFilter {
		if mp, ok, err := tracer.Manager.GetMap(bmapProcFilter); err == nil && ok {
			return func(pid int) {
				k := uint32(pid)
				v := ProcFilterInfoC{
					disable: 1,
				}
				if err := mp.Update(&k, &v, ebpf.UpdateAny); err != nil {
					log.Info(err)
				}
			}
		}
		return func(pid int) {}
	}()

	tracer.processFilter.setKernelProcFilter(fn)

	pses, err := pr.Processes()
	if err != nil {
		return nil
	}
	for _, p := range pses {
		if err := tracer.attachProcess(p); err != nil {
			log.Debug(err)
		}
	}

	return nil
}

func (tracer *SchedTracer) Stop() error {
	if err := tracer.Manager.Stop(ebpfmanager.CleanAll); err != nil {
		return err
	}

	return nil
}

func NewSchedManger(handler perfHandler) (*ebpfmanager.Manager, error) {
	m := &ebpfmanager.Manager{
		Probes: []*ebpfmanager.Probe{
			{
				ProbeIdentificationPair: ebpfmanager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sched_process_fork",
				},
			},
			{
				ProbeIdentificationPair: ebpfmanager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sched_process_exec",
				},
			},
			{
				ProbeIdentificationPair: ebpfmanager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sched_process_exit",
				},
			},
		},
		PerfMaps: []*ebpfmanager.PerfMap{
			{
				Map: ebpfmanager.Map{
					Name: "process_sched_event",
				},
				PerfMapOptions: ebpfmanager.PerfMapOptions{
					PerfRingBufferSize: 32 * os.Getpagesize(),
					DataHandler:        handler,
				},
			},
		},
	}
	mOpts := ebpfmanager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}

	buf, err := dkebpf.ProcessSchedBin()
	if err != nil {
		return nil, fmt.Errorf("process_sched.o: %w", err)
	}

	if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, fmt.Errorf("init process sched tracer: %w", err)
	}

	return m, nil
}

func (tracer *SchedTracer) ProcessSchedHandler(cpu int, data []byte,
	perfmap *ebpfmanager.PerfMap, manager *ebpfmanager.Manager) {
	evetC := (*ProcessSchedC)(unsafe.Pointer(&data[0]))

	switch evetC.status {
	case SchedFork:
	case SchedExec:
		p, err := pr.NewProcess(int32(evetC.nxt_pid))
		if err != nil {
			break
		}

		if err := tracer.attachProcess(p); err != nil {
			log.Debug(err)
		}
	case SchedExit:
		if tracer.processFilter != nil {
			if procInfo, ok := tracer.processFilter.GetProcInfo(int(evetC.nxt_pid)); ok {
				if procInfo.binPath != 0 {
					if binPath, shouldDetach := tracer.attachInfo.fileUpdater.Forget(procInfo.binPath); shouldDetach {
						uid := ShortID(binPath)
						for _, fnName := range execGoFnName {
							if err := tracer.Manager.DetachHook(ebpfmanager.ProbeIdentificationPair{
								UID:          uid,
								EBPFFuncName: fnName,
							}); err != nil {
								log.Warn(err)
							} else {
								log.Infof("DetachHook: %s, ShortID: %s, name: %s",
									binPath, uid, procInfo.name)
							}
						}
					}
				}
			}
			tracer.processFilter.Delete(int(evetC.nxt_pid))
		}
	default:
		return
	}
}

func (tracer *SchedTracer) attachProcess(p *pr.Process) error {
	if p.Pid <= 0 {
		return fmt.Errorf("pid <= 0")
	}

	binPath, procInfo, err := tracer.processFilter.TryAdd(int(p.Pid))
	if err != nil {
		return err
	}

	if procInfo.deleted {
		return nil
	}

	if tracer.Manager == nil {
		return nil
	}

	if !procInfo.TraceFilterd() {
		return nil
	}

	if binPath == "" {
		return nil
	}

	rec, ok, err := tracer.attachInfo.fileUpdater.Check(binPath, procInfo.binPath)
	if err != nil {
		return err
	} else if !ok {
		if rec.inj != nil {
			emap, ok, err := tracer.Manager.GetMap(bmapProcInject)
			if err != nil {
				return fmt.Errorf("get bpf map bmap_proc_inject failed: %w", err)
			}
			if !ok {
				log.Warn("get bpf map bmap_proc_inject failed")
			}
			pidU32 := (uint32)(p.Pid)
			if err := emap.Update(unsafe.Pointer(&pidU32), unsafe.Pointer(&rec.inj), ebpf.UpdateAny); err != nil {
				return err
			}
		}
		return nil
	}

	// clean old hook
	if rec.inj != nil {
		uid := ShortID(binPath)
		log.Info("DetachHook: file modfied: ", binPath, " ShortID: ", uid)
		for _, fnName := range execGoFnName {
			p, ok := tracer.Manager.GetProbe(ebpfmanager.ProbeIdentificationPair{
				UID:          uid,
				EBPFFuncName: fnName,
			})
			if !ok {
				continue
			}
			if err := tracer.Manager.DetachHook(ebpfmanager.ProbeIdentificationPair{
				UID:          uid,
				EBPFFuncName: fnName,
			}); err != nil {
				log.Error(err)
			}
			if pp := p.Program(); pp != nil {
				if err := pp.Close(); err != nil {
					log.Warn(err)
				}
			}
		}
	}

	var goVer = [2]int{}
	inf, err := buildinfo.ReadFile(binPath)
	if err != nil {
		log.Debug(err)
	} else {
		goVer, _ = parseGoVersion(inf.GoVersion)
	}

	elfFile, err := elf.Open(binPath)
	if err != nil {
		return fmt.Errorf("failed to open elf file %s: %w", binPath, err)
	}
	defer elfFile.Close() // nolint:errcheck

	sym, err := getGoUprobeSymbolFromPCLN(elfFile, goVer[1] >= 20, "runtime.execute")
	if err != nil {
		log.Debug(err)
		return nil
	}
	symbolAddr := sym.Start

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

	emap, ok, err := tracer.Manager.GetMap(bmapProcInject)
	if err != nil {
		return fmt.Errorf("get bpf map bmap_proc_inject failed: %w", err)
	}
	if !ok {
		log.Warn("get bpf map bmap_proc_inject failed")
	}
	pidU32 := (uint32)(p.Pid)
	if err := emap.Update(unsafe.Pointer(&pidU32), unsafe.Pointer(&val), ebpf.UpdateAny); err != nil {
		return err
	}

	tracer.attachInfo.fileUpdater.Inject(procInfo.binPath, &val)

	uid := ShortID(binPath)
	for _, fnName := range execGoFnName {
		if err := tracer.Manager.AddHook("", &ebpfmanager.Probe{
			ProbeIdentificationPair: ebpfmanager.ProbeIdentificationPair{
				UID:          uid,
				EBPFFuncName: fnName,
			},
			UprobeOffset: symbolAddr,
			BinaryPath:   intern.String(binPath),
		}); err != nil {
			log.Warn(err)
		} else {
			log.Infof("AddHooK: %s, ShortID: %s, name: %s, pid: %d",
				binPath, uid, procInfo.name, pidU32)
		}
	}

	return nil
}
