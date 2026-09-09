// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pipeline implement datakit's logging pipeline.
package pipeline

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	// it will use this embedded information in time/tzdata.
	_ "time/tzdata"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	plmanager "github.com/GuanceCloud/pipeline-go/manager"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
	plstats "github.com/GuanceCloud/pipeline-go/stats"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"

	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	plval "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	l          = logger.DefaultSLogger("pipeline")
	rec        *plstats.RecStats
	jitInitMu  sync.RWMutex
	jitRunners jitRunnerRegistry
)

type jitBatchProcessor interface {
	Projection(string) (pljit.InputProjection, error)
	Process(string, []byte) (pljit.Batch, error)
}

type jitProcessor interface {
	jitBatchProcessor
	Prepare(string) error
	Check(string) pljit.CheckResult
	Invalidate(string)
	MinBatchSize() int
	Close() error
}

type jitRunnerGeneration struct {
	registry  *jitRunnerRegistry
	runner    jitProcessor
	control   *jitControl
	refs      int
	retired   bool
	drained   chan struct{}
	closeOnce sync.Once
	closeErr  error
	pending   bool
	closed    bool
}

type jitRunnerRegistry struct {
	mu      sync.Mutex
	current *jitRunnerGeneration
	pending int
}

type jitRunnerLease struct {
	registry   *jitRunnerRegistry
	generation *jitRunnerGeneration
	once       sync.Once
}

func (registry *jitRunnerRegistry) acquire() *jitRunnerLease {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if registry.current == nil || registry.current.runner == nil {
		return nil
	}
	registry.current.refs++
	return &jitRunnerLease{registry: registry, generation: registry.current}
}

func (lease *jitRunnerLease) runner() jitProcessor {
	if lease == nil || lease.generation == nil {
		return nil
	}
	return lease.generation.runner
}

func (lease *jitRunnerLease) release() {
	if lease == nil || lease.registry == nil || lease.generation == nil {
		return
	}
	lease.once.Do(func() {
		lease.registry.mu.Lock()
		lease.generation.refs--
		if lease.generation.retired && lease.generation.refs == 0 {
			close(lease.generation.drained)
		}
		lease.registry.mu.Unlock()
	})
}

func (registry *jitRunnerRegistry) replace(runner jitProcessor, controls ...*jitControl) *jitRunnerGeneration {
	var next *jitRunnerGeneration
	if runner != nil {
		next = &jitRunnerGeneration{runner: runner, registry: registry, drained: make(chan struct{})}
		if len(controls) != 0 {
			next.control = controls[0]
		}
	}

	registry.mu.Lock()
	previous := registry.current
	registry.current = next
	if previous != nil {
		if previous.control != nil {
			previous.control.active.Store(false)
		}
		previous.retired = true
		if previous.refs == 0 {
			close(previous.drained)
		}
	}
	registry.mu.Unlock()
	return previous
}

func closeJITGeneration(generation *jitRunnerGeneration) error {
	if generation == nil {
		return nil
	}
	<-generation.drained
	generation.closeOnce.Do(func() {
		generation.closeErr = generation.runner.Close()
		if registry := generation.registry; registry != nil {
			registry.mu.Lock()
			generation.closed = true
			if generation.pending {
				registry.pending--
				generation.pending = false
			}
			registry.mu.Unlock()
		}
	})
	return generation.closeErr
}

// retireJITGeneration lets publication complete without making new batches
// wait for an unrelated old in-flight batch. The retired generation remains
// owned by this goroutine; its leases keep the runner, immutable services and
// host registrations alive until the final old batch releases them. Runner
// Close then performs aggregate retirement drain and native resource cleanup.
func retireJITGeneration(generation *jitRunnerGeneration) {
	if generation == nil {
		return
	}
	if registry := generation.registry; registry != nil {
		registry.mu.Lock()
		if !generation.closed && !generation.pending {
			generation.pending = true
			registry.pending++
		}
		registry.mu.Unlock()
	}
	go func() {
		timer := time.AfterFunc(30*time.Second, func() {
			jitRetirementTimeouts.Inc()
			l.Warn("pipeline JIT retirement exceeds 30s; preserving in-use resources")
		})
		defer timer.Stop()
		if err := closeJITGeneration(generation); err != nil {
			l.Warnf("close retired pipeline JIT runner: %s", err)
		}
	}()
}

func InitPipeline(cfg *plval.PipelineCfg,
	upFn plmap.UploadFunc,
	gTags map[string]string,
	installDir string,
) error {
	l = logger.SLogger("pipeline")
	if cfg != nil {
		if err := cfg.JIT.Validate(); err != nil {
			return err
		}
	}

	// Serialize preparation/publication, including changes to shared services.
	jitInitMu.Lock()
	defer jitInitMu.Unlock()
	// Load once before InitPlVal mutates shared state, then bind the freshly
	// initialized host to that same unpublished runner.
	if cfg != nil && cfg.JIT != nil && cfg.JIT.Enabled {
		if err := checkJITRetirementCapacity(); err != nil {
			return err
		}
		started := time.Now()
		control, err := prepareJITControl(cfg, jitRuntimePath(cfg, installDir))
		var candidate *pljit.Runner
		if err == nil {
			candidate, err = pljit.NewRunnerWithServices(jitRuntimePath(cfg, installDir),
				"pipeline-go-1.4.3-datakit", cfg.JIT.MaxCachedPrograms, nil, jitServiceConfig(cfg))
		}
		if err == nil && !candidate.CanBindUnpublishedHost() {
			err = errors.New("runtime cannot bind unpublished host")
		}
		if err == nil {
			err = candidate.SetAggregateEventHandler(uploadJITAggregateEvents)
		}
		if err == nil {
			control.aggregatePrepared = true
		}
		if err != nil {
			if candidate != nil {
				err = errors.Join(err, candidate.Close())
			}
			observeJITRuntimeInit("error", time.Since(started))
			if jitCanInitiallyDegrade(cfg) {
				l.Warnf("pipeline JIT initial load failed; selecting Go before execution: %s", err)
				if err := initPipelineGoLocked(jitDisabledConfig(cfg), upFn, gTags, installDir); err != nil {
					return err
				}
				jitControlStateVec.WithLabelValues("init_degraded").Set(1)
				return nil
			}
			return fmt.Errorf("prepare pipeline JIT: %w", err)
		}
		committed := false
		defer func() {
			if !committed {
				_ = candidate.Close()
			}
		}()
		services := plval.PreparePlValServices(cfg)
		defer services.Abort()
		host := pljit.NewPipelineGoHostWithReferObserved(services.IPDB(), services.ReferTables(), observeJITHostCallback)
		if err := candidate.BindUnpublishedHost(host); err != nil {
			// Capability was preflighted above; keep this defensive check for
			// non-native Runtime implementations used by embedders/tests.
			observeJITRuntimeInit("error", time.Since(started))
			return fmt.Errorf("bind pipeline JIT host: %w", err)
		}
		if err := plval.InitPlValPrepared(cfg, upFn, gTags, installDir, services); err != nil {
			observeJITRuntimeInit("error", time.Since(started))
			return err
		}
		plstats.SetStats(rec)
		observeJITRuntimeInit("ready", time.Since(started))
		if err := publishJIT(candidate, jitRuntimePath(cfg, installDir), control); err != nil {
			return err
		}
		committed = true
		return nil
	}

	return initPipelineGoLocked(cfg, upFn, gTags, installDir)
}

func initPipelineGoLocked(cfg *plval.PipelineCfg, upFn plmap.UploadFunc, gTags map[string]string, installDir string) error {
	plstats.SetStats(rec)
	if err := plval.InitPlVal(cfg, upFn, gTags, installDir); err != nil {
		return err
	}
	return initJITLocked(cfg, installDir)
}

func jitRuntimePath(cfg *plval.PipelineCfg, installDir string) string {
	if cfg.JIT.RuntimePath != "" {
		return cfg.JIT.RuntimePath
	}
	return filepath.Join(installDir, "lib", "libplatypus_jit.so")
}

func jitServiceConfig(cfg *plval.PipelineCfg) pljit.ServiceConfig {
	return pljit.ServiceConfig{
		DisableHTTPRequestFunc: cfg.DisableHTTPRequestFunc,
		EnableMixedArrayField:  point.EnableMixedArrayField,
		Version:                1,
		HTTP: pljit.HTTPServiceConfig{
			DisableInternalNetwork: cfg.HTTPRequestDisableInternalNet,
			HostWhitelist:          cfg.HTTPRequestHostWhitelist,
			CIDRWhitelist:          cfg.HTTPRequestCIDRWhitelist,
		},
	}
}

func initJIT(cfg *plval.PipelineCfg, installDir string) error {
	jitInitMu.Lock()
	defer jitInitMu.Unlock()
	return initJITLocked(cfg, installDir)
}

func initJITLocked(cfg *plval.PipelineCfg, installDir string) error {
	if cfg != nil {
		if err := cfg.JIT.Validate(); err != nil {
			return err
		}
	}
	if cfg == nil || cfg.JIT == nil || !cfg.JIT.Enabled {
		// Do not take manager's write lease while old batches still hold it.
		// The retired control rejects future load-time checks; no new batch
		// can select JIT once the runner registry has published nil.
		retired := jitRunners.replace(nil)
		jitControlStateVec.WithLabelValues("enabled").Set(0)
		jitControlStateVec.WithLabelValues("init_degraded").Set(0)
		retireJITGeneration(retired)
		return nil
	}
	runtimePath := jitRuntimePath(cfg, installDir)
	if err := checkJITRetirementCapacity(); err != nil {
		return err
	}
	initStarted := time.Now()
	control, err := prepareJITControl(cfg, runtimePath)
	if err == nil {
		if reused, reuseErr := reuseJITForPolicy(control); reused {
			return reuseErr
		}
	}
	var runner *pljit.Runner
	if err == nil {
		runner, err = pljit.NewRunnerWithServices(runtimePath, "pipeline-go-1.4.3-datakit", cfg.JIT.MaxCachedPrograms, currentJITHost(), jitServiceConfig(cfg))
	}
	if err != nil {
		observeJITRuntimeInit("error", time.Since(initStarted))
		return fmt.Errorf("initialize pipeline JIT: %w", err)
	}
	observeJITRuntimeInit("ready", time.Since(initStarted))
	return publishJIT(runner, runtimePath, control)
}

func currentJITHost() *pljit.HostCompat {
	var referTables refertable.PlReferTables
	if value, ok := plval.GetRefTb(); ok && value != nil {
		referTables = value.Tables()
	}
	if value, ok := plval.GetIPDB(); ok {
		return pljit.NewPipelineGoHostWithReferObserved(value, referTables, observeJITHostCallback)
	}
	return pljit.NewPipelineGoHostWithReferObserved(nil, referTables, observeJITHostCallback)
}

// Caller holds jitInitMu and transfers candidate ownership to the registry.
func publishJIT(runner *pljit.Runner, runtimePath string, controls ...*jitControl) error {
	var control *jitControl
	if len(controls) > 0 {
		control = controls[0]
	}
	if control == nil || !control.aggregatePrepared {
		if err := runner.SetAggregateEventHandler(uploadJITAggregateEvents); err != nil {
			return errors.Join(fmt.Errorf("configure pipeline JIT aggregate event handler: %w", err), runner.Close())
		}
	}
	plval.SetJITModuleHooks(nil, nil)
	plval.SetJITInvalidate(runner.Invalidate)
	retired := jitRunners.replace(runner, control)
	plval.SetJITCheck(func(source string) bool {
		if control != nil {
			return false
		}
		started := time.Now()
		result := runner.Check(source)
		elapsed := time.Since(started)
		observeJITCheckDuration(result, elapsed)
		logJITCheck(source, result, elapsed)
		return result.Route == pljit.RouteJITNative || result.Route == pljit.RouteJITWithHost
	})
	plval.SetJITModuleHooks(func(snapshot *pljit.ModuleSnapshot) bool {
		if control != nil {
			return false
		}
		if !snapshot.HasDependencies() {
			return false
		}
		started := time.Now()
		result := runner.CheckModules(snapshot)
		observeJITCheckDuration(result, time.Since(started))
		observeJITCheck(result)
		l.Debugf("pipeline JIT module snapshot %x: route=%s detail=%s", snapshot.Identity(), result.Route, result.Detail)
		return result.Route == pljit.RouteJITNative || result.Route == pljit.RouteJITWithHost
	}, runner.InvalidateModules)
	if manager, ok := plval.GetManager(); ok && manager != nil {
		if control == nil {
			manager.SetJITAdmission(nil)
		} else {
			control.owner = manager
			manager.SetJITAdmission(func(snapshot *pljit.ModuleSnapshot) *pljit.Admission { return control.admission(runner, snapshot) })
		}
	}
	jitControlStateVec.WithLabelValues("enabled").Set(1)
	jitControlStateVec.WithLabelValues("init_degraded").Set(0)
	retireJITGeneration(retired)
	l.Infof("pipeline JIT enabled with runtime %s", runtimePath)
	if control != nil {
		l.Infof("pipeline JIT runtime SHA-256 %x; policy %s", control.runtimeSHA, control.policy.Mode())
	}
	return nil
}

func uploadJITAggregateEvents(values []pljit.Point) (err error) {
	started := time.Now()
	defer func() { observeJITAggregateDelivery(len(values), time.Since(started), err) }()
	manager, ok := plval.GetManager()
	if !ok || manager == nil {
		return fmt.Errorf("pipeline manager is unavailable for %d aggregate points", len(values))
	}
	type aggregateGroup struct {
		category point.Category
		bucket   string
		points   []*point.Point
	}
	groups := make(map[string]*aggregateGroup)
	order := make([]string, 0)
	for index := range values {
		category := point.CatString(values[index].Category)
		if category == point.UnknownCategory {
			return fmt.Errorf("aggregate point %d has unknown category %q", index, values[index].Category)
		}
		converted, err := pointFromJIT(&values[index])
		if err != nil {
			return fmt.Errorf("convert aggregate point %d: %w", index, err)
		}
		key := category.String() + "\x00" + values[index].Measurement
		group := groups[key]
		if group == nil {
			group = &aggregateGroup{category: category, bucket: values[index].Measurement}
			groups[key] = group
			order = append(order, key)
		}
		group.points = append(group.points, converted)
	}
	for _, key := range order {
		group := groups[key]
		if err := manager.UploadAggregates(group.category, group.bucket, group.points); err != nil {
			return fmt.Errorf("upload aggregate bucket %s/%s: %w", group.category, group.bucket, err)
		}
	}
	return nil
}

const maximumJITCheckDetailBytes = 512

func logJITCheck(source string, result pljit.CheckResult, elapsed time.Duration) {
	sourceHash := sha256.Sum256([]byte(source))
	observeJITCheck(result)
	l.Debugf("pipeline JIT check route=%s reason=%s backend=%s execution_tier=%s execution_mode=%s duration=%s source_hash=%x detail=%q",
		result.Route, result.Reason, result.Capabilities.Backend,
		result.Capabilities.ExecutionTier, result.Capabilities.ExecutionMode,
		elapsed, sourceHash, boundedJITCheckDetail(result.Detail))
}

func boundedJITCheckDetail(detail string) string {
	detail = strings.ToValidUTF8(detail, "\ufffd")
	if len(detail) <= maximumJITCheckDetailBytes {
		return detail
	}
	end := maximumJITCheckDetailBytes
	for end > 0 && !utf8.RuneStart(detail[end]) {
		end--
	}
	return detail[:end] + "..."
}

func NewPlScriptSampleFromFile(category point.Category, path string, buks ...*plmap.AggBuckets) (*platypus.PlScript, error) {
	name, script, err := plmanager.ReadScript(path)
	if err != nil {
		return nil, err
	}

	return NewPlScriptSimple(category, name, script, buks...)
}

func NewPlScriptSimple(category point.Category, name, script string, buks ...*plmap.AggBuckets) (*platypus.PlScript, error) {
	var bkt *plmap.AggBuckets
	if len(buks) > 0 {
		bkt = buks[0]
	}
	scs, errs := platypus.NewScripts(map[string]string{name: script},
		lang.WithCat(category),
		lang.WithCache(),
		lang.WithPtWindow(),
		lang.WithAggBktUser(bkt),
	)

	if v, ok := errs[name]; ok {
		return nil, v
	}

	if sc, ok := scs[name]; ok {
		return sc, nil
	}

	return nil, fmt.Errorf("unknown error")
}

func NewPipelineMulti(category point.Category, scripts map[string]string, buks *plmap.AggBuckets,
) (map[string]*platypus.PlScript, map[string]error) {
	return platypus.NewScripts(scripts,
		lang.WithCat(category),

		lang.WithCache(),
		lang.WithPtWindow(),
		lang.WithAggBktUser(buks),
	)
}

type Pipeline struct {
	Script *platypus.PlScript
}

func (p *Pipeline) Run(cat point.Category, pt *point.Point, plOpt *lang.LogOption,
	signal plruntime.Signal,
) (ptinput.PlInputPt, error) {
	if p.Script == nil || p.Script.Engine() == nil {
		return nil, fmt.Errorf("pipeline engine not initialized")
	}

	if pt == nil {
		return nil, fmt.Errorf("no data")
	}

	plpt := ptinput.PtWrap(cat, pt)

	if v, ok := plval.GetRefTb(); ok {
		plpt.SetPlReferTables(v.Tables())
	}
	if v, ok := plval.GetIPDB(); ok {
		plpt.SetIPDB(v)
	}

	if err := p.Script.Run(plpt, signal, plOpt); err != nil {
		return nil, err
	} else {
		return plpt, nil
	}
}

// GbToUtf8 Gb to UTF-8.
// http/api_pipeline.go.
func GbToUtf8(s []byte, encoding string) ([]byte, error) {
	var t transform.Transformer
	switch encoding {
	case "gbk":
		t = simplifiedchinese.GBK.NewDecoder()
	case "gb18030":
		t = simplifiedchinese.GB18030.NewDecoder()
	}
	reader := transform.NewReader(bytes.NewReader(s), t)
	d, e := io.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func DecodeContent(content []byte, encode string) (string, error) {
	var err error
	if encode != "" {
		encode = strings.ToLower(encode)
	}
	switch encode {
	case "gbk", "gb18030":
		content, err = GbToUtf8(content, encode)
		if err != nil {
			return "", err
		}
	case "utf8", "utf-8":
	default:
	}
	return string(content), nil
}

func init() { //nolint:gochecknoinits
	rec = plstats.NewRecStats("datakit", plstats.DefaultSubSystem, nil, 128)
	metrics.MustRegister(rec.Metrics()...)
	metrics.MustRegister(jitMetrics()...)
}
