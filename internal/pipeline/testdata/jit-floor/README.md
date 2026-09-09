# Adapter floor probes

These four scripts isolate local assignment, Point reads, constant writes and
scalar conversion. They use the existing real-script benchmark and include
the same Point construction and group adapter work as other end-to-end rows.
The local assignment is a minimal-work probe, not an assertion of zero native
instructions. The message includes its final newline.

Set JIT_BENCH_LOG_MATRIX=1 and JIT_BENCH_LOG_DIRECTORY to this directory, then
select `BenchmarkPipelineJITRealScriptsCrossover/floor-/batch-(001|010)/...`.
Use a precompiled benchmark binary or the generic quick script; the log-matrix
shell wrapper deliberately selects the separate JSON/Grok fixtures.

Compare pipeline-go-e2e and jit-group-e2e. Other phases use different harnesses
and cannot be added together as an exact group cost. Required output values
and complete Go/JIT Point equivalence are checked before group timing.
