# Kubernetes with Metric Server
---

操作系统支持：Linux

## 视图预览
Kubernetes 性能指标展示：包括 pod 数量、deployment 数量、job 数量、endpoint 数量、service 数量、CPU、内存、Pod 分布等。<br />
![image.png](imgs/input-kube-metric-server-01.png)<br />
![image.png](imgs/input-kube-metric-server-02.png)<br />
![image.png](imgs/input-kube-metric-server-03.png)<br />
![1650000133(1).png](imgs/input-kube-metric-server-04.png)<br />
![1650000156(1).png](imgs/input-kube-metric-server-05.png)<br />
![1650000171(1).png](imgs/input-kube-metric-server-06.png)<br />
![1650000207(1).png](imgs/input-kube-metric-server-07.png)

## 安装部署
说明：示例 Kubernetes 版本为：1.22.6

### 前置条件

- Kubernetes 集群。
- 采集 Kubernetes Pod 指标数据，[需要 Kubernetes 安装 Metrics-Server 组件](https://github.com/kubernetes-sigs/metrics-server#installation)。


### 配置实施

#### 部署 Metric-Server  (必选)
新建 metric-server.yaml ，在 kubernetes 集群执行
```
kubectl apply -f metric-server.yaml 
```
metric-server.yaml  完整内容如下：
```
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-app: metrics-server
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-view: "true"
  name: system:aggregated-metrics-reader
rules:
- apiGroups:
  - metrics.k8s.io
  resources:
  - pods
  - nodes
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-app: metrics-server
  name: system:metrics-server
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  - nodes/stats
  - namespaces
  - configmaps
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server-auth-reader
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server:system:auth-delegator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: system:metrics-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:metrics-server
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
spec:
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: https
  selector:
    k8s-app: metrics-server
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: metrics-server
  strategy:
    rollingUpdate:
      maxUnavailable: 0
  template:
    metadata:
      labels:
        k8s-app: metrics-server
    spec:      
      containers:
      - args:
        - --cert-dir=/tmp
        - --secure-port=4443
        - --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname
        - --kubelet-use-node-status-port
        - --metric-resolution=15s
        - --kubelet-insecure-tls 
        # image: k8s.gcr.io/metrics-server/metrics-server:v0.5.2
        image: bitnami/metrics-server:0.5.2
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /livez
            port: https
            scheme: HTTPS
          periodSeconds: 10
        name: metrics-server
        ports:
        - containerPort: 4443
          name: https
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /readyz
            port: https
            scheme: HTTPS
          initialDelaySeconds: 20
          periodSeconds: 10
        resources:
          requests:
            cpu: 100m
            memory: 200Mi
        securityContext:
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1000
        volumeMounts:
        - mountPath: /tmp
          name: tmp-dir
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: system-cluster-critical
      serviceAccountName: metrics-server
      volumes:
      - emptyDir: {}
        name: tmp-dir
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  labels:
    k8s-app: metrics-server
  name: v1beta1.metrics.k8s.io
spec:
  group: metrics.k8s.io
  groupPriorityMinimum: 100
  insecureSkipTLSVerify: true
  service:
    name: metrics-server
    namespace: kube-system
  version: v1beta1
  versionPriority: 100

```


#### Daemonset 部署 DataKit (必选)
登录[观测云](https://console.guance.com/)，【集成】->【DataKit】-> 【Kubernetes】，下载 `datakit.yaml`（命名无要求）。

1. 修改 `datakit.yaml` 中的 dataway 配置

      进入【管理】模块，找到下图中 token。

![1648545757(1).png](imgs/input-kube-metric-server-08.png)<br />替换 datakit.yaml 文件中的 ENV_DATAWAY 环境变量的 value 值中的 <your-token>。
```
        - name: ENV_DATAWAY
          value: https://openway.guance.com?token=<your-token>
```

在 datakit.yaml 文件中的 ENV_GLOBAL_TAGS 环境变量值最后增加 cluster_name_k8s=k8s-prod，其中  k8s-prod 为指标设置的全局 tag，即指标所在的集群名称。
```
        - name: ENV_GLOBAL_TAGS
          value: host=__datakit_hostname,host_ip=__datakit_ip,cluster_name_k8s=k8s-prod
```
更多环境变量请参考 [DataKit 环境变量设置](https://www.yuque.com/dataflux/datakit/datakit-daemonset-deploy)。

2. 增加 ENV_NAMESPACE 环境变量 

      修改 `datakit.yaml`，增加 ENV_NAMESPACE 环境变量，这个环境变量是为了区分不同集群的选举，多个集群 value 值不能相同。
```
        - name: ENV_NAMESPACE
          value: xxx
```

3. 定义ConfigMap

『注意』下载的 datakit.yaml 并没有 ConfigMap，定义的 ConfigMap 可一起放到 datakit.yaml 。
```
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:
    #### container
    container.conf: |-  
      [inputs.container]
        docker_endpoint = "unix:///var/run/docker.sock"
        containerd_address = "/var/run/containerd/containerd.sock"

        enable_container_metric = true
        enable_k8s_metric = true
        enable_pod_metric = true

        ## Containers logs to include and exclude, default collect all containers. Globs accepted.
        container_include_log = []
        container_exclude_log = ["image:pubrepo.jiagouyun.com/datakit/logfwd*", "image:pubrepo.jiagouyun.com/datakit/datakit*"]

        exclude_pause_container = true

        ## Removes ANSI escape codes from text strings
        logging_remove_ansi_escape_codes = false

        kubernetes_url = "https://kubernetes.default:443"

        ## Authorization level:
        ##   bearer_token -> bearer_token_string -> TLS
        ## Use bearer token for authorization. ('bearer_token' takes priority)
        ## linux at:   /run/secrets/kubernetes.io/serviceaccount/token
        ## windows at: C:\var\run\secrets\kubernetes.io\serviceaccount\token
        bearer_token = "/run/secrets/kubernetes.io/serviceaccount/token"
        # bearer_token_string = "<your-token-string>"

        [inputs.container.tags]
          # some_tag = "some_value"
          # more_tag = "some_other_value"
```
[inputs.container]参数说明

- enable_container_metric：是否开启 container 指标采集，请设置为true。
- enable_k8s_metric：是否开启 kubernetes 指标采集。
- enable_pod_metric：是否开启 Pod 指标采集。
- container_include_log：须要采集的容器日志。
- container_exclude_log：不须要采集的容器日志。

`container_include_log` 和 `container_exclude_log` 必须以 `image` 开头，格式为 `"image:<glob规则>"`，表示 glob 规则是针对容器 image 生效。[Glob 规则](https://en.wikipedia.org/wiki/Glob_(programming))是一种轻量级的正则表达式，支持 `*` `?` 等基本匹配单元

4. 使用ConfigMap

        在 datakit.yaml 文件中的 volumeMounts 下面增加：
```
        - mountPath: /usr/local/datakit/conf.d/container/container.conf
          name: datakit-conf
          subPath: container.conf
```

5. 部署Datakit
```
kubectl apply -f datakit.yaml
```

#### 日志采集 
默认自动收集输出到控制台的日志，如果采集不输出到控制台且输出文件的日志，请参考<<[Kubernetes 应用的 RUM-APM-LOG 联动分析](https://www.yuque.com/dataflux/bp/k8s-rum-apm-log)>>的日志配置部分。


#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 kubernetes 指标都会带有 tag1 = "val1" 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](https://www.yuque.com/dataflux/bp/tag)>
```
          [inputs.kubernetes.tags]
           #tag1 = "val1"
           #tag2 = "valn"   
```


## 场景视图
场景 - 新建仪表板 - Kubernetes 监控视图<br />相关文档 <[DataFlux 场景管理](https://www.yuque.com/dataflux/doc/trq02t)> 

## 异常检测
暂无

## 指标详解
<[Kubernetes 指标详情](https://www.yuque.com/dataflux/datakit/container#23ae0855)>


## 最佳实践
暂无

## 故障排查
<[无数据上报排查](https://www.yuque.com/dataflux/datakit/why-no-data)>

