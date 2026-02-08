# DataKit Operator 注入 async-profiler

## 前置条件 {#async-profiler-prerequisites}

- 集群已安装 [DataKit](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/){:target="_blank"}。
- [开启 profile](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env){:target="_blank"} 采集器。
- Linux 内核参数 [kernel.perf_event_paranoid](https://www.kernel.org/doc/Documentation/sysctl/kernel.txt){:target="_blank"} 值设置为 2 及以下。

<!-- markdownlint-disable MD046 -->
???+ note

    `async-profiler` 使用 [`perf_events`](https://perf.wiki.kernel.org/index.php/Main_Page){:target="_blank"} 工具来抓取 Linux 的内核调用堆栈，非特权进程依赖内核的相应设置，可以使用以下命令来修改内核参数：
    ```shell
    $ sudo sysctl kernel.perf_event_paranoid=1
    $ sudo sysctl kernel.kptr_restrict=0
    # 或者
    $ sudo sh -c 'echo 1 >/proc/sys/kernel/perf_event_paranoid'
    $ sudo sh -c 'echo 0 >/proc/sys/kernel/kptr_restrict'
    ```
<!-- markdownlint-enable -->

## 配置注入 {#annotation-injection}

在你的 [Pod 控制器](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} 资源配置文件中的
`.spec.template.metadata.annotations` 节点下添加如下 annotation，然后应用该资源配置文件，
DataKit-Operator 会自动在相应的 Pod 中创建一个名为 `datakit-profiler` 的容器来辅助进行 profiling。

以如下 Deployment 资源配置文件为例进行说明：

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

应用配置文件并检查是否生效：

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

稍等几分钟后即可在<<<custom_key.brand_name>>>控制台 [应用性能检监测-Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    - 默认使用命令 `jps -q -J-XX:+PerfDisableSharedMem | head -n 20` 来查找容器中的 JVM 进程，出于性能的考虑，最多只会采集 20 个进程的数据。

    - 可以通过修改 `datakit-operator.yaml` 配置文件中的 `datakit-operator-config` 下的环境变量来配置 profiling 的行为。
    
    | 环境变量              | 说明                                                                                                                                               | 默认值                        |
    | ----                  | --                                                                                                                                                 | -----                         |
    | `DK_PROFILE_SCHEDULE` | profiling 的运行计划，使用与 Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"} 相同的语法，如 `*/10 * * * *` | `0 * * * *`（每小时调度一次） |
    | `DK_PROFILE_DURATION` | 每次 profiling 持续的时间，单位秒                                                                                                                  | 240（4 分钟）                 |


    - 若无法看到数据，可以进入 `datakit-profiler` 容器查看相应日志进行排查：

    ```shell
    $ kubectl exec -it movies-java-784f4bb8c7-59g6s -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->

