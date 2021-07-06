{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# Kubernetes 集群监控指标数据

该方案扩展了 Datakit 集群部署，内置 [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) 服务，用来采集 kubernetes 集群监控指标数据。

> 具体指标，参考 [kubernetes](kubernetes) 中的列表

### 修改配置

修改 `datakit-default.yaml` 中的 dataway 配置

```yaml
  - name: ENV_DATAWAY
		value: https://openway.dataflux.cn?token=<your-token> # 此处填上具体 token
```

### 应用 yaml 配置

下载下文 yaml， 直接应用即可：

```shell
kubectl apply path/to/your.yaml
```

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
          value: https://openway.dataflux.cn?token=<your-token>
        - name: ENV_GLOBAL_TAGS
          value: host=__datakit_hostname,host_ip=__datakit_ip
        - name: ENV_ENABLE_INPUTS
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
        - mountPath: /usr/local/datakit/conf.d/container/container.conf
          name: datakit-conf
          subPath: container.conf
        - mountPath: /usr/local/datakit/conf.d/kubernetes/kubernetes.conf
          name: datakit-conf
          subPath: kubernetes.conf
        - mountPath: /usr/local/datakit/conf.d/kubernetes/kube-state-metric.conf
          name: datakit-conf
          subPath: kube-state-metric.conf
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
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.1.0
  name: kube-state-metrics
  namespace: datakit
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: kube-state-metrics
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kube-state-metrics
        app.kubernetes.io/version: 2.1.0
    spec:
      containers:
      - image: pubrepo.jiagouyun.com/metrics-server/kube-state-metrics:v2.1.0
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          timeoutSeconds: 5
        name: kube-state-metrics
        ports:
        - containerPort: 8080
          name: http-metrics
        - containerPort: 8081
          name: telemetry
        readinessProbe:
          httpGet:
            path: /
            port: 8081
          initialDelaySeconds: 5
          timeoutSeconds: 5
        securityContext:
          runAsUser: 65534
      nodeSelector:
        kubernetes.io/os: linux
      serviceAccountName: datakit
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.1.0
  name: kube-state-metrics
  namespace: datakit
spec:
  type: NodePort
  ports:
  - name: http-metrics
    port: 8080
    targetPort: http-metrics
    nodePort: 30022
  - name: telemetry
    port: 8081
    targetPort: telemetry
  selector:
    app.kubernetes.io/name: kube-state-metrics
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
    #### kube-state-metric
    kube-state-metric.conf: |-
      [[inputs.prom]]
        ## kube-state-metrics Exporter 地址(该ip为任意节点ip地址)
        url = "http://kube-state-metrics.datakit:8080/metrics"

        ## 采集器别名
        #input_alias = "kube-state-metric"

        # 只采集 counter 和 gauge 类型的指标
        metric_types = ["counter", "gauge"]

        ## 指标名称过滤
        metric_name_filter = [
          "kube_daemonset_status_current_number_scheduled",
          "kube_daemonset_status_desired_number_scheduled",
          "kube_daemonset_status_number_available",
          "kube_daemonset_status_number_misscheduled",
          "kube_daemonset_status_number_ready",
          "kube_daemonset_status_number_unavailable",
          "kube_daemonset_updated_number_scheduled",

          "kube_deployment_spec_paused",
          "kube_deployment_spec_strategy_rollingupdate_max_unavailable",
          "kube_deployment_spec_strategy_rollingupdate_max_surge",
          "kube_deployment_status_replicas",
          "kube_deployment_status_replicas_available",
          "kube_deployment_status_replicas_unavailable",
          "kube_deployment_status_replicas_updated",
          "kube_deployment_status_condition",
          "kube_deployment_spec_replicas",

          "kube_endpoint_address_available",
          "kube_endpoint_address_not_ready",

          "kube_persistentvolumeclaim_status_phase",
          "kube_persistentvolumeclaim_resource_requests_storage_bytes",
          "kube_persistentvolumeclaim_access_mode",

          "kube_persistentvolume_status_phase",
          "kube_persistentvolume_capacity_bytes",

          "kube_secret_type",

          "kube_replicaset_status_replicas",
          "kube_replicaset_status_fully_labeled_replicas",
          "kube_replicaset_status_ready_replicas",
          "kube_replicaset_status_observed_generation",

          "kube_statefulset_status_replicas",
          "kube_statefulset_status_replicas_current",
          "kube_statefulset_status_replicas_ready",
          "kube_statefulset_status_replicas_updated",
          "kube_statefulset_status_observed_generation",
          "kube_statefulset_replicas",

          "kube_hpa_spec_max_replicas",
          "kube_hpa_spec_min_replicas",
          "kube_hpa_spec_target_metric",
          "kube_hpa_status_current_replicas",
          "kube_hpa_status_desired_replicas",
          "kube_hpa_status_condition",

          "kube_cronjob_status_active",
          "kube_cronjob_spec_suspend",
          "kube_cronjob_status_last_schedule_time",

          "kube_job_status_succeeded",
          "kube_job_status_failed",
          "kube_job_status_active",
          "kube_job_complete",
        ]

        interval = "10s"

        ## 自定义指标集名称
        [[inputs.prom.measurements]]
          ## daemonset
          prefix = "kube_daemonset_"
          name = "kube_daemonset"

        [[inputs.prom.measurements]]
          ## daemonset
          prefix = "kube_deployment_"
          name = "kube_deployment"

        [[inputs.prom.measurements]]
          ## endpoint
          prefix = "kube_endpoint_"
          name = "kube_endpoint"

        [[inputs.prom.measurements]]
          ## persistentvolumeclaim
          prefix = "kube_persistentvolumeclaim_"
          name = "kube_persistentvolumeclaim"

        [[inputs.prom.measurements]]
          ## persistentvolumeclaim
          prefix = "kube_persistentvolume_"
          name = "kube_persistentvolume"

        [[inputs.prom.measurements]]
          ## secret
          prefix = "kube_secret_"
          name = "kube_secret"

        [[inputs.prom.measurements]]
          ## replicaset
          prefix = "kube_replicaset_"
          name = "kube_replicaset"

        [[inputs.prom.measurements]]
          ## statefulset
          prefix = "kube_statefulset_"
          name = "kube_statefulset"

        [[inputs.prom.measurements]]
          ## hpa
          prefix = "kube_hpa_"
          name = "kube_hpa"

        [[inputs.prom.measurements]]
          ## cronjob
          prefix = "kube_cronjob_"
          name = "kube_cronjob"

        [[inputs.prom.measurements]]
          ## job
          prefix = "kube_job_"
          name = "kube_job"

        ## 自定义Tags
        [inputs.prom.tags]
          #tag1 = "value1"
          #tag2 = "value2"
```
