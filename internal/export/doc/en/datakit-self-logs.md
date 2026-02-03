# Best Practices for DataKit Self-Log Collection

DataKit does not collect its own logs by default, but provides a configuration entry to manually enable it. The following describes two ways to collect DataKit self-logs in different deployment environments.

## Overview {#overview}

DataKit self-logs are divided into two categories:

- **Main logs**: Runtime logs of DataKit, recording collector running status, error information, etc.
- **GIN logs**: Access logs of DataKit HTTP service, recording API request information

By default, the DataKit main log path is `/var/log/datakit/log`, and the GIN log path is `/var/log/datakit/gin.log`.

<!-- markdownlint-disable MD046 -->

???+ warning "Do not enable debug log mode"

    Never enable DataKit debug logs while configuring self-log collection. Enabling debug mode will generate a large amount of logs, which may cause log loop collection and disk space exhaustion.

## Configuration Methods {#config-methods}

=== "Host Deployment"

    In host deployment environments, you need to manually create a logging collector configuration file to collect DataKit self-logs.

    Navigate to the `conf.d/log` directory under the DataKit installation directory, create a `logging-dk-self.conf` file (the filename can be customized), and configure it as follows:

    ```toml
    [[inputs.logging]]
      logfiles = [
        "/var/log/datakit/log",
      ]
    
      # ========== Log Processing Configuration ==========
      source = "datakit"
      service = "datakit"

      # Optional: custom tags
      # [inputs.logging.tags]
      #   env = "production"

    [[inputs.logging]]
      logfiles = [
        "/var/log/datakit/gin.log",
      ]
    
      # ========== Log Processing Configuration ==========
      source = "datakit-gin"
      service = "datakit"
    ```

    After configuration, [restart the DataKit service](datakit-service-how-to.md#manage-service) to make the configuration take effect.

    ???+ tip "Verify Configuration"

        After restarting, you can verify whether the configuration is effective through the `datakit monitor` command:

        ```shell
        datakit monitor -V
        ```

=== "Kubernetes"

    In Kubernetes environments, DataKit configures self-log collection through Pod Annotation `datakit/logs`.

    Edit the `datakit.yaml` file, find the `annotations` section in the Pod Template, and modify the `datakit/logs` Annotation to change `"disable": true` to `"disable": false` to enable log collection:

    ```yaml
    apiVersion: apps/v1
    kind: DaemonSet
    metadata:
      name: datakit
      namespace: datakit
    spec:
      template:
        metadata:
          annotations:
            datakit/logs: |
              [
                {
                  "disable":       false,
                  "type":          "file",
                  "path":          "/var/log/datakit/log",
                  "source":        "datakit",
                  "service":       "datakit",
                  "storage_index": ""
                },
                {
                  "disable":       false,
                  "type":          "file",
                  "path":          "/var/log/datakit/gin.log",
                  "source":        "datakit-gin",
                  "service":       "datakit",
                  "storage_index": ""
                }
              ]
    ```

    After modification, apply the configuration:

    ```shell
    kubectl apply -f datakit.yaml
    ```

    DataKit Pods will automatically restart to apply the new configuration.


## Configuration Details {#config-details}

For detailed configuration field descriptions and advanced configuration options, see the [Log Collection Configuration Documentation](inputs/datakit-logging.md#config).

## Notes {#notes}

### Log Loop Collection Risk {#log-loop-risk}

???+ warning "Avoid Log Loop Collection"

    Improper configuration may cause log loop collection:

    - DataKit collects its own logs → writes to log files → DataKit collects again → infinite loop

    To avoid this problem:

    1. **Do not enable debug mode**: Debug mode will generate a large amount of logs
    1. **Configure log level appropriately**: Set `[logging] level = "info"` in `datakit.conf`
    1. **Monitor log volume**: Regularly check log collection volume, and investigate promptly if abnormal growth is found

