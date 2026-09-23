---
title     : 'Flameshot'
summary   : 'Automated profiling tool'
tags      :
  - 'java'
  - 'async-profiler'
  - 'py-spy'
  - 'python'
  - 'profiling'
  - 'flameshot'
---

Flameshot is a lightweight automated profiling tool running in Sidecar mode. It monitors the resource usage (CPU/Memory) of target processes and automatically triggers underlying Profilers (such as `async-profiler` and `py-spy`) when preset thresholds are reached, enabling non-intrusive on-site snapshot collection.

---

## Core Concepts {#core-concepts}

### Operating Mode {#running-mode}

Flameshot is deployed using the **Sidecar Container** pattern. It must run in the same Pod as the main business container (Main Container) and have **PID namespace sharing** enabled.

1. **Monitor**: Flameshot continuously polls the resource levels of target processes within the main container.
1. **Trigger**: When thresholds are met (e.g., CPU > 80%) or an HTTP API request is received, a collection task is triggered.
1. **Execute**: Based on the configured language type (currently supporting Java, Go, and Python), it invokes the corresponding profiler workflow for the target process.
1. **Collect**: The generated Profile files (e.g., `.jfr`, `.pprof`, or py-spy raw collapsed text) are subsequently uploaded to the data observability center.
1. **Timed**: After configuring `FLAMESHOT_AUTO_PROFILING`, it periodically collects profiling data for all matched processes. The sample duration defaults to 30 seconds and can be adjusted through `FLAMESHOT_AUTO_PROFILING_DURATION`.
1. **OOM Summary**: When a container `oom_kill` increment is detected, Flameshot tries to automatically parse `-XX:+HeapDumpOnOutOfMemoryError` and `-XX:HeapDumpPath=...` from the target Java process arguments. If the dump file is generated inside the shared volume, Flameshot finds the corresponding `.hprof` and uploads a summary log.

### Use Cases {#use-cases}

- **Production Safety Net**: Automatically preserve on-site evidence before a service crashes due to CPU spikes or memory leaks.
- **Performance Stress Test Analysis**: Cooperate with stress testing platforms to automatically collect performance hotspots under high load.

---

## Configuration {#configuration}

All Flameshot behaviors are controlled via environment variables. Configuration is divided into **Global Settings** and **Profiling Policies**.

### Global Environment Variables {#global-env}

These variables control the basic behavior of the Sidecar container.

| Variable Name                | Required | Default Value | Description                                                                                                                                                                               |
|:-----------------------------|:---------|:--------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `FLAMESHOT_DATAKIT_ADDR`     | **Yes**  | -             | DataKit's Profiling data receiving interface address.                                                                                                                                     |
| `FLAMESHOT_PROFILING_PATH`   | **Yes**  | `/data`       | **Shared directory path**. Used to store tools and generated temporary files; must match the mount path in the main container. Python relative `pyspy_output_path` values are written under this directory. |
| `FLAMESHOT_MONITOR_INTERVAL` | No       | `1`           | Monitoring polling interval (seconds).                                                                                                                                                    |
| `FLAMESHOT_LOG_LEVEL`        | No       | `info`        | Log level. Options: `debug`, `info`, `warn`, `error`.                                                                                                                                     |
| `FLAMESHOT_PROFILING_ENABLED` | No | `true` | Enable JFR Profiling. Set to `false` to disable timed, threshold, cgroup high-watermark, and HTTP manual Profiling while keeping OOM detection, hprof upload, and proactive Heap Dump. |
| `FLAMESHOT_AUTO_PROFILING`   | No       | -             | Collect Profiling data at regular intervals for all matched processes. The minimum interval must not be less than one minute, such as five minutes: "5m" or one hour: "1h" |
| `FLAMESHOT_AUTO_PROFILING_DURATION` | No | `30s` | Sample duration for timed profiling mode. |
| `FLAMESHOT_OOM_HPROF_ENABLED` | No      | `false`       | Enable the post-OOM `.hprof` summary recovery path. Java only. The target JVM must explicitly enable `-XX:+HeapDumpOnOutOfMemoryError` and configure `-XX:HeapDumpPath=...` inside a shared volume. It is recommended to set this explicitly in deployment manifests. |
| `FLAMESHOT_OOM_HPROF_MATCH_WINDOW` | No | `2m` | Time window used to match an OOM event with the `.hprof` file modification time. |
| `FLAMESHOT_HPROF_UPLOAD_ENABLED` | No | `false` | Upload matched or generated `.hprof` files to object storage. |
| `FLAMESHOT_HPROF_UPLOAD_PROVIDER` | No | - | Object storage provider: `oss` or `s3`. |
| `FLAMESHOT_HPROF_UPLOAD_AUTH_TYPE` | No | `static` | hprof upload authentication type. `static` directly uses AK/SK with an optional STS SecurityToken. `assume_role` uses source AK/SK to call Alibaba Cloud STS AssumeRole and obtain refreshable temporary credentials. `assume_role` is supported for OSS only. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ENDPOINT` | No | - | OSS/S3 endpoint. |
| `FLAMESHOT_HPROF_UPLOAD_REGION` | No | `us-east-1` for S3 | OSS/S3 region. In AssumeRole mode, this is also used to derive the STS region when no STS endpoint is configured. |
| `FLAMESHOT_HPROF_UPLOAD_BUCKET` | No | - | Target bucket. |
| `FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_ID` | No | - | Object storage access key ID. |
| `FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_SECRET` | No | - | Object storage access key secret. |
| `FLAMESHOT_HPROF_UPLOAD_SECURITY_TOKEN` | No | - | SecurityToken for Alibaba Cloud OSS STS credentials. When set with temporary AK/SK, STS authentication is used. Restart the Pod with renewed credentials before they expire. :octicons-tag-24: Version-0.2.3 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_ARN` | Required for `assume_role` | - | Target RAM Role ARN. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SESSION_NAME` | No | SDK default | AssumeRole session name. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_DURATION_SECONDS` | No | `3600` | Validity period of the STS credentials returned by AssumeRole, in seconds. The minimum value is 900. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_POLICY` | No | - | Optional inline policy used to further restrict the returned STS credentials. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_EXTERNAL_ID` | No | - | Optional ExternalId for cross-account or confused deputy prevention scenarios. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_STS_ENDPOINT` | No | SDK default | Custom STS endpoint, for example `sts.cn-hangzhou.aliyuncs.com`. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SOURCE_ACCESS_KEY_ID` | Required for `assume_role` | - | Source identity AK used to call STS AssumeRole. Grant only the minimum required `sts:AssumeRole` permission. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SOURCE_ACCESS_KEY_SECRET` | Required for `assume_role` | - | Source identity SK used to call STS AssumeRole. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SOURCE_SECURITY_TOKEN` | No | - | Source identity SecurityToken when the source identity itself uses temporary credentials. :octicons-tag-24: Version-0.2.4 |
| `FLAMESHOT_HPROF_UPLOAD_PATH_TEMPLATE` | No | `{service}/{pod_name}/{timestamp}/{filename}` | Object key template. Supports service / pod_name / pod_namespace / host / pid / timestamp / filename variables. |
| `FLAMESHOT_HPROF_DOWNLOAD_URL_TEMPLATE` | No | - | Optional download URL template. When configured, events include `hprof_download_url` rendered from it. |
| `FLAMESHOT_HEAP_DUMP_ENABLED` | No | `false` | Enable proactive Java Heap Dump when an emergency memory threshold is reached. |
| `FLAMESHOT_HEAP_DUMP_PATH_TEMPLATE` | No | `{profiling_path}/dumps/{service}_{pod_name}_{pid}_{timestamp}.hprof` | Local Heap Dump output path template. |
| `FLAMESHOT_HEAP_DUMP_JMAP_PATH` | No | `jmap` | `jmap` executable path. Official Sidecar images do not include a JVM/JDK by default; provide an available `jmap` explicitly before enabling proactive Heap Dump. |
| `FLAMESHOT_HEAP_DUMP_TIMEOUT` | No | `120s` | Heap Dump command timeout. |
| `FLAMESHOT_HEAP_DUMP_COOLDOWN` | No | `10m` | Per-process proactive Heap Dump cooldown. |
| `FLAMESHOT_POD_MEM_LIMIT` | No | - | Pod memory limit in Mi. When configured, Flameshot prefers Pod-limit memory percentage over host memory percentage. |
| `FLAMESHOT_POD_CPU_LIMIT` | No | - | Pod CPU limit in millicores. |
| `FLAMESHOT_HTTP_LOCAL_IP`    | **Yes**  | `-`           | The Sidecar's own HTTP service listening host.                                                                                                                                            |
| `FLAMESHOT_HTTP_LOCAL_PORT`  | **Yes**  | `8089`        | The Sidecar's own HTTP service listening port.                                                                                                                                            |
| `FLAMESHOT_SERVICE`          | No       | -             | Will replace the 'service' configuration in 'FLAMESHOT_PROCESSES'                                                                                                                         |
| `FLAMESHOT_TAGS`             | No       | -             | Suggest configuring `host` `pod_name` `pod_namespace`, such as: "host: host_name,pod_name:pod_a"                                                                                          |

### AssumeRole temporary authentication {#hprof-upload-assume-role}

Use `assume_role` mode when the customer already has an Alibaba Cloud STS AssumeRole flow and wants Flameshot to obtain temporary credentials at runtime:

```shell
FLAMESHOT_HPROF_UPLOAD_PROVIDER=oss
FLAMESHOT_HPROF_UPLOAD_AUTH_TYPE=assume_role
FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_ARN=acs:ram::<account-id>:role/<role-name>
FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SOURCE_ACCESS_KEY_ID=<source-access-key-id>
FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SOURCE_ACCESS_KEY_SECRET=<source-access-key-secret>
```

In this mode, `FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_ID`, `FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_SECRET`, and `FLAMESHOT_HPROF_UPLOAD_SECURITY_TOKEN` are no longer used as OSS upload credentials. Flameshot uses the source AK/SK to call Alibaba Cloud STS `AssumeRole`, obtains temporary AK/SK/SecurityToken, uploads to OSS, and lets the SDK refresh the credentials before expiration.

Note that `AssumeRole` is not an anonymous API. The source AK/SK is still the source identity credential used to call STS. Grant it only the minimum required `sts:AssumeRole` permission, and grant the target RAM Role only the minimum write permission for the target OSS bucket/prefix. If the STS call fails, the configuration is incomplete, or the returned credentials are incomplete, the upload fails and does not fall back to the default credential chain, node role, or anonymous upload.

If DataKit is deployed as a DaemonSet and exposes port `9529` through `hostNetwork`/`hostPort`, it is recommended that Flameshot connects directly to the DataKit on the **same node as the application Pod** instead of using a normal Service domain that may route requests to another node:

```yaml
- name: NODE_IP
  valueFrom:
    fieldRef:
      fieldPath: status.hostIP
- name: FLAMESHOT_DATAKIT_ADDR
  value: "http://$(NODE_IP):9529/profiling/v1/input"
```

This keeps each application Pod uploading Profile data to the DataKit instance on its own node, which makes troubleshooting clearer and preserves the node-local collection model.

### Profiling Policy Configuration {#profiling-policy}

Target monitoring rules are defined via the `FLAMESHOT_PROCESSES` environment variable. The value must be a standard **JSON Array** string.

To maintain readability in Kubernetes YAML, it is **strongly recommended** to use YAML's block scalar syntax (`|`) for writing the JSON configuration, as shown below:

```yaml
    env:
      # ... other environment variables ...
      - name: FLAMESHOT_PROCESSES
        value: |
          [
            {
              "service": "user-service",
              "language": "java",
              "command": "^java.*user-service\\.jar$",
              "duration": "60s",
              "events": "cpu,alloc",
              "cpu_usage_percent": 80,
              "mem_usage_percent": 80,
              "mem_usage_mb": 1024,
              "mem_usage_percent_emergency": 92,
              "mem_usage_mb_emergency": 1536,
              "heap_dump_on_memory_emergency": true,
              "emergency_duration": "10s",
              "tags": [
                "env:prod",
                "version:v1.2"
              ]
            }
          ]
```

**Common Field Descriptions:**

- **`service`** (String): Service name reported to the observability center.
- **`language`** (String): Target process language. Currently supports `java`, `go`, `golang`, and `python`.
- **`command`** (String): Regular expression to match the process command line.
- **`duration`** (String): Duration of a single collection (e.g., `30s`, `1m`). **Note**: To avoid execution timeouts, it is recommended not to exceed 5 minutes.
- **`emergency_duration`** (String): Shorter profiling duration used after an emergency memory hit. `10s` or `15s` is recommended.
- **`pprof_url`** (String): Go pprof HTTP base URL, for example `http://127.0.0.1:6060`. Required when `language` is `go` or `golang`.
- **`pprof_types`** (List): Go pprof types. Supported values are `cpu`, `goroutine`, `heap`, `mutex`, and `block`, matching the existing Go pull mode in the profile input.
- **`pprof_timeout`** (String): Go pprof request timeout. It should be greater than the CPU profile `duration`.
- **`pyspy_path`** (String): Python `py-spy` executable path. Defaults to `py-spy`.
- **`pyspy_output_path`** (String): Python raw output path. When omitted, Flameshot generates a local temporary file; relative paths are written under `FLAMESHOT_PROFILING_PATH` or the default output directory.
- **`pyspy_rate`** (Int): Python sample rate. Defaults to `100`.
- **`pyspy_subprocesses`** (Bool): Whether to sample subprocesses. Defaults to `false`.
- **`pyspy_idle`** (Bool): Whether to include idle threads. Defaults to `false`.
- **`tags`** (List): List of custom tags; recommended to include meta-information like `env`, `version`.
- **`cpu_usage_percent`** (Int): CPU trigger threshold (0-N). Values may exceed 100 in multi-core environments.
- **`mem_usage_percent`** (Int): Average memory-percentage threshold (0-100), evaluated by the latest 5 points.
- **`mem_usage_mb`** (Int): Average RSS threshold in MB, evaluated by the latest 5 points.
- **`mem_usage_percent_emergency`** (Int): Instant emergency memory-percentage threshold (0-100). A single hit triggers immediately.
- **`mem_usage_mb_emergency`** (Int): Instant emergency RSS threshold in MB. A single hit triggers immediately.
- **`heap_dump_on_memory_emergency`** (Bool): Whether this process rule may trigger proactive Heap Dump on emergency memory hits. When omitted, it defaults to enabled if `FLAMESHOT_HEAP_DUMP_ENABLED=true`.
- **`cpu_usage_percent`**, **`mem_usage_percent`**, and **`mem_usage_mb`** skip threshold checks when omitted or set to 0.
- When `FLAMESHOT_POD_MEM_LIMIT` is configured, `mem_usage_percent` and `mem_usage_percent_emergency` prefer Pod-limit memory percentage instead of host memory percentage.
- When `FLAMESHOT_HEAP_DUMP_ENABLED=true`, emergency memory hits enqueue a `jmap` Heap Dump task. This requires the Sidecar to execute the `jmap` referenced by `FLAMESHOT_HEAP_DUMP_JMAP_PATH`. If hprof object storage upload is configured, the generated `.hprof` is uploaded and the event includes `hprof_object_key`, `hprof_download_url`, and upload status.

---

## Language Specifics {#language-specifics}

Flameshot invokes different underlying tools depending on the technology stack of the monitored application.

<!-- markdownlint-disable MD046 -->
=== "Java"

    ### Java Profiling {#java-profiling}
    
    For Java applications, Flameshot includes **async-profiler** (supporting `linux-amd64` / `linux-arm64`).
    
    **Key Configuration Fields (`FLAMESHOT_PROCESSES`):**
    
    - **`language`**: Must be set to `java`.
    - **`events`**: Supports `cpu` (CPU cycles), `alloc` (memory allocation), `lock` (lock contention), `cache-misses`, `nativemem`. Defaults to `all`.
    - **`jdk_version`**: (Optional) JDK version used for metadata display.
    
    **Notes:**

    - No reliance on JVM Safepoint; extremely low overhead.
    - If you want Flameshot to automatically discover and upload a post-OOM `.hprof` summary, the JVM must explicitly enable `-XX:+HeapDumpOnOutOfMemoryError` and configure `-XX:HeapDumpPath=...`. Setting `FLAMESHOT_OOM_HPROF_ENABLED=true` alone does not modify the target JVM startup options.
    - If `FLAMESHOT_HEAP_DUMP_ENABLED=true` is enabled, Flameshot executes `jmap -dump:format=b,file=<path> <pid>` when an emergency memory threshold is reached. Official Sidecar images do not include a JVM/JDK by default; use a custom image, mounted tool, or another explicit mechanism to provide a `jmap` compatible with the target JVM, and point `FLAMESHOT_HEAP_DUMP_JMAP_PATH` to it.
    - `HeapDumpPath` must point to a directory or file path inside a volume shared by both the application container and the Flameshot Sidecar. A stable and process-distinguishable dump path is recommended.
    - Declare the `.hprof` summary recovery feature flags explicitly in deployment manifests instead of relying on implicit defaults.

=== "Go"

    ### Go Profiling {#go-profiling}

    For Go applications, Flameshot pulls `.pprof` data from the `net/http/pprof` HTTP endpoint exposed by the business process and uploads it to DataKit.

    **Key Configuration Fields (`FLAMESHOT_PROCESSES`):**

    - **`language`**: Must be set to `go` or `golang`.
    - **`pprof_url`**: Business process pprof HTTP base URL, for example `http://127.0.0.1:6060`.
    - **`pprof_types`**: Supports `cpu`, `goroutine`, `heap`, `mutex`, and `block`.
    - **`duration`**: CPU profile duration, mapped to `/debug/pprof/profile?seconds=<duration>`.
    - **`pprof_timeout`**: pprof request timeout. It should be greater than `duration`.

    **Go application requirement:**

    ```go
    import (
        "net/http"
        _ "net/http/pprof"
    )

    func main() {
        go http.ListenAndServe("127.0.0.1:6060", nil)
    }
    ```

    `pprof_url` must be an address that the Flameshot Sidecar can actually reach. Common cases:

    - If pprof listens on `127.0.0.1:6060` or `0.0.0.0:6060`, configure `http://127.0.0.1:6060`.
    - If pprof only listens on the Pod IP, for example `10.x.x.x:6060`, inject the Pod IP through the Downward API and configure `http://$(POD_IP):6060`.

    ```yaml
    - name: POD_IP
      valueFrom:
        fieldRef:
          fieldPath: status.podIP
    - name: FLAMESHOT_PROCESSES
      value: |
        [
          {
            "service": "go-app",
            "language": "go",
            "command": "^/app/go-app$",
            "pprof_url": "http://$(POD_IP):6060",
            "pprof_types": ["cpu", "goroutine", "heap", "mutex", "block"],
            "duration": "30s",
            "pprof_timeout": "45s"
          }
        ]
    ```

    Configuration example:

    ```json
    {
      "service": "go-app",
      "language": "go",
      "command": "^/app/go-app",
      "pprof_url": "http://127.0.0.1:6060",
      "pprof_types": ["cpu", "goroutine", "heap", "mutex", "block"],
      "duration": "30s",
      "pprof_timeout": "45s",
      "tags": ["env:prod", "version:v1"]
    }
    ```

    **Notes:**

    - `heap`, `mutex`, and `block` are uploaded as delta profiles. The first sample is stored as the baseline and is not uploaded for those delta types.
    - `mutex` and `block` do not collect useful data by default. The application must explicitly enable `runtime.SetMutexProfileFraction` and `runtime.SetBlockProfileRate`.
    - The pprof endpoint can expose sensitive runtime details. Listen only on a Pod-local address and do not expose it through a Service or public network.
    - If `netstat -anp | grep 6060` only shows something like `10.x.x.x:60602 ... ESTABLISHED`, it is not the pprof listening port. A working pprof server should show `:6060 ... LISTEN`.

=== "Python"

    ### Python Profiling {#python-profiling}

    For Python applications, Flameshot uses the official `py-spy` to attach to the target process and runs `py-spy record --format raw` to generate collapsed text for upload to DataKit. The uploaded event uses `family=python`, `format=collapse`, `profiler=pyspy`, and the attachment name is fixed to `prof`.

    **Current data type limitations:**

    - The current py-spy integration only produces and displays **CPU Samples**. This value is the number of times a Python stack was sampled, not an exact CPU utilization or CPU time measurement.
    - A raw collapsed file contains only stacks and sample counts, so the integration currently cannot provide other Profiling data types such as CPU Time, Wall Time, Heap, Allocation, Lock, Exception, or I/O.
    - `pyspy_rate` only changes the sampling frequency, and `pyspy_subprocesses` only expands the set of sampled processes. Neither option adds data types.
    - `pyspy_idle=true` makes py-spy include idle threads, but the result is still displayed as CPU Samples and does not add a Wall Time data type.
    - `cpu_usage_percent`, `mem_usage_percent`, and `mem_usage_mb` only control when collection is triggered. A memory threshold does not make py-spy produce a Heap or Allocation Profile.

    **Key Configuration Fields (`FLAMESHOT_PROCESSES`):**

    - **`language`**: Must be set to `python`.
    - **`pyspy_path`**: `py-spy` executable path. Defaults to `py-spy`.
    - **`pyspy_output_path`**: Raw output path. When omitted, Flameshot generates a local temporary file; relative paths are written under `FLAMESHOT_PROFILING_PATH` or the default output directory.
    - **`pyspy_rate`**: Samples per second. Defaults to `100`.
    - **`pyspy_subprocesses`**: Whether to sample subprocesses.
    - **`pyspy_idle`**: Whether to include idle threads.

    Configuration example:

    ```json
    {
      "service": "python-api",
      "language": "python",
      "command": "^python\\b.*app\\.py$",
      "duration": "30s",
      "pyspy_rate": 100,
      "cpu_usage_percent": 80,
      "tags": ["env:prod", "version:v1"]
    }
    ```

    **Notes:**

    - Thresholds, timed collection, and HTTP manual triggers only control when a `py-spy` run starts; sampling duration after attach is controlled by `duration`.
    - The Sidecar must contain `py-spy` and have permission to attach to the target Python process, typically `SYS_PTRACE`.
    - The DataKit profile input forwards py-spy collapsed data and does not parse it through the Python pprof metric extractor.
<!-- markdownlint-enable MD046 -->

---

## Deployment {#deployment}

### Kubernetes Sidecar Deployment {#k8s-sidecar}

For Flameshot to work correctly, the Pod configuration must meet the following three conditions:

1. **Shared Process Namespace** (`shareProcessNamespace: true`).
1. **Shared Storage Volume** (EmptyDir).
1. **System Capabilities** (Capabilities).

**YAML Example:**

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: java-app-profiled
spec:
  # 1. [Core] Enable PID sharing so Sidecar can see the Java process
  shareProcessNamespace: true
  
  volumes:
  - name: shared-data
    emptyDir: {}

  containers:
  # Business Container
  - name: my-app
    image: my-app:latest
    volumeMounts:
    - name: shared-data
      mountPath: /data # Must match Sidecar configuration

  # Flameshot Sidecar
  - name: flameshot
    image: pubrepo.jiagouyun.com/datakit/flameshot:latest
    env:
      - name: FLAMESHOT_PROFILING_PATH
        value: "/data"
      # ... other environment variables ...
    
    # 2. [Core] Grant ptrace capability
    securityContext:
      capabilities:
        add: ["SYS_PTRACE"]
    
    # 3. [Core] Mount the same directory
    volumeMounts:
    - name: shared-data
      mountPath: /data
```

### DataKit DaemonSet Notes {#datakit-daemonset}

When DataKit is deployed as a DaemonSet, Flameshot should preferably upload Profile data through the current node IP:

```yaml
- name: NODE_IP
  valueFrom:
    fieldRef:
      fieldPath: status.hostIP
- name: FLAMESHOT_DATAKIT_ADDR
  value: "http://$(NODE_IP):9529/profiling/v1/input"
```

Also make sure DataKit satisfies all of the following:

1. **The profile input is enabled** and registers the Profiling upload endpoint:

    ```toml
    [[inputs.profile]]
      endpoints = ["/profiling/v1/input"]
    ```

1. **The Profiling API is allowed for non-localhost access**. If DataKit has HTTP API allow-listing enabled, add `/profiling/v1/input` to `ENV_HTTP_PUBLIC_APIS`:

    ```yaml
    - name: ENV_HTTP_PUBLIC_APIS
      value: /otel/v1/traces,/otel/v1/metrics,/otel/v1/logs,/profiling/v1/input
    ```

    If this variable already contains other APIs, append `/profiling/v1/input` to the existing list instead of replacing the whole value. Otherwise, DataKit may return:

    ```text
    datakit.publicAccessDisabled: api /profiling/v1/input disabled from external IP, only loopback(localhost) allowed
    ```

1. **The DataKit listen address is reachable through the node IP**. A common DaemonSet configuration is `hostNetwork: true`, `hostPort: 9529`, and `ENV_HTTP_LISTEN=0.0.0.0:9529`.

### OOM HProf Summary Requirements {#oom-hprof}

If you want Flameshot to automatically recover `.hprof` summary information after a Java OOM, all of the following conditions must be met:

1. Enable `-XX:+HeapDumpOnOutOfMemoryError` in the application JVM arguments.
1. Configure `-XX:HeapDumpPath=/data/...` in the JVM arguments, and make sure the path is inside the shared volume.
1. Enable `FLAMESHOT_OOM_HPROF_ENABLED=true` for the Flameshot Sidecar.
1. It is recommended to set `FLAMESHOT_OOM_HPROF_MATCH_WINDOW` explicitly so the matching window is operationally unambiguous.

For example:

```bash
java \
  -XX:+HeapDumpOnOutOfMemoryError \
  -XX:HeapDumpPath=/data/dumps/app.hprof \
  -jar app.jar
```

Notes:

- Flameshot now discovers `HeapDumpPath` directly from the target Java process arguments. There is no separate configuration item for the `.hprof` path.
- `FLAMESHOT_OOM_HPROF_ENABLED` only enables the Flameshot-side recovery workflow; it does not inject HeapDump-related flags into the target JVM.
- If the target process does not enable `HeapDumpOnOutOfMemoryError`, or if `HeapDumpPath` is not inside a shared volume, Flameshot can only record the OOM event and cannot locate the `.hprof` file.
- If the container is terminated before the dump is fully written, `.hprof` may still be unavailable.

### Docker Local Testing {#docker-testing}

If you need to test in a local Docker environment, use the following command to start Flameshot and monitor the target container.

**Prerequisites:**

- Use `--pid="container:<target_id>"` or shared volumes (depending on the specific Docker version).

**Test Image:** `pubrepo.jiagouyun.com/datakit/flameshot:1.85.1-testing_testing-iss-2876`

**Startup Command Example:**

```bash
docker run -d \
  --name flameshot-debug \
  --volumes-from <YOUR_JAVA_APP_CONTAINER> \
  -e FLAMESHOT_DATAKIT_ADDR="http://datakit:9529/profiling/v1/input" \
  -e FLAMESHOT_PROCESSES='[{"service":"local-test","command":"java","language":"java","cpu_usage_percent":10}]' \
  pubrepo.jiagouyun.com/datakit/flameshot:1.85.1-testing_testing-iss-2876
```

---

## API Reference {#api-reference}

Flameshot provides an HTTP interface allowing users or automated O&M scripts to **manually trigger** collection tasks.

### Manual Triggering {#manual-trigger}

**Interface Address**: `GET /v1/profile`

> **Semantic Explanation**: This interface is used to generate a Profile dataset on demand, not to retrieve monitoring metrics.

**Request Parameters:**

| Parameter | Required | Description | Example |
| :--- | :--- | :--- | :--- |
| `pid` | **One of two** | Target Process ID. Takes precedence over `command`. | `1234` |
| `command` | **One of two** | Target process name regex. Used to match the target process. | `^java.*app.jar$` |
| `duration` | No | Collection duration. Defaults to `30s`. | `30s` |
| `events` | No | Java collection event types. Defaults to `all`. Go collection prefers `pprof_types`. | `cpu,alloc` |

**Usage Examples:**

1. **Trigger collection by PID**:

    ```bash
    # Collect CPU and memory allocation data for the process with PID 1234 for 30 seconds
    curl "http://localhost:8089/v1/profile?pid=1234&duration=30s&events=cpu,alloc"
    ```

1. **Trigger collection by process name regex**:

    ```bash
    # Collect data for the process matching tmall.jar for the default duration
    curl "http://localhost:8089/v1/profile?command=^java\\b.*tmall\\.jar$"
    ```

---

## JFR format {#jfr-format}

**async-profiler** events notes:

| Event Type     | Command Flag | Mechanism                                                                                      | Best Use Case                                                                                                 | Key Note                                                                                     |
|:---------------|:-------------|:-----------------------------------------------------------------------------------------------|:--------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------|
| CPU Time       | cpu          | Uses kernel sampling or itimer to see which code is currently on the CPU.                      | "Performance Tuning: Finding ""hotspots"" in calculation-heavy logic or algorithms."                          | Only tracks time when the thread is actively running on a CPU.                               |
| Wall-clock     | wall         | "Samples all threads at fixed intervals regardless of their state (running,sleeping,blocked)." | "Latency Diagnosis: Finding delays in I/O, database calls or external network requests."                      | "Shows what threads are doing while they are ""waiting."""                                   |
| Allocation     | alloc        | Samples `TLAB` (Thread Local Allocation Buffer) refills and large object allocations.            | Memory Optimization: Reducing GC pressure by finding code that creates excessive temporary objects.           | "Measures the rate of allocation ,not the current heap usage/liveness."                      |
| Lock           | lock         | Tracks contention and wait time on intrinsic JVM monitors (synchronized).                      | Concurrency Bottlenecks: Identifying lock contention or threads blocked by synchronization.                   | Usually filtered to record only events exceeding a certain duration threshold.               |
| Cache Misses   | cache-misses | Utilizes Hardware Performance Counters (PMU) to track L1/L2/L3 cache misses.                   | "Low-level Tuning: Optimizing data structures for CPU cache friendliness (e.g., avoiding false sharing)."     | Requires Linux perf_events support and specific hardware access.                             |
| Context Switch | cs           | Tracks how often the OS scheduler swaps threads in and out of the CPU.                         | Resource Scaling: Identifying if you have too many active threads for your CPU core count.                    | "High context switching leads to ""wasted"" CPU cycles spent on management."                 |
| Java Methods   | itimer       | A timer-based sampling approach provided by the OS kernel.                                     | "Compatibility Mode: Used when perf_events is unavailable (e.g. in some restricted Docker/K8s environments)." | "Good fallback for CPU profiling,though slightly less precise than hardware-based sampling." |


---

## Troubleshooting {#troubleshooting}

1. **Cannot collect data?**

    - Check if `shareProcessNamespace: true` is enabled in the Pod.
    - Check if the Sidecar has `SYS_PTRACE` capability.
    - For Go applications, check if `pprof_url` is reachable from the Flameshot Sidecar.
    - For Go applications, first verify the pprof endpoint from inside the Flameshot container:

        ```bash
        curl -v "http://<pprof-host>:6060/debug/pprof/"
        curl -v "http://<pprof-host>:6060/debug/pprof/goroutine?debug=0"
        ```

    - If the log contains `connect: connection refused`, the target IP:Port is not listening. Check inside the Pod:

        ```bash
        netstat -lntp | grep ':6060'
        ss -ltnp | grep ':6060'
        ```

      If there is no `LISTEN`, the application has not enabled pprof, or it is listening on a different port. Seeing only `60602 ... ESTABLISHED` does not mean port 6060 is listening.
    - If pprof listens on the Pod IP, configure `pprof_url` as `http://$(POD_IP):6060`; if it listens on loopback or all addresses, `http://127.0.0.1:6060` can be used.

1. **File not uploaded?**

    - Check if `FLAMESHOT_PROFILING_PATH` is correctly mounted between the two containers.
    - The system automatically manages file life cycles and will attempt to delete temporary files after collection is complete.
    - Go profiling uploads `.pprof` files as in-memory attachments and usually does not depend on local files. If the Go profiling log already shows `upload to DataKit err`, check the HTTP status code and response body returned by DataKit first.
    - If DataKit returns `403` with `datakit.publicAccessDisabled`, `/profiling/v1/input` has not been allowed for non-localhost access. Add it to DataKit configuration:

        ```yaml
        - name: ENV_HTTP_PUBLIC_APIS
          value: /otel/v1/traces,/otel/v1/metrics,/otel/v1/logs,/profiling/v1/input
        ```

    - If DataKit returns `input "profile" is not enabled for API "/profiling/v1/input"`, the profile input is not enabled. Enable it with:

        ```toml
        [[inputs.profile]]
          endpoints = ["/profiling/v1/input"]
        ```

## Changelog {#changelog}

### 0.2.4 (2026-7-22) {#cl-0.2.4}

#### New Features {#cl-0.2.4-new}

- **Add config**
    - hprof upload to Alibaba Cloud OSS now supports the `assume_role` authentication mode. Flameshot can actively call Alibaba Cloud STS AssumeRole to obtain and refresh temporary credentials before uploading to OSS. (#3152)

### 0.2.3 (2026-7-20) {#cl-0.2.3}

#### New Features {#cl-0.2.3-new}

- **Add config**
    - hprof upload to Alibaba Cloud OSS now supports the SecurityToken from STS temporary credentials, allowing temporary AK/SK to be used without configuring long-term static AK/SK in Flameshot. (#3152)

### 0.2.2 (2026-5-12) {#cl-0.2.2}

#### Bug Fixes {#cl-0.2.2-fix}

- **Fix**
    - Fixed duplicated `host`, `env`, `version`, `service`, and other tags in uploaded Profiling `tags_profiler` metadata.
- **Change**
    - Removed high-watermark `jcmd` snapshot support and related configuration items.
    - Added cgroup memory-pressure diagnostic fields to make threshold, current cgroup memory, and cgroup limit easier to verify.

### 0.2.1 (2026-2-11) {#cl-0.2.1}

#### Optimize {#cl-0.2.1-new}

- **optimize**
    - In a container environment, use the configured resource size as the base value for threshold calculation.

### 0.2.0 (2026-2-4) {#cl-0.2.0}

#### New Features {#cl-0.2.0-new}

- **Add config**
    - Support configuring scheduled Profiling execution via the environment variable `FLAMESHOT_AUTO_PROFILING`
- **optimize**
    - Optimize the threshold processing logic for configuration.

### 0.1.0 (2025-12-17) {#cl-0.1.0}

The first official release of Flameshot, focusing on providing automated profiling capabilities for Java applications in containerized environments.

#### New Features {#cl-0.1.0-new}

- **Core Architecture**:
    - Support for Kubernetes **Sidecar Mode** deployment, utilizing shared PID namespaces for non-intrusive monitoring.
    - Support for **Linux AMD64** and **ARM64** multi-architecture execution.
- **Language Support**:
    - **Java**: Deep integration with `async-profiler`, supporting various event collections like CPU, Alloc, Lock, etc.
    - Automatic detection and adaptation to the target container's JDK environment.
- **Trigger Mechanism**:
    - **Threshold Trigger**: Support for automatic triggering based on CPU usage (`cpu_usage_percent`) and memory usage/amount (`mem_usage_percent`/`mem_usage_mb`).
    - **API Trigger**: Provided HTTP interface `GET /v1/monitor` (Note: should be `/v1/profile` as per API section), supporting manual trigger by PID or regex process name matching.
- **Data Integration**:
    - Support for automatically reporting generated `.jfr` or flame graph data to **DataKit**.
    - Support for flexible multi-process monitoring policies and tags (`tags`) via the `FLAMESHOT_PROCESSES` environment variable.
