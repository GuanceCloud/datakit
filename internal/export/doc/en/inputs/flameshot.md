---
title     : 'Flameshot'
summary   : 'Automated profiling tool'
tags      :
  - 'java'
  - 'async-profiler'
  - 'profiling'
  - 'flameshot'
---

Flameshot is a lightweight automated profiling tool running in Sidecar mode. It monitors the resource usage (CPU/Memory) of target processes and automatically triggers underlying Profilers (such as `async-profiler`) when preset thresholds are reached, enabling non-intrusive on-site snapshot collection.

---

## Core Concepts {#core-concepts}

### Operating Mode {#running-mode}

Flameshot is deployed using the **Sidecar Container** pattern. It must run in the same Pod as the main business container (Main Container) and have **PID namespace sharing** enabled.

1. **Monitor**: Flameshot continuously polls the resource levels of target processes within the main container.
1. **Trigger**: When thresholds are met (e.g., CPU > 80%) or an HTTP API request is received, a collection task is triggered.
1. **Execute**: Based on the configured language type (currently supporting Java), it invokes the corresponding Profiler tool to attach to the target process.
1. **Collect**: The generated Profile files (e.g., `.jfr`) are stored in a shared volume and subsequently uploaded to the data observability center.
1. **Timed**: After configuring `FLAMESHOT_AUTO_PROFILING`, it will periodically collect a 30-second Profiling data for all matched processes.

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
| `FLAMESHOT_PROFILING_PATH`   | **Yes**  | `/data`       | **Shared directory path**. Used to store tools and generated temporary files; must match the mount path in the main container.                                                            |
| `FLAMESHOT_MONITOR_INTERVAL` | No       | `1`           | Monitoring polling interval (seconds).                                                                                                                                                    |
| `FLAMESHOT_LOG_LEVEL`        | No       | `info`        | Log level. Options: `debug`, `info`, `warn`, `error`.                                                                                                                                     |
| `FLAMESHOT_AUTO_PROFILING`   | No       | -             | Collect Profiling data for 30 seconds at regular intervals for all matched processes. The minimum interval must not be less than one minute, such as five minutes: "5m" or one hour: "1h" |
| `FLAMESHOT_HTTP_LOCAL_IP`    | **Yes**  | `-`           | The Sidecar's own HTTP service listening host.                                                                                                                                            |
| `FLAMESHOT_HTTP_LOCAL_PORT`  | **Yes**  | `8089`        | The Sidecar's own HTTP service listening port.                                                                                                                                            |
| `FLAMESHOT_SERVICE`          | No       | -             | Will replace the 'service' configuration in 'FLAMESHOT_PROCESSES'                                                                                                                         |
| `FLAMESHOT_TAGS`             | No       | -             | Suggest configuring `host` `pod_name` `pod_namespace`, such as: "host: host_name,pod_name:pod_a"                                                                                          |


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
              "tags": [
                "env:prod",
                "version:v1.2"
              ]
            }
          ]
```

**Common Field Descriptions:**

- **`service`** (String): Service name reported to the observability center.
- **`language`** (String): Target process language. Currently supports `java`.
- **`command`** (String): Regular expression to match the process command line.
- **`duration`** (String): Duration of a single collection (e.g., `30s`, `1m`). **Note**: To avoid execution timeouts, it is recommended not to exceed 5 minutes.
- **`tags`** (List): List of custom tags; recommended to include meta-information like `env`, `version`.
- **`cpu_usage_percent`** (Int): CPU trigger threshold (0-N). Values may exceed 100 in multi-core environments.
- **`mem_usage_percent`** (Int): Memory usage percentage trigger threshold (0-100).
- **`mem_usage_mb`** (Int): Memory usage absolute value trigger threshold (MB).
- These three configurations: **`cpu_usage_percent`**, **`mem_usage_percent`**, **`mem_usage_mb`** will skip the threshold check for this item if not configured or set to 0.

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
    - If using a non-standard JDK image, ensure the Sidecar mounts `/tmp` or the corresponding Java library path from the main container.

=== "Go (Coming Soon)"

    ### Go Profiling {#go-profiling}
    
    *Planned*: Integration with the `pprof` toolchain.
    
    **Expected Features:**

    - Support for Goroutine blocking analysis.
    - Support for Heap memory snapshots.

=== "Python (Coming Soon)"

    ### Python Profiling {#python-profiling}
    
    *Planned*: Integration with non-intrusive tools like `py-spy`.
<!-- markdownlint-enable -->

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

### Docker Local Testing {#docker-testing}

If you need to test in a local Docker environment, use the following command to start Flameshot and monitor the target container.

**Prerequisites:**

- The main container and Flameshot container must share `/opt/java/openjdk` (or the actual JDK path).
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
| `events` | No | Collection event types. Defaults to `all`. | `cpu,alloc` |

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

1. **File not uploaded?**

    - Check if `FLAMESHOT_PROFILING_PATH` is correctly mounted between the two containers.
    - The system automatically manages file life cycles and will attempt to delete temporary files after collection is complete.

## Changelog {#changelog}


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
