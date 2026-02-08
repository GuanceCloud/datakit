# Injecting async-profiler via DataKit Operator

## Prerequisites {#async-profiler-prerequisites}

- [DataKit](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/){:target="_blank"} is installed in the cluster.
- The [profile](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env){:target="_blank"} collector is enabled.
- The Linux kernel parameter [kernel.perf_event_paranoid](https://www.kernel.org/doc/Documentation/sysctl/kernel.txt){:target="_blank"} is set to 2 or lower.

<!-- markdownlint-disable MD046 -->
???+ note

    `async-profiler` uses the [`perf_events`](https://perf.wiki.kernel.org/index.php/Main_Page){:target="_blank"} tool to capture Linux kernel call stacks. Non-privileged processes rely on corresponding kernel settings. You can use the following commands to modify kernel parameters:
    ```shell
    $ sudo sysctl kernel.perf_event_paranoid=1
    $ sudo sysctl kernel.kptr_restrict=0
    # Or
    $ sudo sh -c 'echo 1 >/proc/sys/kernel/perf_event_paranoid'
    $ sudo sh -c 'echo 0 >/proc/sys/kernel/kptr_restrict'
    ```
<!-- markdownlint-enable -->

## Configuration Injection {#annotation-injection}

Add the following annotation under the `.spec.template.metadata.annotations` node in your [Pod Controller](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} resource configuration file, then apply the resource configuration file. DataKit-Operator will automatically create a container named `datakit-profiler` in the corresponding Pod to assist with profiling.

Taking the following Deployment resource configuration file as an example:

```yaml hl_lines="17"
kind: Deployment
metadata:
  name: movies-java
  labels:
    app: movies-java
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-java
  template:
    metadata:
      name: movies-java
      labels:
        app: movies-java
      annotations:
        admission.datakit/java-profiler.version: "{{.K8sProfilersAsyncProfileVersion}}" # <-- add annotation here
    spec:
      containers:
        - name: movies-java
          image: your/app:v1.2.3
          imagePullPolicy: IfNotPresent
          securityContext:
            seccompProfile:
              type: Unconfined
          env:
            - name: JAVA_OPTS
              value: ""

      restartPolicy: Always
```

Apply the configuration file and check if it takes effect:

```shell
$ kubectl apply -f deployment-movies-java.yaml

$ kubectl get pods | grep movies-java
movies-java-784f4bb8c7-59g6s   2/2     Running   0          47s

$ kubectl describe pod movies-java-784f4bb8c7-59g6s | grep datakit-profiler
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler:
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler-volume:
  Normal  Created    12m   kubelet            Created container datakit-profiler
  Normal  Started    12m   kubelet            Started container datakit-profiler
```

Wait a few minutes, and you can view the application performance data on the <<<custom_key.brand_name>>> console [Application Performance Monitoring - Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} page.

<!-- markdownlint-disable MD046 -->
???+ note

    - By default, the command `jps -q -J-XX:+PerfDisableSharedMem | head -n 20` is used to find JVM processes in the container. For performance reasons, data for at most 20 processes will be collected.

    - You can configure profiling behavior by modifying environment variables under `datakit-operator-config` in the `datakit-operator.yaml` configuration file.
    
    | Environment Variable  | Description                                                                                                                                                    | Default Value                 |
    | ----                  | --                                                                                                                                                             | -----                         |
    | `DK_PROFILE_SCHEDULE` | The profiling schedule, using the same syntax as Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"}, e.g., `*/10 * * * *` | `0 * * * *` (Once every hour) |
    | `DK_PROFILE_DURATION` | Duration of each profiling session, in seconds                                                                                                                 | 240 (4 minutes)               |


    - If you cannot see data, you can enter the `datakit-profiler` container to view relevant logs for troubleshooting:

    ```shell
    $ kubectl exec -it movies-java-784f4bb8c7-59g6s -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->
