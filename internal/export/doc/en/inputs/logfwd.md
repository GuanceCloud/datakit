---
title     : 'Sidecar for Pod Logging'
summary   : 'Collect pod logging via Sidecar'
tags:
  - 'KUBERNETES'
  - 'LOG'
  - 'CONTAINER'
__int_icon      : 'icon/kubernetes'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

":material-kubernetes:"

---

In order to collect the log of application container in Kubernetes Pod, a lightweight log collection client is provided, which is mounted in Pod in sidecar mode and sends the collected log to DataKit.

## Use {#using}

It is divided into two parts, one is to configure DataKit to start the corresponding log receiving function, and the other is to configure and start logfwd collection.

### DataKit Configuration {#datakit-conf}


<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    You need to open [logfwdserver](logfwdserver.md), go to the `conf.d/samples` directory under the DataKit installation directory, copy `logfwdserver.conf.sample` and name it  `logfwdserver.conf`. Examples are as follows:
    
    ``` toml hl_lines="1"
    [inputs.logfwdserver] # Note that this is the configuration of logfwdserver
      ## logfwd receiver listens for addresses and ports
      address = "0.0.0.0:9533"
    
      [inputs.logfwdserver.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [injecting logfwdserver collector configuration in ConfigMap mode](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### logfwd Usage and Configuration (1.86.0 and later) {#config-1-86-0}
<!-- markdownlint-enable -->

> logfwd is recommended for use in Kubernetes Serverless environments. If DaemonSet DataKit is already deployed, using logfwd may result in duplicate data.

Since logfwd version 1.86.0, the overall usage has been further simplified, and some cumbersome configurations have been removed. The main new capabilities are as follows:

- Support pulling `ClusterLoggingConfig` CRD through DataKit-Operator, automatically matching Pods and hot-loading collection configurations;
- Compatible with manual environment variable configuration (`LOGFWD_LOG_CONFIGS`) for scenarios without DataKit-Operator or debugging;
- Collection tasks communicate with DataKit `inputs.logfwdserver` via WebSocket, with automatic reconnection on connection failure (retry every second);
- Automatically supplement Pod metadata (`pod_name`, `namespace`, `pod_ip`) and target Labels, seamlessly compatible with the old volume/mount solution.

#### Startup Method Overview {#config-1-86-0-overview}

| Scenario                          | Key Variables                                          | Description                                                                                                                      |
| :---                              | :---                                                   | :---                                                                                                                             |
| With DataKit-Operator (Recommended) | `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` + Pod metadata     | DataKit-Operator returns matching CRD JSON, logfwd automatically creates/refreshes tailers; log paths, pipelines, etc. need to be declared in `ClusterLoggingConfig`. |
| Manual Configuration              | `LOGFWD_LOG_CONFIGS`                                   | Consistent with the old JSON semantics, but passed through environment variables, suitable for development/transition scenarios, can coexist with DataKit-Operator (manual configuration has higher priority). |

You still need to prepare shared `volume`/`volumeMount` for log files as in the old version; logfwd only monitors and does not create mounts.

#### Global Environment Variables {#config-1-86-0-globals}

| Environment Variable Name          | Configuration Item Meaning                                                                                                                                                                                                           |
| :---                               | :---                                                                                                                                                                                                                                 |
| `LOGFWD_LOG_LEVEL`                 | Runtime log level, default `info`, set to `debug` to see more debug output.                                                                                                                                                          |
| `LOGFWD_DATAKIT_HOST`              | DataKit instance address (IP or resolvable domain name).                                                                                                                                                                             |
| `LOGFWD_DATAKIT_PORT`              | DataKit `logfwdserver` listening port, e.g., `9533`.                                                                                                                                                                                 |
| `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` | DataKit-Operator Endpoint, such as `datakit-operator.datakit.svc:443` or `https://datakit-operator.datakit.svc:443`, used to query CRD configuration; leave empty to skip pulling. Supports automatic addition of `https://` prefix. |
| `LOGFWD_GLOBAL_SOURCE`             | Global `source`, priority higher than the `source` field in individual configurations.                                                                                                                                               |
| `LOGFWD_GLOBAL_SERVICE`            | Global `service`, if not specified in individual configuration, use global value; if global value is also empty, fall back to `source`.                                                                                              |
| `LOGFWD_GLOBAL_STORAGE_INDEX`      | Global `storage_index`, priority higher than the `storage_index` field in individual configurations.                                                                                                                                 |
| `LOGFWD_POD_NAME`                  | Automatically writes `pod_name` tag, usually injected via Downward API.                                                                                                                                                              |
| `LOGFWD_POD_NAMESPACE`             | Automatically writes `namespace` tag.                                                                                                                                                                                                |
| `LOGFWD_POD_IP`                    | Automatically writes `pod_ip` tag for locating container instances.                                                                                                                                                                  |

> Tip: If you need to attach more tags, you can mount the `/etc/podinfo/labels` file in the Pod (automatically added when DataKit-Operator injects logfwd sidecar), and logfwd will parse and align with `podTargetLabels` in the CRD.

#### Collection Configuration {#config-1-86-0-log-configs}

logfwd supports two configuration methods, in order of priority from high to low:

1. **Manual Configuration (`LOGFWD_LOG_CONFIGS`)**: Pass JSON string through environment variables, structure is basically consistent with the old `loggings` sub-items. When manual configuration exists, logfwd immediately creates tailers and maintains this configuration during the process lifetime; after deleting the variable or clearing the content, the container needs to be restarted to release.
1. **DataKit-Operator CRD**: When `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` is specified, logfwd calls the DataKit-Operator API once per minute, and determines whether hot update is needed by MD5 verification of configuration content. After configuration changes, tailers are automatically recreated without restarting the container.

> **Note**: If both manual configuration and CRD configuration exist and point to the same log path, duplicate collection will occur. It is recommended to prioritize CRD configuration, and manual configuration is only for debugging or special scenarios.

The `LOGFWD_LOG_CONFIGS` field structure example is as follows:

``` json
[
  {
    "type": "file",
    "disable": false,
    "source": "nginx-access",
    "service": "nginx",
    "path": "/var/log/nginx/access.log",
    "pipeline": "nginx-access.p",
    "storage_index": "app-logs",
    "multiline_match": "^\\d{4}-\\d{2}-\\d{2}",
    "remove_ansi_escape_codes": false,
    "from_beginning": false,
    "character_encoding": "utf-8",
    "tags": {
      "env": "production",
      "team": "backend"
    }
  }
]
```

| Field                           | Type    | Required               | Description                                                                                                                               | Example                   |
| ------                          | ------  | ------                 | ------                                                                                                                                    | ------                    |
| `type`                          | string  | Yes                    | logfwd collection type can only be `"file"`                                                                                               | `"file"`                  |
| `disable`                       | boolean | No                     | Whether to disable this collection configuration                                                                                          | `false`                   |
| `source`                        | string  | Yes                    | Log source identifier, used to distinguish different log streams                                                                          | `"nginx-access"`          |
| `service`                       | string  | No                     | Service to which the log belongs, default value is the log source (source)                                                                | `"nginx"`                 |
| `path`                          | string  | Conditionally Required | Log file path (supports glob patterns), required when type=file                                                                           | `"/var/log/nginx/*.log"`  |
| `multiline_match`               | string  | No                     | Regular expression for the start line of multi-line logs, note that backslashes need to be escaped in JSON                                | `"^\\d{4}-\\d{2}-\\d{2}"` |
| `pipeline`                      | string  | No                     | Log parsing pipeline configuration file name (needs to be configured on the DataKit side)                                                 | `"nginx-access.p"`        |
| `storage_index`                 | string  | No                     | Log storage index name                                                                                                                    | `"app-logs"`              |
| `remove_ansi_escape_codes`      | boolean | No                     | Whether to remove ANSI escape characters (color codes, etc.) from log data                                                                | `false`                   |
| `from_beginning`                | boolean | No                     | Whether to collect logs from the beginning of the file (default starts from the end of the file)                                          | `false`                   |
| `from_beginning_threshold_size` | int     | No                     | When a file is discovered, if the file size is less than this value, start reading from the beginning of the file, in bytes, default 20MB | `1000`                    |
| `character_encoding`            | string  | No                     | Character encoding, supports `utf-8`, `utf-16le`, `utf-16be`, `gbk`, `gb18030` or empty string (auto-detect). Default is empty.           | `"utf-8"`                 |
| `tags`                          | object  | No                     | Additional tag key-value pairs that will be attached to each log record                                                                   | `{"env": "prod"}`         |

When `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` is configured, logfwd will make requests to DataKit-Operator based on `LOGFWD_POD_NAMESPACE`, `LOGFWD_POD_NAME`, and `pod_labels` (optional, requires mounting `/etc/podinfo/labels` file). As long as a `ClusterLoggingConfig` CRD rule matches the current Pod, the corresponding `configs` JSON will be returned and hot update will be triggered.

CRD Configuration Example:

```yaml
apiVersion: logging.datakits.io/v1alpha1
kind: ClusterLoggingConfig
metadata:
  name: nginx-logs
spec:
  selector:
    namespaceRegex: "^(default|production)$"
    podRegex: "^(nginx-.*)$"
    podLabelSelector: "app=nginx,env=production"
    containerRegex: "^(nginx|app)$"
  
  podTargetLabels:
    - app
    - version
    - team
  
  configs:
    - type: "file"
      source: "nginx-access"
      path: "/var/log/nginx/access.log"
      pipeline: "nginx-access.p"
      storage_index: "app-logs"
      tags:
        log_type: "access"
        component: "nginx"
    
    - type: "file"
      source: "nginx-error"
      path: "/var/log/nginx/error.log"
      pipeline: "nginx-error.p"
      storage_index: "app-logs"
      tags:
        log_type: "error"
        component: "nginx"
```

CRD Selector Description:

| Field               | Type   | Required | Description                                      | Example                                 |
| ------              | ------ | ------   | ------                                           | ------                                  |
| `namespaceRegex`    | string | No       | Namespace name regex match (all conditions are AND relationship) | `"^(default\|production)$"`             |
| `podRegex`          | string | No       | Pod name regex match                             | `"^(nginx-.*)$"`                        |
| `podLabelSelector`  | string | No       | Pod label selector (comma-separated key=value pairs) | `"app=nginx,environment=production"`    |
| `containerRegex`    | string | No       | Container name regex match                       | `"^(nginx\|app-container)$"`            |

`podTargetLabels`*: Specifies the list of label keys to extract from Pod Labels and attach to logs. logfwd will read the `/etc/podinfo/labels` file (injected by Downward API or DataKit-Operator), extract matching labels and add them to the log `tags`.

Configuration Hot Update Mechanism:

- logfwd polls the DataKit-Operator API once per minute
- Determines if there are changes by calculating the MD5 value of the configuration content
- After configuration changes, automatically stops old tailers and creates new tailers without restarting the container
- Configuration changes usually take effect within 1 minute

<!-- markdownlint-disable MD046 -->
??? topic

    - Log directories need to be shared in advance using `volumes`/`volumeMounts` in the business Pod/sidecar (such as `emptyDir`), otherwise logfwd cannot access log files.
    - `LOGFWD_LOG_CONFIGS` and CRD configurations are independent of each other. If both point to the same path, duplicate collection will occur.
    - DataKit-Operator supports automatic injection of logfwd sidecar and mounts for target Pods. For details, please refer to the DataKit-Operator [documentation](../datakit/datakit-operator.md).
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
#### ClusterLoggingConfig CRD Selector Support {#config-1-86-0-crd-selector}
<!-- markdownlint-enable -->

When logfwd queries the `ClusterLoggingConfig` CRD through DataKit-Operator, it supports the following selector fields to match target CRDs and log collection configurations:

| Selector Field      | Description                                                                                                                         | Example                                    |
| :------------------ | :-------------------------------------------------------------------------------------------------------                            | :----------------------------------------- |
| `namespaceRegex`    | Namespace name regex matching, using the `LOGFWD_POD_NAMESPACE` environment variable of the logfwd container as the query parameter | `"^(default)$"`                            |
| `podNameRegex`      | Pod name regex matching, using the `LOGFWD_POD_NAME` environment variable of the logfwd container as the query parameter            | `"^(nginx-app.*)$"`                        |
| `podLabelSelector`  | Pod label selector (prerequisite: the `/etc/podinfo/labels` file in the logfwd container has labels content)                        | `"app=nginx,environment=production"`       |

<!-- markdownlint-disable MD046 -->
??? topic

    - logfwd **does not support** the `containerRegex` selector. Since logfwd runs as a Pod Sidecar, it only collects log files and cannot distinguish container names.
    - The use of `podLabelSelector` depends on the existence of the `/etc/podinfo/labels` file. DataKit-Operator automatically mounts this file when injecting the logfwd sidecar (via Downward API). If this file does not exist or is empty, `podLabelSelector` will not take effect.
    - All selector conditions have an **AND** relationship, meaning all specified selectors must match for a Pod to be selected.
<!-- markdownlint-enable -->

#### Example: Kubernetes Pod Configuration {#config-1-86-0-example}

- Using DataKit-Operator CRD Configuration

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-app
  namespace: default
  labels:
    app: nginx
    version: v1.0
spec:
  containers:
  - name: nginx
    image: nginx:latest
    volumeMounts:
    - name: nginx-logs
      mountPath: /var/log/nginx
  - name: logfwd
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:{{ .Version }}
    env:
    - name: LOGFWD_LOG_LEVEL
      value: "info"  # Optional: debug to see detailed logs
    - name: LOGFWD_DATAKIT_HOST
      valueFrom:
        fieldRef:
          fieldPath: status.hostIP
    - name: LOGFWD_DATAKIT_PORT
      value: "9533"
    - name: LOGFWD_DATAKIT_OPERATOR_ENDPOINT
      value: datakit-operator.datakit.svc:443
    - name: LOGFWD_POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    - name: LOGFWD_POD_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    - name: LOGFWD_POD_IP
      valueFrom:
        fieldRef:
          fieldPath: status.podIP
    volumeMounts:
    - name: podinfo
      mountPath: /etc/podinfo
      readOnly: true
    - name: nginx-logs
      mountPath: /var/log/nginx
      readOnly: true
  volumes:
  - name: podinfo
    downwardAPI:
      items:
      - path: "labels"
        fieldRef:
          fieldPath: metadata.labels
  - name: nginx-logs
    emptyDir: {}
```

Corresponding `ClusterLoggingConfig` CRD configuration:

```yaml
apiVersion: logging.datakits.io/v1alpha1
kind: ClusterLoggingConfig
metadata:
  name: nginx-logs
spec:
  selector:
    namespaceRegex: "^default$"
    podLabelSelector: "app=nginx"
  podTargetLabels:
    - app
    - version
  configs:
    - type: "file"
      source: "nginx-access"
      path: "/var/log/nginx/access.log"
      pipeline: "nginx-access.p"
    - type: "file"
      source: "nginx-error"
      path: "/var/log/nginx/error.log"
      pipeline: "nginx-error.p"
```

- Using Manual Configuration

If you need to temporarily use manual configuration or debugging, you can add the `LOGFWD_LOG_CONFIGS` environment variable:

```yaml
spec:
  containers:
  - name: logfwd
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:{{ .Version }}
    env:
    - name: LOGFWD_DATAKIT_HOST
      valueFrom:
        fieldRef:
          fieldPath: status.hostIP
    - name: LOGFWD_DATAKIT_PORT
      value: "9533"
    - name: LOGFWD_POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    - name: LOGFWD_POD_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    - name: LOGFWD_LOG_CONFIGS
      value: |
        [
          {
            "type": "file",
            "source": "app-logs",
            "path": "/var/log/app/*.log",
            "pipeline": "app.p",
            "from_beginning": false,
            "tags": {
              "env": "production"
            }
          }
        ]
    volumeMounts:
    - name: app-logs
      mountPath: /var/log/app
      readOnly: true
  volumes:
  - name: app-logs
    emptyDir: {}
```

The mounting patterns, `volumes`/`volumeMounts` syntax, resource limits, etc. remain consistent with versions before 1.86.0, and you can continue to refer to the old version examples in the next section.

<!-- markdownlint-disable MD013 -->
### logfwd Usage and Configuration (Before 1.86.0) {#config-before-1-86-0}
<!-- markdownlint-enable -->

The logfwd main configuration is in JSON format, and the following is a configuration example:

``` json
[
    {
        "datakit_addr": "127.0.0.1:9533",
        "loggings": [
            {
                "logfiles":      ["<your-logfile-path>"],
                "ignore":        [],
                "storage_index": "<your-storage-index>",
                "source":        "<your-source>",
                "service":       "<your-service>",
                "pipeline":      "<your-pipeline.p>",
                "character_encoding": "",
                "multiline_match": "<your-match>",
                "tags": {}
            },
            {
                "logfiles": ["<your-logfile-path-2>"],
                "source": "<your-source-2>"
            }
        ]
    }
]
```

Description of configuration parameters:

- `datakit_addr` is the DataKit logfwdserver address, typically configured with the environment variables `LOGFWD_DATAKIT_HOST` and `LOGFWD_DATAKIT_PORT`

- `loggings` is the primary configuration, an array, and the subitems are basically the same as the [logging](logging.md) collector.
    - `logfiles` list of log files, you can specify absolute paths, support batch specifying using glob rules, and recommend using absolute paths.
    - `ignore` file path filtering, using glob rules, the file will not be collected if any filtering condition is met.
    - `storage_index` set storage index
    - `source` data source; if empty, 'default' is used by default.
    - `service` adds tag; if empty, $source is used by default.
    - `pipeline` Pipeline script path, if empty $source.p will be used, if $source.p does not exist will not use Pipeline (this script file exists on the DataKit side).
    - `character_encoding` # Select the code. If there is a misunderstanding in the code and the data cannot be viewed, it will be empty by default. Support `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030` or ""
    - `multiline_match` multi-line match, as in the [logging](logging.md) configuration, note that "no escape writing" with 3 single quotes is not supported because it is in JSON format, and regular `^\d{4}` needs to be escaped as `^\\d{4}`
    - `tags` adds additional `tag` written in a JSON map, such as `{ "key1":"value1", "key2":"value2" }`

Supported environment variables:

| Environment Variable Name        | Configuration Item Meaning                                                                                                  |
| :---                             | :---                                                                                                                        |
| `LOGFWD_DATAKIT_HOST`            | DataKit Address                                                                                                             |
| `LOGFWD_DATAKIT_PORT`            | DataKit Port                                                                                                                |
| `LOGFWD_GLOBAL_SOURCE`           | Configure the global source with the highest priority                                                                       |
| `LOGFWD_GLOBAL_STORAGE_INDEX`    | Configure the global storage_index with the highest priority                                                                |
| `LOGFWD_GLOBAL_SERVICE`          | Configure the global service with the highest priority                                                                      |
| `LOGFWD_POD_NAME`                | Specifying pod name adds `pod_name` to tags                                                                                 |
| `LOGFWD_POD_NAMESPACE`           | Specifying pod namespace adds `namespace` to tags                                                                           |
| `LOGFWD_ANNOTATION_DATAKIT_LOGS` | Use the annotations `datakit/logs` configuration of the current Pod with higher priority than the logfwd JSON configuration |
| `LOGFWD_JSON_CONFIG`             | Logfwd main configuration, i.e. the JSON-formatted text above                                                               |

#### Installation and Running {#install-run}

The deployment configuration of logfwd in Kubernetes is divided into two parts. One is the configuration of Kubernetes Pod to create `spec.containers`, including injecting environment variables and mounting directories. The configuration is as follows:

```yaml
spec:
  containers:
  - name: logfwd
    env:
    - name: LOGFWD_DATAKIT_HOST
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: status.hostIP
    - name: LOGFWD_DATAKIT_PORT
      value: "9533"
    - name: LOGFWD_ANNOTATION_DATAKIT_LOGS
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.annotations['datakit/logs']
    - name: LOGFWD_POD_NAME
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.name
    - name: LOGFWD_POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: LOGFWD_GLOBAL_SOURCE
      value: nginx-souce-test
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:{{.Version}}
    imagePullPolicy: Always
    resources:
      requests:
        cpu: "200m"
        memory: "128Mi"
      limits:
        cpu: "1000m"
        memory: "2Gi"
    volumeMounts:
    - mountPath: /opt/logfwd/config
      name: logfwd-config-volume
      subPath: config
    workingDir: /opt/logfwd
  volumes:
  - configMap:
      name: logfwd-config
    name: logfwd-config-volume
```

The second configuration is the configuration where logfwd actually runs, the JSON-formatted master configuration mentioned earlier, which exists in Kubernetes as a ConfigMap.

According to the logfwd configuration example, modify `config` as it is. The `ConfigMap` format is as follows:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: logfwd-conf
data:
  config: |
    [
        {
            "loggings": [
                {
                    "logfiles": ["/var/log/1.log"],
                    "source": "log_source",
                    "tags": {}
                },
                {
                    "logfiles": ["/var/log/2.log"],
                    "source": "log_source2"
                }
            ]
        }
    ]
```

By integrating the two configurations into the existing Kubernetes yaml and using `volumes` and `volumeMounts` to share directories within containers, the logfwd container collects log files from other containers.

> Note that you need to use `volumes` and `volumeMounts` to mount and share the log directory of the application container (that is, the `count` container in the example) for normal access in the logfwd container. See `volumes` [doc](https://kubernetes.io/docs/concepts/storage/volumes/){:target="_blank"}

The complete example is as follows:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: logfwd
spec:
  containers:
  - name: count
    image: busybox
    args:
    - /bin/sh
    - -c
    - >
      i=0;
      while true;
      do
        echo "$i: $(date)" >> /var/log/1.log;
        echo "$(date) INFO $i" >> /var/log/2.log;
        i=$((i+1));
        sleep 1;
      done
    volumeMounts:
    - name: varlog
      mountPath: /var/log
  - name: logfwd
    env:
    - name: LOGFWD_DATAKIT_HOST
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: status.hostIP
    - name: LOGFWD_DATAKIT_PORT
      value: "9533"
    - name: LOGFWD_ANNOTATION_DATAKIT_LOGS
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.annotations['datakit/logs']
    - name: LOGFWD_POD_NAME
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.name
    - name: LOGFWD_POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:{{.Version}}
    imagePullPolicy: Always
    resources:
      requests:
        cpu: "200m"
        memory: "128Mi"
      limits:
        cpu: "1000m"
        memory: "2Gi"
    volumeMounts:
    - name: varlog
      mountPath: /var/log
    - mountPath: /opt/logfwd/config
      name: logfwd-config-volume
      subPath: config
    workingDir: /opt/logfwd
  volumes:
  - name: varlog
    emptyDir: {}
  - configMap:
      name: logfwd-config
    name: logfwd-config-volume

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: logfwd-config
data:
  config: |
    [
        {
            "loggings": [
                {
                    "logfiles": ["/var/log/1.log"],
                    "source": "log_source",
                    "tags": {
                        "flag": "tag1"
                    }
                },
                {
                    "logfiles": ["/var/log/2.log"],
                    "source": "log_source2"
                }
            ]
        }
    ]
```

### Performance Test {#bench}

- Environment:

```shell
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
```

- Log file contains 1000w nginx logs, file size 2.2 GB:

```not-set
192.168.17.1 - - [06/Jan/2022:16:16:37 +0000] "GET /google/company?test=var1%20Pl HTTP/1.1" 401 612 "http://www.google.com/" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36" "-"
```

- Results:

It takes**95 seconds** to read and forward all logs, with an average of 10w logs read per second.

The peak single-core CPU utilization rate was 42%, and the following is the `top` record at that time:

```shell
top - 16:32:46 up 52 days,  7:28, 17 users,  load average: 2.53, 0.96, 0.59
Tasks: 464 total,   2 running, 457 sleeping,   0 stopped,   5 zombie
%Cpu(s): 30.3 us, 33.7 sy,  0.0 ni, 34.3 id,  0.1 wa,  0.0 hi,  1.5 si,  0.0 st
MiB Mem :  15885.2 total,    985.2 free,   6204.0 used,   8696.1 buff/cache
MiB Swap:   2048.0 total,      0.0 free,   2048.0 used.   8793.3 avail Mem

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1850829 root      20   0  715416  17500   8964 R  42.1   0.1   0:10.44 logfwd
```

## More Readings {#more-reading}

- [DataKit summary of log collection](datakit-logging.md)
- [Socket Log access best practices](logging_socket.md)
- [Log collection configuration for specifying pod in Kubernetes](container-log.md#logging-with-annotation-or-label)
- [Third-party log access](logstreaming.md)
- [Introduction of DataKit configuration mode in Kubernetes environment](../datakit/k8s-config-how-to.md)
- [Install DataKit as DaemonSet](../datakit/datakit-daemonset-deploy.md)
- [Deploy `logfwdserver` on DataKit](logfwdserver.md)
- [Proper use of regular expressions to configure](../datakit/datakit-input-conf.md#debug-regex)
