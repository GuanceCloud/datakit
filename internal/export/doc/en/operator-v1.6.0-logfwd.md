# Logfwd Usage in Operator < 1.6.0

This configuration method is used in DataKit-Operator v1.6.0 and earlier. Version 1.7.0 adopted a new CRD configuration method, and the CRD + Annotation hybrid scheme introduced in v1.6.0 has been deprecated.

1. [Download and install DataKit-Operator](datakit-operator.md#install) in the target Kubernetes cluster.
1. Add the specified Annotation to the deployment to indicate the need to mount the logfwd sidecar. Note that the Annotation should be added to the template.
    - The key is uniformly `admission.datakit/logfwd.instances`.
    - The value is a JSON string, representing the specific logfwd configuration. An example is as follows:

```json
[
    {
        "datakit_addr": "datakit-service.datakit.svc:9533",
        "loggings": [
            {
                "logfiles":      ["<your-logfile-path>"],
                "ignore":        [],
                "storage_index": "<your-storage-index>",
                "source":        "<your-source>",
                "service":       "<your-service>",
                "pipeline":      "<your-pipeline.p>",
                "character_encoding": "",
                "multiline_match": "<your-match>",
                "tags": {}
            },
            {
                "logfiles": ["<your-logfile-path-2>"],
                "source": "<your-source-2>"
            }
        ]
    }
]
```

Parameter description, refer to [logfwd configuration](../integrations/logfwd.md#config):

- `datakit_addr` is the DataKit logfwdserver address.
- `loggings` is the main configuration, which is an array. Refer to [DataKit logging collector](../integrations/logging.md).
    - `logfiles`: List of log files. Absolute paths can be specified, and glob rules are supported for batch specification. Using absolute paths is recommended.
    - `ignore`: File path filtering using glob rules. Files matching any filtering condition will not be collected.
    - `storage_index`: Specifies the log storage index.
    - `source`: Data source. If empty, defaults to 'default'.
    - `service`: Adds a tag. If empty, defaults to $source.
    - `pipeline`: Pipeline script path. If empty, uses $source.p. If $source.p does not exist, no Pipeline will be used (this script file exists on the DataKit side).
    - `character_encoding`: Select encoding. Incorrect encoding will cause data unreadable. Default is empty. Supports `utf-8/utf-16le/utf-16le/gbk/gb18030`.
    - `multiline_match`: Multiline matching. See [DataKit log multiline configuration](../integrations/logging.md#multiline). Note that since it is JSON format, the triple single quote "non-escaping syntax" is not supported. Regex `^\d{4}` needs to be escaped as `^\\d{4}`.
    - `tags`: Add extra `tag`. Format is JSON map, e.g., `{ "key1":"value1", "key2":"value2" }`.

<!-- markdownlint-disable MD046 -->
???+ note

    When injecting logfwd, DataKit Operator defaults to reusing volumes with the same path to avoid injection errors due to duplicate volume paths.

    Paths ending with a slash and without a slash have different meanings. For example, `/var/log` and `/var/log/` are different paths and cannot be reused.
<!-- markdownlint-enable -->

## Example Case {#datakit-operator-inject-logfwd-example}

Below is a Deployment example that uses shell to continuously write data to a file and configures the collection of that file:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    name: logging-deployment
    labels:
    app: logging
spec:
    replicas: 1
    selector:
    matchLabels:
        app: logging
    template:
    metadata:
        labels:
        app: logging
        annotations:
        admission.datakit/logfwd.instances: '[{"datakit_addr":"datakit-service.datakit.svc:9533","loggings":[{"logfiles":["/var/log/log-test/*.log"],"source":"deployment-logging","tags":{"key01":"value01"}}]}]'
    spec:
        containers:
        - name: log-container
        image: busybox
        args: [/bin/sh, -c, 'mkdir -p /var/log/log-test; i=0; while true; do printf "$(date "+%F %H:%M:%S") [%-8d] Bash For Loop Examples.\\n" $i >> /var/log/log-test/1.log; i=$((i+1)); sleep 1; done']
```

Create resources using the yaml file:

```shell
$ kubectl apply -f logging.yaml
...
```

Verify as follows:

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
logging-deployment-5d48bf9995-vt6bb    1/1     Running   0             4s

$ kubectl get pod logging-deployment-5d48bf9995-vt6bb -o=jsonpath={.spec.containers\[\*\].name}
log-container datakit-logfwd
```

Finally, you can check on the <<<custom_key.brand_name>>> log platform whether logs are being collected.
