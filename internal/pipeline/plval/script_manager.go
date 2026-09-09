// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package plval

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	plmanager "github.com/GuanceCloud/pipeline-go/manager"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	plstats "github.com/GuanceCloud/pipeline-go/stats"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

type scriptSourceState map[string]map[point.Category]map[string]string
type scriptTagState map[string]map[point.Category]map[string]string
type compiledScriptState map[string]map[point.Category]map[string]*platypus.PlScript

type managerState struct {
	raw              *plmanager.Manager
	scripts          compiledScriptState
	sources          scriptSourceState
	tags             scriptTagState
	defaults         map[point.Category]string
	relation         map[point.Category]map[string]string
	relationUpdateAt int64
	routes           map[string]bool
	moduleSnapshots  map[*platypus.PlScript]moduleSnapshotResult
}

type moduleSnapshotResult struct {
	snapshot  *pljit.ModuleSnapshot
	err       error
	eligible  bool
	admission *pljit.Admission
}

// RemoteManagerUpdate describes one remote publication. Fields whose replace
// flag is false retain their currently published value.
type RemoteManagerUpdate struct {
	Scripts          map[point.Category]map[string]string
	ReplaceScripts   bool
	Defaults         map[point.Category]string
	ReplaceDefaults  bool
	Relation         map[point.Category]map[string]string
	RelationUpdateAt int64
	ReplaceRelation  bool
}

// PreparedManagerUpdate owns an unpublished, validated manager generation.
// The caller must call Commit or Abort exactly once.
type PreparedManagerUpdate struct {
	manager        *ScriptManager
	next           *managerState
	replaceScripts bool
	once           sync.Once
	err            error
}

// ScriptManager publishes scripts, defaults, relation and JIT routes as one
// immutable generation. A batch lease pins one generation until the batch is
// complete; publication waits for old leases before reclaiming their scripts.
type ScriptManager struct {
	updateMu sync.Mutex
	rw       sync.RWMutex
	state    *managerState

	upFn                plmap.UploadFunc
	globalTags          [][2]string
	jitCheck            func(string) bool
	jitInvalidate       func(string)
	jitModuleCheck      func(*pljit.ModuleSnapshot) bool
	jitModuleInvalidate func(*pljit.ModuleSnapshot)
	jitAdmission        func(*pljit.ModuleSnapshot) *pljit.Admission
}

type ManagerLease struct {
	manager *ScriptManager
	state   *managerState
	once    sync.Once
	cleanup []func()
}

func NewScriptManager(upFn plmap.UploadFunc, globalTags [][2]string) *ScriptManager {
	raw := plmanager.NewManager(plmanager.NewManagerCfg(upFn, globalTags))
	return &ScriptManager{
		upFn:       upFn,
		globalTags: append([][2]string(nil), globalTags...),
		state: &managerState{
			raw:      raw,
			sources:  scriptSourceState{},
			tags:     scriptTagState{},
			defaults: map[point.Category]string{},
			relation: map[point.Category]map[string]string{},
			routes:   map[string]bool{},
		},
	}
}

// UploadAggregates forwards one native aggregation flush through the same
// callback used by pipeline-go AggBuckets. The upload function is immutable for
// the manager lifetime, while script generations continue to use rw leases.
func (m *ScriptManager) UploadAggregates(category point.Category, bucket string, points []*point.Point) error {
	if m == nil || m.upFn == nil {
		return errors.New("pipeline aggregate upload is not configured")
	}
	return m.upFn(category, bucket, points)
}

func (m *ScriptManager) Acquire() (*ManagerLease, bool) {
	if m == nil {
		return nil, false
	}
	m.rw.RLock()
	if m.state == nil || m.state.raw == nil {
		m.rw.RUnlock()
		return nil, false
	}
	return &ManagerLease{manager: m, state: m.state}, true
}

// AcquireContext leaves no pending lock-taking goroutine after cancellation.
// The returned lease owns the read lock until Release, just like Acquire.
func (m *ScriptManager) AcquireContext(ctx context.Context) (*ManagerLease, error) {
	if ctx == nil {
		return nil, errors.New("nil manager context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("script manager not ready")
	}
	if !m.rw.TryRLock() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ticker.C:
			}
			if m.rw.TryRLock() {
				break
			}
		}
	}
	if err := ctx.Err(); err != nil {
		m.rw.RUnlock()
		return nil, err
	}
	if m.state == nil || m.state.raw == nil {
		m.rw.RUnlock()
		return nil, errors.New("script manager not ready")
	}
	return &ManagerLease{manager: m, state: m.state}, nil
}

func (lease *ManagerLease) Manager() *plmanager.Manager {
	if lease == nil || lease.state == nil {
		return nil
	}
	return lease.state.raw
}

func (lease *ManagerLease) Relation() *plmanager.ScriptRelation {
	if lease == nil || lease.state == nil || lease.state.raw == nil {
		return nil
	}
	return lease.state.raw.GetScriptRelation()
}

func (lease *ManagerLease) JITEligible(source string) bool {
	return lease != nil && lease.state != nil && lease.state.routes[source]
}

func (lease *ManagerLease) JITModuleEligible(script *platypus.PlScript) bool {
	return lease != nil && lease.state != nil && lease.state.moduleSnapshots[script].eligible
}

func (lease *ManagerLease) JITAdmission(script *platypus.PlScript) *pljit.Admission {
	if lease == nil || lease.state == nil {
		return nil
	}
	return lease.state.moduleSnapshots[script].admission
}

// JITModuleSnapshot resolves the selected script against this batch's pinned
// generation, never against the manager's latest state. The returned snapshot
// owns its sources and can outlive the lease. It is prebuilt during materialize,
// not per Point. Route publication remains the caller's responsibility.
func (lease *ManagerLease) JITModuleSnapshot(category point.Category, script *platypus.PlScript) (*pljit.ModuleSnapshot, error) {
	if lease == nil || lease.state == nil || lease.state.raw == nil || script == nil {
		return nil, errors.New("missing manager generation or selected script")
	}
	selected, ok := lease.state.raw.QueryScript(category, script.Name(), struct{}{})
	if !ok || selected != script {
		return nil, errors.New("selected script is not owned by the pinned manager generation")
	}
	result, ok := lease.state.moduleSnapshots[script]
	if !ok {
		return nil, errors.New("selected script has no published module snapshot")
	}
	return result.snapshot, result.err
}

func buildModuleSnapshot(state *managerState, category point.Category, script *platypus.PlScript) (*pljit.ModuleSnapshot, error) {
	sources := state.sources[script.NS()][category]
	source, ok := sources[script.Name()]
	if !ok || source != script.Content() {
		return nil, errors.New("selected script does not belong to the pinned source snapshot")
	}
	closure, err := moduleSourceClosure(script.Engine(), sources)
	if err != nil {
		return nil, err
	}
	snapshot, err := pljit.NewScopedModuleSnapshot(script.NS(), category.String(), script.Name(), closure)
	if err != nil {
		return nil, err
	}
	return bindJITPatternRules(snapshot), nil
}

// Walk Go's linked call graph, including synthetic use calls from after_use.
// Do not reparse source or resolve dependency names using filesystem rules.
func moduleSourceClosure(root *plruntime.Script, sources map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	visited := make(map[*plruntime.Script]bool)
	pending := []*plruntime.Script{root}
	for len(pending) > 0 {
		current := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if current == nil {
			return nil, errors.New("nil module in linked dependency graph")
		}
		if visited[current] {
			continue
		}
		visited[current] = true
		source, ok := sources[current.Name]
		if !ok {
			return nil, fmt.Errorf("linked module %q missing from generation sources", current.Name)
		}
		result[current.Name] = source
		if len(result) > 128 {
			return nil, errors.New("JIT dependency closure exceeds 128 modules")
		}
		for _, call := range current.CallRef {
			if call == nil {
				return nil, errors.New("nil linked module call")
			}
			dependency, ok := call.PrivateData.(*plruntime.Script)
			if !ok || dependency == nil {
				return nil, errors.New("unresolved linked module call")
			}
			pending = append(pending, dependency)
		}
	}
	return result, nil
}

func (lease *ManagerLease) Release() {
	if lease == nil || lease.manager == nil {
		return
	}
	lease.once.Do(func() {
		for index := len(lease.cleanup) - 1; index >= 0; index-- {
			lease.cleanup[index]()
		}
		lease.manager.rw.RUnlock()
	})
}

// AddRelease pins resources acquired with this manager generation until the
// batch releases its single lease. It must be called before the lease is
// published to another goroutine.
func (lease *ManagerLease) AddRelease(release func()) {
	if lease != nil && release != nil {
		lease.cleanup = append(lease.cleanup, release)
	}
}

func (m *ScriptManager) QueryScript(category point.Category, name string,
	disableDefault ...struct{},
) (*platypus.PlScript, bool) {
	lease, ok := m.Acquire()
	if !ok {
		return nil, false
	}
	defer lease.Release()
	return lease.Manager().QueryScript(category, name, disableDefault...)
}

func (m *ScriptManager) ScriptCount(category point.Category) int {
	lease, ok := m.Acquire()
	if !ok {
		return 0
	}
	defer lease.Release()
	return lease.Manager().ScriptCount(category)
}

func (m *ScriptManager) RelationUpdateAt() int64 {
	lease, ok := m.Acquire()
	if !ok {
		return 0
	}
	defer lease.Release()
	if lease.Relation() == nil {
		return 0
	}
	return lease.Relation().UpdateAt()
}

func (m *ScriptManager) UpdateRelation(updateAt int64, relation map[point.Category]map[string]string) {
	prepared, err := m.PrepareRemoteUpdate(RemoteManagerUpdate{
		Relation: relation, RelationUpdateAt: updateAt, ReplaceRelation: true,
	})
	if err != nil {
		l.Errorf("reject pipeline relation update: %s", err)
		return
	}
	if err := prepared.Commit(); err != nil {
		l.Errorf("publish pipeline relation update: %s", err)
	}
}

func (m *ScriptManager) UpdateDefaultScript(defaults map[point.Category]string) {
	prepared, err := m.PrepareRemoteUpdate(RemoteManagerUpdate{
		Defaults: defaults, ReplaceDefaults: true,
	})
	if err != nil {
		l.Errorf("reject default pipeline update: %s", err)
		return
	}
	if err := prepared.Commit(); err != nil {
		l.Errorf("publish default pipeline update: %s", err)
	}
}

func (m *ScriptManager) LoadScriptsFromWorkspace(ns, path string, tags map[string]string) {
	if err := m.loadScriptsFromWorkspace(ns, path, tags); err != nil {
		l.Errorf("publish workspace pipeline: %s", err)
	}
}

func (m *ScriptManager) loadScriptsFromWorkspace(ns, path string, tags map[string]string) error {
	if path == "" {
		return nil
	}
	scripts, _ := plmanager.ReadWorkspaceScripts(path)
	diagnostics, err := m.loadWorkspaceScripts(ns, scripts, tags)
	if diagnostics != nil {
		l.Errorf("workspace pipeline compile failures (valid scripts retained): %s", diagnostics)
	}
	return err
}

func (m *ScriptManager) loadWorkspaceScripts(ns string, scripts map[point.Category]map[string]string, tags map[string]string) (error, error) {
	m.updateMu.Lock()
	// First import of this namespace tolerates bad siblings, even when other
	// namespaces are already loaded. Reloads retain the old set on failure.
	_, published := m.state.sources[ns]
	initial := !published
	sources := cloneScriptSources(m.state.sources)
	sources[ns] = cloneCategoryScripts(scripts)
	nextTags := cloneScriptTags(m.state.tags)
	nextTags[ns] = map[point.Category]map[string]string{}
	for cat := range scripts {
		nextTags[ns][cat] = cloneStringMap(tags)
	}
	next, err := m.buildStateMode(sources, nextTags, m.state.defaults, m.state.relationUpdateAt, m.state.relation, ns, nil, initial)
	if next == nil {
		m.updateMu.Unlock()
		return nil, err
	}
	commitErr := (&PreparedManagerUpdate{manager: m, next: next, replaceScripts: true}).Commit()
	return err, commitErr
}

func (m *ScriptManager) LoadScripts(ns string, scripts map[point.Category]map[string]string,
	tags map[string]string,
) {
	if err := m.LoadScriptsChecked(ns, scripts, tags); err != nil {
		l.Errorf("reject pipeline update for namespace %s: %s", ns, err)
	}
}

func (m *ScriptManager) LoadScriptsChecked(ns string,
	scripts map[point.Category]map[string]string, tags map[string]string,
) error {
	return m.update(ns, nil, scripts, tags)
}

func (m *ScriptManager) LoadScriptWithCat(category point.Category, ns string,
	scripts, tags map[string]string,
) {
	if err := m.LoadScriptWithCatChecked(category, ns, scripts, tags); err != nil {
		l.Errorf("reject pipeline update for namespace %s category %s: %s", ns, category, err)
	}
}

func (m *ScriptManager) LoadScriptWithCatChecked(category point.Category, ns string,
	scripts, tags map[string]string,
) error {
	return m.update(ns, &category, map[point.Category]map[string]string{category: scripts}, tags)
}

func (m *ScriptManager) update(ns string, only *point.Category,
	scripts map[point.Category]map[string]string, tags map[string]string,
) error {
	if m == nil {
		return errors.New("script manager is nil")
	}
	m.updateMu.Lock()
	nextSources := cloneScriptSources(m.state.sources)
	nextTags := cloneScriptTags(m.state.tags)
	categories := point.AllCategories()
	if only != nil {
		categories = []point.Category{*only}
	}
	if nextSources[ns] == nil {
		nextSources[ns] = map[point.Category]map[string]string{}
	}
	if nextTags[ns] == nil {
		nextTags[ns] = map[point.Category]map[string]string{}
	}
	for _, category := range categories {
		nextSources[ns][category] = cloneStringMap(scripts[category])
		nextTags[ns][category] = cloneStringMap(tags)
	}

	next, err := m.buildState(nextSources, nextTags, m.state.defaults,
		m.state.relationUpdateAt, m.state.relation, ns, only)
	if err != nil {
		m.updateMu.Unlock()
		return err
	}
	return (&PreparedManagerUpdate{manager: m, next: next, replaceScripts: true}).Commit()
}

// PrepareRemoteUpdate compiles a complete candidate while serializing writers.
// It intentionally keeps updateMu held until Commit or Abort so a persisted
// candidate can never be committed over a newer in-memory generation.
func (m *ScriptManager) PrepareRemoteUpdate(update RemoteManagerUpdate) (*PreparedManagerUpdate, error) {
	if m == nil {
		return nil, errors.New("script manager is nil")
	}
	m.updateMu.Lock()
	nextSources := cloneScriptSources(m.state.sources)
	nextTags := cloneScriptTags(m.state.tags)
	if update.ReplaceScripts {
		nextSources[constants.NSRemote] = cloneCategoryScripts(update.Scripts)
		delete(nextTags, constants.NSRemote)
	}
	nextDefaults := cloneCategoryStringMap(m.state.defaults)
	if update.ReplaceDefaults {
		nextDefaults = cloneCategoryStringMap(update.Defaults)
	}
	nextRelation := cloneCategoryScripts(m.state.relation)
	nextRelationUpdateAt := m.state.relationUpdateAt
	if update.ReplaceRelation {
		nextRelation = cloneCategoryScripts(update.Relation)
		nextRelationUpdateAt = update.RelationUpdateAt
	}
	var next *managerState
	if update.ReplaceScripts {
		var err error
		next, err = m.buildState(nextSources, nextTags, nextDefaults, nextRelationUpdateAt, nextRelation, constants.NSRemote, nil)
		if err != nil {
			m.updateMu.Unlock()
			return nil, err
		}
	} else {
		next = &managerState{
			sources: nextSources, tags: nextTags,
			defaults: nextDefaults, relation: nextRelation, relationUpdateAt: nextRelationUpdateAt,
			routes: cloneBoolMap(m.state.routes),
		}
	}
	return &PreparedManagerUpdate{manager: m, next: next, replaceScripts: update.ReplaceScripts}, nil
}

func (prepared *PreparedManagerUpdate) Commit() error {
	if prepared == nil || prepared.manager == nil {
		return errors.New("prepared manager update is nil")
	}
	prepared.once.Do(func() {
		m := prepared.manager
		if !prepared.replaceScripts {
			m.rw.Lock()
			prepared.next.raw = m.state.raw
			prepared.next.scripts = m.state.scripts
			prepared.next.moduleSnapshots = m.state.moduleSnapshots
			prepared.next.raw.UpdateDefaultScript(prepared.next.defaults)
			prepared.next.raw.GetScriptRelation().UpdateRelation(
				prepared.next.relationUpdateAt, prepared.next.relation)
			m.state = prepared.next
			m.rw.Unlock()
			m.updateMu.Unlock()
			return
		}
		if err := m.materializeState(prepared.next); err != nil {
			prepared.err = err
			prepared.discard()
			m.updateMu.Unlock()
			return
		}
		old := m.state
		retired := retiredSources(old.routes, prepared.next.routes)
		m.rw.Lock()
		m.state = prepared.next
		m.rw.Unlock()
		for _, source := range retired {
			if m.jitInvalidate != nil {
				m.jitInvalidate(source)
			}
		}
		publishScriptChanges(old, prepared.next)
		cleanupRetiredScripts(old, prepared.next)
		m.invalidateRetiredModules(old, prepared.next)
		m.updateMu.Unlock()
	})
	return prepared.err
}

func (prepared *PreparedManagerUpdate) Abort() {
	if prepared == nil || prepared.manager == nil {
		return
	}
	prepared.once.Do(func() {
		prepared.discard()
		prepared.manager.updateMu.Unlock()
	})
}

func (prepared *PreparedManagerUpdate) discard() {
	prepared.manager.invalidateRetiredModules(prepared.next, prepared.manager.state)
	for _, source := range retiredSources(prepared.next.routes, prepared.manager.state.routes) {
		if prepared.manager.jitInvalidate != nil {
			prepared.manager.jitInvalidate(source)
		}
	}
	cleanupRetiredScripts(prepared.next, prepared.manager.state)
}

func (m *ScriptManager) buildState(sources scriptSourceState, tags scriptTagState,
	defaults map[point.Category]string, relationUpdateAt int64,
	relation map[point.Category]map[string]string, replaceNS string, only *point.Category,
) (*managerState, error) {
	return m.buildStateMode(sources, tags, defaults, relationUpdateAt, relation, replaceNS, only, false)
}

func (m *ScriptManager) buildStateMode(sources scriptSourceState, tags scriptTagState,
	defaults map[point.Category]string, relationUpdateAt int64,
	relation map[point.Category]map[string]string, replaceNS string, only *point.Category, partial bool,
) (*managerState, error) {
	next := &managerState{sources: sources, tags: tags, defaults: cloneCategoryStringMap(defaults),
		relation: cloneCategoryScripts(relation), relationUpdateAt: relationUpdateAt, scripts: compiledScriptState{}}
	var compileErrs []error
	for ns, categories := range sources {
		next.scripts[ns] = map[point.Category]map[string]*platypus.PlScript{}
		for category, scripts := range categories {
			// Preserve the legacy replacement boundary: explicit reloads replace
			// every script in their namespace/category, even with identical source.
			// Other namespaces/categories retain their live execution state.
			if m.state != nil && (ns != replaceNS || (only != nil && category != *only)) {
				next.scripts[ns][category] = m.state.scripts[ns][category]
				continue
			}
			compiled, errs := platypus.NewScripts(scripts, lang.WithMeta(tags[ns][category]), lang.WithNS(ns), lang.WithCat(category),
				lang.WithAggBkt(m.upFn, m.globalTags), lang.WithCache(), lang.WithPtWindow())
			if len(errs) > 0 {
				failed := map[string]string{}
				for name, err := range errs {
					failed[name] = scripts[name]
					compileErrs = append(compileErrs, fmt.Errorf("%s/%s/%s: %w", ns, category, name, err))
				}
				recorder := plmanager.NewManager(plmanager.NewManagerCfg(m.upFn, m.globalTags))
				recorder.LoadScriptWithCat(category, ns, failed, tags[ns][category])
				recorder.LoadScriptWithCat(category, ns, nil, nil)
			}
			if partial {
				valid := make(map[string]string, len(compiled))
				for name, script := range compiled {
					valid[name] = script.Content()
				}
				next.sources[ns][category] = valid
			}
			next.scripts[ns][category] = compiled
		}
	}
	if len(compileErrs) > 0 && !partial {
		cleanupRetiredScripts(next, m.state)
		return nil, errors.Join(compileErrs...)
	}
	next.routes = m.nextRoutes(next.sources, true)
	return next, errors.Join(compileErrs...)
}

func (m *ScriptManager) materializeState(state *managerState) error {
	if state == nil {
		return errors.New("manager state is nil")
	}
	var scripts []*platypus.PlScript
	for _, categories := range state.scripts {
		for _, named := range categories {
			for _, script := range named {
				scripts = append(scripts, script)
			}
		}
	}
	raw, err := plmanager.NewCompiledSnapshot(plmanager.NewManagerCfg(m.upFn, m.globalTags), scripts)
	if err != nil {
		return err
	}

	raw.UpdateDefaultScript(state.defaults)
	raw.GetScriptRelation().UpdateRelation(state.relationUpdateAt, state.relation)
	state.raw = raw
	state.moduleSnapshots = make(map[*platypus.PlScript]moduleSnapshotResult)
	previous := make(map[[32]byte]bool)
	previousAdmissions := make(map[[32]byte]*pljit.Admission)
	if m.state != nil {
		for _, result := range m.state.moduleSnapshots {
			if result.snapshot != nil {
				previous[result.snapshot.Identity()] = result.eligible
				previousAdmissions[result.snapshot.BundleIdentity()] = result.admission
			}
		}
	}
	for _, categories := range state.sources {
		for category, sources := range categories {
			for name := range sources {
				script, ok := raw.QueryScript(category, name, struct{}{})
				if !ok || script == nil {
					continue
				}
				if _, exists := state.moduleSnapshots[script]; exists {
					continue
				}
				snapshot, err := buildModuleSnapshot(state, category, script)
				// A JIT resource limit must not invalidate a valid Go generation.
				// Preserve the exact error for the later admission decision.
				eligible := false
				if err == nil {
					var exists bool
					eligible, exists = previous[snapshot.Identity()]
					if !exists && m.jitModuleCheck != nil {
						eligible = m.jitModuleCheck(snapshot)
					}
				}
				result := moduleSnapshotResult{snapshot: snapshot, err: err, eligible: eligible}
				if m.jitAdmission != nil {
					result.admission = &pljit.Admission{Reason: "pre_route"}
					if err == nil {
						if old := previousAdmissions[snapshot.BundleIdentity()]; old != nil {
							result.admission = old
						} else {
							result.admission = m.jitAdmission(snapshot)
						}
					}
				}
				state.moduleSnapshots[script] = result
			}
		}
	}
	return nil
}

// SetJITAdmission changes routing metadata without rebuilding Go scripts or
// their shared state. Caller serializes it with runner publication.
func (m *ScriptManager) SetJITAdmission(check func(*pljit.ModuleSnapshot) *pljit.Admission) {
	_ = m.setJITAdmission(check, false)
}

// ReplaceJITAdmission refuses a policy-only migration of active shared state.
// It preserves the existing manager, Go scripts, native runner and services.
func (m *ScriptManager) ReplaceJITAdmission(check func(*pljit.ModuleSnapshot) *pljit.Admission) error {
	return m.setJITAdmission(check, true)
}

func (m *ScriptManager) setJITAdmission(check func(*pljit.ModuleSnapshot) *pljit.Admission, preserveState bool) error {
	if m == nil {
		return nil
	}
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	updated := make(map[*platypus.PlScript]moduleSnapshotResult, len(m.state.moduleSnapshots))
	for script, result := range m.state.moduleSnapshots {
		previous := result.admission
		result.admission = nil
		if check != nil {
			result.admission = &pljit.Admission{Reason: "pre_route"}
			if result.err == nil && result.snapshot != nil {
				result.admission = check(result.snapshot)
			}
		}
		if preserveState && previous != nil && !previous.Stateless && previous.Reason == "" &&
			(result.admission == nil || result.admission.Reason != "") {
			return fmt.Errorf("JIT policy would migrate shared state for %s; drain and restart explicitly", script.Name())
		}
		if preserveState && previous != nil && previous.Reason != "" && result.admission != nil &&
			result.admission.Reason == "" && !result.admission.Stateless {
			return fmt.Errorf("JIT policy would move Go shared state for %s into native; drain and restart explicitly", script.Name())
		}
		updated[script] = result
	}
	m.rw.Lock()
	m.state.moduleSnapshots = updated
	m.jitAdmission = check
	m.rw.Unlock()
	return nil
}

// SetJITModuleHooks installs load-time admission for module snapshots. Callers
// must serialize runtime replacement with active runner generations.
func (m *ScriptManager) SetJITModuleHooks(check func(*pljit.ModuleSnapshot) bool, invalidate func(*pljit.ModuleSnapshot)) {
	if m == nil {
		return
	}
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.jitModuleCheck, m.jitModuleInvalidate = check, invalidate
	updated := make(map[*platypus.PlScript]moduleSnapshotResult, len(m.state.moduleSnapshots))
	for script, result := range m.state.moduleSnapshots {
		result.eligible = result.err == nil && result.snapshot != nil && check != nil && check(result.snapshot)
		updated[script] = result
	}
	m.rw.Lock()
	m.state.moduleSnapshots = updated
	m.rw.Unlock()
}

func (m *ScriptManager) invalidateRetiredModules(old, next *managerState) {
	if m.jitModuleInvalidate == nil || old == nil {
		return
	}
	retained := make(map[[32]byte]bool)
	if next != nil {
		for _, result := range next.moduleSnapshots {
			if result.snapshot != nil {
				retained[result.snapshot.Identity()] = true
			}
		}
	}
	for _, result := range old.moduleSnapshots {
		if result.snapshot != nil && !retained[result.snapshot.Identity()] {
			m.jitModuleInvalidate(result.snapshot)
		}
	}
}

func (m *ScriptManager) SetJITCheck(check func(string) bool) {
	if m == nil {
		return
	}
	m.updateMu.Lock()
	m.jitCheck = check
	nextRoutes := m.nextRoutes(m.state.sources, false)
	m.rw.Lock()
	m.state.routes = nextRoutes
	m.rw.Unlock()
	m.updateMu.Unlock()
}

func (m *ScriptManager) SetJITInvalidate(invalidate func(string)) {
	if m == nil {
		return
	}
	m.updateMu.Lock()
	m.jitInvalidate = invalidate
	m.updateMu.Unlock()
}

func (m *ScriptManager) nextRoutes(sources scriptSourceState, preserveExisting bool) map[string]bool {
	routes := map[string]bool{}
	for _, namespace := range sources {
		for _, scripts := range namespace {
			for _, source := range scripts {
				if _, exists := routes[source]; exists {
					continue
				}
				// An unrelated publication must not retry a capacity-rejected
				// Go route or move an existing native route to another engine.
				if preserveExisting && m.state != nil {
					if route, exists := m.state.routes[source]; exists {
						routes[source] = route
						continue
					}
				}
				routes[source] = m.jitCheck != nil && m.jitCheck(source)
			}
		}
	}
	return routes
}

// Announce only the selected scripts that actually changed, after publication.
// Candidate construction and Abort have no successful-publication events.
func publishScriptChanges(old, next *managerState) {
	for _, category := range point.AllCategories() {
		names := map[string]bool{}
		for _, state := range []*managerState{old, next} {
			if state != nil {
				for _, cats := range state.sources {
					for name := range cats[category] {
						names[name] = true
					}
				}
			}
		}
		for name := range names {
			var before, after *platypus.PlScript
			if old != nil && old.raw != nil {
				before, _ = old.raw.QueryScript(category, name, struct{}{})
			}
			if next != nil && next.raw != nil {
				after, _ = next.raw.QueryScript(category, name, struct{}{})
			}
			if before == after {
				continue
			}
			event := &plstats.ChangeEvent{Name: name, Category: category, Time: time.Now()}
			var meta map[string]string
			if before != nil {
				event.NSOld = before.NS()
				event.ScriptOld = before.Content()
				meta = before.Meta()
				plstats.WriteUpdateTime(meta)
			}
			if after != nil {
				event.NS = after.NS()
				event.Script = after.Content()
				meta = after.Meta()
				plstats.WriteUpdateTime(meta)
			}
			switch {
			case before == nil:
				event.Op = plstats.EventOpIndex
			case after == nil:
				event.Op = plstats.EventOpIndexDelete
				event.NS = before.NS()
				event.Script = before.Content()
			case plmanager.NSFindPriority(before.NS()) > plmanager.NSFindPriority(after.NS()):
				event.Op = plstats.EventOpIndexDeleteAndBack
			default:
				event.Op = plstats.EventOpIndexUpdate
			}
			plstats.WriteEvent(event, meta)
		}
	}
}

func cleanupManagerState(state *managerState) {
	if state == nil {
		return
	}
	if state.raw == nil {
		cleanupRetiredScripts(state, nil)
		return
	}
	// A retired manager owns its entire current state, unlike two successive
	// states of a live manager. Clear its indexes as well as its resources.
	for ns := range state.sources {
		state.raw.LoadScripts(ns, nil, nil)
	}
}

func cleanupRetiredScripts(old, next *managerState) {
	if old == nil {
		return
	}
	retained := map[*platypus.PlScript]bool{}
	if next != nil {
		for _, cats := range next.scripts {
			for _, scripts := range cats {
				for _, script := range scripts {
					retained[script] = true
				}
			}
		}
	}
	for _, cats := range old.scripts {
		for _, scripts := range cats {
			for _, script := range scripts {
				if !retained[script] {
					script.Cleanup()
					retained[script] = true
				}
			}
		}
	}
}

func retiredSources(old, next map[string]bool) []string {
	retired := make([]string, 0)
	for source := range old {
		if _, ok := next[source]; !ok {
			retired = append(retired, source)
		}
	}
	return retired
}

func cloneScriptSources(sources scriptSourceState) scriptSourceState {
	clone := make(scriptSourceState, len(sources))
	for ns, categories := range sources {
		clone[ns] = cloneCategoryScripts(categories)
	}
	return clone
}

func cloneScriptTags(tags scriptTagState) scriptTagState {
	clone := make(scriptTagState, len(tags))
	for ns, categories := range tags {
		copyCategories := make(map[point.Category]map[string]string, len(categories))
		for category, values := range categories {
			copyCategories[category] = cloneStringMap(values)
		}
		clone[ns] = copyCategories
	}
	return clone
}

func cloneCategoryScripts(values map[point.Category]map[string]string) map[point.Category]map[string]string {
	clone := make(map[point.Category]map[string]string, len(values))
	for category, scripts := range values {
		clone[category] = cloneStringMap(scripts)
	}
	return clone
}

func cloneCategoryStringMap(values map[point.Category]string) map[point.Category]string {
	clone := make(map[point.Category]string, len(values))
	for category, value := range values {
		clone[category] = value
	}
	return clone
}

func cloneBoolMap(values map[string]bool) map[string]bool {
	clone := make(map[string]bool, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func cloneStringMap(values map[string]string) map[string]string {
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}
