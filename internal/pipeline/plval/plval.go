// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package plval store pipeline private values.
package plval

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/offload"
)

var (
	l = logger.DefaultSLogger("pl-val")

	g = goroutine.NewGroup(goroutine.Option{
		Name: "pipeline",
	})
)

var (
	// pipeline mamager.
	_managerIns atomic.Pointer[ScriptManager]

	// refer table.
	_referTable atomic.Pointer[refertable.ReferTable]
	referWorker struct {
		sync.Mutex
		current *referWorkerGeneration
	}

	// offload.
	offloadWorker struct {
		sync.Mutex
		current *offloadWorkerGeneration
	}

	enableAppendRunInfo atomic.Bool
	grokFastPathEnabled uint32
	externalRequestsOff atomic.Bool
	originalHTTPRequest = funcs.FuncsMap["http_request"]
)

type referWorkerGeneration struct {
	cancel context.CancelFunc
	done   chan struct{}
	table  *refertable.ReferTable
	users  sync.RWMutex
	close  sync.Once
}

type ReferLease struct {
	generation *referWorkerGeneration
	once       sync.Once
}

type offloadWorkerGeneration struct {
	worker *offload.OffloadWorker
	users  sync.RWMutex
	stop   sync.Once
}

type OffloadLease struct {
	generation *offloadWorkerGeneration
	once       sync.Once
}

func (lease *OffloadLease) Worker() *offload.OffloadWorker {
	if lease == nil || lease.generation == nil {
		return nil
	}
	return lease.generation.worker
}

func (lease *OffloadLease) Release() {
	if lease == nil || lease.generation == nil {
		return
	}
	lease.once.Do(lease.generation.users.RUnlock)
}

// PreparedPlValServices owns unpublished IPDB/refer resources. A JIT runner
// can bind these exact instances before any global generation changes. The
// caller must Commit or Abort exactly once; both operations are idempotent.
type PreparedPlValServices struct {
	mu             sync.Mutex
	ipdb           ipdb.IPdb
	refer          *refertable.ReferTable
	offload        *offload.OffloadWorker
	replaceOffload bool
	completed      bool
}

func PreparePlValServices(cfg *PipelineCfg) *PreparedPlValServices {
	prepared := &PreparedPlValServices{replaceOffload: true}
	offload.InitOffload()
	if value, err := InitIPdb(datakit.DataDir, cfg); err != nil {
		l.Warnf("init ipdb error: %s", err)
	} else {
		prepared.ipdb = value
	}
	if cfg != nil && cfg.Offload != nil && cfg.Offload.Receiver != "" &&
		len(cfg.Offload.Addresses) != 0 {
		worker, err := offload.NewOffloader(&offload.OffloadConfig{
			Receiver: cfg.Offload.Receiver, Addresses: append([]string(nil), cfg.Offload.Addresses...),
		})
		if err != nil {
			// Preserve the previous worker, matching the historical reload
			// behavior when a replacement configuration cannot be constructed.
			l.Errorf("init offload worker, error: %v", err)
			prepared.replaceOffload = false
		} else {
			prepared.offload = worker
		}
	}
	if cfg == nil || cfg.ReferTableURL == "" {
		return prepared
	}
	duration, err := time.ParseDuration(cfg.ReferTablePullInterval)
	if err != nil {
		l.Warnf("refer table pull interval %s, err: %v", cfg.ReferTablePullInterval, err)
		duration = 5 * time.Minute
	}
	prepared.refer, err = refertable.NewReferTable(refertable.RefTbCfg{
		URL: cfg.ReferTableURL, Interval: duration,
		UseSQLite: cfg.UseSQLite, SQLiteMemMode: cfg.SQLiteMemMode,
		DBPath: filepath.Join(datakit.DataDir, "reftable_sqlite"),
	})
	if err != nil {
		l.Errorf("init refer table: %v", err)
		prepared.refer = nil
	}
	return prepared
}

func (prepared *PreparedPlValServices) IPDB() ipdb.IPdb {
	if prepared == nil {
		return nil
	}
	prepared.mu.Lock()
	defer prepared.mu.Unlock()
	return prepared.ipdb
}

func (prepared *PreparedPlValServices) ReferTables() refertable.PlReferTables {
	if prepared == nil {
		return nil
	}
	prepared.mu.Lock()
	defer prepared.mu.Unlock()
	if prepared.refer == nil {
		return nil
	}
	return prepared.refer.Tables()
}

func (prepared *PreparedPlValServices) Commit() (*offload.OffloadWorker, bool) {
	if prepared == nil {
		return nil, false
	}
	prepared.mu.Lock()
	if prepared.completed {
		prepared.mu.Unlock()
		return nil, false
	}
	prepared.completed = true
	db, table, offloadWorker, replaceOffload := prepared.ipdb, prepared.refer, prepared.offload, prepared.replaceOffload
	prepared.ipdb, prepared.refer, prepared.offload = nil, nil, nil
	prepared.mu.Unlock()
	SetIPDB(db)
	replaceReferWorker(table)
	if replaceOffload {
		replaceOffloadWorker(offloadWorker)
	}
	return offloadWorker, replaceOffload
}

func (prepared *PreparedPlValServices) Abort() {
	if prepared == nil {
		return
	}
	prepared.mu.Lock()
	if prepared.completed {
		prepared.mu.Unlock()
		return
	}
	prepared.completed = true
	db, table, offloadWorker := prepared.ipdb, prepared.refer, prepared.offload
	prepared.ipdb, prepared.refer, prepared.offload = nil, nil, nil
	prepared.mu.Unlock()
	if table != nil {
		if err := table.Close(); err != nil {
			l.Warnf("close aborted pipeline refer table: %s", err)
		}
	}
	if closer, ok := db.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			l.Warnf("close aborted pipeline IPDB: %s", err)
		}
	}
	if offloadWorker != nil {
		offloadWorker.Stop()
	}
}

func (lease *ReferLease) Tables() refertable.PlReferTables {
	if lease == nil || lease.generation == nil || lease.generation.table == nil {
		return nil
	}
	return lease.generation.table.Tables()
}

func (lease *ReferLease) Release() {
	if lease == nil || lease.generation == nil {
		return
	}
	lease.once.Do(lease.generation.users.RUnlock)
}

// AcquireRefTb pins the published refer service generation. Replacement can
// stop its refresh worker immediately, but cannot close the backing SQLite
// pool while a pipeline batch still uses its immutable generation.
func AcquireRefTb() (*ReferLease, bool) {
	referWorker.Lock()
	generation := referWorker.current
	if generation == nil || generation.table == nil {
		referWorker.Unlock()
		return nil, false
	}
	generation.users.RLock()
	referWorker.Unlock()
	return &ReferLease{generation: generation}, true
}

func init() { //nolint:gochecknoinits
	installDisableGrokFastPath()
	funcs.FuncsMap["http_request"] = func(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
		if externalRequestsOff.Load() {
			ctx.Regs.ReturnAppend(nil, ast.Nil)
			return nil
		}
		return originalHTTPRequest(ctx, funcExpr)
	}
}

func EnableAppendRunInfo() bool {
	return enableAppendRunInfo.Load()
}

func setGrokFastPathEnabled(enabled bool) {
	if enabled {
		atomic.StoreUint32(&grokFastPathEnabled, 1)
		return
	}
	atomic.StoreUint32(&grokFastPathEnabled, 0)
}

func isGrokFastPathEnabled() bool {
	return atomic.LoadUint32(&grokFastPathEnabled) != 0
}

func SetManager(m *ScriptManager) {
	previous := _managerIns.Swap(m)
	if previous == nil || previous == m {
		return
	}
	go func(retired *ScriptManager) {
		retired.rw.Lock()
		defer retired.rw.Unlock()
		cleanupManagerState(retired.state)
	}(previous)
}

func GetManager() (*ScriptManager, bool) {
	m := _managerIns.Load()
	return m, m != nil
}

func AcquireManager() (*ManagerLease, bool) {
	m := _managerIns.Load()
	if m == nil {
		return nil, false
	}
	return m.Acquire()
}

func AcquireManagerContext(ctx context.Context) (*ManagerLease, error) {
	return _managerIns.Load().AcquireContext(ctx)
}

func SetJITCheck(check func(string) bool) {
	if m := _managerIns.Load(); m != nil {
		m.SetJITCheck(check)
	}
}

func SetJITModuleHooks(check func(*pljit.ModuleSnapshot) bool, invalidate func(*pljit.ModuleSnapshot)) {
	if m := _managerIns.Load(); m != nil {
		m.SetJITModuleHooks(check, invalidate)
	}
}

func SetJITInvalidate(invalidate func(string)) {
	if m := _managerIns.Load(); m != nil {
		m.SetJITInvalidate(invalidate)
	}
}

func SetIPDB(db ipdb.IPdb) {
	var next *ipdbInstance
	if db == nil {
		next = nil
	} else {
		next = &ipdbInstance{db: db}
	}
	previous := _ipdb.Swap(next)
	if previous != nil {
		go func(retired *ipdbInstance) {
			retired.users.Lock()
			defer retired.users.Unlock()
			retired.close.Do(func() {
				if closer, ok := retired.db.(interface{ Close() error }); ok {
					if err := closer.Close(); err != nil {
						l.Warnf("close retired pipeline IPDB: %s", err)
					}
				}
			})
		}(previous)
	}
}

func SetRefTb(tb *refertable.ReferTable) {
	_referTable.Store(tb)
}

func GetRefTb() (*refertable.ReferTable, bool) {
	table := _referTable.Load()
	return table, table != nil
}

// Bind the worker to its creating generation, not the global pointer at the
// later moment when the goroutine happens to start.
func referPullTask(table *refertable.ReferTable) func(context.Context) error {
	return func(ctx context.Context) error {
		if table == nil {
			return fmt.Errorf("pipeline refertable not ready")
		}
		table.PullWorker(ctx)
		return nil
	}
}

// replaceReferWorker publishes one pull worker generation and drains the old
// one. ReferTable snapshots already captured by in-flight pipeline batches
// remain valid for reads; only their obsolete refresh loop is stopped.
func replaceReferWorker(table *refertable.ReferTable) {
	var next *referWorkerGeneration
	if table != nil {
		ctx, cancel := context.WithCancel(context.Background())
		next = &referWorkerGeneration{cancel: cancel, done: make(chan struct{}), table: table}
		g.Go(func(groupCtx context.Context) error {
			stop := context.AfterFunc(groupCtx, cancel)
			defer stop()
			defer close(next.done)
			return referPullTask(table)(ctx)
		})
	}

	referWorker.Lock()
	previous := referWorker.current
	referWorker.current = next
	_referTable.Store(table)
	referWorker.Unlock()
	if previous != nil {
		previous.cancel()
		<-previous.done
		go func(retired *referWorkerGeneration) {
			retired.users.Lock()
			defer retired.users.Unlock()
			retired.close.Do(func() {
				if err := retired.table.Close(); err != nil {
					l.Warnf("close retired pipeline refer table: %s", err)
				}
			})
		}(previous)
	}
}

func SetOffload(offl *offload.OffloadWorker) {
	replaceOffloadWorker(offl)
}

func GetOffload() (*offload.OffloadWorker, bool) {
	lease, ok := AcquireOffload()
	if !ok {
		return nil, false
	}
	worker := lease.Worker()
	lease.Release()
	return worker, worker != nil
}

func replaceOffloadWorker(worker *offload.OffloadWorker) {
	var next *offloadWorkerGeneration
	if worker != nil {
		next = &offloadWorkerGeneration{worker: worker}
	}
	offloadWorker.Lock()
	previous := offloadWorker.current
	offloadWorker.current = next
	offloadWorker.Unlock()
	if previous != nil && previous.worker != worker {
		if previous.users.TryLock() {
			previous.stop.Do(previous.worker.Stop)
			previous.users.Unlock()
			return
		}
		go func(retired *offloadWorkerGeneration) {
			retired.users.Lock()
			defer retired.users.Unlock()
			retired.stop.Do(retired.worker.Stop)
		}(previous)
	}
}

func AcquireOffload() (*OffloadLease, bool) {
	offloadWorker.Lock()
	generation := offloadWorker.current
	if generation == nil || generation.worker == nil {
		offloadWorker.Unlock()
		return nil, false
	}
	generation.users.RLock()
	offloadWorker.Unlock()
	return &OffloadLease{generation: generation}, true
}

const maxCustomer = 16

var (
	localDefaultPipelineMu sync.RWMutex
	localDefaultPipeline   map[point.Category]string
)

func GetLocalDefaultPipeline() map[point.Category]string {
	localDefaultPipelineMu.RLock()
	defer localDefaultPipelineMu.RUnlock()
	if localDefaultPipeline == nil {
		return nil
	}
	result := make(map[point.Category]string, len(localDefaultPipeline))
	for category, source := range localDefaultPipeline {
		result[category] = source
	}
	return result
}

func setLocalDefaultPipeline(defaults map[point.Category]string) {
	localDefaultPipelineMu.Lock()
	defer localDefaultPipelineMu.Unlock()
	if defaults == nil {
		localDefaultPipeline = nil
		return
	}
	localDefaultPipeline = make(map[point.Category]string, len(defaults))
	for category, source := range defaults {
		localDefaultPipeline[category] = source
	}
}

func PreferLocalDefaultPipeline(m map[point.Category]string) map[point.Category]string {
	result := map[point.Category]string{}
	for k, v := range m {
		result[k] = v
	}
	for k, v := range GetLocalDefaultPipeline() {
		result[k] = v
	}

	return result
}

func DisableExternalRequestsFunc() {
	externalRequestsOff.Store(true)
}

func EnableExternalRequestsFunc() {
	externalRequestsOff.Store(false)
}

func InitPlVal(cfg *PipelineCfg, upFn plmap.UploadFunc, gTags map[string]string,
	installDir string,
) error {
	prepared := PreparePlValServices(cfg)
	defer prepared.Abort()
	return InitPlValPrepared(cfg, upFn, gTags, installDir, prepared)
}

func InitPlValPrepared(cfg *PipelineCfg, upFn plmap.UploadFunc, gTags map[string]string,
	installDir string, prepared *PreparedPlValServices,
) error {
	if prepared == nil {
		return fmt.Errorf("pipeline service candidate is nil")
	}
	l = logger.SLogger("plval")

	setGrokFastPathEnabled(cfg != nil && cfg.EnableGrokFastPath)

	if cfg != nil {
		if cfg.DisableHTTPRequestFunc {
			l.Info("Pipeline disable http_request function")
			DisableExternalRequestsFunc()
		} else {
			EnableExternalRequestsFunc()
		}
		if cfg.HTTPRequestDisableInternalNet || len(cfg.HTTPRequestCIDRWhitelist) > 0 ||
			len(cfg.HTTPRequestHostWhitelist) > 0 {
			l.Info("Pipeline http_request set filter: %v %v %v", cfg.HTTPRequestDisableInternalNet,
				cfg.HTTPRequestCIDRWhitelist, cfg.HTTPRequestHostWhitelist)
		}
		funcs.ReplaceNetFilter(cfg.HTTPRequestDisableInternalNet,
			cfg.HTTPRequestCIDRWhitelist, cfg.HTTPRequestHostWhitelist)
	} else {
		EnableExternalRequestsFunc()
		funcs.ReplaceNetFilter(false, nil, nil)
	}

	pipeline.InitLog()

	// load grok pattern
	if err := LoadPatterns(datakit.PipelinePatternDir); err != nil {
		l.Warnf("load pattern from directory failed: %w", err)
	}

	var gTagsLi [][2]string

	for k, v := range gTags {
		gTagsLi = append(gTagsLi, [2]string{k, v})
	}

	// init script manager
	managerIns := NewScriptManager(upFn, gTagsLi)
	if err := managerIns.loadScriptsFromWorkspace(constants.NSDefault, filepath.Join(installDir, "pipeline"), nil); err != nil {
		cleanupManagerState(managerIns.state)
		return fmt.Errorf("initialize pipeline scripts: %w", err)
	}
	SetManager(managerIns)

	// Publish the exact service instances that an unpublished JIT runner was
	// allowed to bind. No service constructor runs after this point.
	offloadWorker, replaceOffload := prepared.Commit()

	enableAppendRunInfo.Store(cfg != nil && cfg.EnableDebugFields)

	var defaults map[point.Category]string
	if cfg != nil && len(cfg.DefaultPipeline) > 0 {
		defaults = map[point.Category]string{}
		for k, v := range cfg.DefaultPipeline {
			defaults[point.CatString(k)] = v
		}
		managerIns.UpdateDefaultScript(defaults)
		l.Infof("set default pipeline: %v", defaults)
	}
	setLocalDefaultPipeline(defaults)

	// Start consumers only after the candidate worker has been published. Bind
	// each goroutine to that generation so a later reload cannot make an old
	// goroutine consume from the new worker.
	if replaceOffload && offloadWorker != nil {
		nworkers := int(math.Ceil(float64(datakit.AvailableCPUs) * 1.5))
		l.Infof("start %d offload workers...", nworkers)
		for i := 0; i < nworkers && i < maxCustomer; i++ {
			worker := offloadWorker
			g.Go(func(ctx context.Context) error {
				return worker.Customer(ctx, point.Logging)
			})
		}
	}

	return nil
}
