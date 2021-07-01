{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# kubernetes 集群中 Prometheus Exporter 指标采集

## 介绍

该方案可以在 kubernetes 集群中通过配置，自定义收集 Prometheus Exporter 数据

以下以一个 Demo 的方式进行展开

## 使用demo

### 场景

现有一个服务 dummy-server 需要被 datakit 收集其通过 Prometheus Exporter 接口提供的 Metric

### 部署demo Deployment

dummy-server 服务是使用 deployment 部署的3个 pod 副本的服务，在该服务中，打有特征标记label datakit

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dummy-exporter-deployment
  labels:
    app: dummy-exporter
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dummy-exporter
  template:
    metadata:
      labels:
        app: dummy-exporter
        datakit: prom-dev
    spec:
      containers:
      - name: dummy-exporter
        image: pubrepo.jiagouyun.com/demo/dummy-exporter:latest
        ports:
        - containerPort: 12345
```

### 创建pod注解

通过上文定义的标签特征对 pod 添加 annotate 注解

```
kubectl annotate --overwrite pods -l datakit exporter_url.dummy_server='http://$ip:12345/metric'
```

**注意**

约定annotate key/value 数据格式规范： `exporter_url.<service>='http://$ip:<port>/<metric path>'`

参数说明:

- <service>: 服务名，该配置要和下文中 prom 配置相统一
- <port>: pod 中 exporter 的端口
- <metric path>: exporter 的路由, 如：`/metric`

### 禁用采集

关闭对自定义exporter的指标收集

```
kubectl annotate --overwrite pods -l datakit exporter_url.dummy_server='off'
```

### 编写自定义 prom Config

详情参见 [prom采集器](prom)

### 修改k8s daemonSet

- 追加mountPath

注意：以下配置中变量内容，为上文创建 pod 注解中的 `service` 值保持一致

示例:

```
- mountPath: /usr/local/datakit/conf.d/prom/<dummy_server>.conf
  name: datakit-conf
  subPath: <dummy_server>.conf
```

- 追加编写的configMap

注意：以下配置中变量内容，为上文创建 pod 注解中的`service`值保持一致

示例:

```
  #### dummy-server
  <dummy_server>.conf: |-
    [[inputs.prom]]
      ## Exporter 地址
      url = "/usr/local/datakit/data/exporter_urls/<dummy_server>.json"

      # 默认只采集 counter 和 gauge 类型的指标
      metric_types = ["counter", "gauge"]

      ## 采集间隔
      interval = "10s"

      tags_ignore = ["pod"]

      ## 自定义Tags
      [inputs.prom.tags]
        service = "dummy-exporter"
```

完整部署yaml示例：

```
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
  - ""
  resources:
  - nodes
  - nodes/proxy
  - namespaces
  - pods
  - services
  - endpoints
  - persistentvolumes
  - persistentvolumeclaims
  - ingresses
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
  - extensions
  resources:
  - ingresses
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
          value: https://openway.dataflux.cn?token=<xxxxx>
        - name: ENV_GLOBAL_TAGS
          value: host=__datakit_hostname,host_ip=__datakit_ip
        - name: ENV_ENABLE_INPUTS
          value: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,kubernetes,container
        - name: ENV_ENABLE_ELECTION
          value: enable
        - name: ENV_HTTP_LISTEN
          value: 0.0.0.0:9529
        image: pubrepo.jiagouyun.com/demo/datakit:<xxxxx>
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
        - mountPath: /usr/local/datakit/conf.d/container/container.conf
          name: datakit-conf
          subPath: container.conf
        - mountPath: /usr/local/datakit/conf.d/kubernetes/kubernetes.conf
          name: datakit-conf
          subPath: kubernetes.conf
        - mountPath: /usr/local/datakit/conf.d/prom/dummy_server.conf
          name: datakit-conf
          subPath: dummy_server.conf
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
      hostNetwork: true
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
        endpoint = "unix:///var/run/docker.sock"

        enable_metric = false
        enable_object = true
        enable_logging = true

        metric_interval = "10s"

        ## TLS Config
        # tls_ca = "/path/to/ca.pem"
        # tls_cert = "/path/to/cert.pem"
        # tls_key = "/path/to/key.pem"
        ## Use TLS but skip chain & host verification
        # insecure_skip_verify = false

        [inputs.container.kubelet]
          kubelet_url = "http://127.0.0.1:10255"

          ## Use bearer token for authorization. ('bearer_token' takes priority)
          ## If both of these are empty, we'll use the default serviceaccount:
          ## at: /run/secrets/kubernetes.io/serviceaccount/token
          # bearer_token = "/path/to/bearer/token"
          ## OR
          # bearer_token_string = "abc_123"

          ## Optional TLS Config
          # tls_ca = /path/to/ca.pem
          # tls_cert = /path/to/cert.pem
          # tls_key = /path/to/key.pem
          ## Use TLS but skip chain & host verification
          # insecure_skip_verify = false

        #[[inputs.container.logfilter]]
        #  filter_message = [
        #    '''<this-is-message-regexp''',
        #    '''<this-is-another-message-regexp''',
        #  ]
        #  source = "<your-source-name>"
        #  service = "<your-service-name>"
        #  pipeline = "<pipeline.p>"

        [inputs.container.tags]
          # some_tag = "some_value"
          # more_tag = "some_other_value"


    #### kubernetes
    kubernetes.conf: |-
      [[inputs.kubernetes]]
          # required
          interval = "10s"
          ## URL for the Kubernetes API
          url = "https://kubernetes.default:443"
          ## Use bearer token for authorization. ('bearer_token' takes priority)
          ## at: /run/secrets/kubernetes.io/serviceaccount/token
          bearer_token = "/run/secrets/kubernetes.io/serviceaccount/token"

          ## Set http timeout (default 5 seconds)
          timeout = "5s"

           ## Optional TLS Config
          tls_ca = "/run/secrets/kubernetes.io/serviceaccount/ca.crt"

          ## Use TLS but skip chain & host verification
          insecure_skip_verify = false

          [inputs.kubernetes.tags]
           #tag1 = "val1"
           #tag2 = "valn"

    #### prom_dummy-exporter
    dummy_server.conf: |-
      [[inputs.prom]]
        ## Exporter 地址
        url = "/usr/local/datakit/data/exporter_urls/dummy_server.json"

        # 默认只采集 counter 和 gauge 类型的指标
        metric_types = ["counter", "gauge"]

        measurement_name = "dummy_server"

        ## 采集间隔
        interval = "10s"

        tags_ignore = ["pod"]

        ## 自定义Tags
        [inputs.prom.tags]
          service = "dummy-exporter"
```




