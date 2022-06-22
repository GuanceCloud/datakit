# Istio
---

## 视图预览

Istio 性能指标展示：Incoming Request Volume、Incoming Success Rate、Incoming Requests By Source And Response Code、Outgoing Requests By Destination And Response Code 等。

![1650359370(1).png](../imgs/istio-1.png)

## 版本支持

Istio 版本： [istio](https://github.com/istio/istio)。

## 前置条件

- 已部署 [Kubernetes](https://kubernetes.io/docs/setup/production-environment/tools/)。

- 已部署 DataKit，请参考  [Daemonset 部署 Datakit](kube-metric-server.md) 。

## 安装配置

说明：示例 Istio 版本为 1.11.2。

### 安装 Istio

#### 1 下载 Istio 

[下载](https://github.com/istio/istio/releases ) **Source Code **和 **istio-1.11.2-linux-amd64.tar.gz。**

#### 2 安装 Istio 

上传 istio-1.11.2-linux-amd64.tar.gz 到 /usr/local/df-demo/ 目录。

```
cd /usr/local/df-demo/
tar zxvf istio-1.11.2-linux-amd64.tar.gz  
cd /usr/local/df-demo/istio-1.11.2
export PATH=$PWD/bin:$PATH$ 
cp -ar /usr/local/df-demo/istio-1.11.2/bin/istioctl /usr/bin/

istioctl install --set profile=demo 
```

#### 3 验证安装

部署成功后，ingressgateway、egressgateway、istiod 会处于 Running 状态。

```
kubectl get pods -n istio-system 
```

![1650364329(1).png](../imgs/istio-2.png)

### 部署实施

#### 指标采集 (必选)

1、 开通 Sidecar 注入

为集群中的 namespace 设置 sidecar 自动注入，在该 namespace 下，新创建的 Pod 就会注入一个 Envoy 容器用来接管流量。开通方式是为 namespace 添加标签，下面以 prod 名称空间为例。

```
kubectl create namespace prod
kubectl label namespace prod istio-injection=enabled
```

2、 开启 Envoy 指标采集器

在业务 Pod 处添加如下 annotations（具体路径 spec.template.metadata 下），这样即可采集 Envoy 的指标数据。

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: productpage-v1
spec:  
  template:
    metadata:
    ...
      annotations:  # 下面是新增部分
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:15020/stats/prometheus"
            source = "bookinfo-istio-product"
            metric_types = ["counter", "gauge", "histogram"]
            interval = "10s"
            tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
            #measurement_prefix = ""
            measurement_name = "istio_prom"
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"         
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
```

参数说明

- url：Exporter 地址

- source：采集器名称
- metric_types：指标类型过滤
- measurement_name：采集后的指标集名称
- interval：采集指标频率，s秒
- $IP：通配 Pod 的内网 IP
- $NAMESPACE：Pod所在命名空间
- tags_ignore:  忽略的 tag。

3、 部署 bookinfo 项目

下面部署 istio 自带的 bookinfo，里面的 Pod 都要增加 annotations。

```
/usr/local/df-demo/istio-1.11.2/samples/bookinfo/platform/kube/bookinfo.yaml
```

bookinfo.yaml 修改后的完整内容：

```
apiVersion: v1
kind: Service
metadata:
  name: details
  namespace: prod
  labels:
    app: details
    service: details
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: details
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookinfo-details
  namespace: prod
  labels:
    account: details
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: details-v1
  namespace: prod
  labels:
    app: details
    version: v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: details
      version: v1
  template:
    metadata:
      labels:
        app: details
        version: v1
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:15020/stats/prometheus"
            source = "bookinfo-istio-details"
            metric_types = ["counter", "gauge"]
            interval = "10s"
            tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
            #measurement_prefix = ""
            measurement_name = "istio_prom"
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"         
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
    spec:
      serviceAccountName: bookinfo-details
      containers:
      - name: details
        image: docker.io/istio/examples-bookinfo-details-v1:1.16.2
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9080
        securityContext:
          runAsUser: 1000
---
##################################################################################################
# Ratings service
##################################################################################################
apiVersion: v1
kind: Service
metadata:
  name: ratings
  namespace: prod
  labels:
    app: ratings
    service: ratings
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: ratings
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookinfo-ratings
  namespace: prod
  labels:
    account: ratings
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ratings-v1
  namespace: prod
  labels:
    app: ratings
    version: v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ratings
      version: v1
  template:
    metadata:
      labels:
        app: ratings
        version: v1
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:15020/stats/prometheus"
            source = "bookinfo-istio-ratings"
            metric_types = ["counter", "gauge"]
            interval = "10s"
            tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
            #measurement_prefix = ""
            measurement_name = "istio_prom"
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"         
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
    spec:
      serviceAccountName: bookinfo-ratings
      containers:
      - name: ratings
        image: docker.io/istio/examples-bookinfo-ratings-v1:1.16.2
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9080
        securityContext:
          runAsUser: 1000
---
##################################################################################################
# Reviews service
##################################################################################################
apiVersion: v1
kind: Service
metadata:
  name: reviews
  namespace: prod
  labels:
    app: reviews
    service: reviews
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: reviews
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookinfo-reviews
  namespace: prod
  labels:
    account: reviews
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: reviews-v1
  namespace: prod
  labels:
    app: reviews
    version: v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: reviews
      version: v1
  template:
    metadata:
      labels:
        app: reviews
        version: v1
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:15020/stats/prometheus"
            source = "bookinfo-istio-review1"
            metric_types = ["counter", "gauge"]
            interval = "10s"
			tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
            #measurement_prefix = ""
            measurement_name = "istio_prom"
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"         
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
    spec:
      serviceAccountName: bookinfo-reviews
      containers:
      - name: reviews
        image: docker.io/istio/examples-bookinfo-reviews-v1:1.16.2
        imagePullPolicy: IfNotPresent
        env:
        - name: LOG_DIR
          value: "/tmp/logs"
        ports:
        - containerPort: 9080
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: wlp-output
          mountPath: /opt/ibm/wlp/output
        securityContext:
          runAsUser: 1000
      volumes:
      - name: wlp-output
        emptyDir: {}
      - name: tmp
        emptyDir: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: reviews-v2
  namespace: prod
  labels:
    app: reviews
    version: v2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: reviews
      version: v2
  template:
    metadata:
      labels:
        app: reviews
        version: v2
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:15020/stats/prometheus"
            source = "bookinfo-istio-review2"
            metric_types = ["counter", "gauge"]
            interval = "10s"
			tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
            #measurement_prefix = ""
            measurement_name = "istio_prom"
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"         
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
    spec:
      serviceAccountName: bookinfo-reviews
      containers:
      - name: reviews
        image: docker.io/istio/examples-bookinfo-reviews-v2:1.16.2
        imagePullPolicy: IfNotPresent
        env:
        - name: LOG_DIR
          value: "/tmp/logs"
        ports:
        - containerPort: 9080
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: wlp-output
          mountPath: /opt/ibm/wlp/output
        securityContext:
          runAsUser: 1000
      volumes:
      - name: wlp-output
        emptyDir: {}
      - name: tmp
        emptyDir: {}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: reviews-v3
  namespace: prod
  labels:
    app: reviews
    version: v3
spec:
  replicas: 1
  selector:
    matchLabels:
      app: reviews
      version: v3
  template:
    metadata:
      labels:
        app: reviews
        version: v3
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:15020/stats/prometheus"
            source = "bookinfo-istio-review3"
            metric_types = ["counter", "gauge"]
            interval = "10s"
			tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
            #measurement_prefix = ""
            measurement_name = "istio_prom"
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"         
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
    spec:
      serviceAccountName: bookinfo-reviews
      containers:
      - name: reviews
        image: docker.io/istio/examples-bookinfo-reviews-v3:1.16.2
        imagePullPolicy: IfNotPresent
        env:
        - name: LOG_DIR
          value: "/tmp/logs"
        ports:
        - containerPort: 9080
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: wlp-output
          mountPath: /opt/ibm/wlp/output
        securityContext:
          runAsUser: 1000
      volumes:
      - name: wlp-output
        emptyDir: {}
      - name: tmp
        emptyDir: {}
---
##################################################################################################
# Productpage services
##################################################################################################
apiVersion: v1
kind: Service
metadata:
  name: productpage
  namespace: prod
  labels:
    app: productpage
    service: productpage
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: productpage
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookinfo-productpage
  namespace: prod
  labels:
    account: productpage
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: productpage-v1
  namespace: prod
  labels:
    app: productpage
    version: v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: productpage
      version: v1
  template:
    metadata:
      labels:
        app: productpage
        version: v1
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:15020/stats/prometheus"
            source = "bookinfo-istio-product"
            metric_types = ["counter", "gauge", "histogram"]
            interval = "10s"
			tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
            #measurement_prefix = ""
            measurement_name = "istio_prom"
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"         
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
    spec:
      serviceAccountName: bookinfo-productpage
      containers:
      - name: productpage
        image: docker.io/istio/examples-bookinfo-productpage-v1:1.16.2
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9080
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        securityContext:
          runAsUser: 1000
      volumes:
      - name: tmp
        emptyDir: {}

```
```
kubectl apply -f bookinfo.yaml
```
         创建 bookinfo gateway 资源和虚拟服务
```
kubectl apply -f /usr/local/df-demo/istio-1.11.2/samples/bookinfo/networking/bookinfo-gateway.yaml
```
        bookinfo-gateway.yaml 增加了 namespace，完整内容如下：
```
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: bookinfo-gateway
  namespace: prod
spec:
  selector:
    istio: ingressgateway # use istio default controller
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: bookinfo
  namespace: prod
spec:
  hosts:
  - "*"
  gateways:
  - bookinfo-gateway
  http:
  - match:
    - uri:
        exact: /productpage
    - uri:
        prefix: /static
    - uri:
        exact: /login
    - uri:
        exact: /logout
    - uri:
        prefix: /api/v1/products
    route:
    - destination:
        host: productpage
        port:
          number: 9080

```

4、 采集 istiod、ingressgateway、egressgateway pod 指标

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:    # 下面是新增部分
    prom_istiod.conf: |-    
      [[inputs.prom]] 
        url = "http://istiod.istio-system.svc.cluster.local:15014/metrics"
        source = "prom-istiod"
        metric_types = ["counter", "gauge", "histogram"]
        interval = "10s"
        tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
        #measurement_prefix = ""
        measurement_name = "istio_prom"
        #[[inputs.prom.measurements]]
        # prefix = "cpu_"
        # name ="cpu"
        [inputs.prom.tags]
          app_id="istiod"
    #### ingressgateway
    prom-ingressgateway.conf: |- 
        [[inputs.prom]] 
          url = "http://istio-ingressgateway-ext.istio-system.svc.cluster.local:15020/stats/prometheus"
          source = "prom-ingressgateway"
          metric_types = ["counter", "gauge", "histogram"]
          interval = "10s"
          tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
          #measurement_prefix = ""
          measurement_name = "istio_prom"
          #[[inputs.prom.measurements]]
          # prefix = "cpu_"
          # name ="cpu"
    #### egressgateway
    prom-egressgateway.conf: |- 
        [[inputs.prom]] 
          url = "http://istio-egressgateway-ext.istio-system.svc.cluster.local:15020/stats/prometheus"
          source = "prom-egressgateway"
          metric_types = ["counter", "gauge", "histogram"]
          interval = "10s"
          tags_ignore = ["cache","cluster_type","component","destination_app","destination_canonical_revision","destination_canonical_service","destination_cluster","destination_principal","group","grpc_code","grpc_method","grpc_service","grpc_type","reason","request_protocol","request_type","resource","responce_code_class","response_flags","source_app","source_canonical_revision","source_canonical-service","source_cluster","source_principal","source_version","wasm_filter"]
          #measurement_prefix = ""
          measurement_name = "istio_prom"
          #[[inputs.prom.measurements]]
          # prefix = "cpu_"
          # name ="cpu"		  
```
```
apiVersion: apps/v1
kind: DaemonSet
...
spec:
  template
    spec:
      containers:
      - env:
        volumeMounts: # 下面是新增部分
        - mountPath: /usr/local/datakit/conf.d/prom/prom_istiod.conf
          name: datakit-conf
          subPath: prom_istiod.conf
        - mountPath: /usr/local/datakit/conf.d/prom/prom-ingressgateway.conf
          name: datakit-conf
          subPath: prom-ingressgateway.conf
        - mountPath: /usr/local/datakit/conf.d/prom/prom-egressgateway.conf
          name: datakit-conf
          subPath: prom-egressgateway.conf
```
       重新部署 DataKit
```
kubectl delete -f datakit.yaml
kubectl apply -f  datakit.yaml
```

5、 访问 bookinfo 

查看 ingresgateway 对外暴露的端口。

![](../imgs/istio-3.png)

浏览器访问 [http://8.136.193.105:32156/productpage](http://8.136.193.105:32156/productpage)，即可访问 productpage。

指标预览

![1649829879(1).png](../imgs/istio-4.png)

#### APM 采集 (必选)

1、开启 Zipkin 采集器

修改 datakit.yaml，通过 ConfigMap 把 zipkin.conf 挂载到 datakit 的 /usr/local/datakit/conf.d/zipkin/zipkin.conf 目录，下面修改 datakit.yaml。

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:    # 下面是新增部分
    zipkin.conf: |-
      [[inputs.zipkin]]
        pathV1 = "/api/v1/spans"
        pathV2 = "/api/v2/spans"
```
```
apiVersion: apps/v1
kind: DaemonSet
...
spec:
  template
    spec:
      containers:
      - env:
        volumeMounts: # 下面是新增部分
        - mountPath: /usr/local/datakit/conf.d/zipkin/zipkin.conf
          name: datakit-conf
          subPath: zipkin.conf
```
```
kubectl delete -f datakit.yaml
kubectl apply -f  datakit.yaml
```
部署完 Istio 后，链路数据会被打到** **zipkin.istio-system的 Service上，且上报端口是 9411。在部署 DataKit 时已开通链路指标采集的 Zipkin 采集器，由于 DataKit 服务的名称空间是 datakit，端口是 9529，所以这里需要做一下转换，详情请参考[Kubernetes 集群使用 ExternalName 映射 DataKit 服务](/best-practices/guance-skill/kubernetes-external-name )。创建后的 Service 如下图：

![1650367040(1).png](../imgs/istio-5.png)

链路预览

![](../imgs/istio-6.png)

#### 日志采集 (非必选)

DataKit 默认的配置，采集容器输出到 /dev/stdout 的日志。更多关于日志的配置，请参考文章末尾的**进一步阅读**内容。

日志预览

![1649829817(1).png](../imgs/istio-7.png)

#### 插件标签 (非必选)

暂无

## 场景视图

场景 - 新建仪表板 - Istio Workload 监控视图

场景 - 新建仪表板 - Istio Control Plane 监控视图

场景 - 新建仪表板 - Istio Mesh 监控视图

场景 - 新建仪表板 - Istio Service 监控视图


## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```


## 指标详解

| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| istio_agent_process_virtual_memory_bytes | Virtual memory size in bytes | int | B |
| istio_agent_go_memstats_alloc_bytes | Number of bytes allocated and still in use. | int | B |
| istio_agent_go_memstats_heap_inuse_bytes | Number of heap bytes that are in use. | int | B |
| istio_agent_go_memstats_stack_inuse_bytes | Number of bytes in use by the stack allocator. | int | B |
| istio_agent_go_memstats_last_gc_time_seconds | Number of seconds since 1970 of last garbage collection | int | s |
| istio_agent_go_memstats_next_gc_bytes | Number of heap bytes when next garbage collection will take place. | int | B |
| istio_agent_process_cpu_seconds_total | Total user and system CPU time spent in seconds. | int | count |
| istio_agent_outgoing_latency | The latency of outgoing requests (e.g. to a token exchange server, CA, etc.) in milliseconds. | int | count |
| istio_requests_total | requests total. | int | <br /> |
| istio_agent_pilot_xds | Number of endpoints connected to this pilot using XDS. | int | count |
| istio_agent_pilot_xds_pushes | Pilot build and send errors for lds, rds, cds and eds. | int | count |
| istio_agent_pilot_xds_expired_nonce | Total number of XDS requests with an expired nonce. | int | count |
| istio_agent_pilot_push_triggers | Total number of times a push was triggered, labeled by reason for the push. | int | count |
| istio_agent_pilot_endpoint_not_ready | Endpoint found in unready state. | int | count |
| envoy_cluster_upstream_cx_total | envoy cluster upstream cx total | int | count |
| istio_agent_pilot_conflict_inbound_listener | Number of conflicting inbound listeners | int | count |
| istio_agent_pilot_conflict_outbound_listener_http_over_current_tcp | Number of conflicting wildcard http listeners with current wildcard tcp listener. | int | count |
| istio_agent_pilot_conflict_outbound_listener_tcp_over_current_tcp | Number of conflicting tcp listeners with current tcp listener. | int | count |
| istio_agent_pilot_conflict_outbound_listener_tcp_over_current_http | Number of conflicting wildcard tcp listeners with current wildcard http listener. | int | count |


## 常见问题排查

- [无数据上报排查](why-no-data.md)

## 进一步阅读

- [基于 Istio 实现微服务可观测最佳实践](/best-practices/cloud-native/istio.md)

- [Pod 日志采集最佳实践](/best-practices/logs/pod-log.md)

- [Kubernetes 集群中日志采集的几种玩法](/best-practices/logs/k8s-logs.md)
