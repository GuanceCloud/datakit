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

This page retains legacy `datakit/prom.instances` configuration examples and compatibility instructions for auto-discovery environment variables.

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
<!-- markdownlint-enable MD046 -->

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
<!-- markdownlint-enable MD046 -->


- Create a resource with the new yaml

```shell
kubectl apply -f deployment.yaml
```

At this point, Annotations has been added. DataKit later reads the Pod's Annotations and collects the metrics exposed on `url`.

<!-- markdownlint-disable MD013 -->
## Auto-discovery of Prometheus Metrics from Pods/Services {#auto-discovery-metrics-with-prometheus}
<!-- markdownlint-enable MD013 -->

For configuration instructions and annotation parameters, see the [KubernetesPrometheus collector](kubernetesprometheus.md#auto-discovery-metrics-with-prometheus). The original environment variables remain supported, so existing configurations need no changes. The following options are disabled by default:

- **`ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS`**: Set to `"true"` to enable auto-discovery based on Pod annotations.
- **`ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS`**: Set to `"true"` to enable auto-discovery based on Service annotations.

For detailed environment variable configuration, please refer to the [container documentation](container.md#config-using-env).

## Extended Reading {#more-readings}

- [Prometheus Exporter Data Collection](prom.md)
