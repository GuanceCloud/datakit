#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_PATH="${BIN_PATH:-/tmp/datakit-ebpf-procwatch-repro}"
LOG_PATH="${LOG_PATH:-/tmp/datakit-ebpf-procwatch-repro.log}"
PIDFILE="${PIDFILE:-/tmp/datakit-ebpf-procwatch-repro.pid}"
PPROF_PORT="${PPROF_PORT:-6073}"
TRACE_SERVER="${TRACE_SERVER:-http://127.0.0.1:9529}"
IMAGE_NAME="${IMAGE_NAME:-datakit-procwatch-repro:local}"
CONTAINER_PREFIX="${CONTAINER_PREFIX:-procwatch-repro}"
TRACE_PROC_NAME="${TRACE_PROC_NAME:-smrepro}"
ITERATIONS="${ITERATIONS:-5}"
CONTAINER_TIMEOUT="${CONTAINER_TIMEOUT:-20}"
DETACH_SETTLE_TIMEOUT="${DETACH_SETTLE_TIMEOUT:-45}"
DETACH_POLL_INTERVAL="${DETACH_POLL_INTERVAL:-1}"
KEEP_REPRO_ARTIFACTS="${KEEP_REPRO_ARTIFACTS:-0}"

run_sudo() {
  if [[ -n "${SUDO_PASSWORD:-}" ]]; then
    printf '%s\n' "${SUDO_PASSWORD}" | sudo -S -p '' "$@"
  else
    sudo "$@"
  fi
}

dump_runtime_state() {
  set +e
  echo "---- runtime state ----"
  run_sudo sh -c "tail -n 80 '${LOG_PATH}'" || true
  echo "---- bpftool prog show ----"
  run_sudo bpftool prog show | grep -E 'tracepoint__sched_process_|uprobe__go_runtime_execute' || true
  echo "---- bpftool link show ----"
  run_sudo bpftool link show | grep -E 'sched_process|runtime.execute' || true
  echo "-----------------------"
}

cleanup() {
  local exit_code="${1:-0}"
  set +e
  if [[ "${exit_code}" != "0" ]]; then
    dump_runtime_state
  fi
  if [[ "${KEEP_REPRO_ARTIFACTS}" == "1" && "${exit_code}" != "0" ]]; then
    echo "keeping repro artifacts for inspection"
    return
  fi
  if [[ -f "${PIDFILE}" ]]; then
    run_sudo kill "$(cat "${PIDFILE}")" >/dev/null 2>&1 || true
  fi
  docker rm -f $(docker ps -aq --filter "name=^${CONTAINER_PREFIX}-") >/dev/null 2>&1 || true
}
trap 'cleanup $?' EXIT

echo "[1/8] build datakit-ebpf -> ${BIN_PATH}"
(
  cd "${ROOT_DIR}"
  go build -o "${BIN_PATH}" ./cmd/datakit-ebpf
)

tmpdir="$(mktemp -d)"
cat >"${tmpdir}/main.go" <<'EOF'
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("procwatch repro start")
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			time.Sleep(250 * time.Millisecond)
			fmt.Printf("worker-%d\n", idx)
		}(i)
	}
	wg.Wait()
	time.Sleep(5 * time.Second)
	fmt.Println("procwatch repro exit")
}
EOF

echo "[2/8] build local repro binary"
(
  cd "${tmpdir}"
  CGO_ENABLED=0 GOOS=linux GOARCH="$(go env GOARCH)" go build -o "${TRACE_PROC_NAME}" main.go
)

cat >"${tmpdir}/Dockerfile" <<EOF
FROM scratch
COPY ${TRACE_PROC_NAME} /usr/local/bin/${TRACE_PROC_NAME}
ENTRYPOINT ["/usr/local/bin/${TRACE_PROC_NAME}"]
EOF

echo "[3/8] build repro image -> ${IMAGE_NAME}"
docker build -t "${IMAGE_NAME}" "${tmpdir}" >/dev/null

echo "[4/8] start datakit-ebpf with procwatch tracing"
run_sudo sh -c "'${BIN_PATH}' run \
  --enabled ebpf-net,ebpf-trace \
  --trace-uprobe \
  --trace-server '${TRACE_SERVER}' \
  --trace-name-list '${TRACE_PROC_NAME}' \
  --trace-env-list DD_SERVICE \
  --pprof-host 127.0.0.1 \
  --pprof-port '${PPROF_PORT}' \
  --interval 10s \
  --log-level debug \
  --log '${LOG_PATH}' \
  --pidfile '${PIDFILE}' >/tmp/datakit-ebpf-procwatch-repro.stdout 2>&1 &"
sleep 6

if [[ ! -f "${PIDFILE}" ]]; then
  echo "datakit-ebpf did not start" >&2
  exit 1
fi

echo "[5/8] run one local short-lived Go process"
DD_SERVICE=procwatch-repro "${tmpdir}/${TRACE_PROC_NAME}" >/tmp/${TRACE_PROC_NAME}.out 2>/tmp/${TRACE_PROC_NAME}.err
sleep 2

echo "[6/8] run ${ITERATIONS} short-lived Go containers"
for i in $(seq 1 "${ITERATIONS}"); do
  name="${CONTAINER_PREFIX}-${i}"
  cid="$(docker run -d --rm --name "${name}" -e DD_SERVICE=procwatch-repro "${IMAGE_NAME}")"
  start_ts="$(date +%s)"
  while :; do
    if ! docker ps -q --no-trunc | grep -q "${cid}"; then
      break
    fi
    if (( $(date +%s) - start_ts > CONTAINER_TIMEOUT )); then
      echo "container ${name} did not exit within ${CONTAINER_TIMEOUT}s" >&2
      docker logs "${cid}" || true
      exit 1
    fi
    sleep 1
  done
done

echo "[7/8] wait for detach hook balance to settle"
if run_sudo grep -q "target-process uprobe attach is disabled by default" "${LOG_PATH}"; then
  echo "dynamic procwatch uprobe path was not enabled; make sure --trace-uprobe, --trace-server, and trace allowlists are all set" >&2
  exit 1
fi

deadline="$(( $(date +%s) + DETACH_SETTLE_TIMEOUT ))"
while :; do
  add_count="$(run_sudo grep -c "AddHooK: .*${TRACE_PROC_NAME}" "${LOG_PATH}" || true)"
  detach_count="$(run_sudo grep -c "DetachHook: .*${TRACE_PROC_NAME}" "${LOG_PATH}" || true)"
  if [[ "${add_count}" == "${detach_count}" ]]; then
    break
  fi
  if (( $(date +%s) >= deadline )); then
    break
  fi
  sleep "${DETACH_POLL_INTERVAL}"
done

echo "[8/8] validate hook balance"

echo "AddHooK count   : ${add_count}"
echo "DetachHook count: ${detach_count}"

if [[ "${add_count}" != "${detach_count}" ]]; then
  echo "hook attach/detach count mismatch" >&2
  exit 1
fi

if [[ "${add_count}" == "0" ]]; then
  echo "no dynamic procwatch uprobes were attached; repro did not cover the target path" >&2
  exit 1
fi

echo "procwatch repro passed"
