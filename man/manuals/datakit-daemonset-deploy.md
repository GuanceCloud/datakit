{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# DaemonSet 安装 DataKit 

本文档介绍如何在 K8s 中通过 DaemonSet 方式安装 DataKit。

## 安装步骤 

先下载本文档尾部的 yaml 配置，保存为 `datakit-default.yaml`（命名无要求）。在该配置中，有两个采集器可以配置：

- kubernetes：用来采集 Kubernetes 中心指标，需要填写 kubernetes 中心采集地址
- container：用来采集 Node 上的容器对象以及运行指标（如果要采集容器运行指标，则需要修改配置）

其它主机相关的采集器都是默认开启的（`cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes`），且无需额外配置。

### 修改配置

修改 `datakit-default.yaml` 中的 dataway 配置

```yaml
	- name: ENV_DATAWAY
		value: https://openway.guance.com?token=<your-token> # 此处填上 dataway 真实地址
```

#### container 配置

默认情况下，container 采集器没有开启指标采集，如需开启指标采集，修改 `datakit-default.yaml` 中如下配置：

```yaml
	[inputs.container]
		endpoint = "unix:///var/run/docker.sock"

		enable_metric = true # 将此处设置成 true
		enable_object = true
```

详情参见 [容器采集配置](container)

### 安装 yaml

```shell
kubectl apply -f datakit-default.yaml
```

### 查看运行状态：

安装完后，会创建一个 datakit 的 DaemonSet 部署：

```shell
kubectl get pod -n datakit
```

### DataKit 中其它环境变量设置

在 DaemonSet 模式中，DataKit 支持多个环境变量配置，如下表所示：

| 环境变量名称                 | 默认值                     | 是否必须 | 说明                                                                                                               |
| ---------                    | ---                        | ------   | ----                                                                                                               |
| ENV_DATAWAY                | 无                         | 是       | 可配置多个 dataway，以英文逗号分割，如 `https://openway.guance.com?token=xxx,https://openway.guance.com?token=yyy` |
| ENV_GLOBAL_TAGS            | 无                         | 否       | 全局 tag，多个 tag 之间以英文逗号分割，如 `tag1=val,tag2=val2`                                                     |
| ENV_LOG_LEVEL              | `info`                     | 否       | 可选值 `info/debug`                                                                                                |
| ENV_LOG                    | `/var/log/datakit/log`     | 否       | 如果改成 `stdout`，日志将不写文件，而是终端输出                                                                    |
| ENV_GIN_LOG                | `/var/log/datakit/gin.log` | 否       | 如果改成 `stdout`，日志将不写文件，而是终端输出                                                                    |
| ENV_HTTP_LISTEN            | `localhost:9529`           | 否       | 可修改改地址，使得外部可以调用 [DataKit 接口](apis)                                                                |
| ENV_RUM_ORIGIN_IP_HEADER   | `X-Forwarded-For`          | 否       | RUM 专用                                                                                                           |
| ENV_DISABLE_404PAGE        | 无                         | 否       | 禁用 DataKit 404 页面（公网部署 DataKit RUM 时常用）                                                               |
| ENV_DEFAULT_ENABLED_INPUTS | 无                         | 否       | 默认开启采集器列表，以英文逗号分割，如 `cpu,mem,disk`。                                                              |
| ENV_ENABLE_ELECTION        | 默认不开启                 | 否       | 开启[选举](election)，默认不开启，如需开启，给该环境变量任意一个非空字符串值即可                                   |
| ENV_NAMESPACE              | 无                         | 否       | DataKit 所在的命名空间，默认为空表示不区分命名空间，接收任意非空字符串，如 `dk-namespace-example`                  |
| ENV_HOSTNAME               | 无                         | 否       | 默认为本地主机名，可安装时指定，如， `dk-your-hostname`                  |


> 注意，`ENV_ENABLE_INPUTS` 已被弃用（但仍有效），建议使用 `ENV_DEFAULT_ENABLED_INPUTS`。如果俩个环境变量同时指定，则**只有后者生效**。

### yaml 配置

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: datakit
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterroles
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  - nodes/proxy
  - namespaces
  - pods
  - events
  - services
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - deployments
  - daemonsets
  - statefulsets
  - replicasets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - get
  - list
  - watch
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]

---

apiVersion: v1
kind: ServiceAccount
metadata:
  name: datakit
  namespace: datakit

---

apiVersion: v1
kind: Service
metadata:
  name: datakit-service
  namespace: datakit
spec:
  selector:
    app: daemonset-datakit
  ports:
    - protocol: TCP
      port: 9529
      targetPort: 9529

---

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: datakit
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: datakit
subjects:
- kind: ServiceAccount
  name: datakit
  namespace: datakit

---

apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: daemonset-datakit
  name: datakit
  namespace: datakit
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: daemonset-datakit
  template:
    metadata:
      labels:
        app: daemonset-datakit
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - env:
        - name: HOST_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: ENV_DATAWAY
          value: https://openway.guance.com?token=<your-token> # 此处填上 dataway 真实地址
        - name: ENV_GLOBAL_TAGS
          value: host=__datakit_hostname,host_ip=__datakit_ip
        - name: ENV_DEFAULT_ENABLED_INPUTS
          value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,kubernetes,container
        - name: ENV_ENABLE_ELECTION
          value: enable
        - name: ENV_HTTP_LISTEN
          value: 0.0.0.0:9529
        image: pubrepo.jiagouyun.com/datakit/datakit:{{.Version}}
        imagePullPolicy: Always
        name: datakit
        ports:
        - containerPort: 9529
          hostPort: 9529
          name: port
          protocol: TCP
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /var/run/docker.sock
          name: docker-socket
          readOnly: true
        - mountPath: /host/proc
          name: proc
          readOnly: true
        - mountPath: /host/dev
          name: dev
          readOnly: true
        - mountPath: /host/sys
          name: sys
          readOnly: true
        - mountPath: /rootfs
          name: rootfs
        workingDir: /usr/local/datakit
      hostIPC: true
      hostPID: true
      restartPolicy: Always
      serviceAccount: datakit
      serviceAccountName: datakit
      volumes:
      - configMap:
          name: datakit-conf
        name: datakit-conf
      - hostPath:
          path: /var/run/docker.sock
        name: docker-socket
      - hostPath:
          path: /proc
          type: ""
        name: proc
      - hostPath:
          path: /dev
          type: ""
        name: dev
      - hostPath:
          path: /sys
          type: ""
        name: sys
      - hostPath:
          path: /
          type: ""
        name: rootfs
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
```

> 注意：默认情况下，我们在该 yaml 中开启了如下采集器：

- `cpu`
- `disk`
- `diskio`
- `mem`
- `swap`
- `system`
- `hostobject`
- `net`
- `host_processes`
- `kubernetes`
- `container`

如需开启更多其它采集器，如开启 ddtrace，直接在如下配置中追加即可。当然也可以将某些采集器从这个列表中删掉。

```yaml
        - name: ENV_DEFAULT_ENABLED_INPUTS
          value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,kubernetes,container
```
