---
title     : 'CouchDB'
summary   : '采集 CouchDB 的指标数据'
__int_icon      : 'icon/couchdb'
dashboard :
  - desc  : 'CouchDB'
    path  : 'dashboard/zh/couchdb'
monitor   :
  - desc  : 'CouchDB'
    path  : 'monitor/zh/couchdb'
---

<!-- markdownlint-disable MD025 -->
# CouchDB
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

CouchDB 采集器用于采集 CouchDB 相关的指标数据，目前只支持 Prometheus 格式的数据

已测试的版本：

- [x] CouchDB 3.3.2
- [x] CouchDB 3.2
- [ ] CouchDBCouchDB 3.1 及以下版本不支持


## 前置条件 {#requirements}

- 安装 CouchDB 服务
  
参见[官方安装文档](https://docs.couchdb.org/en/stable/install/index.html){:target="_blank"}

- 验证是否正确安装

  在浏览器访问网址 `<ip>:5984/_utils/` 可以进入 CouchDB 管理界面。

- 打开 CouchDB Prometheus 端口
  
  找到并编辑 CouchDB 启动配置文件，通常是在 `/opt/couchdb/etc/local.ini`

  ```ini
  [prometheus]
  additional_port = false
  bind_address = 127.0.0.1
  port = 17986
  ```

  改为

  ```ini
  [prometheus]
  additional_port = true
  bind_address = 0.0.0.0
  port = 17986
  ```

  参见[官方配置文档](https://docs.couchdb.org/en/stable/config/misc.html#configuration-of-prometheus-endpoint){:target="_blank"}
  
- 重启 CouchDB 服务

<!-- markdownlint-disable MD046 -->
???+ tip

    - 采集数据需要用到 `5984` `17986` 几个端口，远程采集的时候，被采集服务器这些端口需要打开。
    - bind_address = 127.0.0.1 如果是本地采集，就不需要修改。
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
