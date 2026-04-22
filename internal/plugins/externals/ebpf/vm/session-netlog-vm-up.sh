#!/usr/bin/env bash
set -euo pipefail

VM_DIR="${EBPF_VM_DIR:-/tmp/ebpf-netlog-session-vm}"
SSH_PORT="${EBPF_VM_SSH_PORT:-2222}"
VM_NAME="${EBPF_VM_NAME:-ebpf-netlog-lab}"
BASE_IMG="${VM_DIR}/noble-server-cloudimg-amd64.img"
OVERLAY_IMG="${VM_DIR}/${VM_NAME}.qcow2"
SEED_IMG="${VM_DIR}/${VM_NAME}-seed.img"
PIDFILE="${VM_DIR}/${VM_NAME}.pid"
SERIAL_LOG="${VM_DIR}/${VM_NAME}-serial.log"
QEMU_BIN="${QEMU_BIN:-qemu-system-x86_64}"
BASE_URL="${EBPF_VM_BASE_URL:-https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img}"
SSH_KEY="${EBPF_VM_SSH_KEY:-$HOME/.ssh/id_ed25519.pub}"
VM_USER="${EBPF_VM_USER:-vircoys}"
SSH_OPTS=(-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=5 -p "${SSH_PORT}")

mkdir -p "${VM_DIR}"

if [[ ! -f "${SSH_KEY}" ]]; then
	echo "missing ssh public key: ${SSH_KEY}" >&2
	exit 1
fi

if [[ ! -f "${BASE_IMG}" ]]; then
	curl -L --fail --progress-bar "${BASE_URL}" -o "${BASE_IMG}"
fi

if [[ ! -f "${OVERLAY_IMG}" ]]; then
	qemu-img create -f qcow2 -F qcow2 -b "${BASE_IMG}" "${OVERLAY_IMG}" 80G >/dev/null
fi

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
    groups: [adm, sudo]
    shell: /bin/bash
    ssh_authorized_keys:
      - $(<"${SSH_KEY}")
ssh_pwauth: false
package_update: true
packages:
  - qemu-guest-agent
  - stress-ng
  - bpftool
  - clang
  - llvm
  - make
  - gcc
  - git
  - curl
  - jq
runcmd:
  - systemctl enable --now qemu-guest-agent
  - sysctl -w kernel.perf_event_paranoid=1
  - sysctl -w kernel.kptr_restrict=0
final_message: "ebpf netlog lab is ready after \$UPTIME seconds"
EOF

cloud-localds "${SEED_IMG}" "${VM_DIR}/user-data" "${VM_DIR}/meta-data" >/dev/null

if [[ -f "${PIDFILE}" ]]; then
	if kill -0 "$(cat "${PIDFILE}")" 2>/dev/null; then
		if ssh "${SSH_OPTS[@]}" "${VM_USER}@127.0.0.1" "echo up" >/dev/null 2>&1; then
			echo "VM already running on ssh port ${SSH_PORT}"
			exit 0
		fi
	else
		rm -f "${PIDFILE}"
	fi
fi

"${QEMU_BIN}" \
	-enable-kvm \
	-daemonize \
	-name "${VM_NAME}" \
	-pidfile "${PIDFILE}" \
	-display none \
	-serial "file:${SERIAL_LOG}" \
	-cpu host \
	-smp "${EBPF_VM_CPUS:-4}" \
	-m "${EBPF_VM_MEMORY_MB:-8192}" \
	-device virtio-rng-pci \
	-netdev "user,id=net0,hostfwd=tcp:127.0.0.1:${SSH_PORT}-:22" \
	-device virtio-net-pci,netdev=net0 \
	-drive "if=virtio,format=qcow2,file=${OVERLAY_IMG}" \
	-drive "if=virtio,format=raw,file=${SEED_IMG}"

for _ in $(seq 1 180); do
	if ssh "${SSH_OPTS[@]}" "${VM_USER}@127.0.0.1" "cloud-init status --wait >/dev/null 2>&1 || true; echo ok" >/dev/null 2>&1; then
		echo "VM ready: ssh ${VM_USER}@127.0.0.1 -p ${SSH_PORT}"
		exit 0
	fi
	sleep 2
done

echo "timed out waiting for VM to become reachable" >&2
exit 1
