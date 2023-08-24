# DataKit Metrics

---

{{.AvailableArchs}} · [:octicons-tag-24: Version-1.10.0](../datakit/changelog.md#cl-1.10.0)

---

This Input used to collect Datakit exported metrics, such as runtime/CPU/memory and various other metrics of each modules.

## Configuration {#config}

After Datakit startup, it will expose a lot of [Prometheus metrics](datakit-metrics.md), and the input `dk` can scrap
these metrics.

<!-- markdownlint-disable MD046 -->
=== "*dk.conf*"

    
    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit]().

=== "Kubernetes"

    Kubernetes supports modifying configuration parameters in the form of environment variables:

    | Environment Name                  | Description                                                            | Examples                                                                               |
    | :---                              | ---                                                                    | ---                                                                                    |
    | `ENV_INPUT_DK_ENABLE_ALL_METRICS` | Enable all metrics, this may collect more than 300+ metrics on Datakit | `on/yes/`                                                                              |
    | `ENV_INPUT_DK_ADD_METRICS`        | Add extra metrics (JSON array)                                         | `["datakit_io_.*", "datakit_pipeline_.*"]`, Available metrics list [here](../datakit/datakit-metrics.md) |
    | `ENV_INPUT_DK_ONLY_METRICS`       | **Only** enalbe specified metrics(JSON array)                          | `["datakit_io_.*", "datakit_pipeline_.*"]`                                             |
<!-- markdownlint-enable -->

## Measurements {#metric}

Datakit exported Prometheus metrics, see [here](../datakit/datakit-metrics.md) for full metric list.
