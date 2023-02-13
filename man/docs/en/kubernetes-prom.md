
# Prometheus Exportor Metric Collection
---

:fontawesome-brands-linux: :material-kubernetes:

---

## Introduction {#intro}

This document describes how to capture Prometheus metrics exposed by custom Pods in Kubernetes clusters in two ways:

- Expose the pointer interface to the DataKit through Annotations
- Expose the metric interface to the DataKit by automatically discovering Kubernetes endpoint services to prometheus

The usage of the two methods will be explained in detail below.

## Open Metrics Interface with Annotations {#annotations-of-prometheus}

You need to add specific template annotations to the Kubernetes deployment to capture the metrics exposed by the Pod it creates. Annotations requires the following:

- Key is fixed `datakit/prom.instances`
- Value is the full configuration of [prom collector](prom.md), for example:

```toml
[[inputs.prom]]
  ## Exporter 地址
  url = "http://$IP:9100/metrics"

  source = "<your-service-name>"
  metric_types = ["counter", "gauge"]

  measurement_name = "prom"
  # metric_name_filter = ["cpu"]
  # measurement_prefix = ""
  #tags_ignore = ["xxxx"]

  interval = "10s"

  #[[inputs.prom.measurements]]
  # prefix = "cpu_"
  # name = "cpu"

  [inputs.prom.tags]
    # namespace = "$NAMESPACE"
    # pod_name = "$PODNAME"
    # node_name = "$NODENAME"
```

The following wildcard characters are supported:

- `$IP`: Intranet IP of the Pod
- `$NAMESPACE`: Pod Namespace
- `$PODNAME`: Pod Name
- `$NODENAME`: The name of the Node where the Pod is located

!!! tip

    Instead of automatically adding tags such as `namespace` and `pod_name`, the Prom collector can add additional tags using wildcards in the config above, for example:

    ``` toml
      [inputs.prom.tags]
        namespace = "$NAMESPACE"
        pod_name = "$PODNAME"
        node_name = "$NODENAME"
    ```

### Select Specified Pod IP {#pod-ip}

In some cases, there will be multiple IPs on the Pod, and it is inaccurate to get the Exporter address only by `$IP`. Selecting Pod IP by configuring Annotations is supported.

- Key is fixed `datakit/prom.instances.ip_index`
- Value is a natural number, such as `0` `1` `2` and so on, which is the subscript of the IP to be used in the entire IP array (Pod IPs).

If this Annotations Key is not available, the default Pod IP is used.

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
        datakit/prom.instances.ip_index: 2
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:9100/metrics"
          
            source = "<your-service-name>"
            metric_types = ["counter", "gauge"]
            # metric_name_filter = ["cpu"]
            # measurement_prefix = ""
            # measurement_name = "prom"
          
            interval = "10s"
          
            # tags_ignore = ["xxxx"]
          
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"
          
            [inputs.prom.tags] # 视情况开启下面的 Tags
            #namespace = "$NAMESPACE"
            #pod_name = "$PODNAME"
            #node_name = "$NODENAME"
```

???+ attention

    The `annotations` must be added under the `template` field so that the Pod created by *deployment.yaml* carries `datakit/prom.instances`.


- Create a resource with the new yaml

```shell
kubectl apply -f deployment.yaml
```

At this point, Annotations has been added. DataKit later reads the Pod's Annotations and collects the metrics exposed on `url`.

## Automatically Discover the Service Exposure Metrics Interface {#auto-discovery-of-service-prometheus}

The Service needs to be bound to the Pod, and the Service adds the specified Annotations, which Datakit automatically discovers and accesses to obtain the prometheus metric.

For example, use the following yaml configuration to create a Pod and a Service, and add Annotations such as `prometheus.io/scrape` to the Service:

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
    prometheus.io/port: "8080"
spec:
  selector:
    app.kubernetes.io/name: proxy
  ports:
  - name: name-of-service-port
    protocol: TCP
    port: 8080
    targetPort: http-web-svc
```

Datakit automatically discovers a Service with `prometheus.io/scrape: "true"` and builds a prom collection based on its additional configuration items:

- `prometheus.io/scrape`: Only services as "true" are collected, required
- `prometheus.io/port`: Specify the metrics port, required
- `prometheus.io/scheme`: Select `https` and `http` according to metrics endpoint, default is `http`
- `prometheus.io/path`: Configure the metrics path, default to `/metrics`

Eventually Datakit accesses `http://nginx-service.ns-testing:8080/metrics` to collect prometheus metrics, taking the Service yaml configuration above as an example.

The collection interval is 1 minute.

### Measurements and Tags {#measurement-and-tags}

Automatically discover Service prometheus, whose measurement name is parsed by Datakit. By default, the metric name will be cut with an underscore `_`. The first field after cutting will be the measurement name, and the remaining fields will be the current metric name.

For example, the following prometheus raw data:

```
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
```

Distinguished by the first underscore, `promhttp` on the left is the metric set name, and `metric_handler_errors_total` on the right is the field name.

In addition, Datakit adds two tags, `service` and `namespace`, whose values are the Service name and the Service's Namespace, to locate the Service in the Kubernetes cluster.

## Extended Reading {#more-readings}

- [Prometheus Exportor Data Collection](prom.md)
