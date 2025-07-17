# 变更事件

本文档提供了系统支持的对象变更类型及其配置模板，帮助用户了解和管理以下几类资源变更：

1. **Kubernetes 资源对象**：包括 Pod、Deployment、Service 等 K8s 核心资源的变更

当前系统支持以下对象变更类型，每种变更都对应特定的 manifest 配置模板。

{{ if eq .ChangeManifests.K8sManifest nil }}
{{ else }}
## Kubernetes {#k8s}

当前版本： {{  .ChangeManifests.K8sManifest.Version }}

{{.ChangeManifests.K8sManifest.MDTable "zh" }}

{{ end }}

{{ if eq .ChangeManifests.HostManifest nil }}
{{ else }}
## 主机 {#host}

当前版本： {{  .ChangeManifests.HostManifest.Version }}

{{.ChangeManifests.HostManifest.MDTable "zh" }}

{{ end }}
