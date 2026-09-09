// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	plmanager "github.com/GuanceCloud/pipeline-go/manager"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
	plstats "github.com/GuanceCloud/pipeline-go/stats"
	"github.com/GuanceCloud/platypus/pkg/ast"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	plval "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

const (
	plTagName    = "_pl_script"
	plTagService = "_pl_service"
	plTagNS      = "_pl_ns"
	plStatus     = "_pl_status"
	plFieldCost  = "_pl_cost" // data type: float64, unit: second

	svcName = "datakit"

	sOk     = "ok"
	sFailed = "failed"
)

type ScriptResult struct {
	pts        []*point.Point
	ptsOffload []*point.Point
	ptsCreated map[point.Category][]*point.Point
	offload    *plval.OffloadLease
}

type jitBatchDecision struct {
	reason    string
	admission *pljit.Admission
	eligible  bool
	blocked   bool
}

type pointRun struct {
	admission     *pljit.Admission
	ipdbSnapshot  ipdb.IPdb
	referSnapshot refertable.PlReferTables
	point         *point.Point
	script        *platypus.PlScript
	started       time.Time
	output        *point.Point
	created       map[point.Category][]*point.Point
	ipdbCaptured  bool
	referCaptured bool
	dropped       bool
	offload       bool
}

func (r *ScriptResult) Pts() []*point.Point {
	return r.pts
}

func (r *ScriptResult) PtsOffload() []*point.Point {
	return r.ptsOffload
}

func (r *ScriptResult) PtsCreated() map[point.Category][]*point.Point {
	return r.ptsCreated
}

func (r *ScriptResult) SendOffload(category point.Category) error {
	if r == nil || len(r.ptsOffload) == 0 {
		return nil
	}
	if r.offload == nil || r.offload.Worker() == nil {
		return fmt.Errorf("offload generation unavailable")
	}
	return r.offload.Worker().Send(category, r.ptsOffload)
}

func (r *ScriptResult) Release() {
	if r != nil && r.offload != nil {
		r.offload.Release()
	}
}

func RunPl(category point.Category, pts []*point.Point,
	plOpt *lang.LogOption,
) (reslt *ScriptResult, retErr error) {
	return RunPlContext(context.Background(), category, pts, plOpt)
}

// RunPlContext returns the partial result together with cancellation errors.
// Callers must not replay the original batch: scripts may have side effects.
func RunPlContext(ctx context.Context, category point.Category, pts []*point.Point,
	plOpt *lang.LogOption,
) (reslt *ScriptResult, retErr error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil pipeline context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("run pl: %s", err)
		}
	}()

	managerLease, jitLease, offloadLease, referSnapshot, ipdbSnapshot, err := acquirePipelineLeasesContext(ctx)
	if err != nil {
		return nil, err
	}
	defer managerLease.Release()
	sManager := managerLease.Manager()
	if jitLease != nil {
		defer jitLease.release()
	}
	retainOffloadLease := false
	defer func() {
		if !retainOffloadLease && offloadLease != nil {
			offloadLease.Release()
		}
	}()
	runner := jitLease.runner()
	if sManager.ScriptCount(category) < 1 {
		return &ScriptResult{
			pts: pts,
		}, nil
	}

	runs := make([]pointRun, len(pts))
	for i := range runs {
		runs[i].ipdbSnapshot = ipdbSnapshot
		runs[i].ipdbCaptured = true
		runs[i].referSnapshot = referSnapshot
		runs[i].referCaptured = true
	}
	groups := make(map[*platypus.PlScript][]int)
	jitDecisions := make(map[*platypus.PlScript]jitBatchDecision)
	var policyCounts [7]int
	moduleBindings := make(map[*platypus.PlScript]*pljit.BoundModules)
	moduleBindingErrors := make(map[*platypus.PlScript]error)
	compileFallbacks := 0
	runnerUnavailableFallbacks := 0
	statusMappingFallbacks := 0
	subPt := make(map[point.Category][]*point.Point)
	for index, pt := range pts {
		runs[index].point = pt
		if ctx.Err() != nil {
			runs[index].output = pt
			continue
		}
		var sMap map[string]string
		if plOpt != nil {
			sMap = plOpt.ScriptMap
		}
		script, ok := searchScript(sManager, managerLease, category, pt, sMap)

		if !ok || script == nil {
			runs[index].output = pt
			continue
		}
		runs[index].script = script

		if offloadLease != nil && offloadLease.Worker() != nil &&
			script.NS() == constants.NSRemote &&
			category == point.Logging {
			runs[index].offload = true
			continue
		}

		runs[index].started = time.Now()
		if plval.EnableAppendRunInfo() {
			pt.AddTag(plTagName, script.Name())
			pt.AddTag(plTagService, svcName)
			pt.AddTag(plTagNS, script.NS())
		}

		if runner != nil && !(category == point.Logging && plOpt != nil && plOpt.DisableAddStatusField) {
			decision, checked := jitDecisions[script]
			if !checked {
				decision.eligible = managerLease.JITEligible(script.Content())
				decision.admission = managerLease.JITAdmission(script)
				if decision.admission != nil {
					decision.reason, decision.blocked = decision.admission.Decision()
					decision.eligible = decision.reason == ""
				}
				if len(script.Engine().CallRef) > 0 && (decision.admission == nil || decision.eligible) {
					moduleEligible := managerLease.JITModuleEligible(script) || (decision.admission != nil && decision.eligible)
					decision.eligible = false
					if moduleEligible {
						if moduleRunner, ok := runner.(interface {
							BindModules(*pljit.ModuleSnapshot) (*pljit.BoundModules, error)
						}); ok {
							snapshot, bindErr := managerLease.JITModuleSnapshot(category, script)
							if bindErr == nil {
								binding, err := moduleRunner.BindModules(snapshot)
								if err == nil {
									moduleBindings[script] = binding
									defer binding.Close()
									decision.eligible = true
								} else {
									bindErr = err
								}
							}
							if bindErr != nil {
								moduleBindingErrors[script] = bindErr
								l.Warnf("pipeline JIT module binding %s failed before execution: %v", script.Name(), bindErr)
							}
						} else {
							moduleBindingErrors[script] = fmt.Errorf("active JIT runner lacks admitted module binding")
						}
					}
				}
				jitDecisions[script] = decision
				if !decision.eligible {
					l.Debugf("pipeline JIT pre-route %s to pipeline-go: script is not JIT eligible", script.Name())
				}
			}
			runs[index].admission = decision.admission
			if decision.blocked {
				failJITExecution(category, &runs[index], "quarantined_stateful", fmt.Errorf("JIT shared-state program is quarantined; explicit drain and recovery required"), 0)
				continue
			}
			if decision.reason != "" && decision.reason != "pre_route" {
				for i, reason := range jitPolicyReasons {
					if reason == decision.reason {
						policyCounts[i]++
						break
					}
				}
				runPipelineGoContext(ctx, category, &runs[index], plOpt)
				continue
			}
			if bindErr := moduleBindingErrors[script]; bindErr != nil {
				failJITExecution(category, &runs[index], "module_binding", bindErr, 0)
				continue
			}
			if decision.eligible {
				groups[script] = append(groups[script], index)
			} else {
				compileFallbacks++
				runPipelineGoContext(ctx, category, &runs[index], plOpt)
			}
		} else {
			if runner == nil {
				runnerUnavailableFallbacks++
			} else {
				admission := managerLease.JITAdmission(script)
				if _, blocked := admission.Decision(); blocked {
					runs[index].admission = admission
					failJITExecution(category, &runs[index], "quarantined_stateful", fmt.Errorf("quarantined shared state cannot use the status-mapping Go route"), 0)
					continue
				}
				statusMappingFallbacks++
			}
			runPipelineGoContext(ctx, category, &runs[index], plOpt)
		}
	}
	// These records were rejected by the immutable route decision before any
	// native batch was attempted. Keep this distinct from runtime fallback
	// reasons (process/record/apply), so the metric does not imply a failed
	// native execution.
	for i, count := range policyCounts {
		observeJITRoute(category, "direct_go", jitPolicyReasons[i], count)
	}
	observeJITFallback(category, "pre_route", compileFallbacks)
	observeJITRoute(category, "direct_go", "pre_route", compileFallbacks)
	observeJITRoute(category, "direct_go", "runner_unavailable", runnerUnavailableFallbacks)
	observeJITRoute(category, "direct_go", "status_mapping_disabled", statusMappingFallbacks)

	for script, indexes := range groups {
		var processor jitBatchProcessor = runner
		if binding := moduleBindings[script]; binding != nil {
			processor = binding
		}
		if pljit.HasExecutionControl(ctx) {
			processor = contextJITProcessor{ctx: ctx, processor: processor}
		}
		runJITGroup(processor, category, script, indexes, runs, plOpt)
	}

	ret := make([]*point.Point, 0, len(pts))
	ptsOffload := make([]*point.Point, 0)
	for index := range runs {
		run := &runs[index]
		mergeCreatedPoints(subPt, run.created)
		if run.offload {
			ptsOffload = append(ptsOffload, run.point)
		} else if !run.dropped && run.output != nil {
			ret = append(ret, run.output)
		}
	}

	retainOffloadLease = len(ptsOffload) > 0
	return &ScriptResult{
		pts:        ret,
		ptsOffload: ptsOffload,
		ptsCreated: subPt,
		offload:    offloadLease,
	}, ctx.Err()
}

// Initialization publishes manager/hooks and runner under jitInitMu. Hold a
// read lock only while acquiring both leases, never for script execution.
func acquirePipelineLeases() (*plval.ManagerLease, *jitRunnerLease, bool) {
	manager, runner, offload, _, _, err := acquirePipelineLeasesContext(context.Background())
	if offload != nil {
		offload.Release()
	}
	return manager, runner, err == nil
}

func acquirePipelineLeasesContext(ctx context.Context) (*plval.ManagerLease, *jitRunnerLease, *plval.OffloadLease, refertable.PlReferTables, ipdb.IPdb, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	// No timer allocation on the ordinary uncontended path. On contention,
	// retry without a goroutine that could outlive cancellation holding a lock.
	if !jitInitMu.TryRLock() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil, nil, nil, nil, nil, ctx.Err()
			case <-ticker.C:
			}
			if jitInitMu.TryRLock() {
				break
			}
		}
	}
	defer jitInitMu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	manager, err := plval.AcquireManagerContext(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err := ctx.Err(); err != nil {
		manager.Release()
		return nil, nil, nil, nil, nil, err
	}
	offload, _ := plval.AcquireOffload()
	var referSnapshot refertable.PlReferTables
	if lease, ok := plval.AcquireRefTb(); ok {
		referSnapshot = lease.Tables()
		manager.AddRelease(lease.Release)
	} else if table, ok := plval.GetRefTb(); ok {
		// Compatibility for tests and embedders that publish a table directly
		// without owning a pull-worker generation.
		referSnapshot = table.Tables()
	}
	var db ipdb.IPdb
	if lease, ok := plval.AcquireIPDB(); ok {
		db = lease.DB()
		manager.AddRelease(lease.Release)
	} else {
		db, _ = plval.GetIPDB()
	}
	return manager, jitRunners.acquire(), offload, referSnapshot, db, nil
}

type contextJITProcessor struct {
	ctx       context.Context
	processor jitBatchProcessor
}

func (p contextJITProcessor) Projection(source string) (pljit.InputProjection, error) {
	if err := p.ctx.Err(); err != nil {
		return pljit.InputProjection{}, err
	}
	return p.processor.Projection(source)
}

func (p contextJITProcessor) Process(source string, input []byte) (pljit.Batch, error) {
	if err := p.ctx.Err(); err != nil {
		return pljit.Batch{}, err
	}
	processor, ok := p.processor.(interface {
		ProcessContext(context.Context, string, []byte) (pljit.Batch, error)
	})
	if !ok {
		return pljit.Batch{}, fmt.Errorf("JIT processor lacks context cancellation")
	}
	return processor.ProcessContext(p.ctx, source, input)
}

func (p contextJITProcessor) ProcessInto(source string, input []byte, scratch *pljit.Batch) (pljit.Batch, error) {
	if processor, ok := p.processor.(interface {
		ProcessContextInto(context.Context, string, []byte, *pljit.Batch) (pljit.Batch, error)
	}); ok {
		return processor.ProcessContextInto(p.ctx, source, input, scratch)
	}
	return p.Process(source, input)
}

func runJITGroup(
	runner jitBatchProcessor,
	category point.Category,
	script *platypus.PlScript,
	indexes []int,
	runs []pointRun,
	plOpt *lang.LogOption,
) {
	if len(indexes) == 0 {
		return
	}
	groupStarted := time.Now()
	for _, index := range indexes {
		runs[index].started = groupStarted
	}
	bindTarget := runner
	contextRunner, hasContext := runner.(contextJITProcessor)
	if hasContext {
		bindTarget = contextRunner.processor
	}
	if binder, ok := bindTarget.(interface {
		BindSource(string) (*pljit.BoundModules, error)
	}); ok {
		bound, err := binder.BindSource(script.Content())
		if err != nil {
			for _, index := range indexes {
				failJITExecution(category, &runs[index], "projection", err, 0)
			}
			return
		}
		defer bound.Close()
		runner = bound
		if hasContext {
			runner = contextJITProcessor{ctx: contextRunner.ctx, processor: bound}
		}
	}
	projection, err := runner.Projection(script.Content())
	if err != nil {
		for _, index := range indexes {
			failJITExecution(category, &runs[index], "projection", err, 0)
		}
		return
	}
	metrics := groupJITMetrics(category)
	encoder := projectedJITEncoders.Get().(*projectedJITEncoder)
	encoder.batch.PreparePointFields()
	defer releaseProjectedJITEncoder(encoder)
	var points []*point.Point
	prepared := !projection.IsAll() || projection.AllowsRawStringValues() || projection.AllowsRawTagValues()
	for chunkStart := 0; chunkStart < len(indexes); {
		var chunkEnd int
		var encodeStarted time.Time
		if prepared {
			encodeStarted = time.Now()
			chunkEnd = encoder.prepareChunk(category, indexes, runs, projection, chunkStart)
		} else {
			chunkEnd = nextJITChunkEnd(category, indexes, runs, projection, chunkStart)
		}
		chunkIndexes := indexes[chunkStart:chunkEnd]

		var input []byte
		for {
			if prepared {
				input, err = encoder.encodePrepared(len(chunkIndexes), projection)
			} else {
				if cap(points) < len(chunkIndexes) {
					points = make([]*point.Point, len(chunkIndexes))
				} else {
					points = points[:cap(points)]
					clear(points)
					points = points[:len(chunkIndexes)]
				}
				for offset, index := range chunkIndexes {
					points[offset] = runs[index].point
				}
				encodeStarted = time.Now()
				input, err = encoder.encode(category, points, projection)
			}
			metrics.encode.observe(time.Since(encodeStarted))
			// The estimator is deliberately cheap and conservative for ordinary
			// scalar fields. Composite values can still exceed it, so split an
			// oversized encoded chunk before crossing the FFI boundary. Encoding
			// errors are also isolated to a single record: no native execution
			// has occurred yet, so splitting cannot duplicate side effects.
			if len(chunkIndexes) > 1 && (err != nil || len(input) > jitChunkTargetBytes) {
				chunkEnd = chunkStart + max(1, len(chunkIndexes)/2)
				chunkIndexes = indexes[chunkStart:chunkEnd]
				encodeStarted = time.Now()
				continue
			}
			break
		}
		if err != nil {
			for _, index := range chunkIndexes {
				failJITExecution(category, &runs[index], "encode", err, 0)
			}
			chunkStart = chunkEnd
			continue
		}

		metrics.submitted.Add(float64(len(chunkIndexes)))
		metrics.inputBytes.Observe(float64(len(input)))
		started := time.Now()
		var batch pljit.Batch
		var processErr error
		if reusable, ok := runner.(interface {
			ProcessInto(string, []byte, *pljit.Batch) (pljit.Batch, error)
		}); ok {
			batch, processErr = reusable.ProcessInto(script.Content(), input, &encoder.batch)
		} else {
			batch, processErr = runner.Process(script.Content(), input)
		}
		elapsed := time.Since(started)
		metrics.native.observe(elapsed)
		if processErr != nil || len(batch.Records) != len(chunkIndexes) {
			reason := "process"
			if processErr == nil {
				processErr = fmt.Errorf("JIT returned %d records for %d inputs", len(batch.Records), len(chunkIndexes))
				reason = "record_count"
			}
			// Process may have committed external side effects before the
			// transport/decoder failed. Never execute this chunk in Go again.
			for _, index := range chunkIndexes {
				failJITExecution(category, &runs[index], reason, processErr, elapsed/time.Duration(len(chunkIndexes)))
			}
			chunkStart = chunkEnd
			continue
		}
		protocol := batch.Protocol
		if protocol == "" {
			protocol = "unknown"
		}
		metrics.batchRecords.Observe(float64(len(chunkIndexes)))
		if protocol == "static-v2" {
			metrics.staticOutputBytes.Observe(float64(batch.EncodedBytes))
		} else if protocol == "dynamic-v3" {
			metrics.dynamicOutputBytes.Observe(float64(batch.EncodedBytes))
		} else {
			jitBytesVec.WithLabelValues(category.String(), "output", protocol).Observe(float64(batch.EncodedBytes))
		}
		for _, diagnostic := range batch.Diagnostics {
			l.Debugf("pipeline JIT %s: %s", script.Name(), diagnostic)
		}

		perPointCost := elapsed / time.Duration(len(chunkIndexes))
		var applyElapsed time.Duration
		var okRecords, droppedRecords, committedErrors int
		for offset, index := range chunkIndexes {
			run := &runs[index]
			record := batch.Records[offset]
			// Validate side outputs before applying any mutation to the input.
			// Native state may already have changed: malformed side output must
			// never cause this record to be executed again in another engine.
			emitted, emittedErr := decodeJITEmitted(record.Emitted)
			if emittedErr != nil {
				failJITExecution(category, run, "emitted_error", emittedErr, perPointCost)
				continue
			}
			switch record.Status { //nolint:exhaustive // Unknown terminals share the default failure path.
			case pljit.TerminalOK:
				var created map[point.Category][]*point.Point
				var dropped bool
				var err error
				applyStarted := time.Now()
				if batch.Static != nil {
					created, dropped, err = applyJITStatic(category, run.point, batch.Static, offset, plOpt)
				} else {
					created, dropped, err = applyJITRecord(category, run.point, record, uint64(offset), plOpt)
				}
				applyElapsed += time.Since(applyStarted)
				if err != nil {
					failJITExecution(category, run, "apply", err, perPointCost)
					continue
				}
				for emittedCategory, points := range emitted {
					if created == nil {
						created = make(map[point.Category][]*point.Point)
					}
					created[emittedCategory] = append(created[emittedCategory], points...)
				}
				run.created = created
				ptinput.PtWrap(category, run.point).KeyTime2Time()
				run.output = run.point
				run.dropped = dropped
				appendJITRunInfo(category, run)
				if dropped {
					plstats.WriteMetric(script.Meta(), 1, 1, 0, perPointCost)
				} else {
					plstats.WriteMetric(script.Meta(), 1, 0, 0, perPointCost)
				}
				okRecords++
			case pljit.TerminalDropped:
				run.created = emitted
				run.dropped = true
				appendJITRunInfo(category, run)
				plstats.WriteMetric(script.Meta(), 1, 1, 0, perPointCost)
				droppedRecords++
			case pljit.TerminalError:
				if record.CommitPrefixError {
					applyStarted := time.Now()
					var err error
					if batch.Static != nil {
						_, _, err = applyJITStatic(category, run.point, batch.Static, offset, plOpt)
					} else {
						_, _, err = applyJITRecord(category, run.point, record, uint64(offset), plOpt)
					}
					applyElapsed += time.Since(applyStarted)
					if err != nil {
						failJITExecution(category, run, "apply", err, perPointCost)
						continue
					}
					// pipeline-go exposes mutations performed before a script runtime
					// error, but does not publish created points/drop state and skips
					// normal status/time finalization.
					run.created = nil
					run.dropped = false
					run.output = run.point
					appendJITFailedRunInfo(run)
					plstats.WriteMetric(script.Meta(), 1, 0, 1, perPointCost)
					l.Warn(committedJITError(script.Name(), record.Error))
					committedErrors++
					continue
				}
				failJITExecution(category, run, "record_error", fmt.Errorf("%s", record.Error), perPointCost)
			case pljit.TerminalCancelled:
				failJITExecution(category, run, "record_cancelled", fmt.Errorf("%s", record.Error), perPointCost) //nolint:misspell // Preserve the existing metric reason.
			default:
				failJITExecution(category, run, "invalid_terminal", fmt.Errorf("invalid terminal %d", record.Status), perPointCost)
			}
		}
		// Aggregate terminal counters once per chunk; preserve exact record counts
		// without allocating Prometheus label slices for every Point.
		metrics.ok.Add(float64(okRecords))
		metrics.dropped.Add(float64(droppedRecords))
		metrics.committedError.Add(float64(committedErrors))
		metrics.apply.observe(applyElapsed)
		chunkStart = chunkEnd
	}
}

// failJITExecution preserves the available point for diagnosis, but explicitly
// marks it failed even when optional run-info tags are disabled. It must not
// imply that native execution was rolled back or safe to retry.
func failJITExecution(category point.Category, run *pointRun, reason string, cause error, cost time.Duration) {
	engineFault, runtimeWide := pljit.EngineFault(cause)
	switch reason {
	case "record_count", "invalid_terminal", "emitted_error", "apply":
		engineFault = true
	}
	if engineFault && run.admission.Quarantine(runtimeWide) {
		scope := "script"
		if runtimeWide {
			scope = "runtime"
		}
		jitQuarantinesVec.WithLabelValues(scope).Inc()
		l.Warnf("pipeline JIT quarantined %s scope for script %s: %s", scope, run.script.Name(), reason)
	}
	run.output = run.point
	run.created = nil
	run.dropped = false
	run.point.AddTag(plStatus, sFailed)
	disposition := "execution outcome may be partial; not replayed"
	if reason == "projection" || reason == "encode" {
		disposition = "not submitted to native; selected engine unchanged"
	}
	run.point.Add("pl_msg", fmt.Sprintf("JIT %s failed; %s: %v", reason, disposition, cause))
	plstats.WriteMetric(run.script.Meta(), 1, 0, 1, cost)
	observeJITRoute(category, "native", reason, 1)
	l.Warnf("pipeline JIT execution failed for %s (not replayed): %s: %v", run.script.Name(), reason, cause)
}

func runPipelineGo(
	category point.Category,
	run *pointRun,
	plOpt *lang.LogOption,
) {
	runPipelineGoContext(context.Background(), category, run, plOpt)
}

type pipelineContextSignal struct{ context.Context }

func (signal pipelineContextSignal) ExitSignal() bool { return signal.Err() != nil }

type pipelineTimeoutSignal struct {
	pipelineContextSignal
	started time.Time
	timeout time.Duration
}

func (signal *pipelineTimeoutSignal) ExitSignal() bool {
	return signal.Err() != nil || signal.timedOut()
}
func (signal *pipelineTimeoutSignal) timedOut() bool {
	return time.Since(signal.started) >= signal.timeout
}

func runPipelineGoContext(ctx context.Context, category point.Category, run *pointRun, plOpt *lang.LogOption) {
	run.started = time.Now()
	run.created = nil
	inputData := ptinput.PtWrap(category, run.point)
	if run.referCaptured {
		inputData.SetPlReferTables(run.referSnapshot)
	} else if v, ok := plval.GetRefTb(); ok {
		inputData.SetPlReferTables(v.Tables())
	}
	if run.ipdbCaptured {
		inputData.SetIPDB(run.ipdbSnapshot)
	} else if v, ok := plval.GetIPDB(); ok {
		inputData.SetIPDB(v)
	}
	timeout, limited := pljit.RecordTimeout(ctx)
	var signal interface{ ExitSignal() bool } = pipelineContextSignal{ctx}
	var clock *pipelineTimeoutSignal
	if limited {
		if timeout <= 0 {
			appendJITFailedRunInfo(run)
			run.output = run.point
			return
		}
		clock = &pipelineTimeoutSignal{pipelineContextSignal: pipelineContextSignal{ctx}, started: time.Now(), timeout: timeout}
		signal = clock
	}
	if err := run.script.Run(inputData, signal, plOpt); err != nil {
		l.Warn(err)
		if plval.EnableAppendRunInfo() {
			plCost := time.Since(run.started)
			run.point.AddTag(plStatus, sFailed)
			run.point.Add(plFieldCost, float64(plCost)/float64(time.Second))
		}
		run.output = run.point
		return
	}

	if clock != nil && clock.timedOut() {
		appendJITFailedRunInfo(run)
		run.output = run.point
		return
	}
	// The Go signal may end execution without returning a script error. Treat
	// request cancellation as interruption before publishing child outputs.
	if ctx.Err() != nil {
		run.output = run.point
		return
	}
	if points := inputData.GetSubPoint(); len(points) > 0 {
		for _, subpoint := range points {
			if !subpoint.Dropped() {
				if run.created == nil {
					run.created = make(map[point.Category][]*point.Point)
				}
				run.created[subpoint.Category()] = append(run.created[subpoint.Category()], subpoint.Point())
			}
		}
	}
	if plval.EnableAppendRunInfo() {
		plCost := time.Since(run.started)
		_ = inputData.SetTag(plStatus, sOk, ast.String)
		_ = inputData.Set(plFieldCost, float64(plCost)/float64(time.Second), ast.Float)
	}
	if ctxPts := inputData.CallbackPtWinMove(); len(ctxPts) > 0 {
		if run.created == nil {
			run.created = make(map[point.Category][]*point.Point)
		}
		run.created[category] = append(run.created[category], ctxPts...)
	}
	run.dropped = inputData.Dropped()
	if !run.dropped {
		run.output = inputData.Point()
	}
}

func appendJITRunInfo(category point.Category, run *pointRun) {
	if !plval.EnableAppendRunInfo() {
		return
	}
	inputData := ptinput.PtWrap(category, run.point)
	_ = inputData.SetTag(plStatus, sOk, ast.String)
	_ = inputData.Set(plFieldCost, float64(time.Since(run.started))/float64(time.Second), ast.Float)
}

func appendJITFailedRunInfo(run *pointRun) {
	appendJITFailedRunInfoEnabled(run, plval.EnableAppendRunInfo())
}

func appendJITFailedRunInfoEnabled(run *pointRun, enabled bool) {
	if !enabled || run == nil || run.point == nil {
		return
	}
	run.point.AddTag(plStatus, sFailed)
	run.point.Add(plFieldCost, float64(time.Since(run.started))/float64(time.Second))
}

type jitRuntimeDiagnostic struct {
	Message    string `json:"message"`
	SourceName string `json:"source_name"`
	Span       *struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"span"`
}

func committedJITError(scriptName string, payload []byte) error {
	var diagnostic jitRuntimeDiagnostic
	if json.Unmarshal(payload, &diagnostic) == nil && diagnostic.Message != "" {
		name := scriptName
		if name == "" {
			name = diagnostic.SourceName
		}
		if diagnostic.Span != nil && name != "" {
			return fmt.Errorf("%s:%d:%d: %s", name, diagnostic.Span.Line, diagnostic.Span.Column, diagnostic.Message)
		}
		if name != "" {
			return fmt.Errorf("%s: %s", name, diagnostic.Message)
		}
		return fmt.Errorf("%s", diagnostic.Message)
	}
	return fmt.Errorf("pipeline JIT script error: %s", payload)
}

func mergeCreatedPoints(target, source map[point.Category][]*point.Point) {
	for category, points := range source {
		target[category] = append(target[category], points...)
	}
}

func searchScript(center *plmanager.Manager, lease *plval.ManagerLease, cat point.Category,
	pt *point.Point, scriptMap map[string]string,
) (*platypus.PlScript, bool) {
	if pt == nil {
		return nil, false
	}
	if center == nil || lease == nil {
		return nil, false
	}

	relat := lease.Relation()
	scriptName, ok := plmanager.ScriptName(relat, cat, pt, scriptMap)
	if !ok {
		return nil, false
	}

	sc, ok := center.QueryScript(cat, scriptName)
	if ok {
		return sc, true
	} else {
		return nil, false
	}
}
