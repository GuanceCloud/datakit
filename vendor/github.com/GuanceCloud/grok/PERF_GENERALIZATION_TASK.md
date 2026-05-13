# Grok Performance Generalization Task

Date: `2026-04-14`
Owner: `Codex + repository maintainer`
Status: `phase-4-completed`

## Background

The current repository already has two useful optimization directions:

- structured fast path for a subset of Grok patterns
- regexp-path improvements such as typed extraction reuse and cheap prefiltering

These help on real fixtures, but most decisions are still driven by concrete examples in `testdata/datakit_pipeline_cases.json`.

The next task is to make optimization decisions more generic:

- less dependent on fixture-specific heuristics
- more traceable at compile time
- applicable to broader Grok syntax, including user-defined patterns

## Problem Statement

Today, performance behavior is split across:

- direct regexp execution
- structured matcher execution
- ad hoc compile-time heuristics

This makes it hard to answer:

- why a given pattern did or did not use the fast path
- which parts of a pattern are safe to prefilter
- how to extend fast-path coverage without adding case-specific rules
- how to benchmark mismatch-heavy workloads in a reusable way

## Goal

Introduce a generic compilation model for Grok patterns that supports:

- traceable compile-time analysis
- generic prefilter extraction
- generic fast-path eligibility scoring
- broader structured execution on a syntax-subset basis
- benchmark-driven validation for both match and mismatch workloads

## Non-Goals

- changing the exported API
- rewriting the entire engine in one step
- removing regexp fallback
- targeting only known Datakit fixture layouts

## Design Direction

### 1. Add Layered Pattern IR

Compile Grok patterns into internal IR before producing regexp or structured matchers.

The IR should be split into two layers:

- `Grok IR`
  - preserves Grok syntax and naming semantics
  - examples: literal segments, `%{PATTERN:alias:type}` references, optional blocks, alternation, repetition
- `Match IR`
  - normalized execution-oriented representation derived from Grok IR
  - examples: literal nodes, primitive token nodes, branch nodes, optional nodes, repetition nodes, capture metadata

Responsibilities:

- `Grok IR` is the traceable parse product
- `Match IR` is the analysis and execution input

Together they become the source of truth for:

- regexp generation
- structured matcher generation
- prefilter generation
- complexity scoring

### 2. Compile Generic Prefilters from IR

Extract only information that is provably required.

Initial scope must stay conservative:

- fixed prefix
- required literal fragments
- minimum length

Deferred until later phases:

- fixed suffix
- separator inference beyond simple literals
- character-class-derived prefilters

Rules:

- derivation must be conservative
- false negatives are allowed
- false positives are not allowed

The prefilter must never change semantic matching results.

### 3. Add Fast-Path Scoring Beside Existing Heuristics

Current fast-path selection relies on several hand-written heuristics.

Introduce an explainable score computed from IR properties:

- number of greedy segments
- number of optional segments
- alternation count
- nesting depth
- estimated backtracking risk
- number of deterministic delimiter boundaries

Outputs:

- `fast-path preferred`
- `regexp preferred`
- `prefilter only`

Rollout rule:

- phase 3 does not immediately replace existing heuristics
- phase 3 first runs in observe-only mode
- benchmark and fixture comparison decide whether score-based decisions can later take over

### 4. Expand Structured Matcher by Syntax Coverage

Do not optimize by named fixture.
Optimize by syntax capability:

- concatenation with deterministic literals
- bounded optional segments
- bounded alternation
- primitive token parsing
- delimiter-bounded `DATA` and `GREEDYDATA`

When a pattern is partially supported:

- supported regions use structured execution
- unsupported regions remain regexp fallback

Scope rule:

- mixed-mode execution is not part of the default plan for this task
- IR design should allow future mixed-mode work, but phase 4 should assume full-pattern fast path or full regexp fallback

### 5. Add General Benchmark Matrix

Benchmarks must cover more than fixed examples.

Required classes:

- real-world fixtures from `testdata/datakit_pipeline_cases.json`
- synthetic parameterized match cases
- synthetic parameterized mismatch cases
- pipeline-order benchmarks where N patterns are tested before success
- long-line tail-mismatch benchmarks

## Deliverables

### Phase 1: IR Foundation

- internal `Grok IR` types
- internal `Match IR` types
- Grok-pattern-to-IR compiler
- stable IR dump / snapshot helper for tests
- tests that assert IR shape for representative patterns
- no exported API changes
- no execution-path decision changes

Acceptance:

- existing tests still pass
- IR compile is deterministic
- representative patterns can be explained from IR dumps
- dump format is stable enough for snapshot-style assertions

### Phase 2: Generic Prefilter

- IR-based conservative prefilter extraction
- prefilter application before regexp fallback
- correctness tests proving parity with no-prefilter execution

Acceptance:

- no match-order divergence in existing fixture tests
- `BenchmarkDatakitPipelineFirstMatchRegexpPath` improves on mismatch-heavy paths
- no correctness regressions in `go test ./...`
- prefix and minimum-length inference are proven safe before broader literal inference expands

### Phase 3: Fast-Path Scoring Observation

- add IR-driven scoring without removing existing heuristics
- add explainability for why a pattern got a given execution mode

Acceptance:

- compile-time decision is reproducible
- score output can be asserted in tests for representative patterns
- observe-only scoring does not change current execution behavior
- comparison output is sufficient to evaluate later heuristic replacement

### Phase 4: Structured Coverage Expansion

- support more generic syntax combinations
- reduce dependence on regexp for delimiter-bounded wide tokens

Acceptance:

- expanded benchmark wins without introducing fixture-specific branches
- fallback parity remains intact

### Phase 5: Tooling and Traceability

- optional debug dump for pattern analysis
- optional benchmark snapshot update process
- per-pattern decision trace for tests or debug logs

Acceptance:

- a maintainer can inspect why a pattern used regexp or fast path
- benchmark changes can be tied back to a specific phase and metric

## Validation Plan

### Correctness

Always run:

```bash
env GOCACHE=/tmp/grok-go-cache go test ./...
```

Core parity checks:

- fixture parity between current path and forced regexp path
- typed extraction parity
- optional-segment parity
- first-match order parity for pipeline evaluation

### Performance

Keep these benchmark families:

```bash
env GOCACHE=/tmp/grok-go-cache go test -run '^$' -bench '^BenchmarkDatakitFixtures$' -benchmem
env GOCACHE=/tmp/grok-go-cache go test -run '^$' -bench '^BenchmarkDatakitPipelineFirstMatchRegexpPath$' -benchmem
env GOCACHE=/tmp/grok-go-cache go test -run '^$' -bench '^BenchmarkRunWithTypeInfo(|RegexpPath)$' -benchmem
```

Additional benchmark work to add in later phases:

- synthetic mismatch matrix
- synthetic long-line mismatch
- optional-hit-rate sweeps
- pattern-order sweeps

### Performance Gates

Each phase that changes behavior must satisfy:

- `go test ./...` stays green
- `BenchmarkDatakitPipelineFirstMatchRegexpPath` must not regress by more than `5%`
- known strong fast-path cases in `BenchmarkDatakitFixtures` must not regress by more than `10%` without an explicit documented tradeoff
- any optimization that mainly targets mismatch cost must state that explicitly in results
- compile-time cost changes must be recorded when a phase adds new compiler analysis

## Traceability Rules

Every implementation PR or patch for this task should record:

- phase number
- affected files
- correctness commands run
- benchmark commands run
- before/after benchmark numbers
- known non-goals or deferred work

The task document should also be updated after each completed phase with:

- `Status`
- `Decision Log`
- `Observed Results`

Recommended commit/patch footer template:

```text
Perf-Task: generalization-phase-<N>
Validation:
- env GOCACHE=/tmp/grok-go-cache go test ./...
- <benchmark command>
```

## Current Baseline Notes

As of this document:

- `RunWithTypeInfo` already benefits from structured extraction reuse
- `BenchmarkDatakitPipelineFirstMatchRegexpPath` exists to track regexp-path mismatch cost
- structured fast path is still partly controlled by heuristics in `structured_fastpath.go`
- regexp fallback remains the correctness anchor

These are starting points, not the final architecture.

## Observed Results

- `2026-04-14`: task document created; no phase completed yet
- `2026-04-14`: phase 1 implemented with internal `Grok IR`, derived `Match IR`, raw-pattern parser, and stable dump-based tests; no execution-path behavior changes were introduced
- `2026-04-14`: phase 2 implemented IR-based conservative regexp prefilters using prefix, required literal fragments, and minimum match length; `BenchmarkDatakitPipelineFirstMatchRegexpPath` improved from `12043 ns/op` without prefilter to `8626 ns/op` with prefilter in the validation run
- `2026-04-14`: phase 3 implemented observe-only IR-based fast-path scoring with stable dump output and representative verdict tests; runtime behavior remains unchanged, and `BenchmarkDatakitPipelineFirstMatchRegexpPath` stayed in the same range during validation (`9833 ns/op` prefilter, `9890 ns/op` no_prefilter)
- `2026-04-14`: phase 4 expanded structured coverage for bounded group repetition, enabling generic patterns such as `%{HTTPD20_ERRORLOG}` to use the structured path without fixture-specific branches; validation run showed `BenchmarkStructuredBoundedRepeatPatterns/HTTPD20_ERRORLOG` improve from `8785 ns/op` on forced regexp to `716.0 ns/op` on the structured path, while `BenchmarkDatakitPipelineFirstMatchRegexpPath` remained healthy (`6802 ns/op` prefilter, `10819 ns/op` no_prefilter)
- `2026-04-14`: post-phase-4 prefilter follow-up added IR-derived fixed suffix filtering for end-anchored regexp patterns only; validation run showed `BenchmarkRegexpPrefilterTailMismatch` improve from `251472 ns/op` without prefilter to `14.08 ns/op` with prefilter, while `BenchmarkDatakitPipelineFirstMatchRegexpPath` remained healthy (`9937 ns/op` prefilter, `20753 ns/op` no_prefilter)
- `2026-04-14`: post-phase-4 compile-time follow-up added global `Grok IR` / `Match IR` compile caches and removed duplicate IR traversal between prefilter and scoring; representative compile-time benchmark deltas were `simple/grok_ir: 499.2 -> 17.31 ns/op`, `simple/prefilter_ir: 10200 -> 2753 ns/op`, `httpd20_errorlog/prefilter_ir: 2477083 -> 228365 ns/op`, and `postgres_complex/score_ir: 165105 -> 60774 ns/op`
- `2026-04-14`: observe-only comparison output was added for `IR score` versus current structured fast-path heuristic; representative tests now cover aligned cases and one stable divergence sample (`^foo.*bar$`), while runtime behavior remains unchanged
- `2026-04-14`: representative observation catalog and benchmark were added to make score-versus-heuristic drift measurable across multiple pattern families; current observation costs include `simple_structured: 4072 ns/op`, `bounded_greedy_tail: 21403 ns/op`, `postgres_complex: 42640 ns/op`, `raw_regex_prefilter: 2818 ns/op`, `httpd20_errorlog: 116964 ns/op`, and `apache_error_pid_only: 39687 ns/op`

## Risks

- over-aggressive prefilter inference can change match order
- generalized scoring can regress known fast cases if thresholds are poorly tuned
- IR duplication with existing matcher compilation can increase maintenance cost
- mixed-mode execution may be more complex than pure structured or pure regexp execution

## Open Questions

- `Grok IR` should be built from raw Grok syntax
- `Match IR` should be built from `Grok IR` after semantic expansion and normalization
- Do we want a debug-only IR dump API, or only test helpers?
- mixed-mode execution is out of scope for this task; phase 4 stays pure full-pattern fast path versus regexp fallback
- Should benchmark snapshots be committed into `BENCHMARKS.md` each phase, or only at milestones?

## Decision Log

- `2026-04-14`: Chosen direction is IR-first generalization instead of continuing fixture-specific heuristic tuning.
- `2026-04-14`: Prefilter correctness is prioritized over aggressive literal inference.
- `2026-04-14`: No exported API changes are planned for this task.
- `2026-04-14`: Phase 1 is intentionally limited to modeling and traceability; execution behavior changes are deferred.
- `2026-04-14`: Two-layer IR (`Grok IR` and `Match IR`) is preferred over a single mixed-responsibility IR.
- `2026-04-14`: Fast-path scoring should start in observe-only mode before replacing existing heuristics.
- `2026-04-14`: `Grok IR` is sourced from raw Grok syntax; `Match IR` is derived from `Grok IR`, not directly from denormalized regexp text.
- `2026-04-14`: Mixed-mode execution is explicitly deferred beyond this task.
- `2026-04-14`: Performance gates are numerical where practical to reduce review ambiguity.
- `2026-04-14`: Phase 1 implementation keeps IR completely off the runtime matching path; it is currently traceability-only infrastructure.
- `2026-04-14`: Phase 2 derives regexp prefilters from `Match IR` rather than regexp text; rollout remains conservative and correctness-first.
- `2026-04-14`: Phase 3 scoring is analyze-only; it does not participate in matcher selection until a later decision explicitly promotes it.
- `2026-04-14`: Phase 4 is intentionally limited to bounded group repetition that the current matcher model can execute reliably; broader finite-repeat coverage such as char-class expansion is deferred until it shows both correctness and a benchmark win.
- `2026-04-14`: Fixed suffix prefiltering is only safe for end-anchored patterns under `FindStringSubmatchIndex` semantics; unanchored patterns must not use whole-line `HasSuffix` checks.
- `2026-04-14`: IR analysis remains compile-time-only, so cache work should target parse and analysis reuse first before considering runtime IR execution.
- `2026-04-14`: Score-versus-heuristic comparison should remain observe-only until divergence samples are cataloged and benchmarked; the comparison layer exists to explain disagreement, not to override runtime choices yet.
- `2026-04-14`: Representative observation coverage should include both aligned and divergent samples, and at least one case each for heuristic allow/fallback and score fast-path/prefilter-only verdicts.
