{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit DaemonSet 部署最佳实践

由于 [Datakit DaemonSet](datakit-daemonset-deploy) 配置管理非常复杂，此篇文章将介绍配置管理最佳实践。
本篇将描述2种不同部署的方式的配置方法:

- ComfigMap 管理配置

- 启用 git 管理配置

  

## ComfigMap 管理配置

Datakit 部分采集器的开启，可以通过 [ConfigMap](https://kubernetes.io/zh/docs/concepts/configuration/configmap/) 来注入。ComfigMap 注入灵活，但不易管理。

可以分为：

- Helm 安装注入
- yaml 安装注入

以下是 MySQL 和 Java Pipeline 的注入示例

### Helm 安装注入

#### 前提条件

- Kubernetes >= 1.14 

- Helm >= 2.17.0 

  

#### 添加 Helm 仓库

```shell
helm repo add dataflux  https://pubrepo.guance.com/chartrepo/datakit
```



#### 查看 DataKit 版本

```shell
helm search repo datakit
NAME                	CHART VERSION	APP VERSION	DESCRIPTION
dataflux/datakit	1.2.10       	1.2.10     	Chart for the DaemonSet datakit
```



#### 下载 Helm 包

```shell
helm repo update 
helm pull dataflux/datakit --untar
```



#### 修改 values.yaml 配置

修改 datakit/values.yaml 

注意 yaml 格式，dataway_url 和 dkconfig 都要改

<font color=#FF0000 > values.yaml 可以用于下次升级使用 </font>


```yaml
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



#### 安装或升级 DataKit

安装

```shell
cd datakit
helm repo update 
helm install my-datakit dataflux/datakit -f values.yaml -n datakit  --create-namespace 
```

升级

```shell
helm repo update 
helm upgrade my-datakit . -n datakit  -f values.yaml

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



#### 查看是否部署成功

```shell
helm list -n datakit

kubectl get pods -n datakit
```


### yaml 安装注入 

可参考 [DaemonSet 安装](datakit-daemonset-deploy####ConfigMap 设置)



## 启用 git 管理配置

​	由于 ComfigMap 注入灵活，但不易管理特性，我们可以采用 git 仓库来管理我们的配置。启用 [git 管理](datakit-conf###使用 Git 管理 DataKit 配置) ，DataKit 会定时 pull 远程仓库的配置，既不需要频繁修改ComfigMap，也不需要重启DataKit，更重要的是有修改记录，可回滚配置。

### 前提条件

- 已经准备 git 仓库

- 写入以下配置并 push 远程仓库

  `conf.d/mysql.conf`

  ```toml
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
  ```

  `conf.d/java.p`

  ```
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

  

- 目录结构为：
```shell
├── README.md
├── conf.d
│   ├── mysql.conf
└── pipeline
    ├── java.p
```



### Helm 启用 git 管理配置

​	使用Helm 启用 git 管理配置，一个命令就能完成安装和配置，*简单高效*。

注意：启用 git 管理配置，则 ComfigMap 将失效，default_enabled_inputs 不会影响

- 使用密码管理 git

需要修改 dataway_url，git_repos.git_url
```shell
helm repo add dataflux  https://pubrepo.guance.com/chartrepo/datakit

helm repo update 

helm install my-datakit dataflux/datakit -n datakit --set dataway_url="https://openway.guance.com?token=<your-token>" \
--set git_repos.git_url="http://username:password@github.com/path/to/repository.git" \
--create-namespace 
```

- 使用 git key 管理 git

需要修改 dataway_url，git_repos.git_url，git_repos.git_key_path(绝对路径)
```shell
helm repo add dataflux  https://pubrepo.guance.com/chartrepo/datakit
helm repo update 
helm install my-datakit dataflux/datakit -n datakit --set dataway_url="https://openway.guance.com?token=<your-token>" \
--set git_repos.git_url="git@github.com:path/to/repository.git" \
--set-file git_repos.git_key_path="/Users/buleleaf/.ssh/id_rsa" \
--create-namespace 
```


### yaml 启用 git 管理配置

yaml 配置复杂，建议使用 [Helm 部署](###Helm 启用 git 管理配置)

先下载 [datakit.yaml](https://static.guance.com/datakit/datakit.yaml)

- 使用密码管理 git

修改 datakit.yaml，添加 env
```shell
        - name: DK_GIT_URL
          value: "http://username:password@github.com/path/to/repository.git"
        - name: DK_GIT_BRANCH
          value: "master"
        - name: DK_GIT_INTERVAL
          value: "1m"
```

- 使用 git key 管理 git

添加 ComfigMap
```shell
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
添加 env
```shell
        - name: DK_GIT_URL
          value: "git@github.com:path/to/repository.git"
        - name: DK_GIT_KEY_PATH
          value: "/usr/local/datakit/id_rsa"
        - name: DK_GIT_BRANCH
          value: "master"
        - name: DK_GIT_INTERVAL
          value: "1m"
```
添加 volume
```shell
        volumeMounts:
        - mountPath: /usr/local/datakit/id_rsa
          name: id-rsa
      volumes:
      - configMap:
          defaultMode: 420
          name: id-rsa
        name: id-rsa
```



### 验证是否部署成功

登录容器查看 `/usr/local/datakit/gitrepos` 目录是否同步成功

```
kubectl exec -ti datakit-xxxx bash
ls gitrepos
```

