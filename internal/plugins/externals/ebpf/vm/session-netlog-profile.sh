#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
VM_DIR="${EBPF_VM_DIR:-/tmp/ebpf-netlog-session-vm}"
ARTIFACT_DIR="${EBPF_VM_ARTIFACT_DIR:-${VM_DIR}/artifacts}"
SSH_PORT="${EBPF_VM_SSH_PORT:-2222}"
VM_HOST="${EBPF_VM_HOST:-127.0.0.1}"
VM_USER="${EBPF_VM_USER:-vircoys}"
NS_COUNT="${EBPF_VM_NS_COUNT:-20}"
HOST_CONC="${EBPF_VM_HOST_CONC:-24}"
NS_CONC="${EBPF_VM_NS_CONC:-12}"
LOAD_DURATION="${EBPF_VM_LOAD_DURATION:-15s}"
PROFILE_SECS="${EBPF_VM_PROFILE_SECS:-10}"
WAIT_FLUSH="${EBPF_VM_WAIT_FLUSH:-70}"
BASE_NET="${EBPF_VM_BASE_NET:-10.240}"
HOST_MIX="${EBPF_VM_HOST_MIX:-0}"
HOST_ROOT_WORKERS="${EBPF_VM_HOST_ROOT_WORKERS:-8}"
HOST_ROOT_CONC="${EBPF_VM_HOST_ROOT_CONC:-128}"
HOST_ROOT_DURATION="${EBPF_VM_HOST_ROOT_DURATION:-${LOAD_DURATION}}"
HOST_READY_TIMEOUT="${EBPF_VM_HOST_READY_TIMEOUT:-60}"
SSH_OPTS=(-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P "${SSH_PORT}")
RUN_SSH_OPTS=(-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p "${SSH_PORT}")

mkdir -p "${ARTIFACT_DIR}" "${VM_DIR}/bin"

if [[ "${VM_HOST}" == "127.0.0.1" && "${SSH_PORT}" == "2222" ]]; then
	"${ROOT_DIR}/internal/plugins/externals/ebpf/vm/session-netlog-vm-up.sh"
fi

cd "${ROOT_DIR}"

go build -o "${VM_DIR}/bin/datakit-ebpf-netlog" ./internal/plugins/externals/ebpf/cmd/datakit-ebpf
go build -o "${VM_DIR}/bin/ok-server" ./internal/plugins/externals/ebpf/vm/ok_server.go
go build -o "${VM_DIR}/bin/loadblast" ./internal/plugins/externals/ebpf/vm/loadblast.go

ssh "${RUN_SSH_OPTS[@]}" "${VM_USER}@${VM_HOST}" \
	"sudo pkill -f '^/tmp/datakit-ebpf-netlog ' || true; sudo pkill -x ok-server || true; sudo pkill -x loadblast || true; rm -f /tmp/datakit-ebpf-netlog /tmp/ok-server /tmp/loadblast"

scp "${SSH_OPTS[@]}" "${VM_DIR}/bin/datakit-ebpf-netlog" "${VM_DIR}/bin/ok-server" "${VM_DIR}/bin/loadblast" \
	"${VM_USER}@${VM_HOST}:/tmp/"

run_remote_profile() {
ssh "${RUN_SSH_OPTS[@]}" "${VM_USER}@${VM_HOST}" \
	"NS_COUNT='${NS_COUNT}' HOST_CONC='${HOST_CONC}' NS_CONC='${NS_CONC}' LOAD_DURATION='${LOAD_DURATION}' PROFILE_SECS='${PROFILE_SECS}' WAIT_FLUSH='${WAIT_FLUSH}' BASE_NET='${BASE_NET}' bash -s" <<'REMOTE'
set -euo pipefail

cleanup() {
	sudo pkill -f '^/tmp/datakit-ebpf-netlog ' || true
	sudo pkill -x ok-server || true
	sudo pkill -x loadblast || true
	sudo pkill -f 'mock_datakit_api.py' || true
	for ns in $(ip netns list | awk '{print $1}' | grep -E '^ns[0-9]+$' || true); do
		sudo ip netns pids "${ns}" 2>/dev/null | xargs -r sudo kill || true
		sudo ip netns del "${ns}" 2>/dev/null || true
	done
	for dev in $(ip -o link show | awk -F': ' '{print $2}' | sed 's/@.*//' | grep -E '^veth[0-9]+h$' || true); do
		sudo ip link del "${dev}" 2>/dev/null || true
	done
}

cleanup

cat >/tmp/mock_datakit_api.py <<'PY'
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json, time
LOG = "/tmp/mock-datakit-api.jsonl"
class H(BaseHTTPRequestHandler):
    def do_POST(self):
        n = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(n)
        rec = {"ts": time.time(), "path": self.path, "len": len(body)}
        with open(LOG, "a", encoding="utf-8") as f:
            f.write(json.dumps(rec) + "\n")
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ok")
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ok")
    def log_message(self, fmt, *args):
        return
ThreadingHTTPServer(("127.0.0.1", 9529), H).serve_forever()
PY

sudo rm -f \
	/tmp/mock-datakit-api.jsonl \
	/tmp/datakit-ebpf-netlog.log \
	/tmp/datakit-ebpf-netlog.pid \
	/tmp/datakit-ebpf-netlog.heap.pb.gz \
	/tmp/datakit-ebpf-netlog.cpu.pb.gz \
	/tmp/datakit-ebpf-netlog-report.txt \
	/tmp/datakit-ebpf-netlog-load.pids
sudo rm -f /tmp/loadblast-*.json /tmp/server-*.log /tmp/mock-api.log

nohup python3 /tmp/mock_datakit_api.py >/tmp/mock-api.log 2>&1 </dev/null &
nohup /tmp/ok-server -addr :18081 >/tmp/server-root.log 2>&1 </dev/null &
sleep 1
curl -fsS http://127.0.0.1:9529/ok >/dev/null
curl -fsS http://127.0.0.1:18081/ready >/dev/null

for i in $(seq 1 "${NS_COUNT}"); do
	sudo ip netns add "ns${i}"
	sudo ip link add "veth${i}h" type veth peer name "veth${i}n"
	sudo ip addr add "${BASE_NET}.${i}.1/24" dev "veth${i}h"
	sudo ip link set "veth${i}h" up
	sudo ip link set "veth${i}n" netns "ns${i}"
	sudo ip -n "ns${i}" addr add "${BASE_NET}.${i}.2/24" dev "veth${i}n"
	sudo ip -n "ns${i}" link set lo up
	sudo ip -n "ns${i}" link set "veth${i}n" up
	sudo ip -n "ns${i}" route add default via "${BASE_NET}.${i}.1"
	sudo ip netns exec "ns${i}" nohup /tmp/ok-server -addr :18080 >/tmp/server-ns${i}.log 2>&1 </dev/null &
	timeout 15 bash -c "until curl -fsS http://${BASE_NET}.${i}.2:18080/ready >/dev/null 2>&1; do sleep 0.2; done"
	timeout 15 sudo ip netns exec "ns${i}" bash -c "until curl -fsS http://${BASE_NET}.${i}.1:18081/ready >/dev/null 2>&1; do sleep 0.2; done"
done

nohup sudo /tmp/datakit-ebpf-netlog run \
	--enabled bpf-netlog \
	--netlog-log \
	--netlog-metric \
	--pprof-host 127.0.0.1 \
	--pprof-port 6071 \
	--interval 60s \
	--log-level info \
	--datakit-apiserver 127.0.0.1:9529 \
	--log /tmp/datakit-ebpf-netlog.log \
	--pidfile /tmp/datakit-ebpf-netlog.pid \
	>/tmp/datakit-ebpf-netlog.stdout 2>&1 &

for _ in $(seq 1 40); do
	if [ -f /tmp/datakit-ebpf-netlog.pid ] && curl -fsS http://127.0.0.1:6071/debug/pprof/ >/dev/null 2>&1; then
		break
	fi
	sleep 1
done
pid="$(cat /tmp/datakit-ebpf-netlog.pid)"

{
	echo TOPOLOGY
	ip -br link | sed -n '1,120p'
	echo NETNS_COUNT "$(ip netns list | wc -l)"
	echo BASELINE
	ps -o pid,%cpu,%mem,rss,vsz,etime,cmd -p "${pid}"
} >/tmp/datakit-ebpf-netlog-report.txt

: >/tmp/datakit-ebpf-netlog-load.pids
for i in $(seq 1 "${NS_COUNT}"); do
	nohup /tmp/loadblast -name "host-to-ns${i}" -url "http://${BASE_NET}.${i}.2:18080/h${i}" -concurrency "${HOST_CONC}" -duration "${LOAD_DURATION}" >/tmp/loadblast-host-${i}.json 2>/tmp/loadblast-host-${i}.err &
	echo $! >>/tmp/datakit-ebpf-netlog-load.pids
	nohup sudo ip netns exec "ns${i}" /tmp/loadblast -name "ns${i}-to-root" -url "http://${BASE_NET}.${i}.1:18081/n${i}" -concurrency "${NS_CONC}" -duration "${LOAD_DURATION}" >/tmp/loadblast-ns-${i}.json 2>/tmp/loadblast-ns-${i}.err &
	echo $! >>/tmp/datakit-ebpf-netlog-load.pids
done

sleep 1
curl -fsS "http://127.0.0.1:6071/debug/pprof/profile?seconds=${PROFILE_SECS}" >/tmp/datakit-ebpf-netlog.cpu.pb.gz || true
while read -r worker_pid; do
	if [ -n "${worker_pid}" ]; then
		wait "${worker_pid}" || true
	fi
done </tmp/datakit-ebpf-netlog-load.pids
curl -fsS http://127.0.0.1:6071/debug/pprof/heap >/tmp/datakit-ebpf-netlog.heap.pb.gz || true
sleep "${WAIT_FLUSH}"

python3 - <<'PY' >>/tmp/datakit-ebpf-netlog-report.txt
import collections, glob, json, os
attempts = success = fail = 0
sum_qps = 0.0
workers = 0
slow = []
for path in sorted(glob.glob('/tmp/loadblast-*.json')):
    if not os.path.getsize(path):
        continue
    with open(path, 'r', encoding='utf-8') as f:
        rec = json.load(f)
    workers += 1
    attempts += rec.get('attempts', 0)
    success += rec.get('success', 0)
    fail += rec.get('fail', 0)
    sum_qps += rec.get('qps', 0.0)
    slow.append((rec.get('qps', 0.0), rec.get('name', ''), rec.get('fail', 0)))
print('LOAD_WORKERS', workers)
print('LOAD_ATTEMPTS', attempts)
print('LOAD_SUCCESS', success)
print('LOAD_FAIL', fail)
print('LOAD_AGG_QPS', round(sum_qps, 2))
for qps, name, fails in sorted(slow)[:10]:
    print('LOAD_LOWEST_QPS', round(qps, 2), name, fails)
PY

{
	echo AFTER_LOAD
	ps -o pid,%cpu,%mem,rss,vsz,etime,cmd -p "${pid}"
} >>/tmp/datakit-ebpf-netlog-report.txt

python3 - <<'PY' >>/tmp/datakit-ebpf-netlog-report.txt
import json, collections, os
path = '/tmp/mock-datakit-api.jsonl'
count = 0
by_path = collections.Counter()
bytes_by_path = collections.Counter()
if os.path.exists(path):
    with open(path, 'r', encoding='utf-8') as f:
        for line in f:
            if not line.strip():
                continue
            rec = json.loads(line)
            count += 1
            by_path[rec['path']] += 1
            bytes_by_path[rec['path']] += rec['len']
print('POST_TOTAL', count)
for k, v in by_path.most_common():
    print('POST_PATH', k, v, bytes_by_path[k])
PY

{
	echo LOG_TAIL
	sudo tail -n 120 /tmp/datakit-ebpf-netlog.log
} >>/tmp/datakit-ebpf-netlog-report.txt
REMOTE
}

if [[ "${HOST_MIX}" == "1" ]]; then
	HOSTMIX_DIR="${ARTIFACT_DIR}/hostmix"
	REMOTE_LOG="${ARTIFACT_DIR}/remote-netlog-profile.log"
	HOST_PIDS_FILE="${HOSTMIX_DIR}/host-load.pids"
	mkdir -p "${HOSTMIX_DIR}"
	rm -f "${HOSTMIX_DIR}"/host-load-*.json "${HOSTMIX_DIR}"/host-load-*.err "${HOST_PIDS_FILE}" "${REMOTE_LOG}"

	run_remote_profile >"${REMOTE_LOG}" 2>&1 &
	remote_pid=$!

	for _ in $(seq 1 "${HOST_READY_TIMEOUT}"); do
		if curl -fsS "http://${VM_HOST}:18081/ready" >/dev/null 2>&1; then
			break
		fi
		sleep 1
	done

	: >"${HOST_PIDS_FILE}"
	for i in $(seq 1 "${HOST_ROOT_WORKERS}"); do
		nohup "${VM_DIR}/bin/loadblast" \
			-name "hostmix-${i}" \
			-url "http://${VM_HOST}:18081/hostmix${i}" \
			-concurrency "${HOST_ROOT_CONC}" \
			-duration "${HOST_ROOT_DURATION}" \
			>"${HOSTMIX_DIR}/host-load-${i}.json" \
			2>"${HOSTMIX_DIR}/host-load-${i}.err" &
		echo $! >>"${HOST_PIDS_FILE}"
	done

	while read -r worker_pid; do
		if [ -n "${worker_pid}" ]; then
			wait "${worker_pid}" || true
		fi
	done <"${HOST_PIDS_FILE}"

	wait "${remote_pid}"

	HOSTMIX_DIR="${HOSTMIX_DIR}" python3 - <<'PY' >"${HOSTMIX_DIR}/hostmix-summary.txt"
import collections, glob, json, os
attempts = success = fail = 0
sum_qps = 0.0
workers = 0
slow = []
for path in sorted(glob.glob(os.path.join(os.environ["HOSTMIX_DIR"], "host-load-*.json"))):
    if not os.path.getsize(path):
        continue
    with open(path, "r", encoding="utf-8") as f:
        rec = json.load(f)
    workers += 1
    attempts += rec.get("attempts", 0)
    success += rec.get("success", 0)
    fail += rec.get("fail", 0)
    sum_qps += rec.get("qps", 0.0)
    slow.append((rec.get("qps", 0.0), rec.get("name", ""), rec.get("fail", 0)))
print("HOST_LOAD_WORKERS", workers)
print("HOST_ATTEMPTS", attempts)
print("HOST_SUCCESS", success)
print("HOST_FAIL", fail)
print("HOST_AGG_QPS", round(sum_qps, 2))
for qps, name, fails in sorted(slow)[:10]:
    print("HOST_LOWEST_QPS", round(qps, 2), name, fails)
PY
else
	run_remote_profile
fi

scp "${SSH_OPTS[@]}" \
	"${VM_USER}@${VM_HOST}:/tmp/datakit-ebpf-netlog.cpu.pb.gz" \
	"${VM_USER}@${VM_HOST}:/tmp/datakit-ebpf-netlog.heap.pb.gz" \
	"${VM_USER}@${VM_HOST}:/tmp/datakit-ebpf-netlog-report.txt" \
	"${ARTIFACT_DIR}/"

echo "artifacts saved under ${ARTIFACT_DIR}"
