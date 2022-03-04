{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# logfwd 日志采集客户端

## 介绍

为了便于在 Kubernetes Pod 中采集应用容器的日志，提供一个轻量的日志采集客户端，以 sidecar 方式挂载到 Pod 中，并将采集到的日志发送给 DataKit。

## 使用

分成两部分，一是配置 DataKit 开启相应的日志接收功能，二是配置和启动 logfwd 采集。

### DataKit 配置

进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `logfwdserver.conf.sample` 并命名为 `logfwdserver.conf`。示例如下：

``` toml
[inputs.logfwdserver]
  ## logfwd 接收端监听地址和端口
  address = "0.0.0.0:9533"

  [inputs.logfwdserver.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

配置好后，重启 DataKit 即可。

> 注：如果 DataKit 是以 daemonset 方式部署，此段配置需要添加到 `ConfigMap` 并通过 `volumeMounts` 挂载，详见 DataKit daemonset 安装[文档](datakit-daemonset-deploy.md)。

### logfwd 使用和配置

logfwd 主配置是 JSON 格式，以下是配置示例：

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

配置参数说明：

- `datakit_addr` 是 DataKit logfwdserver 地址，通常使用环境变量 `LOGFWD_DATAKIT_HOST` 和 `LOGFWD_DATAKIT_PORT` 进行配置

- `loggings` 为主要配置，是一个数组，子项也基本和 [logging](logging) 采集器相同。
    - `logfiles` 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定，推荐使用绝对路径
    - `ignore` 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
    - `source` 数据来源，如果为空，则默认使用 'default'
    - `service` 新增标记 tag，如果为空，则默认使用 $source
    - `pipeline` pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 pipeline（此脚本文件存在于 DataKit 端）
    - `character_encoding` # 选择编码，如果编码有误会导致数据无法查看，默认为空即可。支持`utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030` or ""
    - `multiline_match` 多行匹配，与 [logging](logging) 该项配置一样，注意因为是 JSON 格式所以不支持 3 个单引号的“不转义写法”，正则 `^\d{4}` 需要添加转义写成 `^\\d{4}`
    - `remove_ansi_escape_codes` 是否删除 ANSI 转义码，例如标准输出的文本颜色等，值为 `true` 或 `false`
    - `tags` 添加额外 `tag`，书写格式是 JSON map，例如 `{ "key1":"value1", "key2":"value2" }`

#### 安装和运行

logfwd 在 Kubernetes 的部署配置分为两部分，一是 Kubernetes Pod 创建 `spec.containers` 的配置，包括注入环境变量和挂载目录。配置如下：

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
    image: pubrepo.jiagouyun.com/datakit/logfwd:{{.Version}}
    imagePullPolicy: Always
    volumeMounts:
    - name: varlog
      mountPath: /var/log
    - mountPath: /opt/logfwd/config
      name: logfwd-config
      subPath: config
      workingDir: /opt/logfwd

```

第二份配置为 logfwd 实际运行的配置，即前文提到的 JSON 格式的主配置，在 Kubernetes 中以 ConfigMap 形式存在。

根据 logfwd 配置示例，按照实际情况修改 `config`。`ConfigMap` 格式如下：

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


将两份配置集成到现有的 Kubernetes yaml 中，并使用 `volumes` 和 `volumeMounts` 将目录在 containers 内部共享，即可实现 logfwd 容器采集其他容器的日志文件。

完整示例如下：

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
    image: pubrepo.jiagouyun.com/datakit/logfwd:{{.Version}}
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
		        "flag": "log_source1"
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

> 注意，需要使用 `volumes` 和 `volumeMounts` 将应用容器（即示例中的 `count` 容器）的日志目录挂载和共享，以便在 logfwd 容器中能够正常访问到。`volumes` 官方说明[文档](https://kubernetes.io/docs/concepts/storage/volumes/)

### 性能测试

- 环境：

```
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
```

- 日志文件内容为 1000w 条 nginx 日志，文件大小 2.2GB：

```
192.168.17.1 - - [06/Jan/2022:16:16:37 +0000] "GET /google/company?test=var1%20Pl HTTP/1.1" 401 612 "http://www.google.com/" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36" "-"
```

- 结果：

耗时**95 秒**将所有日志读取和转发完毕，平均每秒读取 10w 条日志。

单核心 CPU 使用率峰值为 42%，以下是当时的 `top` 记录：

```
top - 16:32:46 up 52 days,  7:28, 17 users,  load average: 2.53, 0.96, 0.59
Tasks: 464 total,   2 running, 457 sleeping,   0 stopped,   5 zombie
%Cpu(s): 30.3 us, 33.7 sy,  0.0 ni, 34.3 id,  0.1 wa,  0.0 hi,  1.5 si,  0.0 st
MiB Mem :  15885.2 total,    985.2 free,   6204.0 used,   8696.1 buff/cache
MiB Swap:   2048.0 total,      0.0 free,   2048.0 used.   8793.3 avail Mem

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1850829 root      20   0  715416  17500   8964 R  42.1   0.1   0:10.44 logfwd
```
