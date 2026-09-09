#!/usr/bin/env python3
"""Summarize the log matrix; incomplete measurements fail, slower rows remain visible."""
import json
import pathlib
import re
import statistics
import sys

raw = pathlib.Path(sys.argv[1])
fixtures = pathlib.Path(sys.argv[2])
batches = [int(x) for x in sys.argv[3].split(',')]
count = int(sys.argv[4])
cases = json.loads((fixtures / 'cases.json').read_text())
pattern = re.compile(r'^BenchmarkPipelineJITRealScriptsCrossover/([^/]+)/batch-(\d+)/(pipeline-go-e2e|jit-group-e2e)\s+\d+\s+([\d.]+) ns/op.*?\s(\d+) B/op\s+(\d+) allocs/op')
samples = {}
for line in raw.read_text().splitlines():
    match = pattern.match(line)
    if match:
        key = (match[1], int(match[2]), match[3])
        samples.setdefault(key, []).append({'ns': float(match[4]), 'bytes': int(match[5]), 'allocs': int(match[6])})
rows = []
for case in cases:
    message = (fixtures / case['input']).read_bytes()
    for batch in batches:
        values = {engine: samples.get((case['name'], batch, engine), []) for engine in ['pipeline-go-e2e', 'jit-group-e2e']}
        if any(len(v) != count for v in values.values()):
            raise SystemExit(f"missing/wrong sample count: {case['name']} batch {batch}")
        go, jit = (statistics.median(x['ns'] for x in values[engine]) for engine in values)
        rows.append({'name': case['name'], 'batch': batch, 'calls': case['calls'], 'kind': case['kind'],
                     'message_bytes': len(message), 'physical_lines': message.count(b'\n')+1,
                     'go_ns': go, 'jit_ns': jit, 'jit_over_go': go/jit,
                     'go_allocs': statistics.median(x['allocs'] for x in values['pipeline-go-e2e']),
                     'jit_allocs': statistics.median(x['allocs'] for x in values['jit-group-e2e']), 'samples': values})
result = {'boundary': 'DataKit group end-to-end', 'samples_per_engine': count, 'rows': rows}
raw.with_name('summary.json').write_text(json.dumps(result, indent=2)+'\n')
lines = ['| Case | Batch | Go us/record | JIT us/record | JIT/Go throughput | Go allocs/batch | JIT allocs/batch |',
         '|---|---:|---:|---:|---:|---:|---:|']
for r in rows:
    lines.append(f"| {r['name']} | {r['batch']} | {r['go_ns']/r['batch']/1000:.3f} | {r['jit_ns']/r['batch']/1000:.3f} | {r['jit_over_go']:.3f}x | {r['go_allocs']} | {r['jit_allocs']} |")
lines += ['', 'Ratios above 1 mean JIT is faster. Medians of wall-clock samples; compilation is excluded.',
          'Allocations count Go-managed memory only. Multiline input is one already assembled Point.',
          'Compare only equivalent extraction work; one whole-stack capture and three selective captures have different outputs.']
raw.with_name('summary.md').write_text('\n'.join(lines)+'\n')
print('\n'.join(lines))
