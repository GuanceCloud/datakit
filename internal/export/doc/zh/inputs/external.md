---
title     : 'External'
summary   : '启动外部程序进行采集'
__int_icon      : 'icon/external'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# External
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

External 采集器可以启动外部程序进行采集。

## 配置 {#config}

### 前置条件 {#requirements}

- 启动命令的程序及其运行环境的依赖完备。比如用 Python 去启动外部 Python 脚本，则该脚本运行所需的引用包等依赖必须要有。

### 采集器配置 {#input-config}

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
