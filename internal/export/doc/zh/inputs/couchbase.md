---
title     : 'Couchbase'
summary   : '采集 Couchbase 服务器相关的指标数据'
__int_icon      : 'icon/couchbase'
dashboard :
  - desc  : 'Couchbase 内置视图'
    path  : 'dashboard/zh/couchbase'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Couchbase
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

Couchbase 采集器用于采集 Couchbase 服务器相关的指标数据。

Couchbase 采集器支持远程采集，可以运行在多种操作系统中。

已测试的版本：

- [x] Couchbase enterprise-7.2.0
- [x] Couchbase community-7.2.0

## 配置 {#config}

### 前置条件 {#requirements}

- 安装 Couchbase 服务
  
[官方文档 - CentOS/RHEL 安装](https://docs.couchbase.com/server/current/install/install-intro.html){:target="_blank"}

[官方文档 - Debian/Ubuntu 安装](https://docs.couchbase.com/server/current/install/ubuntu-debian-install.html){:target="_blank"}

[官方文档 - Windows 安装](https://docs.couchbase.com/server/current/install/install-package-windows.html){:target="_blank"}

- 验证是否正确安装

  在浏览器访问网址 `<ip>:8091` 可以进入 Couchbase 管理界面。

<!-- markdownlint-disable MD046 -->
???+ tip

    - 采集数据需要用到 `8091` `9102` `18091` `19102` 几个端口，远程采集的时候，被采集服务器这些端口需要打开。
<!-- markdownlint-enable -->

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

### TLS 配置 {#tls}

TLS 需要 Couchbase enterprise 版支持

[官方文档 - 配置服务器证书](https://docs.couchbase.com/server/current/manage/manage-security/configure-server-certificates.html){:target="_blank"}

[官方文档 - 配置客户端证书](https://docs.couchbase.com/server/current/manage/manage-security/configure-client-certificates.html){:target="_blank"}

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
