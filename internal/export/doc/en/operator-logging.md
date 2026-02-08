# DataKit Operator Injecting Log Collection Configuration

DataKit Operator can automatically add the configuration required for DataKit Logging collection to specified Pods, including the `datakit/logs` annotation and the corresponding file path volume/volumeMount, simplifying the tedious manual configuration steps. This way, users can automatically enable log collection functionality without manually intervening in each Pod configuration.

Below is a configuration example demonstrating how to implement automatic injection of log collection configuration through the `admission_mutate` configuration of DataKit Operator:

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        # other configurations
    },
    "admission_mutate": {
        "loggings": [
            {
                "namespace_selectors": ["middleware"],
                "label_selectors":     ["app=logging"],
                "config": "[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/**/*.log\",\"source\":\"logging-tmp\"}]"
            }
        ]
    }
}
```

`admission_mutate.loggings`: This is an array of objects containing multiple log collection configurations. Each log configuration includes the following fields:

- `namespace_selectors`: Limits the Namespaces where the matching Pods reside. Multiple Namespaces can be set, and a Pod must match at least one Namespace to be selected. The relationship with `label_selectors` is "OR".
- `label_selectors`: Limits the labels of matching Pods. A Pod must match at least one label selector to be selected. The relationship with `namespace_selectors` is "OR".
- `config`: This is a JSON string that will be added to the Pod's annotations, with the key `datakit/logs`. If the key already exists, it will not be overwritten or added repeatedly. This configuration tells DataKit how to collect logs.

DataKit Operator will automatically parse the `config` configuration and create the corresponding volume and volumeMount for the Pod based on the path (`path`) within it.

Taking the above DataKit Operator configuration as an example, if a Pod is found with Namespace `middleware` or Labels matching `app=logging`, the annotation and mount will be added to the Pod. For example:

```yaml hl_lines="5"
apiVersion: v1
kind: Pod
metadata:
  annotations:
    datakit/logs: '[{"disable":false,"type":"file","path":"/tmp/opt/**/*.log","source":"logging-tmp"}]'
  labels:
    app: logging
  name: logging-test
  namespace: default
spec:
  containers:
  - args:
    - |
      mkdir -p /tmp/opt/log1;
      i=1;
      while true; do
        echo "Writing logs to file ${i}.log";
        for ((j=1;j<=10000000;j++)); do
          echo "$(date +'%F %H:%M:%S')  [$j]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log1/file_${i}.log;
          sleep 1;
        done;
        echo "Finished writing 5000000 lines to file_${i}.log";
        i=$((i+1));
      done
    command:
    - /bin/bash
    - -c
    - --
    image: pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04
    imagePullPolicy: IfNotPresent
    name: demo
    volumeMounts:
    - mountPath: /tmp/opt
      name: datakit-logs-volume-0
  volumes:
  - emptyDir: {}
    name: datakit-logs-volume-0
```

This Pod has the label `app=logging`, which matches, so DataKit Operator adds the `datakit/logs` annotation to it and adds an EmptyDir mount for the path `/tmp/opt`.

After DataKit log collection discovers the Pod, it will perform customized collection based on the content of `datakit/logs`.
