# Injecting Python Profiling via DataKit Operator

## Prerequisites {#prerequisites}

- Currently, only the official Python interpreter (CPython) is supported.

Add the following annotation under the `.spec.template.metadata.annotations` node in your [Pod Controller](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} resource configuration file, then apply the resource configuration file. DataKit-Operator will automatically create a container named `datakit-profiler` in the corresponding Pod to assist with profiling.

> **Annotation Usage Instructions**: For how `check_annotation` configuration affects version annotation behavior, and detailed explanations of various annotations, please refer to [Annotation Configuration Injection](datakit-operator.md#annotation-injection) and [`check_annotation` Configuration Item Explanation](datakit-operator.md#check-annotation-config).

The following uses a `Deployment` resource configuration file named "movies-python" as an example.

```yaml hl_lines="17"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: movies-python
  labels:
    app: movies-python
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-python
  template:
    metadata:
      name: movies-python
      labels:
        app: movies-python
      annotations:
        admission.datakit/python-profiler.version: "{{.K8sProfilersPySpyVersion}}" # <-- add annotation here
    spec:
      containers:
        - name: movies-python
          image: zhangyicloud/movies-python:1.2.3
          imagePullPolicy: Always
          command:
            - "gunicorn"
            - "-w"
            - "4"
            - "--bind"
            - "0.0.0.0:8080"
            - "app:app"
```

Apply the resource configuration and verify if it takes effect:

```shell
$ kubectl apply -f deployment-movies-python.yaml

$ kubectl get pods | grep movies-python
movies-python-78b6cf55f-ptzxf   2/2     Running   0          64s
 
$ kubectl describe pod movies-python-78b6cf55f-ptzxf | grep datakit-profiler
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler:
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler-volume:
  Normal  Created    98s   kubelet            Created container datakit-profiler
  Normal  Started    97s   kubelet            Started container datakit-profiler
```

Wait a few minutes, and you can view the application performance data on the <<<custom_key.brand_name>>> console [Application Performance Monitoring - Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} page.

<!-- markdownlint-disable MD046 -->
???+ note

    - By default, the command `ps -e -o pid,cmd --no-headers | grep -v grep | grep "python" | head -n 20` is used to find `Python` processes in the container. For performance reasons, data for at most 20 processes will be collected.

    - You can configure profiling behavior by modifying environment variables under the ConfigMap `datakit-operator-config` in the `datakit-operator.yaml` configuration file.

    | Environment Variable  | Description                                                                                                                                                    | Default Value                 |
    | ----                  | --                                                                                                                                                             | -----                         |
    | `DK_PROFILE_SCHEDULE` | The profiling schedule, using the same syntax as Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"}, e.g., `*/10 * * * *` | `0 * * * *` (Once every hour) |
    | `DK_PROFILE_DURATION` | Duration of each profiling session, in seconds                                                                                                                 | 240 (4 minutes)               |


    - If you cannot see data, you can enter the `datakit-profiler` container to view relevant logs for troubleshooting:

    ```shell
    $ kubectl exec -it movies-python-78b6cf55f-ptzxf -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->
