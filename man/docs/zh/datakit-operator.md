{{.CSS}}
# Datakit-Operator 使用说明

---

:material-kubernetes:

---

## 概述和安装 {#datakit-operator-install}

Datakit-Operator 是 Datakit 在 Kubernetes 编排的联动项目，旨在协助 Datakit 更方便的部署，以及其他诸如验证、注入的功能。

目前 Datakit-Operator 提供以下功能：

- [x] 针对特殊 Pod，提供注入 `dd-lib` 文件和 environment 的功能，[文档](#datakit-operator-admission-mutate)
- [ ] 负责创建和更新 Datakit 即相关 Pod 的编排
- [ ] 验证 Datakit 的配置

推荐 Kubernetes v1.24.1 及以上版本，且能够访问互联网（下载 yaml 文件和 Image）。

```
$ kubectl create namespace datakit
$ wget https://static.guance.com/datakit-operator/datakit-operator.yaml
$ kubectl apply -f datakit-operator.yaml
$ kubectl get pod -n datakit
NAME                               READY   STATUS    RESTARTS   AGE
datakit-operator-f948897fb-5w5nm   1/1     Running   0          15s
```

*如果出现 `InvalidImageName` 报错，可以手动 pull 镜像。*

## 使用 Operator 在 Pod 中注入 lib 文件和相关的环境变量 {#datakit-operator-inject-lib}

场景是：

用户在 k8s 环境部署了很多 Pod（Deployment/DaemonSet 同理），想开启 trace 采集发给 Datakit，但是 Pod 缺少所需的 dd-lib 文件和环境变量。

是否有简单的办法，能批量在 Pod 中添加 dd-lib 文件？

---

首先，根据 Kubenetes Admission Controller 的机制，这是可行的。

**注意，这是侵入式行为，要修改用户原来的 yaml 文件，把所需数据注入进去，不是所有人都愿意自己的 yaml 被修改。**

具体做法是：

1. 用户在自己的 k8s 集群，下载和安装 datakit-operator.yaml
2. 在所有需要添加 dd-lib 文件的 Pod 添加指定 Annotation
3. 同时，在该 Pod 添加环境变量 `DD_AGENT_HOST` 和 `DD_TRACE_AGENT_PORT` 指定接受地址。以 JAVA 为例，详见 [dd-trace JAVA 启动文档](https://docs.guance.com/datakit/ddtrace-java/#start-options)

datakit-operator 运行后会根据 Annotation 决定是否添加 dd-lib 文件和环境变量。

> 指定的 Annotation Key 是 `admission.datakit/%s-lib.version`，%s 需要替换成指定的语言，目前支持 `java`、`python` 和 `js`

>  Value 是指定版本号，如果为空，就使用默认的稳定版，现在支持的版本有限，推荐使用默认值。

Datakit-Operator v1.0.1 的 dd-lib stable 版本号：

- java: v1.0.1-guance
- python: v1.6.2
- js: v3.9.3

1. [安装 Datkait-Operator](#datakit-operator-install)

2. 修改现有的应用 yaml，以 nginx deployment 为例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
      annotations:
        admission.datakit/js-lib.version: ""
    spec:
      containers:
      - name: nginx
        image: nginx:1.22
        ports:
        - containerPort: 80
        env:
        - name: DD_AGENT_HOST
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        - name: DD_TRACE_AGENT_PORT
          value: 9529
```

其中有 3 项配置需要手动添加：

- 添加 Annotation 的 `admission.datakit/js-lib.version: ""`
- 添加环境变量 `DD_AGENT_HOST`，value 使用了 hostIP
- 添加环境变量 `DD_TRACE_AGENT_PORT`，value 是 Datakit 的端口 9529

使用 yaml 文件创建资源：

```shell
$ kubectl apply -f nginx.yaml
```

此时，datakit-operator 已经为 nginx deployment 的所有 Pod 添加了 java-lib 文件，验证如下：

```shell
$ kubectl get pod
$ NAME                                   READY   STATUS    RESTARTS      AGE
nginx-deployment-7bd8dd85f-fzmt2       1/1     Running   0             4s
$ kubectl get pod nginx-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.initContainers\[0\].name}
$ datakit-lib-init
```

这个名为 `datakit-lib-init` 的 initContainers 就是 Datakit-Operator 添加的，该容器内有 java-lib 文件，并且和主应用容器互通 volume，使得主容器可以访问到该文件。

----

简述相关原理：

[Admission Controller（准入控制器）](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/)是 k8s 的一项功能，它会在请求通过认证和鉴权之后、对象被持久化之前拦截到达 API 服务器的请求。

1. Datakit-Operator 利用这个特性，向 k8s 注册一个 admission mutate，允许自身访问和修改所有 Pod（在这个 Pod 创建时）

2. 当有一个新的 Pod 被创建（CREATE），k8s 中心会将该 Pod 的配置发给 Datakit-Operator，Datakit-Operator 扫描其中的 Annotations 并找到 `admission.datakit/js-lib-verion`，如果找不到，就保持原样返回

3. 对于符合条件的 Pod，Datakit-Operator 会在其配置中，添加一个 `initContainers`，image 是根据 `admission.datakit/js-lib-verion` 转化所得，这个 image 在 `/datadog-lib` 路径下存有一个 lib 文件

4. Datakit-Operator 会将 initContainers 和其他容器的 `/datadog-lib` 路径打通，使得其他容器能访问到该路径的文件

5. Datakit-Operator 会给所有容器添加一个特殊的环境变量，例如 `NODE_OPTIONS` = `--require=/datadog-lib/node_modules/dd-trace/init`

6. 最后，Datakit-Operator 将这份修改过的 Pod 配置，发回 k8s 中心，k8s 会用它创建 Pod，此时这个运行的 Pod 已经被添加 lib 文件

补充，对于 Deployment 和 DaemonSet 来说同样有效，因为 Deployment/DaemonSet 是 Pod 的上层编排，终究还是要创建 Pod。


