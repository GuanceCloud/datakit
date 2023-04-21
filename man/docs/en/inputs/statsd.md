
# Statsd Data Access
---

{{.AvailableArchs}}

---

The statsd collector is used to receive statsd data sent over the network.

## Preconditions {#requrements}

None

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    The collector can now be turned on by [configMap injection collector configuration](datakit-daemonset-deploy.md#configmap-setting).

## Measurement {#measurement}

Statsd has no measurement definition at present, and all metrics are subject to the metrics sent by the network.
