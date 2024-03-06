---
title     : 'ebpftrace'
summary   : '关联 eBPF 采集的链路 span，生成链路'
__int_icon      : 'icon/ebpf'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# ebpftrace
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

## 配置 {#config}

### 前置条件 {#requirements}

采集器默认开启采样，采样率默认值为 `0.1` 即 `10%` 的链路采样。

如果数据量在 1e6 span/min，目前需要至少提供 4C 的 cpu 资源和 4G 的 mem 资源。

`ebpftrace` 采集器用于接收和链接 eBPF span , 最终实现链路 trace_id 的生成，并建立 span 间的父子关系。

请参考以下部署方式（如下图）： 需要使所有 `ebpf` 外部采集器的 [`ebpf-trace`](./ebpf.md#ebpf-trace) 插件生成的数据发送至**同一个开启 `ebpftracing` 采集器的 DataKit**上，该 DataKit 将会所有 eBPF 采集器生成的链路 eBPF span 数据进行再处理后统一上传至观测云。

> 如果一个服务的三个应用 App 1 ～ 3 位于两个不同的节点，`ebpftrace` 目前根据 tcp seq 等来确认进程间的网络调用关系，需要对相关 eBPF span 进行链接以此生成 trace_id 和设置 parent_id。

![img0](./imgs/tracing.png)

### 采集器配置 {#input-config}

完成设置后需要将开启了 `ebpftrace` 采集器的 DataKit 或相关 K8s Service 的 `ip:port` 提供给 eBPF 采集器用于 eBPF Span 的传输。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

=== "Kubernetes"

    该采集器需要部署时需要限定副本数为 1，参考以下 yaml 进行部署，需要设置 yaml 中的 `ENV_DATAWAY` 和 `image` ：
  
    ```yaml
    apiVersion: v1
    kind: Namespace
    metadata:
      name: datakit-ebpftrace
    
    ---
    
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: datakit-ebpftrace
      labels:
        app: deployment-datakit-ebpftrace
      namespace: datakit-ebpftrace
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: deployment-datakit-ebpftrace
      template:
        metadata:
          labels:
            app: deployment-datakit-ebpftrace
        spec:
          containers:
          - name: datakit-ebpftrace
            image: 
            imagePullPolicy: Always
            ports:
            - containerPort: 9529
              protocol: TCP
            - containerPort: 6060
            resources:
              requests:
                cpu: "200m"
                memory: "256Mi"
              limits:
                cpu: "4000m"
                memory: "8Gi"
            env:
            - name: HOST_IP
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: status.hostIP
            - name: ENV_K8S_NODE_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
            - name: ENV_HTTP_LISTEN
              value: 0.0.0.0:9529
            - name: ENV_DATAWAY
              value: https://openway.guance.com?token=<xxx>
            - name: ENV_GLOBAL_TAGS
              value: host=__datakit_hostname,host_ip=__datakit_ip
            - name: ENV_DEFAULT_ENABLED_INPUTS
              value: ebpftrace
            - name: ENV_INPUT_EBPFTRACE_WINDOW
              value: "20s"
            - name: ENV_INPUT_EBPFTRACE_SAMPLING_RATE
              value: "0.1"
            - name: ENV_ENABLE_PPROF
              value: "true"
            - name: ENV_PPROF_LISTEN
              value: "0.0.0.0:6060"
    
    ---
    
    apiVersion: v1
    kind: Service
    metadata:
      name: datakit-ebpftrace-service
      namespace: datakit-ebpftrace
    spec:
      selector:
        app: deployment-datakit-ebpftrace
      ports:
        - protocol: TCP
          port: 9529
          targetPort: 9529
    ```

    通过以下环境变量可以调整 Kubernetes 中 ebpftrace 采集配置：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
