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

???+ tip

  采集数据需要用到 `8091` `9102` `18091` `19102` 几个端口，远程采集的时候，被采集服务器这些端口需要打开。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    支持以环境变量的方式修改配置参数（只在 Datakit 以 K8s DaemonSet 方式运行时生效，主机部署的 Datakit 不支持此功能）：

    | 环境变量名                              | 对应的配置参数项    | 参数示例                                                 |
    | :-----------------------------        | ---               | ---                                                     |
    | `ENV_INPUT_COUCHBASE_INTERVAL`        | `interval`        | `"30s"` (`"10s"` ~ `"60s"`)                             |
    | `ENV_INPUT_COUCHBASE_TIMEOUT`         | `timeout`         | `"5s"`  (`"5s"` ~ `"30s"`)                              |
    | `ENV_INPUT_COUCHBASE_SCHEME`          | `scheme`          | `"http"` or `"https"`                                   |
    | `ENV_INPUT_COUCHBASE_HOST`            | `host`            | `"127.0.0.1"`                                           |
    | `ENV_INPUT_COUCHBASE_PORT`            | `port`            | `8091` or `18091`                                       |
    | `ENV_INPUT_COUCHBASE_ADDITIONAL_PORT` | `additional_port` | `9102` or `19102`                                       |
    | `ENV_INPUT_COUCHBASE_USER`            | `user`            | `"Administrator"`                                       |
    | `ENV_INPUT_COUCHBASE_PASSWORD`        | `password`        | `"123456"`                                              |
    | `ENV_INPUT_COUCHBASE_TLS_OPEN`        | `tls_open`        | `true` or `false`                                       |
    | `ENV_INPUT_COUCHBASE_TLS_CA`          | `tls_ca`          | `""`                                                    |
    | `ENV_INPUT_COUCHBASE_TLS_CERT`        | `tls_cert`        | `"/var/cb/clientcertfiles/travel-sample.pem"`           |
    | `ENV_INPUT_COUCHBASE_TLS_KEY`         | `tls_key`         | `"/var/cb/clientcertfiles/travel-sample.key"`           |
    | `ENV_INPUT_COUCHBASE_TAGS`            | `tags`            | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
    | `ENV_INPUT_COUCHBASE_ELECTION`        | `election`        | `true` or `false`                                       |

    也可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

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
