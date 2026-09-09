#!/usr/bin/env python3
"""Compare native artifacts and optionally Go binaries, in ABBA order.

Compile the benchmark binary before running. Use the existing matrix fixture
environment variables when selecting log cases. No builds run during sampling.
--group-by-workload interleaves engines per workload/batch to shorten the gap
between comparable samples; a separate 1x discovery run is excluded from results.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import statistics
import subprocess


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', required=True, type=Path)
    parser.add_argument('--candidate-binary', type=Path, help='optional Go candidate; defaults to --binary')
    parser.add_argument('--baseline', required=True, type=Path)
    parser.add_argument('--candidate', required=True, type=Path)
    parser.add_argument('--filter', required=True)
    parser.add_argument('--output', required=True, type=Path)
    parser.add_argument('--duration', default='200ms')
    parser.add_argument('--cycles', default=2, type=int, help='ABBA cycles; two samples per variant per cycle')
    parser.add_argument('--cpu', default=2, type=int)
    parser.add_argument('--group-by-workload', action='store_true',
                        help='run ABBA separately for each workload/batch, keeping selected modes together')
    args = parser.parse_args()
    if args.cycles < 1:
        parser.error('--cycles must be positive')
    args.output.mkdir(parents=True, exist_ok=False)
    paths = {name: getattr(args, name).resolve(strict=True) for name in ['binary', 'baseline', 'candidate']}
    if args.candidate_binary is not None:
        paths['candidate_binary'] = args.candidate_binary.resolve(strict=True)
    metadata = {
        'artifacts': {name: {'path': str(path), 'sha256': digest(path)} for name, path in paths.items()},
        'filter': args.filter, 'duration': args.duration, 'order': 'ABBA' * args.cycles,
        'cpu': args.cpu, 'go_cpu': 1,
        'interleave': 'workload-batch' if args.group_by_workload else 'matrix',
        'fixture_environment': {k: v for k, v in os.environ.items() if k in [
            'JIT_BENCH_LOG_MATRIX', 'JIT_BENCH_LOG_DIRECTORY', 'PIPELINE_SCRIPT_DIRECTORY']},
    }
    fixture_dir = os.environ.get('JIT_BENCH_LOG_DIRECTORY')
    if fixture_dir:
        metadata['fixture_sha256'] = {p.name: digest(p) for p in sorted(Path(fixture_dir).iterdir()) if p.is_file()}
    (args.output / 'metadata.json').write_text(json.dumps(metadata, indent=2) + '\n')
    command = ['taskset', '-c', str(args.cpu), str(paths['binary']), '-test.run', '^$',
               '-test.bench', args.filter, '-test.benchtime', args.duration, '-test.count', '1', '-test.cpu', '1']
    pattern = re.compile(r'^(Benchmark\S+)\s+\d+\s+([\d.]+) ns/op', re.M)
    groups = [(args.filter, None)]
    if args.group_by_workload:
        # Discover subbenchmarks with one untimed-for-comparison iteration.
        # Keep the discovery output, but exclude it from all reported samples.
        discovery = command.copy()
        discovery[discovery.index('-test.benchtime') + 1] = '1x'
        env = os.environ.copy()
        env['PLATYPUS_JIT_RUNTIME'] = str(paths['baseline'])
        raw = args.output / 'discovery.txt'
        with raw.open('w') as stream:
            subprocess.run(discovery, env=env, stdout=stream, stderr=subprocess.STDOUT, check=True)
        found = pattern.findall(raw.read_text())
        names = {name for name, _ in found}
        if not names or len(names) != len(found):
            raise RuntimeError(f'empty or duplicate benchmark discovery in {raw}')
        grouped = {}
        for name in sorted(names):
            prefix, separator, mode = name.rpartition('/')
            if not separator:
                raise RuntimeError(f'benchmark has no mode suffix: {name}')
            grouped.setdefault(prefix, []).append((name, mode))
        groups = []
        for prefix, entries in grouped.items():
            # Go treats each slash-separated component as its own regexp.
            selection = '/'.join('^' + re.escape(part) + '$' for part in prefix.split('/'))
            selection += '/^(' + '|'.join(re.escape(mode) for _, mode in entries) + ')$'
            groups.append((selection, {name for name, _ in entries}))
        metadata['execution_groups'] = [sorted(names) for _, names in groups]
        metadata['discovery_duration'] = '1x'
        (args.output / 'metadata.json').write_text(json.dumps(metadata, indent=2) + '\n')
    samples = {}
    for group_index, (selection, expected) in enumerate(groups):
        directory = args.output if not args.group_by_workload else args.output / f'group-{group_index:02d}'
        directory.mkdir(exist_ok=True)
        for index, variant in enumerate(metadata['order']):
            env = os.environ.copy()
            env['PLATYPUS_JIT_RUNTIME'] = str(paths['baseline' if variant == 'A' else 'candidate'])
            run_command = command.copy()
            run_command[run_command.index('-test.bench') + 1] = selection
            if variant == 'B' and 'candidate_binary' in paths:
                run_command[3] = str(paths['candidate_binary'])
            raw = directory / f'{index:02d}-{variant}.txt'
            with raw.open('w') as stream:
                subprocess.run(run_command, env=env, stdout=stream, stderr=subprocess.STDOUT, check=True)
            found = pattern.findall(raw.read_text())
            names = {name for name, _ in found}
            if not names or len(names) != len(found) or (expected is not None and names != expected):
                raise RuntimeError(f'empty, duplicate or inconsistent benchmark selection in {raw}')
            expected = names
            for name, ns in found:
                samples.setdefault(name, {'A': [], 'B': []})[variant].append(float(ns))
            print(f'group {group_index + 1}/{len(groups)} completed {index + 1}/{len(metadata["order"])} variant={variant}', flush=True)
    for name, path in paths.items():
        if digest(path) != metadata['artifacts'][name]['sha256']:
            raise RuntimeError(f'artifact changed during measurement: {path}')
    rows = []
    lines = ['| Benchmark | Baseline ns/op | Candidate ns/op | Speedup | A min/max | B min/max |',
             '|---|---:|---:|---:|---:|---:|']
    for name, values in sorted(samples.items()):
        if any(len(v) != args.cycles * 2 for v in values.values()):
            raise RuntimeError(f'incomplete samples: {name}')
        a, b = (statistics.median(values[k]) for k in ['A', 'B'])
        rows.append({'name': name, 'baseline_ns': a, 'candidate_ns': b, 'speedup': a / b, 'samples': values})
        lines.append(f'| {name} | {a:.1f} | {b:.1f} | {a / b:.3f}x | '
                     f'{min(values["A"]):.0f}/{max(values["A"]):.0f} | '
                     f'{min(values["B"]):.0f}/{max(values["B"]):.0f} |')
    (args.output / 'summary.json').write_text(json.dumps(rows, indent=2) + '\n')
    (args.output / 'summary.md').write_text('\n'.join(lines) + '\n')
    print('\n'.join(lines))


if __name__ == '__main__':
    main()
