---
title     : 'CoreDNS'
summary   : 'Collect CoreDNS metrics and logs'
tags:
  - 'MIDDLEWARE'
__int_icon      : 'icon/coredns'
dashboard :
  - desc  : 'CoreDNS'
    path  : 'dashboard/en/coredns'
monitor   :
  - desc  : 'CoreDNS'
    path  : 'monitor/en/coredns'
---


{{.AvailableArchs}}

---

CoreDNS collector is used to collect metric data related to CoreDNS.

## Configuration {#config}

### Preconditions {#requirements}

- CoreDNS [configuration](https://coredns.io/plugins/metrics/){:target="_blank"}; Enable the `prometheus` plug-in

### Collector Configuration {#input-conifg}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"


    Enable [`kubernetesprometheus`(https://docs.<<<custom_key.brand_main_domain>>>/integrations/kubernetesprometheus/) through DataKit .

    ```yaml
    [inputs.kubernetesprometheus]
      [[inputs.kubernetesprometheus.instances]]
          role       = "pod"
          namespaces = ["kube-system"]
          selector   = "k8s-app=kube-dns"
          port     = "__kubernetes_pod_container_coredns_port_metrics_number"
        [inputs.kubernetesprometheus.instances.custom]
          [inputs.kubernetesprometheus.instances.custom.tags]
            cluster = "demo"
    ```

<!-- markdownlint-enable -->

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
