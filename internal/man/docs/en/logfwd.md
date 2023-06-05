
# Collect Pod Log by Sidecar Mode
---

":material-kubernetes:"

---

In order to collect the log of application container in Kubernetes Pod, a lightweight log collection client is provided, which is mounted in Pod in sidecar mode and sends the collected log to DataKit.

## Use {#using}

It is divided into two parts, one is to configure DataKit to start the corresponding log receiving function, and the other is to configure and start logfwd collection.

### DataKit Configuration {#datakit-conf}


=== "Host Installation"

    You need to open [logfwdserver](logfwdserver.md), go to the `conf.d/log` directory under the DataKit installation directory, copy `logfwdserver.conf.sample` and name it  `logfwdserver.conf`. Examples are as follows:
    
    ``` toml hl_lines="1"
    [inputs.logfwdserver] # Note that this is the configuration of logfwdserver
      ## logfwd receiver listens for addresses and ports
      address = "0.0.0.0:9533"
    
      [inputs.logfwdserver.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
    ```
    
    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [injecting logfwdserver collector configuration in ConfigMap mode](datakit-daemonset-deploy.md#configmap-setting).

### logfwd Usage and Configuration  {#config}

The logfwd main configuration is in JSON format, and the following is a configuration example:

``` json
[
    {
        "datakit_addr": "127.0.0.1:9533",
        "loggings": [
            {
                "logfiles": ["<your-logfile-path>"],
                "ignore": [],
                "source": "<your-source>",
                "service": "<your-service>",
                "pipeline": "<your-pipeline.p>",
                "character_encoding": "",
                "multiline_match": "<your-match>",
                "remove_ansi_escape_codes": false,
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
    - `source` data source; if empty, 'default' is used by default.
    - `service` adds tag; if empty, $source is used by default.
    - `pipeline` pipeline script path, if empty $source.p will be used, if $source.p does not exist will not use pipeline (this script file exists on the DataKit side).
    - `character_encoding` # Select the code. If there is a misunderstanding in the code and the data cannot be viewed, it will be empty by default. Support `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030` or ""
    - `multiline_match` multi-line match, as in the [logging](logging.md) configuration, note that "no escape writing" with 3 single quotes is not supported because it is in JSON format, and regular `^\d{4}` needs to be escaped as `^\\d{4}`
    - `remove_ansi_escape_codes` whether to remove ANSI escape codes, such as the text color of standard output, and so on, with a value of `true` or `false`
    - `tags` adds additional `tag` written in a JSON map, such as `{ "key1":"value1", "key2":"value2" }`

Supported environment variables:

| Environment Variable Name                       | Configuration Item Mmeaning                                                                     |
| :---                             | :---                                                                           |
| `LOGFWD_DATAKIT_HOST`            | Datakit 地址                                                                   |
| `LOGFWD_DATAKIT_PORT`            | Datakit Port                                                                   |
| `LOGFWD_GLOBAL_SOURCE`           | Configure the global source with the highest priority                                                    |
| `LOGFWD_GLOBAL_SERVICE`          | Configure the global service with the highest priority                                                   |
| `LOGFWD_POD_NAME`                | Specifying pod name adds `pod_name` to tags                                       |
| `LOGFWD_POD_NAMESPACE`           | Specifying pod namespace adds `pod_namespace` to tags                            |
| `LOGFWD_ANNOTATION_DATAKIT_LOGS` | Use the annotations `datakit/logs` configuration of the current Pod with higher priority than the logfwd JSON configuration |

#### Installation and Running {#install-run}

The deployment configuration of logfwd in Kubernetes is divided into two parts. One is the configuration of Kubernetes Pod to create `spec.containers`, including injecting environment variables and mounting directories. The configuration is as follows:

```
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
    image: pubrepo.jiagouyun.com/datakit/logfwd:1.5.1
    imagePullPolicy: Always
    volumeMounts:
    - name: varlog
      mountPath: /var/log
    - mountPath: /opt/logfwd/config
      name: logfwd-config
      subPath: config
      workingDir: /opt/logfwd

```

The second configuration is the configuration where logfwd actually runs, the JSON-formatted master configuration mentioned earlier, which exists in Kubernetes as a ConfigMap.

According to the logfwd configuration example, modify `config` as it is. The `ConfigMap` format is as follows:

```
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
    image: pubrepo.jiagouyun.com/datakit/logfwd:1.5.1
    imagePullPolicy: Always
    volumeMounts:
    - name: varlog
      mountPath: /var/log
    - mountPath: /opt/logfwd/config
      name: logfwd-config
      subPath: config
    workingDir: /opt/logfwd
  volumes:
  - name: varlog
    emptyDir: {}
  - configMap:
      name: logfwd-conf
    name: logfwd-config

---

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

```
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
```

- Log file contains 1000w nginx logs, file size 2.2 GB:

```
192.168.17.1 - - [06/Jan/2022:16:16:37 +0000] "GET /google/company?test=var1%20Pl HTTP/1.1" 401 612 "http://www.google.com/" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36" "-"
```

- Results:

It takes**95 seconds** to read and forward all logs, with an average of 10w logs read per second.

The peak single-core CPU utilization rate was 42%, and the following is the `top` record at that time:

```
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
- [Introduction of DataKit configuration mode in Kubernetes environment](k8s-config-how-to.md)
- [Install DataKit as DaemonSet](datakit-daemonset-deploy.md)
- [Deploy `logfwdserver` on DataKit](logfwdserver.md)
- [Proper use of regular expressions to configure](datakit-input-conf.md#debug-regex)
