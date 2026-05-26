# DataKit Operator 注入日志采集配置

DataKit Operator 可以为指定的 Pod 自动添加 DataKit Logging 采集所需的配置，包括 `datakit/logs` 注解和对应的文件路径 volume/volumeMount，简化了手动配置的繁杂步骤。这样，用户无需手动干预每个 Pod 配置即可自动启用日志采集功能。

以下是一个配置示例，展示了如何通过 DataKit Operator 的 `admission_mutate` 配置来实现日志采集配置的自动注入：

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        # 其他配置
    },
    "admission_mutate": {
        "loggings": [
            {
                "namespace_selectors": ["middleware"],
                "label_selectors":     ["app=logging"],
                "config": "[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/**/*.log\",\"source\":\"logging-tmp\"}]"
            }
        ]
    }
}
```

`admission_mutate.loggings`：这是一个对象数组，包含多个日志采集配置。每个日志配置包括以下字段：

- `namespace_selectors`：限定符合条件的 Pod 所在的 Namespacce。可以设置多个 Namespace，Pod 必须匹配至少一个 Namespace 才会被选中。与 `label_selectors` 是“或”的关系。
- `label_selectors`：限定符合条件的 Pod 的 label。Pod 必须匹配至少一个 label selector 才会被选中。与 `namespace_selectors` 是“或”的关系。
- `config`：这是一个 JSON 字符串，它将被添加到 Pod 的注解中，注解的 Key 是 `datakit/logs`。如果该 Key 已经存在，它不会被覆盖或重复添加。这个配置将告诉 DataKit 如何采集日志。

DataKit Operator 会自动解析 `config` 配置，并根据其中的路径（`path`）为 Pod 创建对应的 volume 和 volumeMount。

以上述 DataKit Operator 配置为例，如果发现某个 Pod 的 Namespace 是 `middleware`，或 Labels 匹配 `app=logging`，就在 Pod 新增注解和挂载。例如：

```yaml hl_lines="5"
apiVersion: v1
kind: Pod
metadata:
  annotations:
    datakit/logs: '[{"disable":false,"type":"file","path":"/tmp/opt/**/*.log","source":"logging-tmp"}]'
  labels:
    app: logging
  name: logging-test
  namespace: default
spec:
  containers:
  - args:
    - |
      mkdir -p /tmp/opt/log1;
      i=1;
      while true; do
        echo "Writing logs to file ${i}.log";
        for ((j=1;j<=10000000;j++)); do
          echo "$(date +'%F %H:%M:%S')  [$j]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log1/file_${i}.log;
          sleep 1;
        done;
        echo "Finished writing 5000000 lines to file_${i}.log";
        i=$((i+1));
      done
    command:
    - /bin/bash
    - -c
    - --
    image: pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04
    imagePullPolicy: IfNotPresent
    name: demo
    volumeMounts:
    - mountPath: /tmp/opt
      name: datakit-logs-volume-0
  volumes:
  - emptyDir: {}
    name: datakit-logs-volume-0
```

这个 Pod 存在 label `app=logging`，能够匹配上，于是 DataKit Operator 就给它添加了 `datakit/logs` 注解，并且将路径 `/tmp/opt` 添加 EmptyDir 挂载。

DataKit 日志采集发现到 Pod 后，就会根据 `datakit/logs` 内容进行定制化采集。
