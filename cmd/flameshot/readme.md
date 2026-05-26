# Flameshot User Guide

Flameshot is a profiling sidecar used to capture Java process on-site data before CPU spikes, memory pressure, or OOM events cause the Pod to disappear.

This guide focuses on what users need to configure in real deployments:

- timed profiling
- average-threshold profiling
- emergency memory-triggered profiling
- OOM `.hprof` summary recovery
- high-watermark `jcmd` snapshots

## Deployment Requirements

Flameshot works best when the business container and the Flameshot sidecar satisfy all of the following:

1. They run in the same Pod.
2. `shareProcessNamespace: true` is enabled.
3. They share a writable volume for profiling outputs.
4. The sidecar can access `async-profiler`.
5. The sidecar can access `jcmd`, or the business container already includes `jcmd`.

Recommended shared paths:

- profiling output path: `/flameshot-data`
- Java heap dump path: `/flameshot-data/dumps`

## Global Environment Variables

Common sidecar-level settings:

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `FLAMESHOT_DATAKIT_ADDR` | Yes | - | DataKit profiling upload endpoint, for example `http://datakit-service.datakit:9529/profiling/v1/input` |
| `FLAMESHOT_PROFILING_PATH` | Yes | `/data` | Shared writable directory for profiler tools, JFR output, `jcmd` output, and OOM related files |
| `FLAMESHOT_MONITOR_INTERVAL` | No | `1s` | Polling interval |
| `FLAMESHOT_LOG_LEVEL` | No | `info` | Log level |
| `FLAMESHOT_HTTP_LOCAL_IP` | Yes | - | Sidecar HTTP listen host |
| `FLAMESHOT_HTTP_LOCAL_PORT` | Yes | `8089` | Sidecar HTTP listen port |
| `FLAMESHOT_AUTO_PROFILING` | No | - | Timed profiling interval, minimum 1 minute |
| `FLAMESHOT_AUTO_PROFILING_DURATION` | No | `30s` | Timed profiling sample duration |
| `FLAMESHOT_OOM_HPROF_ENABLED` | No | `false` | Enable post-OOM `.hprof` summary recovery |
| `FLAMESHOT_OOM_HPROF_MATCH_WINDOW` | No | `2m` | Time window for matching OOM events and generated `.hprof` files |
| `FLAMESHOT_JCMD_SNAPSHOT_ENABLED` | No | `true` | Enable high-watermark `jcmd` snapshots |
| `FLAMESHOT_JCMD_TIMEOUT` | No | `10s` | Timeout for each `jcmd` command |
| `FLAMESHOT_POD_MEM_LIMIT` | No | - | Pod memory limit in Mi. When configured, Flameshot prefers Pod-limit memory percent instead of host memory percent |
| `FLAMESHOT_POD_CPU_LIMIT` | No | - | Pod CPU limit in millicores |
| `FLAMESHOT_SERVICE` | No | - | Override `service` in all `FLAMESHOT_PROCESSES` rules |
| `FLAMESHOT_TAGS` | No | - | Global tags, for example `host:node-a,pod_name:demo,pod_namespace:prod` |

## `FLAMESHOT_PROCESSES` Fields

`FLAMESHOT_PROCESSES` must be a JSON array string. Each item defines one process matching rule.

Important fields:

| Field | Description |
| --- | --- |
| `service` | Service name reported to DataKit |
| `command` | Process command-line regex |
| `language` | Currently `java` |
| `events` | Profiling events such as `cpu`, `alloc`, `lock`, `nativemem`, or `all` |
| `duration` | Normal profiling duration |
| `emergency_duration` | Short profiling duration used for emergency memory triggers |
| `cpu_usage_percent` | CPU threshold |
| `mem_usage_percent` | Average memory-percent threshold based on the latest 5 points |
| `mem_usage_mb` | Average RSS threshold in MB based on the latest 5 points |
| `mem_usage_percent_emergency` | Instant emergency memory-percent threshold. A single point hit triggers immediately |
| `mem_usage_mb_emergency` | Instant emergency RSS threshold in MB. A single point hit triggers immediately |
| `tags` | Per-rule custom tags |

Trigger behavior:

- `cpu_usage_percent`, `mem_usage_percent`, and `mem_usage_mb` keep the original "latest 5 points average" behavior.
- `mem_usage_percent_emergency` and `mem_usage_mb_emergency` use single-point immediate triggering.
- When `FLAMESHOT_POD_MEM_LIMIT` is configured, memory percentage is calculated relative to the Pod limit first.
- Emergency memory triggers use `emergency_duration`, which should be shorter than the normal `duration`.

## Recommended Kubernetes Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: java-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: java-app
  template:
    metadata:
      labels:
        app: java-app
    spec:
      shareProcessNamespace: true
      volumes:
        - name: flameshot-data
          emptyDir: {}
      containers:
        - name: app
          image: my-java-app:latest
          volumeMounts:
            - name: flameshot-data
              mountPath: /flameshot-data
        - name: flameshot
          image: pubrepo.jiagouyun.com/datakit/flameshot:latest
          env:
            - name: FLAMESHOT_DATAKIT_ADDR
              value: "http://datakit-service.datakit:9529/profiling/v1/input"
            - name: FLAMESHOT_PROFILING_PATH
              value: "/flameshot-data"
            - name: FLAMESHOT_MONITOR_INTERVAL
              value: "1s"
            - name: FLAMESHOT_AUTO_PROFILING
              value: "10m"
            - name: FLAMESHOT_AUTO_PROFILING_DURATION
              value: "15s"
            - name: FLAMESHOT_JCMD_SNAPSHOT_ENABLED
              value: "true"
            - name: FLAMESHOT_JCMD_TIMEOUT
              value: "20s"
            - name: FLAMESHOT_OOM_HPROF_ENABLED
              value: "true"
            - name: FLAMESHOT_OOM_HPROF_MATCH_WINDOW
              value: "3m"
            - name: FLAMESHOT_POD_MEM_LIMIT
              value: "2048"
            - name: FLAMESHOT_POD_CPU_LIMIT
              value: "1000"
            - name: FLAMESHOT_TAGS
              value: "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
            - name: FLAMESHOT_PROCESSES
              value: |
                [
                  {
                    "service": "java-app",
                    "language": "java",
                    "command": "^java\\b.*app\\.jar$",
                    "events": "cpu,alloc",
                    "duration": "30s",
                    "emergency_duration": "10s",
                    "cpu_usage_percent": 80,
                    "mem_usage_percent": 80,
                    "mem_usage_mb": 1536,
                    "mem_usage_percent_emergency": 92,
                    "mem_usage_mb_emergency": 1900,
                    "tags": [
                      "env:prod",
                      "version:v1"
                    ]
                  }
                ]
          securityContext:
            capabilities:
              add: ["SYS_PTRACE"]
          volumeMounts:
            - name: flameshot-data
              mountPath: /flameshot-data
```

## JVM Parameters for OOM Recovery

If you want Flameshot to recover OOM-related `.hprof` summaries, the target JVM should include:

```bash
java \
  -XX:+HeapDumpOnOutOfMemoryError \
  -XX:HeapDumpPath=/flameshot-data/dumps/app.hprof \
  -jar app.jar
```

Notes:

- `HeapDumpPath` should be inside the shared volume.
- Flameshot parses `HeapDumpPath` directly from the target Java process arguments.
- If the Pod is killed too quickly, `.hprof` may still fail to appear. In that case, high-watermark `jcmd` snapshots are usually more reliable.

## What Gets Captured During High Memory Risk

When the emergency threshold is reached, Flameshot tries to preserve the scene in this order:

1. Trigger a short profiling session using `emergency_duration`.
2. Run `jcmd <pid> GC.class_histogram`.
3. Run `jcmd <pid> Thread.print`.
4. If an `oom_kill` increment is later observed, try to locate the matching `.hprof` and upload an OOM summary log.

Raw files written into `FLAMESHOT_PROFILING_PATH` include:

- `profiler_<timestamp>.jfr`
- `jcmd_gc_class_histogram_<pid>_<timestamp>.txt`
- `jcmd_thread_print_<pid>_<timestamp>.txt`
- `.hprof` files generated by the JVM itself

## Troubleshooting

If profiling does not trigger as expected:

1. Check whether the sidecar sees the target process at all.
2. Check whether `FLAMESHOT_POD_MEM_LIMIT` matches the real Pod limit.
3. Check whether `async-profiler` exists under `/opt/async-profiler`.
4. Check whether `jcmd` is present and usable.
5. Check whether `/tmp` or required JDK paths are shared when running as a sidecar.
6. Check whether `FLAMESHOT_PROFILING_PATH` is shared and writable.

Useful checks:

```bash
ps -ef
env | grep FLAMESHOT_
ls -lah /opt/async-profiler
which jcmd
ls -lah /flameshot-data
cat /var/log/flameshot.log
```
