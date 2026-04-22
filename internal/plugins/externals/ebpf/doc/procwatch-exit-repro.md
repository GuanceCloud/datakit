# Procwatch Container Exit Repro

This repro checks the short-lived container / dynamic uprobe lifecycle handled by `procwatch`.
It also prints `bpftool` state and recent eBPF logs on failure so it doubles as a diagnosis entrypoint.

## What it does

1. builds `datakit-ebpf`
2. starts it with:
   - `ebpf-net`
   - `ebpf-trace`
   - `--trace-uprobe`
   - `--trace-server http://127.0.0.1:9529`
   - `--trace-name-list smrepro`
3. builds a tiny short-lived Go container image
4. runs that container repeatedly
5. fails if:
   - a container does not exit in time
   - hook counts do not converge within the detach grace window
   - `AddHooK` and `DetachHook` counts are not balanced in the eBPF log
6. on failure, prints:
   - the last `datakit-ebpf` log lines
   - `bpftool prog show` snippets for `sched_process_*` / `runtime.execute`
   - `bpftool link show` snippets for the active tracepoint/uprobe links

## Usage

```bash
cd internal/plugins/externals/ebpf
SUDO_PASSWORD='your-sudo-password' ./repro-procwatch-container-exit.sh
```

Optional environment variables:

```bash
ITERATIONS=10
CONTAINER_TIMEOUT=30
PPROF_PORT=6073
TRACE_SERVER=http://127.0.0.1:9529
DETACH_SETTLE_TIMEOUT=45
DETACH_POLL_INTERVAL=1
LOG_PATH=/tmp/datakit-ebpf-procwatch-repro.log
KEEP_REPRO_ARTIFACTS=1
```

## Expected result

- every `procwatch-repro-*` container exits normally
- `AddHooK` count equals `DetachHook` count
- the script prints `procwatch repro passed`

If the script still fails to cover the dynamic uprobe path, set `KEEP_REPRO_ARTIFACTS=1`.
That keeps the current `datakit-ebpf` process, image, and logs on disk for manual inspection.

Current note:
- a runtime check on 2026-04-22 showed the previous harness first missed `--trace-uprobe`, and then still failed the runtime safety gate without a non-empty `--trace-server`
- another runtime check on 2026-04-22 showed immediate `AddHooK` vs `DetachHook` comparison was invalid because procwatch keeps binaries in a 30s detach grace window before final cleanup
- the script now enables both `--trace-uprobe` and `--trace-server` explicitly, and fails fast if logs still show the target-process uprobe path is disabled
- the script now waits for hook counts to settle before deciding there is a detach leak
