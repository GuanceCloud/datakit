#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
VM_DIR="${EBPF_VM_DIR:-/tmp/ebpf-compat-vm}"
VM_NAME="${EBPF_VM_NAME:-ebpf-compat}"
SSH_PORT="${EBPF_VM_SSH_PORT:-2222}"
QEMU_BIN="${QEMU_BIN:-qemu-system-x86_64}"
QEMU_ACCEL="${QEMU_ACCEL:--enable-kvm}"
SSH_KEY="${EBPF_VM_SSH_KEY:-$HOME/.ssh/id_ed25519.pub}"
VM_USER="${EBPF_VM_USER:-vircoys}"
VM_OS_FAMILY="${EBPF_VM_OS_FAMILY:-ubuntu}"
BASE_URL="${EBPF_VM_BASE_URL:?set EBPF_VM_BASE_URL}"
BASE_IMG_NAME="${EBPF_VM_BASE_IMG_NAME:-$(basename "${BASE_URL}")}"
BASE_IMG="${VM_DIR}/${BASE_IMG_NAME}"
OVERLAY_IMG="${VM_DIR}/${VM_NAME}.qcow2"
SEED_IMG="${VM_DIR}/${VM_NAME}-seed.img"
PIDFILE="${VM_DIR}/${VM_NAME}.pid"
SERIAL_LOG="${VM_DIR}/${VM_NAME}-serial.log"
ARTIFACT_DIR="${VM_DIR}/${VM_NAME}-artifacts"
SMOKE_BIN="${EBPF_VM_BIN:-/tmp/datakit-ebpf-smoke}"
REMOTE_SMOKE_BIN="${EBPF_VM_REMOTE_BIN:-/tmp/datakit-ebpf-smoke-$(date +%s)}"
ENABLED_PLUGINS="${EBPF_VM_ENABLED_PLUGINS:-ebpf-net}"
L7_ENABLED="${EBPF_VM_L7_ENABLED:-httpflow}"
REMOTE_PREP="${EBPF_VM_REMOTE_PREP:-}"
REMOTE_OFFSET_PATH="${EBPF_VM_REMOTE_OFFSET_PATH:-/usr/local/datakit/externals/datakit-ebpf.offset}"
SCP_OPTS=(-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -P "${SSH_PORT}")
SSH_OPTS=(-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=5 -p "${SSH_PORT}")

mkdir -p "${VM_DIR}" "${ARTIFACT_DIR}"

if [[ ! -f "${SSH_KEY}" ]]; then
	echo "missing ssh public key: ${SSH_KEY}" >&2
	exit 1
fi

if [[ ! -f "${SMOKE_BIN}" ]]; then
	echo "missing collector binary: ${SMOKE_BIN}" >&2
	exit 1
fi

if [[ ! -f "${BASE_IMG}" ]]; then
	curl -L --fail --progress-bar "${BASE_URL}" -o "${BASE_IMG}"
fi

if [[ ! -f "${OVERLAY_IMG}" ]]; then
	qemu-img create -f qcow2 -F qcow2 -b "${BASE_IMG}" "${OVERLAY_IMG}" 40G >/dev/null
fi

case "${VM_OS_FAMILY}" in
ubuntu)
	VM_GROUPS="[adm, sudo]"
	INIT_CMDS=$(cat <<'EOF'
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y sudo curl python3 gcc libc6-dev procps iproute2
sysctl -w kernel.perf_event_paranoid=1 || true
sysctl -w kernel.kptr_restrict=0 || true
EOF
)
	;;
centos)
	VM_GROUPS="[wheel]"
	INIT_CMDS=$(cat <<'EOF'
yum install -y sudo curl python gcc glibc-devel procps-ng iproute
sysctl -w kernel.perf_event_paranoid=1 || true
sysctl -w kernel.kptr_restrict=0 || true
EOF
)
	;;
*)
	echo "unsupported VM_OS_FAMILY=${VM_OS_FAMILY}" >&2
	exit 1
	;;
esac

cat >"${VM_DIR}/meta-data" <<EOF
instance-id: ${VM_NAME}
local-hostname: ${VM_NAME}
EOF

cat >"${VM_DIR}/user-data" <<EOF
#cloud-config
hostname: ${VM_NAME}
manage_etc_hosts: true
users:
  - default
  - name: ${VM_USER}
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: ${VM_GROUPS}
    shell: /bin/bash
    ssh_authorized_keys:
      - $(<"${SSH_KEY}")
ssh_pwauth: false
runcmd:
  - [ bash, -lc, '${INIT_CMDS//$'\n'/; }' ]
final_message: "ebpf compat vm is ready after \$UPTIME seconds"
EOF

cloud-localds "${SEED_IMG}" "${VM_DIR}/user-data" "${VM_DIR}/meta-data" >/dev/null

cleanup() {
	if [[ -f "${PIDFILE}" ]]; then
		if kill -0 "$(cat "${PIDFILE}")" 2>/dev/null; then
			kill "$(cat "${PIDFILE}")" 2>/dev/null || true
		fi
	fi
}

trap cleanup EXIT

"${QEMU_BIN}" \
	${QEMU_ACCEL} \
	-daemonize \
	-name "${VM_NAME}" \
	-pidfile "${PIDFILE}" \
	-display none \
	-serial "file:${SERIAL_LOG}" \
	-cpu host \
	-smp "${EBPF_VM_CPUS:-2}" \
	-m "${EBPF_VM_MEMORY_MB:-4096}" \
	-device virtio-rng-pci \
	-netdev "user,id=net0,hostfwd=tcp:127.0.0.1:${SSH_PORT}-:22" \
	-device virtio-net-pci,netdev=net0 \
	-drive "if=virtio,format=qcow2,file=${OVERLAY_IMG}" \
	-drive "if=virtio,format=raw,file=${SEED_IMG}"

for _ in $(seq 1 240); do
	if ssh "${SSH_OPTS[@]}" "${VM_USER}@127.0.0.1" "echo ok" >/dev/null 2>&1; then
		break
	fi
	sleep 2
done

ssh "${SSH_OPTS[@]}" "${VM_USER}@127.0.0.1" "cloud-init status --wait >/dev/null 2>&1 || true"
ssh "${SSH_OPTS[@]}" "${VM_USER}@127.0.0.1" "uname -a; cat /etc/os-release || true" | tee "${ARTIFACT_DIR}/os-release.txt"

scp "${SCP_OPTS[@]}" "${SMOKE_BIN}" "${VM_USER}@127.0.0.1:${REMOTE_SMOKE_BIN}" >/dev/null

ssh "${SSH_OPTS[@]}" "${VM_USER}@127.0.0.1" \
	"EBPF_VM_REMOTE_BIN='${REMOTE_SMOKE_BIN}' EBPF_VM_ENABLED_PLUGINS='${ENABLED_PLUGINS}' EBPF_VM_L7_ENABLED='${L7_ENABLED}' EBPF_VM_REMOTE_PREP='${REMOTE_PREP}' EBPF_VM_REMOTE_OFFSET_PATH='${REMOTE_OFFSET_PATH}' bash -s" <<'REMOTE'
set -euo pipefail

REMOTE_SMOKE_BIN="${EBPF_VM_REMOTE_BIN:-/tmp/datakit-ebpf-smoke}"
ENABLED_PLUGINS="${EBPF_VM_ENABLED_PLUGINS:-ebpf-net}"
L7_ENABLED="${EBPF_VM_L7_ENABLED:-httpflow}"
REMOTE_PREP="${EBPF_VM_REMOTE_PREP:-}"
REMOTE_OFFSET_PATH="${EBPF_VM_REMOTE_OFFSET_PATH:-/usr/local/datakit/externals/datakit-ebpf.offset}"

rm -f /tmp/mock_api.py /tmp/web_server.py /tmp/mock-api.jsonl /tmp/collector.log /tmp/collector.stdout
rm -f /tmp/collector.log.copy /tmp/collector.stdout.copy /tmp/collector.exitcode /tmp/collector.exitcode.copy
rm -f /tmp/collector.metrics /tmp/collector.ps /tmp/collector.dmesg /tmp/collector.dmesg.copy
rm -f /tmp/collector.offset /tmp/collector.offset.copy
pkill -f '^python .*mock_api.py' || true
pkill -f '^python3 .*mock_api.py' || true
pkill -f '^python .*web_server.py' || true
pkill -f '^python3 .*web_server.py' || true
sudo pkill -f 'datakit-ebpf-smoke.* run ' || true
sudo chmod +x "${REMOTE_SMOKE_BIN}"

if [ -n "${REMOTE_PREP}" ]; then
    sudo bash -lc "${REMOTE_PREP}"
fi

if printf '%s' ",${ENABLED_PLUGINS}," | grep -q ',ebpf-conntrack,'; then
    sudo modprobe xt_CT >/dev/null 2>&1 || true
    if command -v iptables >/dev/null 2>&1; then
        sudo iptables -t raw -C OUTPUT -j CT >/dev/null 2>&1 || \
            sudo iptables -t raw -A OUTPUT -j CT >/dev/null 2>&1 || true
    fi
fi

cat >/tmp/mock_api.py <<'PY'
from __future__ import print_function
try:
    from http.server import BaseHTTPRequestHandler, HTTPServer
except ImportError:
    from BaseHTTPServer import BaseHTTPRequestHandler, HTTPServer
import base64
import gzip
import json
import time
try:
    from io import BytesIO
except ImportError:
    from StringIO import StringIO as BytesIO
try:
    import zlib
except ImportError:
    zlib = None
LOG = "/tmp/mock-api.jsonl"


def decode_body(headers, body):
    data = body
    encoding = (headers.get("Content-Encoding", "") or "").lower()

    try:
        if encoding == "gzip" or body[:2] == b"\x1f\x8b":
            data = gzip.GzipFile(fileobj=BytesIO(body)).read()
        elif encoding == "deflate" and zlib is not None:
            data = zlib.decompress(body)
    except Exception:
        data = body

    try:
        text = data.decode("utf-8", "replace")
    except Exception:
        text = ""

    return {
        "encoding": encoding,
        "decoded_len": len(data),
        "body_preview": text[:16384],
        "body_b64_prefix": base64.b64encode(data[:256]).decode("ascii"),
        "has_dst_nat": ("dst_nat_ip" in text) or ("dst_nat_port" in text),
    }

class H(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ok")
    def do_POST(self):
        n = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(n)
        rec = {"ts": time.time(), "path": self.path, "len": len(body)}
        rec.update(decode_body(self.headers, body))
        with open(LOG, "a") as f:
            f.write(json.dumps(rec) + "\n")
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ok")
    def log_message(self, fmt, *args):
        return
HTTPServer(("127.0.0.1", 9529), H).serve_forever()
PY

cat >/tmp/web_server.py <<'PY'
from __future__ import print_function
try:
    from http.server import BaseHTTPRequestHandler, HTTPServer
except ImportError:
    from BaseHTTPServer import BaseHTTPRequestHandler, HTTPServer
class H(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ok")
    def log_message(self, fmt, *args):
        return
HTTPServer(("0.0.0.0", 18080), H).serve_forever()
PY

(python3 /tmp/mock_api.py >/tmp/mock-api.out 2>/tmp/mock-api.err || python /tmp/mock_api.py >/tmp/mock-api.out 2>/tmp/mock-api.err) &
(python3 /tmp/web_server.py >/tmp/web.out 2>/tmp/web.err || python /tmp/web_server.py >/tmp/web.out 2>/tmp/web.err) &
sleep 2
GUEST_IP="$(hostname -I | awk '{print $1}')"
if [ -z "${GUEST_IP}" ]; then
    GUEST_IP="$(ip -4 route get 1.1.1.1 | awk '{for (i=1;i<=NF;i++) if ($i == "src") {print $(i+1); exit}}')"
fi
if [ -z "${GUEST_IP}" ]; then
    echo "failed to resolve guest IPv4 address" >&2
    exit 1
fi
curl -fsS http://127.0.0.1:9529/ >/dev/null
curl -fsS "http://${GUEST_IP}:18080/" >/dev/null
if printf '%s' ",${ENABLED_PLUGINS}," | grep -q ',ebpf-conntrack,'; then
    if command -v iptables >/dev/null 2>&1; then
        sudo iptables -t nat -C OUTPUT -p tcp -d "${GUEST_IP}" --dport 18081 -j REDIRECT --to-ports 18080 >/dev/null 2>&1 || \
            sudo iptables -t nat -A OUTPUT -p tcp -d "${GUEST_IP}" --dport 18081 -j REDIRECT --to-ports 18080 >/dev/null 2>&1 || true
    fi
    curl -fsS "http://${GUEST_IP}:18081/" >/dev/null 2>&1 || true
fi

nohup sudo env EBPF_VM_REMOTE_BIN="${REMOTE_SMOKE_BIN}" bash -lc '
  "$EBPF_VM_REMOTE_BIN" run \
    --enabled "'"${ENABLED_PLUGINS}"'" \
    --l7net-enabled "'"${L7_ENABLED}"'" \
    --pprof-host 127.0.0.1 \
    --pprof-port 6071 \
    --interval 5s \
    --datakit-apiserver 127.0.0.1:9529 \
    --log /tmp/collector.log \
    --log-level debug \
    >/tmp/collector.stdout 2>&1
  rc=$?
  printf "%s\n" "$rc" >/tmp/collector.exitcode
' >/dev/null 2>&1 &

for _ in $(seq 1 40); do
    if curl -fsS http://127.0.0.1:6071/debug/pprof/ >/dev/null 2>&1; then
        break
    fi
    sleep 1
done

for _ in $(seq 1 20); do
    curl -fsS "http://${GUEST_IP}:18080/" >/dev/null 2>&1 || true
    if printf '%s' ",${ENABLED_PLUGINS}," | grep -q ',ebpf-conntrack,'; then
        curl -fsS "http://${GUEST_IP}:18081/" >/dev/null 2>&1 || true
    fi
    sleep 0.2
done

for _ in $(seq 1 45); do
    if grep -q 'ebpf-net%2Fhttpflow' /tmp/mock-api.jsonl 2>/dev/null; then
        break
    fi
    curl -fsS "http://${GUEST_IP}:18080/" >/dev/null 2>&1 || true
    if printf '%s' ",${ENABLED_PLUGINS}," | grep -q ',ebpf-conntrack,'; then
        curl -fsS "http://${GUEST_IP}:18081/" >/dev/null 2>&1 || true
    fi
    sleep 1
done

curl -fsS http://127.0.0.1:6071/metrics >/tmp/collector.metrics || true
ps -ef | grep 'datakit-ebpf-smoke.* run' | grep -v grep >/tmp/collector.ps || true
sudo tail -n 200 /tmp/collector.log >/tmp/collector.log.tail || true
sudo tail -n 200 /tmp/collector.stdout >/tmp/collector.stdout.tail || true
sudo cp /tmp/collector.log /tmp/collector.log.copy 2>/dev/null || true
sudo cp /tmp/collector.stdout /tmp/collector.stdout.copy 2>/dev/null || true
sudo cp /tmp/collector.exitcode /tmp/collector.exitcode.copy 2>/dev/null || true
sudo cp "${REMOTE_OFFSET_PATH}" /tmp/collector.offset.copy 2>/dev/null || true
sudo dmesg | egrep -i 'bpf|verifier|tracepoint|prog' >/tmp/collector.dmesg 2>/dev/null || true
cp /tmp/collector.dmesg /tmp/collector.dmesg.copy 2>/dev/null || true
if command -v iptables >/dev/null 2>&1; then
    sudo iptables -t nat -S >/tmp/collector.nat 2>/dev/null || true
    cp /tmp/collector.nat /tmp/collector.nat.copy 2>/dev/null || true
fi
cp /tmp/mock-api.jsonl /tmp/mock-api.jsonl.copy 2>/dev/null || true
REMOTE

scp "${SCP_OPTS[@]}" \
	"${VM_USER}@127.0.0.1:/tmp/collector.metrics" \
	"${VM_USER}@127.0.0.1:/tmp/collector.ps" \
	"${VM_USER}@127.0.0.1:/tmp/collector.log.tail" \
	"${VM_USER}@127.0.0.1:/tmp/collector.stdout.tail" \
	"${VM_USER}@127.0.0.1:/tmp/collector.log.copy" \
	"${VM_USER}@127.0.0.1:/tmp/collector.stdout.copy" \
	"${VM_USER}@127.0.0.1:/tmp/collector.exitcode.copy" \
	"${VM_USER}@127.0.0.1:/tmp/collector.offset.copy" \
	"${VM_USER}@127.0.0.1:/tmp/collector.dmesg.copy" \
	"${VM_USER}@127.0.0.1:/tmp/collector.nat.copy" \
	"${VM_USER}@127.0.0.1:/tmp/mock-api.jsonl.copy" \
	"${ARTIFACT_DIR}/" >/dev/null || true

echo "artifacts: ${ARTIFACT_DIR}"
