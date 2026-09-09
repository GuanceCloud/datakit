# JSON, Grok and script execution matrix

These are deterministic synthetic application-log fixtures, not captured customer
traffic or a production workload distribution. They cover common structured
request events, key=value application logs and Java exception stacks. Scripts,
raw input and required outputs are all committed for inspection and reproduction.

`cases.json` defines 30 cases. A benchmark checks every required output against Go
and then compares the complete JIT Point with Go, excluding only measured pipeline
cost. Empty Grok matches therefore cannot silently count as successful parsing.
No executor/runtime code is changed for this matrix.

- JSON: 1/3/10 distinct path extractions on the same small nested document;
  1/10 on a large compact document; 10 on the same large document formatted with
  indentation/newlines; 3 from independently stored JSON fields.
- GJSON: 1/10 basic paths on a small document and 10 on a large document.
- load_json: parse once then extract 1/10 fields, with small and large inputs;
  parse three times and read the same field. These include local assignment and
  indexing costs, not just the parser.
- Script execution: a 20-step local assignment chain, 20 arithmetic expressions,
  20 branches, 100/1000-iteration loops, and loops with branches or break/continue. Seed values come from Point fields to
  avoid constant-only expressions. Numeric seeds are floats (JSON fixture loading);
  the loop counter is integer. Each script publishes one checked result. These
  measure whole scripts, not isolated operator latency.
- Single-line Grok: 1/3/10 calls extracting distinct fields; one call extracting
  all 10 fields; two regex misses followed by a full match.
- Multiline Grok: Java exception events at two sizes. One call captures the whole
  stack; three calls select the header, exception and root cause. These last two
  scripts intentionally have different output sizes and must not be presented
  as equivalent one-call/three-call implementations.

A multiline log is **one Point whose message contains actual LF bytes**, already
assembled by the collector. A batch is 1/10/32 independent copies of that event.
Pretty JSON contains actual newlines between JSON tokens; it is not JSONL. This
matrix does not measure file reading, multiline assembly or collector scheduling.
Large JSON includes an unselected context field near its end to expose scan and
transport costs; the required keys occur earlier in the document.

Call counts refer to the script source. Adjacent JSON path calls on one source can
be fused by the current JIT compiler. Three independent JSON source fields cannot
be combined into one document parse. There is no source/fixture-specific runtime
change or fallback to Go inside the measured JIT route.

Run from the DataKit root with the current native library:

```sh
PLATYPUS_JIT_RUNTIME=../platplusplus/target/jit-over-go/artifacts/branches.so \
  JIT_BENCH_CPU=2 scripts/benchmark-pipeline-jit-log-matrix.sh quick
PLATYPUS_JIT_RUNTIME=../platplusplus/target/jit-over-go/artifacts/branches.so \
  JIT_BENCH_CPU=2 scripts/benchmark-pipeline-jit-log-matrix.sh full
```

Quick: batch 1/10, 100 ms x3. Full: batch 1/10/32, 300 ms x5. Each round runs
Go and JIT consecutively per case, reducing long-run drift between engines.
The summary fails on missing samples; it retains ratios below 1 as results of
this exploratory matrix rather than treating them as a test failure. Compilation
is excluded from warm throughput; Point creation, transport, execute, apply and
metrics are included. Allocation counters cover Go-managed memory only.

The assignment/call controls also include 20 string copies, 20 `len` return-value
assignments, 20 constant `add_key` calls, and 20 alternating case conversions.
The `len` fixture publishes a boolean equality check so its expected JSON value
retains the exact Go type. These separate call preparation and local-variable
costs from JSON/grok parsing; all still use the regular pipeline adapter.
