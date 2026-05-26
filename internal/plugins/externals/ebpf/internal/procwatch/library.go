//go:build linux
// +build linux

package procwatch

import (
	"context"
	"debug/elf"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
)

const (
	defaultLibraryScanInterval = time.Second * 30
	maxLibraryScanInterval     = time.Minute * 5
	initialLibraryScanDelay    = time.Second * 20
)

type HookRule struct {
	Re     *regexp.Regexp
	Attach func(string) error
	Detach func(string) error
}

type BinaryHook struct {
	AttachDynamicLib bool
	LibraryPattern   *regexp.Regexp

	Function string
	Program  string
}

type BinaryHookBinder struct {
	Runtime          *bpfutil.Runtime
	ProcessHooks     []BinaryHook
	SharedLibHooks   []BinaryHook
	ProbePrefixParts []string
}

type LibraryTracker struct {
	rules []HookRule
	re    *regexp.Regexp

	scanFn func(*regexp.Regexp) map[string]struct{}

	libraries map[string]struct{}
	running   int32
	mu        sync.Mutex
}

func NewBinaryHookBinder(runtime *bpfutil.Runtime, processHooks []BinaryHook, sharedLibHooks []BinaryHook, prefix ...string) *BinaryHookBinder {
	return &BinaryHookBinder{
		Runtime:          runtime,
		ProcessHooks:     processHooks,
		SharedLibHooks:   sharedLibHooks,
		ProbePrefixParts: append([]string(nil), prefix...),
	}
}

func (b *BinaryHookBinder) Attach(path string, sharedLibrary bool) error {
	if b.Runtime == nil || path == "" {
		return nil
	}

	configs := b.ProcessHooks
	if sharedLibrary {
		configs = b.SharedLibHooks
	}

	args, err := buildAttachArgs(path, configs)
	if err != nil {
		return err
	}

	for _, arg := range args {
		uid := shortID(append(b.ProbePrefixParts, arg.path)...)
		hookID := bpfutil.HookID{UID: uid, Program: arg.program}
		_ = b.Runtime.DetachHookSpec(hookID)
		if err := b.Runtime.AttachHook(bpfutil.HookSpec{
			ID:           hookID,
			UprobeOffset: arg.offset,
			BinaryPath:   arg.path,
		}); err != nil {
			return err
		}
	}

	return nil
}

type attachArg struct {
	path    string
	offset  uint64
	program string
}

func buildAttachArgs(path string, configs []BinaryHook) ([]attachArg, error) {
	if path == "" || len(configs) == 0 {
		return nil, nil
	}

	file, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck

	var args []attachArg
	for _, cfg := range configs {
		if cfg.AttachDynamicLib && cfg.LibraryPattern != nil && !cfg.LibraryPattern.MatchString(path) {
			continue
		}
		symbols, err := findSymbol(file, cfg.Function)
		if err != nil {
			log.Debug(err)
			continue
		}
		for _, symbol := range symbols {
			if symbol.Section != elf.SHN_UNDEF {
				args = append(args, attachArg{
					path:    path,
					offset:  symbol.Value,
					program: cfg.Program,
				})
			}
		}
	}

	return args, nil
}

func NewLibraryTracker(rules []HookRule) (*LibraryTracker, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("len(rules) == 0")
	}

	patterns := make([]string, 0, len(rules))
	for _, rule := range rules {
		if rule.Re == nil {
			return nil, fmt.Errorf("%#v", rule)
		}
		patterns = append(patterns, fmt.Sprintf("(%s)", rule.Re.String()))
	}

	re, err := regexp.Compile(strings.Join(patterns, "|"))
	if err != nil {
		return nil, err
	}

	return &LibraryTracker{
		rules:  append([]HookRule(nil), rules...),
		re:     re,
		scanFn: findLoadedLibraryHostPaths,
	}, nil
}

func (t *LibraryTracker) Scan() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	current := t.scanLibraries()
	removed, added := deltaSet(t.libraries, current)
	if len(removed) == 0 && len(added) == 0 {
		return false
	}
	t.libraries = current

	for path := range removed {
		for _, rule := range t.rules {
			if rule.Re.MatchString(path) {
				if err := rule.Detach(path); err != nil {
					log.Error(err)
				}
			}
		}
	}

	for path := range added {
		for _, rule := range t.rules {
			if rule.Re.MatchString(path) {
				if err := rule.Attach(path); err != nil {
					log.Error(err)
				}
			}
		}
	}

	return true
}

func (t *LibraryTracker) Clean() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for path := range t.scanLibraries() {
		for _, rule := range t.rules {
			if rule.Re.MatchString(path) {
				if err := rule.Detach(path); err != nil {
					log.Debug(err)
				}
			}
		}
	}
	t.libraries = nil
}

func (t *LibraryTracker) Run(ctx context.Context, interval time.Duration) {
	if atomic.SwapInt32(&t.running, 1) == 1 {
		log.Warn(".so monitor started")
		return
	}

	baseInterval := interval
	if baseInterval <= 0 {
		baseInterval = defaultLibraryScanInterval
	}

	timer := time.NewTimer(initialLibraryScanDelay)
	go func() {
		defer timer.Stop()
		defer atomic.StoreInt32(&t.running, 0)
		nextInterval := baseInterval
		for {
			select {
			case <-timer.C:
				nextInterval = nextLibraryScanInterval(t.Scan(), nextInterval, baseInterval, maxLibraryScanInterval)
				timer.Reset(nextInterval)
			case <-ctx.Done():
				t.Clean()
				return
			}
		}
	}()
}

func (t *LibraryTracker) scanLibraries() map[string]struct{} {
	if t == nil || t.scanFn == nil {
		return nil
	}
	return t.scanFn(t.re)
}

func nextLibraryScanInterval(changed bool, current, base, max time.Duration) time.Duration {
	if base <= 0 {
		base = defaultLibraryScanInterval
	}
	if max < base {
		max = base
	}
	if changed || current < base {
		return base
	}

	next := current * 2
	if next > max {
		return max
	}
	return next
}

func NewAttachFunc(runtime *bpfutil.Runtime, programs []string) func(string) error {
	progList := append([]string(nil), programs...)
	return func(path string) error {
		uid := shortID(path)
		log.Info("AddHook: ", path, " ShortID: ", uid)
		for _, program := range progList {
			if err := runtime.AttachHook(bpfutil.HookSpec{
				ID:         bpfutil.HookID{UID: uid, Program: program},
				BinaryPath: path,
			}); err != nil {
				log.Warn(err)
			}
		}
		return nil
	}
}

func NewDetachFunc(runtime *bpfutil.Runtime, programs []string) func(string) error {
	progList := append([]string(nil), programs...)
	return func(path string) error {
		uid := shortID(path)
		log.Info("DetachHook: ", path, " ShortID: ", uid)
		for _, program := range progList {
			hookID := bpfutil.HookID{UID: uid, Program: program}
			hook, ok := runtime.LookupHook(hookID)
			if !ok {
				continue
			}
			prog := hook.Program()
			if err := runtime.DetachHookSpec(hookID); err != nil {
				log.Error(err)
			}
			if prog != nil {
				if err := prog.Close(); err != nil {
					log.Warn(err)
				}
			}
		}
		return nil
	}
}
