---
title     : 'Neo4j'
summary   : '采集 Neo4j 的指标数据'
__int_icon      : 'icon/neo4j'
dashboard :
  - desc  : 'Neo4j'
    path  : 'dashboard/zh/neo4j'
monitor   :
  - desc  : 'Neo4j'
    path  : 'monitor/zh/neo4j'
---

<!-- markdownlint-disable MD025 -->
# Neo4j
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

Neo4j 采集器用于采集 Neo4j 相关的指标数据，目前只支持 Prometheus 格式的数据

已测试的版本：

- [x] Neo4j 5.11.0 enterprise
- [x] Neo4j 4.4.0 enterprise
- [x] Neo4j 3.4.0 enterprise
- [ ] Neo4j 3.3.0 enterprise 及以下版本不支持
- [ ] Neo4j 5.11.0 community community 版本均不支持


## 前置条件 {#requirements}

- 安装 Neo4j 服务
  
参见[官方安装文档](https://neo4j.com/docs/operations-manual/current/installation/){:target="_blank"}

- 验证是否正确安装

  在浏览器访问网址 `<ip>:7474` 可以进入 Neo4j 管理界面。

- 打开 Neo4j Prometheus 端口
  
  找到并编辑 Neo4j 启动配置文件，通常是在 `/etc/neo4j/neo4j.conf`

  尾部追加

  ```ini
  # Enable the Prometheus endpoint. Default is false.
  server.metrics.prometheus.enabled=true
  # The hostname and port to use as Prometheus endpoint.
  # A socket address is in the format <hostname>, <hostname>:<port>, or :<port>.
  # If missing, the port or hostname is acquired from server.default_listen_address.
  # The default is localhost:2004.
  server.metrics.prometheus.endpoint=0.0.0.0:2004
  ```

  参见[官方配置文档](https://neo4j.com/docs/operations-manual/current/monitoring/metrics/expose/#_prometheus){:target="_blank"}
  
- 重启 Neo4j 服务

<!-- markdownlint-disable MD046 -->
???+ tip

    - 采集数据需要用到 `2004` 端口，远程采集的时候，被采集服务器这些端口需要打开。
    - 0.0.0.0:2004 如果是本地采集，可以改为 localhost:2004。
<!-- markdownlint-enable -->

## 配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

<!-- markdownlint-enable -->

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
