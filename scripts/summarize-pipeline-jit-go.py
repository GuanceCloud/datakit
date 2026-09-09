#!/usr/bin/env python3
"""Fail closed unless every expected DK workload exceeds pipeline-go throughput."""
import json
import pathlib
import re
import statistics
import sys

path = pathlib.Path(sys.argv[1])
mode = sys.argv[2]
if mode not in {"quick", "accept"}:
    raise SystemExit("mode must be quick or accept")
expected_samples = int(sys.argv[3]) if len(sys.argv) > 3 else (3 if mode == "quick" else 7)
if expected_samples < 1:
    raise SystemExit("sample count must be positive")
workloads = ["light-drop-cast", "json-3-cast", "nginx-native", "default-time-native",
             "elasticsearch-native", "tdengine-duration", "tdengine-error", "consul-native"]
batches = [1, 10] if mode == "quick" else [1, 4, 10, 32]
pattern = re.compile(r"^BenchmarkPipelineJITRealScriptsCrossover/([^/]+)/batch-(\d+)/"
                     r"(pipeline-go-e2e|jit-group-e2e)\s+\d+\s+([\d.]+) ns/op")
observed = {}
for line in path.read_text().splitlines():
    match = pattern.match(line)
    if match:
        key = (match[1], int(match[2]), match[3])
        observed.setdefault(key, []).append(float(match[4]))
rows = []
failed = []
for workload in workloads:
    for batch in batches:
        samples = {engine: observed.get((workload, batch, engine), [])
                   for engine in ["pipeline-go-e2e", "jit-group-e2e"]}
        if any(len(v) != expected_samples for v in samples.values()):
            failed.append(f"{workload}/{batch}: expected {expected_samples} samples per engine")
            continue
        go = statistics.median(samples["pipeline-go-e2e"])
        jit = statistics.median(samples["jit-group-e2e"])
        ratio = go / jit
        rows.append({"workload": workload, "batch": batch, "go_ns": go,
                     "jit_ns": jit, "jit_over_go": ratio, "pass": ratio > 1,
                     "samples": samples})
        if ratio <= 1:
            failed.append(f"{workload}/{batch}: {ratio:.3f}x")
result = {"mode": mode, "boundary": "DataKit runJITGroup end-to-end", "rows": rows,
          "expected_samples": expected_samples, "passed": not failed, "failures": failed}
path.with_name("go-comparison.json").write_text(json.dumps(result, indent=2) + "\n")
lines = ["| Workload | Batch | Go us/batch | JIT us/batch | JIT/Go throughput |",
         "|---|---:|---:|---:|---:|"]
for row in rows:
    lines.append(f"| {row['workload']} | {row['batch']} | {row['go_ns']/1000:.3f} | "
                 f"{row['jit_ns']/1000:.3f} | {row['jit_over_go']:.3f}x |")
lines.append("\nPASS" if not failed else "\nNOT PASSED: " + "; ".join(failed))
lines.append("\nWall-clock medians; inspect raw samples and machine load before claiming a gain. "
             "Quick mode guides iteration. Native-only AOT results do not satisfy this gate.")
report = "\n".join(lines) + "\n"
path.with_name("go-comparison.md").write_text(report)
print(report)
sys.exit(2 if failed else 0)
