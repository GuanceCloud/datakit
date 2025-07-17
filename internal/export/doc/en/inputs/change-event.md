# Change Event

This document provides the types of object changes supported by the system and their configuration templates, helping users understand and manage the following categories of resource changes:

1. **Kubernetes Resource Objects**: Including changes to core K8s resources such as Pods, Deployments, Services, etc.

The system currently supports the following object change types, each corresponding to specific manifest configuration templates.

{{ if eq .ChangeManifests.K8sManifest nil }}
{{ else }}
## Kubernetes {#k8s}

Current version: {{  .ChangeManifests.K8sManifest.Version }}

{{.ChangeManifests.K8sManifest.MDTable "en" }}

{{ end }}

{{ if eq .ChangeManifests.HostManifest nil }}
{{ else }}
## Host {#host}

Current version: {{  .ChangeManifests.HostManifest.Version }}

{{.ChangeManifests.HostManifest.MDTable "en" }}

{{ end }}
