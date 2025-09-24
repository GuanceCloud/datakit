---
title     : 'Kubernetes Prometheus Exporter'
summary   : 'Collect Prometheus metrics among Kubernetes Pod'
tags      :
  - 'PROMETHEUS'
  - 'KUBERNETES'
__int_icon: 'icon/kubernetes'
---

:fontawesome-brands-linux: :material-kubernetes:

---

## Introduction {#intro}

**Deprecated, related functionality moved to [KubernetesPrometheus Collector](kubernetesprometheus.md).**

This document describes how to capture Prometheus metrics exposed by custom Pods in Kubernetes clusters in two ways:

- Expose the pointer interface to the DataKit through Annotations
- Expose the metric interface to the DataKit by automatically discovering Kubernetes endpoint services to Prometheus

The usage of the two methods will be explained in detail below.

## Open Metrics Interface with Annotations {#annotations-of-prometheus}

You need to add specific template annotations to the Kubernetes deployment to capture the metrics exposed by the Pod it creates. Annotations requires the following:

- Key is fixed `datakit/prom.instances`
- Value is the full configuration of [prom collector](prom.md), for example:

```toml
[[inputs.prom]]
  urls   = ["http://$IP:9100/metrics"]
  source = "<your-service-name>"
  measurement_name = "<measurement-metrics>"
  interval = "30s"

  [inputs.prom.tags]
    # namespace = "$NAMESPACE"
    # pod_name  = "$PODNAME"
    # node_name = "$NODENAME"
```

The following wildcard characters are supported:

- `$IP`: Intranet IP of the Pod
- `$NAMESPACE`: Pod Namespace
- `$PODNAME`: Pod Name
- `$NODENAME`: The name of the Node where the Pod is located

<!-- markdownlint-disable MD046 -->
!!! tip

    Instead of automatically adding tags such as `namespace` and `pod_name`, the Prom collector can add additional tags using wildcards in the config above, for example:

    ``` toml
      [inputs.prom.tags]
        namespace = "$NAMESPACE"
        pod_name = "$PODNAME"
        node_name = "$NODENAME"
    ```
<!-- markdownlint-enable -->

### Action Steps {#annotations-of-prometheus-steps}

- Log on to Kubernetes' host
- Open `deployment.yaml` and add the template annotations example as follows:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prom-deployment
  labels:
    app: prom
spec:
  template:
    metadata:
      labels:
        app: prom
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            urls   = ["http://$IP:9100/metrics"]
            source = "<your-service-name>"
            interval = "30s"
            [inputs.prom.tags]
              namespace = "$NAMESPACE"
              pod_name  = "$PODNAME"
              node_name = "$NODENAME"
```

<!-- markdownlint-disable MD046 -->
???+ note

    The `annotations` must be added under the `template` field so that the Pod created by *deployment.yaml* carries `datakit/prom.instances`.
<!-- markdownlint-enable -->


- Create a resource with the new yaml

```shell
kubectl apply -f deployment.yaml
```

At this point, Annotations has been added. DataKit later reads the Pod's Annotations and collects the metrics exposed on `url`.

<!-- markdownlint-disable MD013 -->
## Auto-discovery of Prometheus Metrics from Pods/Services {#auto-discovery-metrics-with-prometheus}
<!-- markdownlint-enable -->

**Note: The complete documentation and latest configuration for this feature have been moved to [KubernetesPrometheus Collector - "Auto-discovery of Prometheus Metrics via Declarative Annotations"](kubernetesprometheus.md). This document retains only the basic environment variable configuration instructions. It is recommended to refer to the new document for detailed configuration examples and best practices.**

### Overview {#auto-discovery-metrics-overview}

This feature automatically discovers specific annotations on Kubernetes Pods or Services and dynamically generates collection configurations for Prometheus metrics based on the annotation content. When a Pod or Service is annotated with predefined labels, DataKit automatically constructs an HTTP URL and creates a corresponding Prometheus metrics collection task, eliminating the need for manual collector configuration changes.

### Enablement {#auto-discovery-metrics-enablement}

This feature is disabled by default. It must be enabled in DataKit by setting the following environment variables, which act as global switches controlling the enablement status of the auto-discovery functionality:

- **`ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS`**: Set to `"true"` to enable auto-discovery based on Pod annotations.
- **`ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS`**: Set to `"true"` to enable auto-discovery based on Service annotations.

For detailed environment variable configuration, please refer to the [container documentation](container.md#config-using-env).

### Key Features {#auto-discovery-metrics-features}

- **Automatic Discovery**: No need to restart DataKit; automatically detects newly created or updated Pods/Services.
- **Dynamic Configuration**: Dynamically generates collection configurations based on annotation content, offering flexibility for different applications.
- **Resource Filtering**: Supports precise control over collection targets through annotation values.
- **Configuration Inheritance**: Pod-level configurations have higher priority than Service-level configurations.

### Compatibility Note {#auto-discovery-metrics-note}

The original method of enabling via environment variables remains fully compatible; existing configurations require no modification. However, future feature enhancements and configuration options will be updated in the [KubernetesPrometheus Collector documentation](kubernetesprometheus.md). Users are advised to migrate to the new configuration method for more comprehensive functionality.

## Extended Reading {#more-readings}

- [Prometheus Exporter Data Collection](prom.md)
