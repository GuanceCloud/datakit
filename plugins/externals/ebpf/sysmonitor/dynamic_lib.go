//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package sysmonitor

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DataDog/ebpf/manager"
	"github.com/DataDog/gopsutil/process/so"
	"github.com/GuanceCloud/cliutils/logger"
)

var l = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

func diff(old, cur map[string]struct{}) (map[string]struct{}, map[string]struct{}) {
	add := map[string]struct{}{}
	del := map[string]struct{}{}
	for k := range cur {
		if _, ok := old[k]; !ok {
			add[k] = struct{}{}
		}
	}

	for k := range old {
		if _, ok := cur[k]; !ok {
			del[k] = struct{}{}
		}
	}

	return del, add
}

func ShortID(binPath string) string {
	sha1Val := sha256.Sum256([]byte(binPath))
	return strconv.FormatUint(
		binary.BigEndian.Uint64(sha1Val[:]), 36)
}

type UprobeRegRule struct {
	Re         *regexp.Regexp
	Register   func(string) error
	UnRegister func(string) error
}

type UprobeProcessRegister struct{}

func NewUprobeDyncLibRegister(rules []UprobeRegRule) (*UprobeDynamicLibRegister, error) {
	r := &UprobeDynamicLibRegister{}
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

type UprobeDynamicLibRegister struct {
	rules []UprobeRegRule
	re    *regexp.Regexp

	libPaths map[string]struct{}

	scanInterval time.Duration

	run int32 // 0 or 1(true)

	sync.Mutex
}

func (register *UprobeDynamicLibRegister) ScanAndUpdate() {
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
				if err := r.Register(k); err != nil {
					l.Error(err)
				}
			}
		}
	}
}

func (register *UprobeDynamicLibRegister) CleanAll() {
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

func (register *UprobeDynamicLibRegister) Monitor(ctx context.Context, scanInterval time.Duration) {
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

func NewRegisterFunc(m *manager.Manager, sections []string) registerFunc {
	secs := []string{}
	secs = append(secs, sections...)
	return func(binPath string) error {
		uid := ShortID(binPath)
		l.Info("AddHook: ", binPath, " ShortID: ", uid)
		for _, sec := range secs {
			if err := m.AddHook("", manager.Probe{
				UID:        uid,
				Section:    sec,
				BinaryPath: binPath,
			}); err != nil {
				l.Warn(err)
			}
		}
		return nil
	}
}

func NewUnRegisterFunc(m *manager.Manager, sections []string) unregisterFunc {
	secs := []string{}
	secs = append(secs, sections...)
	return func(binPath string) error {
		uid := ShortID(binPath)
		l.Info("DetachHook: ", binPath, " ShortID: ", uid)
		for _, sec := range secs {
			p, ok := m.GetProbe(manager.ProbeIdentificationPair{
				UID:     uid,
				Section: sec,
			})
			if !ok {
				continue
			}
			pp := p.Program()
			if err := m.DetachHook(sec, uid); err != nil {
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
