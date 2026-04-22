#!/usr/bin/env bash
set -euo pipefail

VM_DIR="${EBPF_VM_DIR:-/tmp/ebpf-netlog-session-vm}"
VM_NAME="${EBPF_VM_NAME:-ebpf-netlog-lab}"
PIDFILE="${VM_DIR}/${VM_NAME}.pid"

if [[ -f "${PIDFILE}" ]]; then
	pid="$(cat "${PIDFILE}")"
	if kill -0 "${pid}" 2>/dev/null; then
		kill "${pid}"
		for _ in $(seq 1 30); do
			if ! kill -0 "${pid}" 2>/dev/null; then
				break
			fi
			sleep 1
		done
	fi
	rm -f "${PIDFILE}"
fi
