---
title     : 'W32Time'
summary   : '采集 Windows Time Service 指标'
tags:
  - 'WINDOWS'
  - 'TIME'
__int_icon: 'icon/windows'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

W32Time 采集器通过 Windows Performance Data Helper 计数器监控本机 Windows Time Service，不执行或解析 `w32tm` 输出。

## 配置 {#config}

### 前置条件 {#requirements}

- Windows Server 2016 及以上版本，或 Windows 10 周年更新及以上版本、Windows 11
- Windows 64 位 DataKit（`windows/amd64`）
- 系统已注册 Windows Time（`W32Time`）服务，无需额外安装 `w32tm` 或其他程序

当 `W32Time` 服务未运行时，采集器仍会上报 `service_running=0`，其他时间同步指标只有在服务运行且对应性能计数器可用时才会上报。

### 采集器配置 {#input-config}

进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`，配置如下：

```toml
{{.InputSample}}
```

配置完成后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service)。

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
