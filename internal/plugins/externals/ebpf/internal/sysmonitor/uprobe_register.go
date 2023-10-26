//go:build linux
// +build linux

package sysmonitor

import (
	"context"
	"debug/elf"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/DataDog/gopsutil/process/so"
)

type UprobeRegRule struct {
	Re         *regexp.Regexp
	Register   func(string) error
	UnRegister func(string) error
}

type UprobeConf struct {
	AttachDynamicLib bool
	DynamicLibNameRe *regexp.Regexp

	FuncName           string
	UprobeProgFuncName string
}

type UprobeAttachTyp int32

const (
	AttachUnknown UprobeAttachTyp = iota
	AttachProcess
	AttachDynamicLib
)

type ProcessUprobeRegister struct {
	Manager                  *manager.Manager
	ProgSymbols              []UprobeConf
	DynamicLibSymbols        []UprobeConf
	AttchDynamicLibOrProcess UprobeAttachTyp
	ProbeIDPrefix            string
}

type uprobeAttachArg struct {
	binPath string

	symbol       string
	symbolOffset uint64

	uprobeFunc string
}

func (reg *ProcessUprobeRegister) Register(binPath string, dynamic bool) error {
	if reg.Manager == nil {
		return nil
	}

	var args []uprobeAttachArg
	var err error

	if dynamic {
		args, err = getUpAttachArgs(binPath, reg.ProgSymbols)
	} else {
		args, err = getUpAttachArgs(binPath, reg.DynamicLibSymbols)
	}

	if err != nil {
		return err
	}

	for _, arg := range args {
		shortid := ShortID(reg.ProbeIDPrefix, arg.binPath)
		probeID := manager.ProbeIdentificationPair{
			UID:          shortid,
			EBPFFuncName: arg.uprobeFunc,
		}

		if err := reg.Manager.DetachHook(probeID); err != nil {
			l.Error(err)
		}

		if err := reg.Manager.AddHook("", &manager.Probe{
			ProbeIdentificationPair: probeID,
			UprobeOffset:            arg.symbolOffset,
			BinaryPath:              arg.binPath,
		}); err != nil {
			return err
		}
	}

	return nil
}

func getUpAttachArgs(binPath string, conf []UprobeConf) ([]uprobeAttachArg, error) {
	if len(conf) == 0 || binPath == "" {
		return nil, nil
	}

	f, err := elf.Open(binPath)
	if err != nil {
		return nil, err
	}
	var upArgs []uprobeAttachArg
	for _, conf := range conf {
		if conf.AttachDynamicLib {
			if conf.DynamicLibNameRe != nil {
				if !conf.DynamicLibNameRe.MatchString(binPath) {
					return nil, nil
				}
			}
		}
		if syms, err := FindSymbol(f, conf.FuncName); err != nil {
			l.Debug(err)
		} else {
			for _, sym := range syms {
				if sym.Section != elf.SHN_UNDEF {
					upArgs = append(upArgs, uprobeAttachArg{
						binPath:      binPath,
						symbol:       sym.Name,
						symbolOffset: sym.Value,
						uprobeFunc:   conf.UprobeProgFuncName,
					})
				}
			}
		}
	}

	return upArgs, nil
}

func NewProcessUProbeRegister(m *manager.Manager, progSymConf, dlSymConf []UprobeConf, idPrefix ...string) *ProcessUprobeRegister {
	var prefix string
	if len(idPrefix) > 0 {
		prefix = idPrefix[0]
	}
	return &ProcessUprobeRegister{
		Manager:           m,
		ProgSymbols:       progSymConf,
		DynamicLibSymbols: dlSymConf,
		ProbeIDPrefix:     prefix,
	}
}

func NewUprobeDyncLibRegister(rules []UprobeRegRule) (*UprobeRegister, error) {
	r := &UprobeRegister{}
	r.rules = append(r.rules, rules...)
	allRe := []string{}
	if len(rules) == 0 {
		return nil, fmt.Errorf("len(rules) == 0")
	}
	for _, v := range rules {
		if v.Re == nil {
			return nil, fmt.Errorf("%#v", v)
		}
		allRe = append(allRe, fmt.Sprintf("(%s)", v.Re.String()))
	}
	var err error
	r.re, err = regexp.Compile(strings.Join(allRe, "|"))
	if err != nil {
		return nil, err
	}
	return r, nil
}

type UprobeRegister struct {
	rules []UprobeRegRule
	re    *regexp.Regexp

	libPaths map[string]struct{}

	scanInterval time.Duration

	run int32 // 0 or 1(true)

	sync.Mutex
}

func (register *UprobeRegister) ScanAndUpdate() {
	register.Lock()
	defer register.Unlock()
	allLibs := map[string]struct{}{}
	for _, v := range so.Find(register.re) {
		allLibs[v.HostPath] = struct{}{}
	}
	del, add := diff(register.libPaths, allLibs)
	if len(del) == 0 && len(add) == 0 {
		return
	}

	register.libPaths = allLibs

	for k := range del {
		for _, r := range register.rules {
			if r.Re.MatchString(k) {
				if err := r.UnRegister(k); err != nil {
					l.Error(err)
				}
			}
		}
	}

	for k := range add {
		for _, r := range register.rules {
			if r.Re.MatchString(k) {
				l.Info(k)
				if err := r.Register(k); err != nil {
					l.Error(err)
				}
			}
		}
	}
}

func (register *UprobeRegister) CleanAll() {
	register.Lock()
	defer register.Unlock()

	allLibs := map[string]struct{}{}
	for _, v := range so.Find(register.re) {
		allLibs[v.HostPath] = struct{}{}
	}

	for k := range allLibs {
		for _, r := range register.rules {
			if r.Re.MatchString(k) {
				if err := r.UnRegister(k); err != nil {
					l.Debug(err)
				}
			}
		}
	}
}

func (register *UprobeRegister) Monitor(ctx context.Context, scanInterval time.Duration) {
	if old := atomic.SwapInt32(&register.run, 1); old == 1 {
		l.Warn(".so monitor started")
		return
	}
	register.scanInterval = scanInterval
	ticker := time.NewTicker(register.scanInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				register.ScanAndUpdate()
			case <-ctx.Done():
				register.CleanAll()
				return
			}
		}
	}()
}

type (
	registerFunc   func(string) error
	unregisterFunc func(string) error
)

func NewRegisterFunc(m *manager.Manager, bpfFuncName []string) registerFunc {
	bfunc := []string{}
	bfunc = append(bfunc, bpfFuncName...)
	return func(binPath string) error {
		uid := ShortID(binPath)
		l.Info("AddHook: ", binPath, " ShortID: ", uid)
		for _, fnName := range bfunc {
			if err := m.AddHook("", &manager.Probe{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					UID:          uid,
					EBPFFuncName: fnName,
				},
				BinaryPath: binPath,
			}); err != nil {
				l.Warn(err)
			}
		}
		return nil
	}
}

func NewUnRegisterFunc(m *manager.Manager, bpfFuncName []string) unregisterFunc {
	bfunc := []string{}
	bfunc = append(bfunc, bpfFuncName...)
	return func(binPath string) error {
		uid := ShortID(binPath)
		l.Info("DetachHook: ", binPath, " ShortID: ", uid)
		for _, fnName := range bfunc {
			p, ok := m.GetProbe(manager.ProbeIdentificationPair{
				UID:          uid,
				EBPFFuncName: fnName,
			})
			if !ok {
				continue
			}
			pp := p.Program()
			if err := m.DetachHook(manager.ProbeIdentificationPair{
				UID:          uid,
				EBPFFuncName: fnName,
			}); err != nil {
				l.Error(err)
			}
			if pp != nil {
				if err := pp.Close(); err != nil {
					l.Warn(err)
				}
			}
		}
		return nil
	}
}
