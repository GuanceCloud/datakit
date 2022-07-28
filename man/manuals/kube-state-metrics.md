# Kube State Metrics
---


## 视图预览
Kubernetes 性能指标展示：Pod desired、Pod desired、Pods ready、Pods Available、Pod Unavailable等。<br />
![1649841130(1).png](imgs/input-kube-state-metrics-01.png)<br />
![1649841147(1).png](imgs/input-kube-state-metrics-02.png)<br />
![1649841167(1).png](imgs/input-kube-state-metrics-03.png)<br />
![1649841184(1).png](imgs/input-kube-state-metrics-04.png)<br />
![1649841202(1).png](imgs/input-kube-state-metrics-05.png)<br />
![1649841214(1).png](imgs/input-kube-state-metrics-06.png)

## 版本支持
kube-state-metrics/ Kubernetes<br />![1649843207(1).png](imgs/input-kube-state-metrics-07.png)

## 前置条件

- Kubernetes 集群 <[安装 Datakit](https://www.yuque.com/dataflux/integrations/kubernetes)>
- 下载 [kube-state-metrics-2.3.0.zip](https://github.com/kubernetes/kube-state-metrics/releases/tag/v2.3.0)

## 安装配置
说明：示例 kube-state-metrics 版本为 2.3.0，Kubernetes 集群 1.22.6。

### 部署实施
(Kubernetes 集群)

#### 指标采集 (必选)

1. 修改镜像

      解压 kube-state-metrics-2.3.0.zip，部署使用 kube-state-metrics-2.3.0\examples\standard\ 目录下的文件，打开 deployment.yaml 修改 k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.3.0 为 bitnami/kube-state-metrics:2.3.0。<br />注意：如果原始镜像可访问，不必修改。

2. 开通 prom 采集器

       打开 deployment.yaml 增加 annotations
```
spec:
  template:
    metadata:
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            urls = ["http://$IP:8080/metrics"]
            source = "prom_state_metrics"
            metric_types = ["counter", "gauge"]
            interval = "30s"
            tags_ignore = ["access_mode","branch","claim_namespace","cluster_ip","condition","configmap","container","container_id","container_runtime_version","created_by_kind","created_by_name","effect","endpoint","external_name","goversion","host_network","image","image_id","image_spec","ingress","ingressclass","internal_ip","job_name","kernel_version","key","kubelet_version","kubeproxy_version","lease","mutatingwebhookconfiguration","name","networkpolicy","node","node_name","os_image","owner_is_controller","owner_kind","owner_name","path","persistentvolume","persistentvolumeclaim","pod_cidr","pod_ip","poddisruptionbudget","port_name","port_number","port_protocol","priority_class","reason","resource","result","revision","role","secret","service","service_name","service_port","shard_ordinal","status","storageclass","system_uuid","type","uid","unit","version","volume","volumename"]
            metric_name_filter = ["kube_pod_status_phase","kube_pod_container_status_restarts_total","kube_daemonset_status_desired_number_scheduled","kube_daemonset_status_number_ready","kube_deployment_spec_replicas","kube_deployment_status_replicas_available","kube_deployment_status_replicas_unavailable","kube_replicaset_status_ready_replicas","kube_replicaset_spec_replicas","kube_pod_container_status_running","kube_pod_container_status_waiting","kube_pod_container_status_terminated","kube_pod_container_status_ready"]
            #measurement_prefix = ""
            measurement_name = "prom_state_metrics"
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"         
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
            pod_name = "$PODNAME"
```

3. 部署
```
kubectl apply -f service-account.yaml
kubectl apply -f cluster-role.yaml  
kubectl apply -f cluster-role-binding.yaml  
kubectl apply -f deployment.yaml 
```

4. 查看监控数据 
```
kubectl get pods -n kube-system -owide
```
![1649843621(1).png](imgs/input-kube-state-metrics-08.png)
```
curl http://10.244.36.66:8080/metrics 
```
![1649843709(1).png](imgs/input-kube-state-metrics-09.png)<br />指标预览<br />
![1649849431(1).png](imgs/input-kube-state-metrics-10.png)

#### 插件标签 (非必选）
暂无

## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Kubernetes Overview with Kube State Metrics 监控视图><br /><场景 - 新建仪表板 - 内置模板库 - Kubernetes Overview by Pods 监控视图>

## 指标详解
prom_state_metrics：

| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| kube_daemonset_status_desired_number_scheduled | The number of nodes that should be running the daemon pod. | int | count |
| kube_deployment_spec_replicas | Number of desired pods for a deployment. | int | count |
| kube_daemonset_status_number_ready | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready. | int | count |
| kube_deployment_status_replicas_available | The number of available replicas per deployment. | int | count |
| kube_deployment_status_replicas_unavailable | The number of unavailable replicas per deployment. | int | count |
| kube_replicaset_status_ready_replicas | The number of ready replicas per ReplicaSet. | int | count |
| kube_replicaset_spec_replicas | Number of desired pods for a ReplicaSet. | int | count |
| kube_pod_container_status_running | Describes whether the container is currently in running state. | int | count |
| kube_pod_container_status_waiting | Describes whether the container is currently in waiting state. | int | count |
| kube_pod_container_status_terminated | Describes whether the container is currently in terminated state. | int | count |
| kube_pod_container_status_ready | Describes whether the containers readiness check succeeded. | int | count |

system：

| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| load15_per_core | 15分钟负载 | int | count |

net：

| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| bytes_sent | 出流量 | int | B |
| bytes_recv | 入流量 | int | B |

mem：

| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| total | 总内存 | int | B |
| used | 已使用内存 | int | B |
| free | 剩余内存 | int | B |
| cached | 缓冲 | int | B |
| buffered | 缓存 | int | B |

dist：

| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| total | 总磁盘空间 | int | B |
| used | 已使用磁盘空间 | int | B |
| free | 剩余磁盘空间 | int | B |


## 常见问题排查
<[无数据上报排查](https://www.yuque.com/dataflux/datakit/why-no-data)>

## 进一步阅读
暂无
