
{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

DataKit DaemonSet 支持两种方式安装，Helm 部署和 yaml 文件部署。此篇文章将讲述两种部署的升级方式。

## Helm 升级 DataKit

### 前提条件

- DataKit DaemonSet 是 Helm 部署方式部署的 

### 添加 Helm 仓库

直接执行如下命令，添加 Helm 仓库
```shell
$ helm repo add datakit  https://pubrepo.guance.com/chartrepo/datakit
```

### 更新 Helm 仓库

直接执行如下命令，更新 Helm 仓库

```shell
$ helm repo update 
```

### 升级 DataKit 

- 如果没有通过修改 `values.yaml` 安装 DataKit，直接执行如下命令，升级 DataKit 

```shell
helm update <RELEASE_NAME> datakit/datakit -n datakit --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" 
```

- 如果是通过修改 `values.yaml` 安装 DataKit，直接执行如下命令，升级 DataKit 

```shell
helm update <RELEASE_NAME> datakit/datakit -f values.yaml -n datakit 
```

## yaml 文件方式升级 DataKit

先下载 [datakit.yaml](https://static.guance.com/datakit/datakit.yaml)，其中开启了很多[默认采集器](datakit-input-conf#764ffbc2)，无需配置。

> 如果要修改这些采集器的默认配置，可通过 [Configmap 方式挂载单独的 conf](k8s-config-how-to#ebf019c2) 来配置。部分采集器可以直接通过环境变量的方式来调整，具体参见具体采集器的文档（[容器采集器示例](container#5cf8fecf)）。总而言之，不管是默认开启的采集器，还是其它采集器，在 DaemonSet 方式部署 DataKit 时，==通过 [Configmap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) 来配置采集器总是生效的==



### 修改配置

修改 `datakit.yaml` 中的 dataway 配置

```yaml
	- name: ENV_DATAWAY
		value: https://openway.guance.com?token=<your-token> # 此处填上 DataWay 真实地址
```

如果选择的是其它节点，此处更改对应的 DataWay 地址即可，如 AWS 节点：

```yaml
	- name: ENV_DATAWAY
		value: https://aws-openway.guance.com?token=<your-token> 
```



### 升级 DataKit

```shell
kubectl apply -f datakit.yaml
```
