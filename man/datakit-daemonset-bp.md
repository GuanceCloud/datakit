{{.CSS}}
# DataKit DaemonSet 部署最佳实践
---

由于 [Datakit DaemonSet](../datakit/datakit-daemonset-deploy.md) 配置管理非常复杂，此篇文章将介绍配置管理最佳实践。本篇将以配置 MySQL 和 Java Pipeline 为演示案例。

本篇将描述以下2种不同的管理方法:

- ConfigMap 管理配置
- 启用 Git 管理配置

## ConfigMap 管理配置 {#configmap}

Datakit 部分采集器的开启，可以通过 [ConfigMap](https://kubernetes.io/zh/docs/concepts/configuration/configmap/){:target="_blank"} 来注入。ConfigMap 注入灵活，但不易管理。

ConfigMap 注入，可以分为以下几种：

- Helm 安装注入
- yaml 安装注入

### Helm 安装 {#helm-install}

前提条件

- Kubernetes 1.14+
- Helm 3.0+

添加 Helm 仓库

```shell
$ helm repo add datakit  https://pubrepo.guance.com/chartrepo/datakit
``` 

查看 DataKit 版本

```shell
$ helm search repo datakit
NAME                	CHART VERSION	APP VERSION	DESCRIPTION
datakit/datakit	1.2.10       	1.2.10     	Chart for the DaemonSet datakit
```

下载 Helm 包

```shell
$ helm repo update 
$ helm pull datakit/datakit --untar
``` 

修改 values.yaml 配置

修改 `datakit/values.yaml` 的 `datakit.dataway_url` 和 `dkconfig`数组。 `datakit.dataway_url` 为 dataway 地址， `dkconfig.path` 为挂载路径， `dkconfig.name` 为配置名称， ` dkconfig.value` 为配置内容。

> 注：`values.yaml` 可以用于下次升级使用

```yaml
datakit:
  dataway_url: https://openway.guance.com?token=<your-token>

... 
dkconfig:
  - path: "/usr/local/datakit/conf.d/db/mysql.conf"
    name: mysql.conf
    value: |
      # {"version": "1.1.9-rc7.1", "desc": "do NOT edit this line"}
      [[inputs.mysql]]
        host = "xxx.xxx.xxx.xxx"
        user = "xxxx"
        pass = "xxxxxxxx"
        port = 3306
        interval = "10s"
        innodb = true
        tables = []
        users = []
        [inputs.mysql.dbm_metric]
          enabled = true
        ## 监控采样配置
        [inputs.mysql.dbm_sample]
          enabled = true
        [inputs.mysql.tags]
          # some_tag = "some_value"
          # more_tag = "some_other_value"
  - path: "/usr/local/datakit/pipeline/java.p"
    name: java.p
    value: |
      json(_, recorder, 'recorder')
      if recorder == "gunicorn" {
        drop_key(func_name)
        json(_, func_name, 'func_name')
        json(_, remote_addr, 'remote_addr')
        json(_, time_local, 'time_local')
        json(_, time_unix, 'time_unix')
        json(_, level, 'level')
        json(_, method, 'method')
        json(_, path, 'path')
        json(_, cost_time, 'cost_time')
        json(_, res_status, 'res_status')
        json(_, res_length, 'res_length')
        json(_, trace_id, 'trace_id')
        add_key(__type, "gunicorn")
      }
      if recorder == "Forethought" {
        json(_, source, 'resource')
        json(_, app, 'app')
        json(_, level, 'level')
        json(_, cc_timestamp, 'cc_timestamp')
        json(_, name, 'name')
        json(_, log_time, 'log_time')
        json(_, message, 'remessage')
              json(_, trace_id, 'trace_id')
        add_key(__type, "Forethought")
      }
      lowercase(level)
      group_in(level, ["error", "panic", "dpanic", "fatal"], "error", status)
      group_in(level, ["info", "debug"], "info", status)
      group_in(level, ["warn", "warning"], "warning", status)

```

#### 安装或升级 DataKit {#install-upgrade}

安装

```shell
$ cd datakit # 此目录为 helm pull 的目录
$ helm repo update 
$ helm install datakit datakit/datakit -f values.yaml -n datakit  --create-namespace 
```

升级

```shell
$ helm repo update 
$ helm upgrade datakit . -n datakit  -f values.yaml

Release "datakit" has been upgraded. Happy Helming!
NAME: datakit
LAST DEPLOYED: Sat Apr  2 15:33:55 2022
NAMESPACE: datakit
STATUS: deployed
REVISION: 10
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace datakit -l "app.kubernetes.io/name=datakit,app.kubernetes.io/instance=datakit" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace datakit $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  echo "Visit http://127.0.0.1:9527 to use your application"
  kubectl --namespace datakit port-forward $POD_NAME 9527:$CONTAINER_PORT

```

查看是否部署成功

```shell
$ helm list -n datakit
$ kubectl get pods -n datakit
```

### yaml 安装注入 {#yaml-install}

可参考 [DaemonSet ConfigMap 设置](../datakit/datakit-daemonset-deploy.md#configmap-setting)

## 启用 Git 管理配置 {#enable-git}

由于 ConfigMap 注入灵活，但不易管理特性，我们可以采用 Git 仓库来管理我们的配置。启用 [Git 管理](../datakit/datakit-conf.md#using-gitrepo) ，DataKit 会定时 pull 远程仓库的配置，既不需要频繁修改 ConfigMap，也不需要重启 DataKit，更重要的是有修改记录，可回滚配置。

> 注意：
> - 如果启用 Git 管理配置，则 ConfigMap 将失效
> - 由于会[自动启动一些采集器](../datakit/datakit-input-conf.md#default-enabled-inputs)，故在 Git 仓库中，不要再放置这些自启动的采集器配置，不然会导致这些数据的多份采集

### 前提条件 {#git-requirements}

- 已经准备 Git 仓库
- 写入配置并 push 远程仓库

以 MySQL 采集器 mysql.conf 和 Java 日志的 Pipeline 脚本（java.p）为例

> 参见 [Git 仓库中目录结构约束](../datakit/datakit-conf.md#gitrepo-limitation)

Git 仓库目录结构为：

```
path/to/local/git/repo
  ├── README.md
  ├── conf.d
  │   ├── mysql.conf
  └── pipeline
      ├── java.p
``` 

### Helm 启用 Git 管理配置 {#git-helm}

使用 Helm 启用 Git 管理配置，一个命令就能完成安装和配置，简单高效。

#### 使用密码管理 Git {#git-pwd}

需要修改如下两个字段：

- `datakit.dataway_url`
- `git_repos.git_url`


```shell
$ helm repo add datakit  https://pubrepo.guance.com/chartrepo/datakit
$ helm repo update 
$ helm install datakit datakit/datakit -n datakit --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" \
--set git_repos.git_url="http://username:password@github.com/path/to/repository.git" \
--create-namespace 
```

#### Helm 中使用用户名密码访问 Git {#helm-git-pwd} 

需要修改 

- `datakit.dataway_url`
- `git_repos.git_url`
- `git_repos.git_key_path`（绝对路径）

```shell
$ helm repo add datakit  https://pubrepo.guance.com/chartrepo/datakit
$ helm repo update 
$ helm install datakit datakit/datakit -n datakit \
  --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" \
  --set git_repos.git_url="git@github.com:path/to/repository.git" \
  --set-file git_repos.git_key_path="/Users/buleleaf/.ssh/id_rsa" \
  --create-namespace 
``` 

### yaml 启用 Git 管理配置 {#yaml-git-pwd}

yaml 配置复杂，建议使用 [Helm 部署](#helm-install)。先下载 [datakit.yaml](https://static.guance.com/datakit/datakit.yaml){:target="_blank"}

#### 使用密码管理 Git

修改 datakit.yaml，添加各种 env 配置：

```yaml
        - name: DK_GIT_URL
          value: "http://username:password@github.com/path/to/repository.git"
        - name: DK_GIT_BRANCH
          value: "master"
        - name: DK_GIT_INTERVAL
          value: "1m"
```

安装 yaml

```shell
$ kubectl apply -f datakit.yaml
```

#### 使用 SSH Key 访问 Git

- 在 *datakit.yaml* 中添加 ConfigMap

```yaml
apiVersion: v1
data:
  id-rsa: |-
    -----BEGIN OPENSSH PRIVATE KEY-----
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxr
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxr
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxr
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxr
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    -----END OPENSSH PRIVATE KEY-----
kind: ConfigMap
metadata:
  name: id-rsa
  namespace: datakit
```

- 添加 env 配置：

```yaml
        - name: DK_GIT_URL
          value: "git@github.com:path/to/repository.git"
        - name: DK_GIT_KEY_PATH
          value: "/usr/local/datakit/id_rsa"
        - name: DK_GIT_BRANCH
          value: "master"
        - name: DK_GIT_INTERVAL
          value: "1m"
```

- 添加 volume 挂载

```yaml
        volumeMounts:
        - mountPath: /usr/local/datakit/id_rsa
          name: id-rsa
      volumes:
      - configMap:
          defaultMode: 420
          name: id-rsa
        name: id-rsa
```

启用 yaml：

```shell
$ kubectl apply -f datakit.yaml
```

验证是否部署成功，登录容器查看 `/usr/local/datakit/gitrepos` 目录是否同步成功

```shell
$ kubectl exec -ti datakit-xxxx bash
$ ls gitrepos
```

## 更多阅读 {#more-readings}

- [DataKit 采集器配置](../datakit/datakit-input-conf.md)
- [DataKit 主配置](../datakit/datakit-conf.md)
- [DataKit Daemonset 部署](../datakit/datakit-daemonset-deploy.md)
