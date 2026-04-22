#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="${TMPDIR:-/tmp}/flameshot-orbstack-mock-$(date +%Y%m%d-%H%M%S)"
NETWORK="flameshot-mock-$RANDOM"
APP_CONTAINER="flameshot-mock-app"
SIDECAR_CONTAINER="flameshot-mock-sidecar"
MOCK_CONTAINER="flameshot-mock-datakit"
BINARY_PATH="$WORK_DIR/flameshot-linux-arm64"
SHARED_DIR="$WORK_DIR/shared"
OUT_DIR="$WORK_DIR/out"
MOCK_ASYNC_DIR="$WORK_DIR/mock-async-profiler"
MANUAL_JCMD_DIR="$SHARED_DIR/manual-jcmd"
APP_DIR="$WORK_DIR/app"

cleanup() {
  if [[ "${KEEP_CONTAINERS:-0}" == "1" ]]; then
    return
  fi
  docker rm -f "$SIDECAR_CONTAINER" "$APP_CONTAINER" "$MOCK_CONTAINER" >/dev/null 2>&1 || true
  docker network rm "$NETWORK" >/dev/null 2>&1 || true
}

trap cleanup EXIT

mkdir -p "$SHARED_DIR" "$OUT_DIR" "$MOCK_ASYNC_DIR/bin" "$MANUAL_JCMD_DIR" "$APP_DIR"
cp "$ROOT_DIR/testdata/flameshot/orbstack/mock_async_profiler.sh" "$MOCK_ASYNC_DIR/bin/asprof"
chmod +x "$MOCK_ASYNC_DIR/bin/asprof"
cp "$ROOT_DIR/testdata/flameshot/orbstack/MockMemoryApp.java" "$APP_DIR/MockMemoryApp.java"

echo "[1/6] building linux/arm64 flameshot binary"
(
  cd "$ROOT_DIR"
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o "$BINARY_PATH" ./cmd/flameshot
)

echo "[2/6] creating docker network on OrbStack engine"
docker network create "$NETWORK" >/dev/null

echo "[3/6] starting mock DataKit receiver"
docker run -d --rm \
  --name "$MOCK_CONTAINER" \
  --network "$NETWORK" \
  -v "$OUT_DIR:/out" \
  -v "$ROOT_DIR/testdata/flameshot/orbstack/mock_datakit.py:/mock_datakit.py:ro" \
  python:3.11-slim \
  python /mock_datakit.py >/dev/null

echo "    compiling mock JVM app"
docker run --rm \
  -v "$APP_DIR:/app" \
  eclipse-temurin:17-jdk \
  javac /app/MockMemoryApp.java >/dev/null

echo "[4/6] starting Java mock app"
docker run -d --rm \
  --name "$APP_CONTAINER" \
  --network "$NETWORK" \
  --cgroupns host \
  --memory 256m \
  -v "$SHARED_DIR:/data" \
  -v "$APP_DIR:/app:ro" \
  -e ALLOC_MB=8 \
  -e ALLOC_STEPS=18 \
  -e ALLOC_SLEEP_MS=500 \
  eclipse-temurin:17-jdk \
  sh -lc 'mkdir -p /data/dumps && exec java -XX:+StartAttachListener -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=/data/dumps/app.hprof -Xms64m -Xmx160m -cp /app MockMemoryApp' >/dev/null

echo "    waiting for mock JVM to be ready"
for _ in $(seq 1 30); do
  if docker exec "$APP_CONTAINER" sh -lc 'ps -eo pid,comm,args | awk '"'"'"'"'"'"'"'"'$2 == "java" && $0 ~ /MockMemoryApp/ {found=1} END {exit found ? 0 : 1}'"'"'"'"'"'"'"'"'' >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "[5/6] starting flameshot sidecar"
docker run -d --rm \
  --name "$SIDECAR_CONTAINER" \
  --network "$NETWORK" \
  --pid "container:$APP_CONTAINER" \
  --cgroupns host \
  -v "$SHARED_DIR:/data" \
  -v "$BINARY_PATH:/work/flameshot:ro" \
  -v "$MOCK_ASYNC_DIR:/opt/async-profiler:ro" \
  -e FLAMESHOT_DATAKIT_ADDR="http://$MOCK_CONTAINER:9529/profiling/v1/input" \
  -e FLAMESHOT_PROFILING_PATH="/data" \
  -e FLAMESHOT_MONITOR_INTERVAL="1s" \
  -e FLAMESHOT_POD_MEM_LIMIT="256" \
  -e FLAMESHOT_LOG_LEVEL="debug" \
  -e FLAMESHOT_OOM_HPROF_ENABLED="true" \
  -e FLAMESHOT_OOM_HPROF_MATCH_WINDOW="2m" \
  -e FLAMESHOT_JCMD_SNAPSHOT_ENABLED="true" \
  -e FLAMESHOT_JCMD_TIMEOUT="30s" \
  -e FLAMESHOT_PROCESSES='[{"service":"mock-java","command":"MockMemoryApp","duration":"5s","emergency_duration":"5s","events":"cpu","language":"java","mem_usage_percent_emergency":40,"tags":["env:mock","pod_name:flameshot-mock"]}]' \
  eclipse-temurin:17-jdk \
  sh -lc '/work/flameshot -config /dev/null' >/dev/null

echo "[6/6] waiting for jcmd artifacts and summary logs"
for _ in $(seq 1 40); do
  if ls "$SHARED_DIR"/jcmd_*.txt >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "    running in-container jcmd control check"
docker exec "$APP_CONTAINER" sh -lc '
  pid=$(ps -eo pid,comm,args | awk '"'"'$2 == "java" && $0 ~ /MockMemoryApp/ {print $1; exit}'"'"')
  if [ -z "$pid" ]; then
    echo "java pid not found" >&2
    exit 1
  fi

  jcmd "$pid" GC.class_histogram > /data/manual-jcmd/gc_class_histogram.txt
  jcmd "$pid" Thread.print > /data/manual-jcmd/thread_print.txt
' >/dev/null

echo
echo "work dir: $WORK_DIR"
echo "shared artifacts:"
ls -1 "$SHARED_DIR" || true
echo
echo "manual jcmd artifacts:"
find "$MANUAL_JCMD_DIR" -maxdepth 1 -type f | sort || true
echo
echo "mock DataKit requests:"
find "$OUT_DIR" -maxdepth 2 -type f | sort || true
echo
echo "java app logs:"
docker logs "$APP_CONTAINER" 2>&1 | tail -n 80 || true
echo
echo "sidecar logs:"
docker logs "$SIDECAR_CONTAINER" 2>&1 | tail -n 80 || true
