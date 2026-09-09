# Built-in script JIT / Go E2E — final delivery branch, 2026-09-09

本报告保留最终测量结论与适用范围。原始数据、profile、复现脚本及 fixtures 已从当前源码目录移出，可在[历史实验归档](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/tree/e309a3734f71b06b909643019c68a54392018019/internal/pipeline/benchmarks)查看。下文中的 results.json、metadata.json、preflight.txt 和脚本名均指归档中的对应文件；最终测量目录为 [jit-builtin-final-20260909](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/tree/e309a3734f71b06b909643019c68a54392018019/internal/pipeline/benchmarks/jit-builtin-final-20260909)，采集器样例位于 [jit-builtin-20260908](https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/tree/e309a3734f71b06b909643019c68a54392018019/internal/pipeline/benchmarks/jit-builtin-20260908)。这些是历史提交的测量结果，并非最新提交重新测试的数据。


DK benchmark source c8946f4cc8; Rust fdd63c39, scalar-slots. Both on feat-jit-aot80. A newly built DK test binary was frozen before measurement; normal runtime SHA-256 is 8620ae54ffabede3a0d7091e178c3eae47d9b0d925701ec8d46ad072597b8ba4. Exact binary and diagnostic hashes are in results.json/runtime-results.json.

## Scope and correctness

Eight unmodified DK collector PipelineConfig scripts and ten original LogExamples, verified against current source using ../jit-builtin-20260908/verify_sources.py. Includes single-line logs and a multiline MySQL slow-log record. These are repository examples repeatedly replayed, not a diverse customer production corpus. Script and fixture hashes are in metadata.json; originals are in ../jit-builtin-20260908/fixtures.

Complete pipeline E2E includes Point construction, RunPlContext routing, Go/native execution, writeback and result release. Excludes collector acquisition, network upload and cold compilation. Uses Background context, JIT auto route. Each benchmark warmup compares Point contents in both directions excluding only _pl_cost and verifies native submission for JIT. Timed iterations check errors and output record count. All accepted invocations passed. Full preflight/control results are in preflight.txt.

## Normal E2E

Four alternating JIT/Go rounds per case, 500ms per benchmark, CPU2, GOMAXPROCS1. Median wall us/record. Go/JIT >1 means JIT faster. Batch10 JIT medians lead on 9/10 cases; Redis trails. Batch1 leads on 7/10; Apache, ordinary MySQL and Redis trail. This is a sample-set result, not a production workload-weighted speedup.

| Case / batch | JIT us/record | Go us/record | Go/JIT |
|---|---:|---:|---:|
| nginx-access-log/batch-001 | 22.859 | 27.255 | 1.19x |
| nginx-access-log/batch-010 | 21.477 | 26.404 | 1.23x |
| nginx-error-log1/batch-001 | 23.444 | 29.618 | 1.26x |
| nginx-error-log1/batch-010 | 19.962 | 35.416 | 1.77x |
| apache-access-log/batch-001 | 9.796 | 9.457 | 0.97x |
| apache-access-log/batch-010 | 7.359 | 8.938 | 1.21x |
| mysql-log/batch-001 | 12.430 | 10.212 | 0.82x |
| mysql-log/batch-010 | 8.605 | 9.877 | 1.15x |
| mysql-slow-log/batch-001 | 17.627 | 30.477 | 1.73x |
| mysql-slow-log/batch-010 | 14.807 | 29.148 | 1.97x |
| redis-log/batch-001 | 7.914 | 5.873 | 0.74x |
| redis-log/batch-010 | 5.559 | 5.243 | 0.94x |
| elasticsearch-search-slow-log/batch-001 | 19.930 | 49.733 | 2.50x |
| elasticsearch-search-slow-log/batch-010 | 15.758 | 48.550 | 3.08x |
| mongodb-log/batch-001 | 8.530 | 10.127 | 1.19x |
| mongodb-log/batch-010 | 6.201 | 9.437 | 1.52x |
| consul-log/batch-001 | 8.268 | 24.447 | 2.96x |
| consul-log/batch-010 | 5.915 | 25.224 | 4.26x |
| tdengine-log-204/batch-001 | 9.413 | 16.088 | 1.71x |
| tdengine-log-204/batch-010 | 6.858 | 15.337 | 2.24x |

## Runtime diagnostics

Separate phase-profile library; three 8192-batch repetitions, six 4096-batch execution windows per case; Go RunStmts three 250ms repetitions with input preparation and status/time finalization outside timing. All ten JIT runtime medians lead Go in this run. These are instrumented independent measurements: do not subtract them from normal E2E to compute an adapter duration. In particular phase/host variation can put a Go runtime sample above a separately measured E2E sample.

| Case (batch 10) | Go runtime us | JIT runtime us |
|---|---:|---:|
| nginx-access-log | 23.434 | 12.271 |
| nginx-error-log1 | 22.185 | 13.966 |
| apache-access-log | 7.061 | 4.442 |
| mysql-log | 12.954 | 5.091 |
| mysql-slow-log | 27.088 | 10.360 |
| redis-log | 3.512 | 2.348 |
| elasticsearch-search-slow-log | 38.860 | 11.758 |
| mongodb-log | 7.503 | 3.231 |
| consul-log | 21.371 | 2.916 |
| tdengine-log-204 | 13.147 | 2.834 |

## Variability and rejected samples

Compiler overlap is monitored and rejected. The initial Go runtime pass encountered an external compiler, so that pass was retained as rejected and rerun. Accepted normal E2E and runtime invocations had no observed compiler overlap. Host activity and frequency are not fully controlled, and broad sample ranges remain; precise ratios are exploratory, especially Nginx. No sample was discarded only for being slow.

| Case/batch | JIT min–max us | Go min–max us |
|---|---:|---:|
| nginx-access-log/batch-001 | 20.432–23.889 | 25.400–55.438 |
| nginx-access-log/batch-010 | 17.521–24.143 | 24.881–52.524 |
| nginx-error-log1/batch-001 | 22.069–36.598 | 25.647–50.664 |
| nginx-error-log1/batch-010 | 18.605–21.973 | 24.187–48.827 |
| apache-access-log/batch-001 | 9.639–11.196 | 9.097–16.169 |
| apache-access-log/batch-010 | 7.275–7.521 | 8.634–11.238 |
| mysql-log/batch-001 | 11.673–17.026 | 10.048–10.515 |
| mysql-log/batch-010 | 8.391–13.404 | 9.532–10.280 |
| mysql-slow-log/batch-001 | 17.216–18.497 | 28.631–49.795 |
| mysql-slow-log/batch-010 | 14.552–15.260 | 28.402–29.663 |
| redis-log/batch-001 | 7.719–8.598 | 5.781–6.003 |
| redis-log/batch-010 | 5.238–9.078 | 5.191–5.332 |
| elasticsearch-search-slow-log/batch-001 | 17.558–24.460 | 48.546–50.724 |
| elasticsearch-search-slow-log/batch-010 | 15.455–22.788 | 47.823–70.005 |
| mongodb-log/batch-001 | 8.431–10.310 | 9.491–16.250 |
| mongodb-log/batch-010 | 6.042–6.391 | 9.155–9.606 |
| consul-log/batch-001 | 7.951–13.586 | 21.681–28.981 |
| consul-log/batch-010 | 5.731–7.680 | 22.799–25.854 |
| tdengine-log-204/batch-001 | 9.346–10.786 | 15.510–19.154 |
| tdengine-log-204/batch-010 | 6.592–8.186 | 14.829–25.633 |

## Reproduction

Build the test binary with `go test -c -tags 'pipeline_jit jitbench' -o /tmp/jit-builtin-final-20260909/pipeline.test ./internal/pipeline` from DK root. Run bench.py with JIT_BENCH_OUTPUT, JIT_BENCH_BINARY, and JIT_BENCH_FIXTURES pointing to the output directory, frozen binary and verified built-in fixture directory. bench.py records the frozen normal library path. runtime.py uses the separately frozen phase-profile library; runtime_retry.py retains the accepted JIT phase samples and reruns Go only after overlap. All scripts preserve the local paths used for this exact run. Do not compile during timed samples.

No production code was changed in this verification. Passing local tests does not establish CI publication, deployment or long-running production behavior.
