# Release small-batch baseline

Linux amd64, AMD Ryzen 7 9700X, `-cpu=1` (GOMAXPROCS=1, not CPU affinity),
release library SHA256 `11169f3c46e7aa00eaa85e889a74691abe2aa037ecf4b75cd28e001bf234ec66`.
Warm cached scripts. Both sides include Point construction; JIT includes encode,
native processing, output decoding and applying mutations. Not the entire DataKit
ingestion chain. Go allocation metrics exclude Rust allocations. No AOT measured.

Command (set PLATYPUS_JIT_RUNTIME to the release library):

```sh
go test -mod=mod -tags 'pipeline_jit jitbench' ./internal/pipeline -run '^$' -bench '^BenchmarkPipelineJITRealScriptsCrossover/(json-3-cast|nginx-native)/batch-(001|008)/(pipeline-go-e2e|jit-e2e)$' -benchtime=1s -count=3 -cpu=1
```

Raw ns/op observations, per batch:

| Workload | Batch | Go observations | JIT observations | Go/JIT median ratio |
| --- | --- | --- | --- | --- |
| json-3-cast | 1 | 3537,3470,3374 | 5824,6758,6241 | 0.56 |
| json-3-cast | 8 | 26449,25793,25560 | 30583,30488,32073 | 0.84 |
| nginx-native | 1 | 27330,27949,27784 | 31235,31466,30961 | 0.89 |
| nginx-native | 8 | 227283,226113,222615 | 225312,218274,241256 | 1.00 |

All JIT samples reported zero host callbacks and zero timestamp callbacks.
No stable small-batch speedup is demonstrated: JSON is slower, nginx single-point
is slower, nginx batch eight is approximately tied with noticeable variation.
This does not support the historical 4–6x claim for the current integration.
Next: phase measurements and profiles before changing the implementation; preserve
correctness and repeat under controlled CPU placement/load. Three observations
are exploratory evidence, not tail latency or production performance acceptance.

## Phase probe

Same library/environment, benchmark selector final component changed to
`(encode|native|decode|apply)$`, `-benchtime=1000x -count=3 -cpu=1`.
Raw ns/op per batch:

| Workload/batch | encode | native | decode | apply |
| --- | --- | --- | --- | --- |
| JSON/1 | 360.7,360.0,269.6 | 3403,4032,3747 | 253.9,199.2,191.1 | 514.4,460.7,474.5 |
| JSON/8 | 5430,5012,5316 | 17718,15318,15481 | 2997,2203,2007 | 3662,2864,2848 |
| nginx/1 | 1072,511.0,319.1 | 22205,21813,25514 | 2205,821.7,1091 | 3173,3928,3509 |
| nginx/8 | 8193,7804,6494 | 214665,206800,195778 | 27737,23095,14955 | 28279,24508,20012 |

This short fixed-iteration probe has substantial variation, especially allocation
phases; use it for direction, not a precise phase percentage. Independent timings
cannot simply be summed or divided by the earlier E2E run to infer CPU shares.
`native` includes Runner cache/lease, read lock, FFI, Rust input decode/execution/
output encoding, and output copy/free. It is NOT pure generated-code execution.
All native samples had zero host callbacks. Input/output bytes per record:
JSON/1 179/87, JSON/8 165/38, nginx/1 306/385, nginx/8 292/336.

Native boundary is the largest measured component in all four scenarios. Next
profiling must separate runtime setup, parsing/grok, generated-code helpers,
serialization and allocation inside that boundary before selecting an optimization.

## Go CPU profile probe

Native sampling tools `perf` and `valgrind` were not found on PATH. Used Go CPU
profiling on nginx batch 8 JIT E2E, `-benchtime=5s -count=1 -cpu=1`, same release
library. Temporary artifacts (not durable repository inputs):
`/tmp/jit-native-profile-KkKFBg/{pipeline.test,cpu.pprof}`.
Run with the same benchmark tags/environment and:

```sh
go test -mod=mod -tags 'pipeline_jit jitbench' ./internal/pipeline -run '^$' -bench '^BenchmarkPipelineJITRealScriptsCrossover/nginx-native/batch-008/jit-e2e$' -benchtime=5s -count=1 -cpu=1 -cpuprofile=/tmp/jit-native-profile-KkKFBg/cpu.pprof -o /tmp/jit-native-profile-KkKFBg/pipeline.test
go tool pprof -top -nodecount=18 /tmp/jit-native-profile-KkKFBg/pipeline.test /tmp/jit-native-profile-KkKFBg/cpu.pprof
```

Profile duration 8.64s, CPU samples 8.22s. `runtime.cgocall` flat 6.12s (74.45%),
cumulative 6.20s (75.43%); `decodeStaticBatch` cumulative 0.40s (4.87%),
`applyJITStatic` 0.35s (4.26%), `encodeProjectedJITPoints` 0.21s (2.55%).
These are sampled attribution buckets, not isolated FFI overhead: Rust processing
is hidden beneath cgocall. The profile includes benchmark setup/calibration too.
Cannot infer grok/JSON/allocator percentages from this profile or attribute 74%
to the transition itself. Native symbolized profiling or opt-in Rust phase
instrumentation remains required. This run measured 237984 ns/batch, 33616 records/s,
zero host callbacks; profile overhead makes it unsuitable as a replacement baseline.
