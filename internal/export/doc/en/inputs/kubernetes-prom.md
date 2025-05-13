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

### Action Steps {#steps}

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
## Automatically Discover the Service Exposure Metrics Interface {#auto-discovery-metrics-with-prometheus}
<!-- markdownlint-enable -->

[:octicons-tag-24: Version-1.5.10](../datakit/changelog.md#cl-1.5.10)

Based on the specified Annotations of Pod or Service, a HTTP URL is constructed and Prometheus metric collection is created.

This feature is disabled by default. To enable it in DataKit, the following two environment variables need to be added as needed, see [container documentation](container.md):

- `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS`: `"true"`
- `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS`: `"true"`

**Note that this feature may generate a large amount of timeline data.**

### Example {#auto-discovery-metrics-with-prometheu-example}

Take adding Annotations in Service as an example. Use the following yaml configuration to create Pod and Service, and add `prometheus.io/scrape` and other Annotations in Service:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: ns-testing
  labels:
    app.kubernetes.io/name: proxy
spec:
  containers:
  - name: nginx
    image: nginx:stable
    ports:
      - containerPort: 80
        name: http-web-svc

---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
  namespace: ns-testing
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "80"
spec:
  selector:
    app.kubernetes.io/name: proxy
  ports:
  - name: name-of-service-port
    protocol: TCP
    port: 8080
    targetPort: http-web-svc
```

DataKit automatically discovers a Service with `prometheus.io/scrape: "true"` and builds a prom collection with `selector` to find a matching Pod:

- `prometheus.io/scrape`: Only services as "true" are collected, required.
- `prometheus.io/port`: Specify the metrics port, required. That this port must be present in the Pod or the collect will fail.
- `prometheus.io/scheme`: Select `https` and `http` according to metrics endpoint, default is `http`.
- `prometheus.io/path`: Configure the metrics path, default to `/metrics`.
- `prometheus.io/param_measurement`：Configure the measurement, default is Pod OwnerReference.

The IP address of the collect target is `PodIP`.

<!-- markdownlint-disable MD046 -->
???+ attention

    DataKit doesn't collects the Service itself, it collects the Pod that the Service is paired with.
<!-- markdownlint-enable -->


The collection interval is 1 minute.

### Measurements and Tags {#measurement-and-tags}

Automatic discovery of Pod/Service Prometheus involves three scenarios for naming metrics, prioritized as follows:

1. Manual configuration of metric sets

    - In Pod/Service Annotations, configure `prometheus.io/param_measurement`, with its value being the specified metric set name. For example:

      ```yaml
      apiVersion: v1
      kind: Pod
      metadata:
        name: testing-prom
        labels:
          app.kubernetes.io/name: MyApp
        annotations:
          prometheus.io/scrape: "true"
          prometheus.io/port: "8080"
          prometheus.io/param_measurement: "pod-measurement"
      ```

      Its Prometheus data metric set would be `pod-measurement`.

    - For Prometheus's PodMonitor/ServiceMonitor CRDs, you can use `params` to specify `measurement`, for example:

      ```yaml
      params:
          measurement:
          - new-measurement
      ```

1. Obtained through data segmentation

    - If the Pod does not have OwnerReferences, the metric name will default to being segmented using an underscore `_`. The first segmented field becomes the metric set name, and the remaining fields become the current metric name.

      For example, consider the following Prometheus raw data:

      ```not-set
      # TYPE promhttp_metric_handler_errors_total counter
      promhttp_metric_handler_errors_total{cause="encoding"} 0
      ```

      Using the first underscore as a delimiter, the left side `promhttp` becomes the metric set name, and the right side `metric_handler_errors_total` becomes the field name.

    - In order to ensure consistency between field names and the original Prom data, the container collector supports the "keep the raw value for prom field names" feature, which can be enabled as follows:

        - In the configuration file: `keep_exist_prometheus_metric_name = true`
        - In the environment variable: `ENV_INPUT_CONTAINER_KEEP_EXIST_PROMETHEUS_METRIC_NAME = "true"`

      Using the `promhttp_metric_handler_errors_total` data as an example, when this feature is enabled, the metric set will be `promhttp`, but the field name will no longer be segmented, and instead will use the raw value `promhttp_metric_handler_errors_total`.

DataKit will add additional tags to locate this resource in the Kubernetes cluster:

- For `Service`, it will add three tags: `namespace`, `service_name`, and `pod_name`.
- For `Pod`, it will add two tags: `namespace` and `pod_name`.

## Extended Reading {#more-readings}

- [Prometheus Exporter Data Collection](prom.md)
