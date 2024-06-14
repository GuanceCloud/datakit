---
title     : 'Prometheus Push Gateway'
summary   : 'Enable Pushgateway API to receive Prometheus metric data'
__int_icon      : 'icon/pushgateway'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Pushgateway
<!-- markdownlint-enable -->
---

{{.AvailableArchs}} · [:octicons-tag-24: Version-1.31.0](../datakit/changelog.md#cl-1.31.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

---

The Pushgateway collector will open the corresponding API interface to receive Prometheus metric data.

## Configuration  {#config}

<!-- markdownlint-disable MD046 -->

=== "Host Deployment"

    Navigate to the `conf.d/{{.Catalog}}` directory in the DataKit installation directory, copy `{{.InputName}}.conf.sample` and rename it to `{{.InputName}}.conf`. The example is as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuring, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    You can enable the collector configuration through the [ConfigMap method injection](../datakit/datakit-daemonset-deploy.md#configmap-setting) or by setting [`ENV_DATAKIT_INPUTS`](../datakit/datakit-daemonset-deploy.md#env-setting).

    It also supports modifying configuration parameters via environment variables (needs to be added as a default collector in `ENV_DEFAULT_ENABLED_INPUTS`):

{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

---

## Example {#example}

The Pushgateway collector follows the [Prometheus Pushgateway](https://github.com/prometheus/pushgateway?tab=readme-ov-file#prometheus-pushgateway) protocol, with some adjustments for DataKit's collection features. Currently, it supports the following functions:

- Receiving Prometheus text data and Protobuf data
- Specifying string labels and base64 labels in the URL
- Decoding gzip data
- Specifying metric set names

Below is a simple example deployed in a Kubernetes cluster:

- Enable the Pushgateway collector. Here, it's enabled as an environment variable in the Datakit YAML.

```yaml
    # ..other..
    spec:
      containers:
      - name: datakit
        env:
        - name: ENV_DEFAULT_ENABLED_INPUTS
          value: dk,cpu,container,pushgateway  # Add pushgateway to enable the collector
        - name: ENV_INPUT_PUSHGATEWAY_ROUTE_PREFIX
          value: /v1/pushgateway               # Optional, specify endpoints route prefix, the target route will become "/v1/pushgateway/metrics"
    # ..other..
```

- Create a Deployment that generates Prometheus data and sends it to the Datakit Pushgateway API.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pushgateway-client
  labels:
    app: pushgateway-client
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pushgateway-client
  template:
    metadata:
      labels:
        app: pushgateway-client
    spec:
      containers:
      - name: client
        image: pubrepo.guance.com/base/curl
        imagePullPolicy: IfNotPresent
        env:
        - name: MY_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: PUSHGATEWAY_ENDPOINT
          value: http://datakit-service.datakit.svc:9529/v1/pushgateway/metrics/job@base64/aGVsbG8=/node/$(MY_NODE_NAME)/pod/$(MY_POD_NAME)/namespace/$(MY_POD_NAMESPACE)
          ## job@base64 specifies that the format is base64, use the command `echo -n hello | base64` to generate the value 'aGVsbG8='
        args:
        - /bin/bash
        - -c
        - >
          i=100;
          while true;
          do
            ## Periodically send data to the Datakit Pushgateway API using the cURL command
            echo -e "# TYPE pushgateway_count counter\npushgateway_count{name=\"client\"} $i" | curl --data-binary @- $PUSHGATEWAY_ENDPOINT;
            i=$((i+1));
            sleep 2;
          done
```

- The metric set seen on the Observability Cloud page is `pushgateway`, with the field being `count`.

## Metric Sets and Tags {#measurement-and-tags}

The Pushgateway collector does not add any tags.

There are two cases for naming metric sets:

1. Use the configuration option `measurement_name` to specify the metric set name.
1. Split the data field names using an underscore `_`, where the first field after splitting becomes the metric set name, and the remaining fields become the current metric name.
