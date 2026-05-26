---
title     : '主机配置变更'
summary   : '监控主机配置的变更，并上报变更事件数据'
tags:
  - '主机'
__int_icon: ''
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

主机配置变更采集器支持监控 Linux 主机上各类配置的变更，构建变更事件数据并上报<<<custom_key.brand_name>>>平台。

> **注意**：此采集器仅支持 Linux 操作系统，不支持 Windows 系统。

## 功能说明 {#features}

主机配置变更采集器支持以下变更检测功能：

| 功能模块 | 说明 |
|---------|------|
| 用户和组变更 | 监控 `/etc/passwd`、`/etc/shadow`、`/etc/group`、`/etc/gshadow` 文件，检测用户和组的创建、删除、属性修改及成员变更 |
| Crontab 变更 | 监控 `/etc/crontab`、`/etc/cron.d/*`、`/var/spool/cron/crontabs/*` 文件，检测定时任务变更 |
| 文件内容变更 | 监控指定文件的内容变更，支持差异对比 |
| 服务变更 | 监控 `systemd` 或 `sysvinit` 服务的创建、删除、属性修改及状态变更 |
| 网络配置变更 | 监控网络接口、DNS 配置、路由配置、防火墙规则、hosts 文件的变更 |

## 配置 {#config}

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 开启采集器。

<!-- markdownlint-enable -->

## 变更事件 {#change-event}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

### 事件字段说明 {#event-fields}

<!-- markdownlint-disable MD024 -->

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "keyevent"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

### 变更事件类型 {#event-types}

- 用户和组变更事件

| 变更 ID | 说明 |
|---------|------|
| `host_change_01_01` | 新增用户 |
| `host_change_01_02` | 删除用户 |
| `host_change_01_03` | 修改用户属性 |
| `host_change_01_04` | 新增组 |
| `host_change_01_05` | 删除组 |
| `host_change_01_06` | 修改组属性 |
| `host_change_01_07` | 组新增成员 |
| `host_change_01_08` | 组删除成员 |

- Crontab 变更事件

| 变更 ID | 说明 |
|---------|------|
| `host_change_02_01` | Crontab 任务变更 |

- 文件变更事件

| 变更 ID | 说明 |
|---------|------|
| `host_change_03_01` | 文件内容变更 |

- 服务变更事件

| 变更 ID | 说明 |
|---------|------|
| `host_change_04_01` | 新增服务 |
| `host_change_04_02` | 删除服务 |
| `host_change_04_03` | 修改服务 |
| `host_change_04_04` | 服务状态变更 |

- 网络配置变更事件

| 变更 ID | 说明 |
|---------|------|
| `host_change_05_01` | 网络接口变更 |
| `host_change_05_02` | DNS 配置变更 |
| `host_change_05_03` | 路由配置变更 |
| `host_change_05_04` | 防火墙规则变更 |
| `host_change_05_05` | Hosts 文件变更 |
