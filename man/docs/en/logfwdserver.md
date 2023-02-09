
# logfwdserver
---

{{.AvailableArchs}}

---

## Introduction {#intro}

Logfwdserver will turn on the websocket function, which is used together with logfwd, and is responsible for receiving and processing the data sent by logfwd.

See [here](logfwd.md) for the use of logfwd.

## Configuration {#datakit-conf}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).
