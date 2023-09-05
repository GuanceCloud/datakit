
# External
---

{{.AvailableArchs}}

---

External Collector launchs outside program to collecting data.

## Configuration {#config}

### Preconditions {#requirements}

- The command and its running environment must have complete dependencies. For example, if Python is used to start an external Python script, the `import` package and other dependencies required for the script to run must be prepared.

### Input configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
