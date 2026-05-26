# DataKit Operator 注入 Python Profiling

## 前置条件 {#prerequisites}

- 当前只支持 Python 官方解释器（CPython）

在你的 [Pod 控制器](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} 资源配置文件中的
`.spec.template.metadata.annotations` 节点下添加如下 annotation，然后应用该资源配置文件，
DataKit-Operator 会自动在相应的 Pod 中创建一个名为 `datakit-profiler` 的容器来辅助进行 profiling。

> **注解使用说明**：关于 `check_annotation` 配置如何影响版本注解的行为，以及各种注解的详细说明，请参考 [Annotation 配置注入](datakit-operator.md#annotation-injection) 和 [`check_annotation` 配置项说明](datakit-operator.md#check-annotation-config)。

接下来将以一个名为 "movies-python" 的 `Deployment` 资源配置文件为例进行说明。

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
        admission.datakit/python-profiler.version: {{.K8sProfilersPySpyVersion}} # <-- add annotation here
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

应用资源配置并验证是否生效：

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

稍等几分钟后即可在<<<custom_key.brand_name>>>控制台 [应用性能检监测-Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    - 默认使用命令 `ps -e -o pid,cmd --no-headers | grep -v grep | grep "python" | head -n 20` 来查找容器中的 `Python` 进程，出于性能考虑，最多只会采集 20 个进程的数据。

    - 可以通过修改 `datakit-operator.yaml` 配置文件中的 ConfigMap `datakit-operator-config`  下的环境变量来配置 profiling 的行为。

    | 环境变量              | 说明                                                                                                                                               | 默认值                        |
    | ----                  | --                                                                                                                                                 | -----                         |
    | `DK_PROFILE_SCHEDULE` | profiling 的运行计划，使用与 Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"} 相同的语法，如 `*/10 * * * *` | `0 * * * *`（每小时调度一次） |
    | `DK_PROFILE_DURATION` | 每次 profiling 持续的时间，单位秒                                                                                                                  | 240（4 分钟）                 |


    - 若无法看到数据，可以进入 `datakit-profiler` 容器查看相应日志进行排查：

    ```shell
    $ kubectl exec -it movies-python-78b6cf55f-ptzxf -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->
