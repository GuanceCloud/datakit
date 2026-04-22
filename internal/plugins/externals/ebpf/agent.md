# eBPF Agent Notes

This subtree contains the external eBPF runtime, helpers, and generated assets used by `datakit-ebpf`.

## Working Rules

- Prefer making and validating changes inside `internal/plugins/externals/ebpf/...` without touching unrelated collectors.
- Keep Go changes compatible with the repository lint rules in `AGENTS_COMMON.md`, especially dead code, unused imports, and explicit error handling.
- For low-risk validation, prefer targeted commands such as `go test ./internal/plugins/externals/ebpf/...`.
- If the task depends on runtime eBPF behavior, use the local KVM/libvirt guest `ebpf-stress-lab` instead of assuming the host environment matches production expectations.

## VM Context

- Libvirt/KVM is installed on the host and the default network is active.
- The VM workspace is `/var/lib/libvirt/images/ebpf-stress`.
- The intended reusable guest name is `ebpf-stress-lab`.
- If `virsh` fails with socket permissions for user `vircoys`, start a fresh shell session so the `libvirt` group membership is picked up.
- For heavier runtime validation, the guest has already been exercised with `8` network namespaces and `8` veth pairs using a local mock DataKit API on `127.0.0.1:9529` and collector pprof on `127.0.0.1:6071`.
- The guest has also been validated with a `40`-namespace, `40`-veth topology using temporary Go binaries `/tmp/ok-server` and `/tmp/loadblast`; on the current `4 vCPU / 8 GiB` VM that harness reached about `57.6k` aggregate QPS.
- The reusable stress guest was later resized to `8 vCPU / 12 GiB`; with host-side load plus guest-internal `40`-namespace traffic, the same topology reached about `240k` combined QPS.
- Under bursty short-lived process churn, `procwatch/catalog.go` currently logs many PID-resolution failures for disappeared `/proc/<pid>` entries; treat that as a known performance-noise hotspot when reading pressure-test logs.
- After the procwatch suppression patch, the disappeared-`/proc` noise was reduced, but a separate verifier failure appeared in one rerun: `ebpf-net` load rejected `tracepoint__sys_exit_writev` with an unbounded memory access error.
- The verifier follow-up required rebuilding the embedded ELF object with `make -C internal/plugins/externals/ebpf apiflow.o`; plain `go build` does not refresh the embedded `apiflow.o`.
- A conservative verifier-safe fix is now in place: `apiflow` uploads one event per perf output and `readv/writev` capture reads only the first useful iovec segment. A VM smoke run confirmed `ebpf-net tracer(ebpf)` and `api tracer starting ...` now load successfully.
