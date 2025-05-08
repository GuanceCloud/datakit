
# 使用 Helm 管理配置
---

本文介绍如何使用 helm 来管理 DataKit 的环境变量和采集配置。我们可以通过维护 helm 管理 DataKit 的配置变更。

## 安装和修改配置 {#instal-config}

### helm 下载 DataKit Charts 包 {#dowbload-config}

```shell
helm pull datakit --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit --untar
```

### 修改 values.yaml {#values-configuration}

<!-- markdownlint-disable MD046 -->
???+ warning "Attention"

     `values.yaml` 在 `datakit` 目录下。
<!-- markdownlint-enable -->

#### 修改 `dataway url`  {#helm-dataway}

```yaml
...
datakit:
  # Datakit will send the indicator data to dataway. Please be sure to change the parameters
  # @param dataway_url - string - optional - default: 'https://<<<custom_key.brand_main_domain>>>'
  # The host of the DataKit intake server to send Agent data to, only set this option
  dataway_url: https://openway.<<<custom_key.brand_main_domain>>>?token=tkn_xxxxxxxxxx
...
```

#### 添加默认采集器  {#helm-default-config}
  
添加 `rum`，在 `default_enabled_inputs` 最后追加参数。

```yaml
..
datakit:
  ...
  # @param default_enabled_inputs - string
  # The default open collector list, format example: input1, input2, input3
  default_enabled_inputs: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,rum
....
```

#### 添加全局 tag {#helm-tag}

添加 `cluster_name_k8s` 全局 tag。

```yaml
datakit:
  ...
  # @param global_tags - string - optional - default: 'host=__datakit_hostname,host_ip=__datakit_ip'
  # It supports filling in global tags in the installation phase. The format example is: Project = ABC, owner = Zhang San (multiple tags are separated by English commas)
  global_tags: host=__datakit_hostname,host_ip=__datakit_ip,cluster_name_k8s=prod  
```

#### 添加 DataKit 环境变量 {#helm-env}

更多环境变量可参考[容器环境变量](datakit-daemonset-deploy.md#using-k8-env)

```yaml
# @param extraEnvs - array - optional
# extra env Add env for customization
# more, see: https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env
# You can add more than one parameter  
extraEnvs:
 - name: ENV_NAMESPACE
   value: government-prod
 - name: ENV_GLOBAL_ELECTION_TAGS
   value: cluster_name_k8s=government-prod
```

#### 挂载采集器配置 {#helm-config}
  
以采集容器主机系统日志为例，`path` 为容器路径，必须在 `/usr/local/datakit/conf.d/` 下。`name` 为配置名称。`value` 为采集配置内容。采集器的 sample 文件，您可以进入容器的 `/usr/local/datakit/conf.d/` 目录下获取。

```yaml
dkconfig:   
 - path: "/usr/local/datakit/conf.d/logging.conf"
   name: logging.conf
   value: |-
     [[inputs.logging]]
       logfiles = [
         "/var/log/syslog",
         "/var/log/message",
       ]
       ignore = [""]
       source = ""
       service = ""
       pipeline = ""
       ignore_status = []
       character_encoding = ""
       auto_multiline_detection = true
       auto_multiline_extra_patterns = []
       remove_ansi_escape_codes = true
       blocking_mode = true
       ignore_dead_log = "1h"
       [inputs.logging.tags]
```

#### 挂载 Pipeline  {#helm-pipeline}

以 `test.p` 为例，`path` 为配置文件绝对路径，必须在 */usr/local/datakit/pipeline/* 下。`name` 为 Pipeline 名称。`value` 为 Pipeline 内容。

```yaml
dkconfig:
 - path: "/usr/local/datakit/pipeline/test.p"
   name: test.p
   value: |-
     # access log
     grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{NUMBER:http_code} ")
     grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"-\" %{NUMBER:http_code} ")
     default_time(time)
     cast(http_code,"int")

     # error log
     grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{GREEDYDATA:pid}:tid %{GREEDYDATA:tid}\\] ")
     grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{INT:pid}\\] ")
     default_time(time)
```

### 安装 DataKit {#datakit-install}

```shell
helm install datakit datakit \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit --create-namespace \
         -f values.yaml
```

输出结果：

```shell
NAME: datakit
LAST DEPLOYED: Tue Apr  4 19:13:29 2023
NAMESPACE: datakit
STATUS: deployed
REVISION: 1
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace datakit -l "app.kubernetes.io/name=datakit,app.kubernetes.io/instance=datakit" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace datakit $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  echo "Visit http://127.0.0.1:9527 to use your application"
  kubectl --namespace datakit port-forward $POD_NAME 9527:$CONTAINER_PORT
```

## 指定版本安装 {#version-install}

```shell
helm install datakit datakit \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit --create-namespace \
         -f values.yaml \
         --version 1.5.x
```

## 升级 {#datakit-upgrade}

<!-- markdownlint-disable MD046 -->
???+ info

    如果 values.yaml 丢失，可执行 `helm -n datakit get  values datakit -o yaml > values.yaml` 获取。
<!-- markdownlint-enable -->

```shell
helm upgrade datakit datakit \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit \
         -f values.yaml
```

## 卸载 {#datakit-uninstall}

```shell
helm uninstall datakit -n datakit 
```

## 配置文件参考 {#config-reference}

<!-- markdownlint-disable MD046 -->
???- note "values.yaml"

    ```yaml
    # Default values for datakit.
    # This is a YAML-formatted file.
    # Declare variables to be passed into your templates.

    datakit:
      # Datakit will send the indicator data to dataway. Please be sure to change the parameters
      # @param dataway_url - string - optional - default: 'https://<<<custom_key.brand_main_domain>>>'
      # The host of the DataKit intake server to send Agent data to, only set this option
      dataway_url: https://openway.<<<custom_key.brand_main_domain>>>?token=tkn_xxxxxxxxxx

      # @param global_tags - string - optional - default: 'host=__datakit_hostname,host_ip=__datakit_ip'
      # It supports filling in global tags in the installation phase. The format example is: Project = ABC, owner = Zhang San (multiple tags are separated by English commas)
      global_tags: host=__datakit_hostname,host_ip=__datakit_ip,cluster_name_k8s=government-prod

      # @param default_enabled_inputs - string
      # The default open collector list, format example: input1, input2, input3
      default_enabled_inputs: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,rum

      # @param enabled_election - boolean
      # When the election is enabled, it is enabled by default. If it needs to be enabled, you can give any non empty string value to the environment variable. (e.g. true / false)
      enabled_election: true

      # @param log - string
      # Set logging verbosity, valid log levels are:
      # info, debug, stdout, warn, error, critical, and off
      log_level: info

      # @param http_listen - string
      # It supports specifying the network card bound to the Datakit HTTP service in the installation phase (default localhost)
      http_listen: 0.0.0.0:9529

    image:
      # @param repository - string - required
      # Define the repository to use:
      #
      repository:  pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit

      # @param tag - string - required
      # Define the Cluster-Agent version to use.
      #
      tag: ""

      # @param pullPolicy - string - optional
      # The Kubernetes [imagePullPolicy][] value
      #
      pullPolicy: Always

    # https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/

    git_repos:
      # use git management DataKit input
      enable: false

      # @param git_url - string - required
      # You Can Set git@github.com:path/to/repository.git or http://username:password@github.com/path/to/repository.git.
      # see https://docs.<<<custom_key.brand_main_domain>>>/best-practices/insight/datakit-daemonset/#git
      git_url: "-"

      # @param git_key_path - string - optional
      # The Git Ssh Key Content,
      # For details,
      # -----BEGIN OPENSSH PRIVATE KEY--
      # ---xxxxx---
      #--END OPENSSH PRIVATE KEY-----
      git_key_path: "-"

      # @param git_key_pw - string - optional
      # The ssh Key Password
      git_key_pw: "-"

      # @param git_url - string - required
      # Specifies the branch to pull. If it is blank, it is the default. The default is the main branch specified remotely, usually the master.
      git_branch: "master"

      # @param git_url - string - required
      # Timed pull interval. (e.g. 1m)
      git_interval: "1m"
      is_use_key: false

    # If true, Datakit install ipdb.
    # ref: https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-tools-how-to/#install-ipdb
    iploc:
      enable: true
      image:
        # @param repository - string - required
        # Define the repository to use:
        #
        repository: "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/iploc"

        # @param tag - string - required
        # Define the Cluster-Agent version to use.
        #
        tag: "1.0"

    # @param extraEnvs - array - optional
    # extra env Add env for customization
    # more, see: https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env
    # You can add more than one parameter
    extraEnvs:
     - name: ENV_NAMESPACE
       value: government-prod
     - name: ENV_GLOBAL_ELECTION_TAGS
       value: cluster_name_k8s=government-prod
     # - name: ENV_NAMESPACE # electoral
     #   value: k8s
     # - name: "NODE_OPTIONS"
     #   value: "--max-old-space-size=1800"


    resources:
      requests:
        cpu: "200m"
        memory: "128Mi" 
      limits:
        cpu: "2000m"
        memory: "4Gi"

    # @param nameOverride - string - optional
    # Override name of app.
    #
    nameOverride: ""

    # @param fullnameOverride - string - optional
    # Override name of app.
    #
    fullnameOverride: ""

    podAnnotations:
      datakit/logs: |
        [{"disable": true}]

    # @param tolerations - array - optional
    # Allow the DaemonSet to schedule on tainted nodes (requires Kubernetes >= 1.6)
    #
    tolerations:
      - operator: Exists

    service:
      type: ClusterIP
      port: 9529

    # @param dkconfig - array - optional
    # Configure Datakit custom input
    #
    dkconfig: 
     - path: "/usr/local/datakit/conf.d/logging.conf"
       name: logging.conf
       value: |-
         [[inputs.logging]]
           logfiles = [
             "/var/log/syslog",
             "/var/log/message",
           ]
           ignore = [""]
           source = ""
           service = ""
           pipeline = ""
           ignore_status = []
           character_encoding = ""
           auto_multiline_detection = true
           auto_multiline_extra_patterns = []
           remove_ansi_escape_codes = true
           blocking_mode = true
           ignore_dead_log = "1h"
           [inputs.logging.tags]
     - path: "/usr/local/datakit/pipeline/test.p"
       name: test.p
       value: |-
         # access log
         grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{NUMBER:http_code} ")
         grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"-\" %{NUMBER:http_code} ")
         default_time(time)
         cast(http_code,"int")

         # error log
         grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{GREEDYDATA:pid}:tid %{GREEDYDATA:tid}\\] ")
         grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{INT:pid}\\] ")

    # If true, deploys the kube-state-metrics deployment.
    # ref: https://github.com/kubernetes/charts/tree/master/stable/kube-state-metrics
    kubeStateMetricsEnabled: true

    # If true, deploys the metrics-server deployment.
    # ref: https://github.com/kubernetes-sigs/metrics-server/tree/master/charts/metrics-server
    MetricsServerEnabled: false
    ```
<!-- markdownlint-enable -->

## FAQ {#faq}

### PodSecurityPolicy 问题 {#pod-security-policy}

`PodSecurityPolicy` 已在 [Kubernetes`1.21`](https://kubernetes.io/blog/2021/04/06/podsecuritypolicy-deprecation-past-present-and-future/){:target="_blank"} 中弃用，并且已在 Kubernetes`1.25` 中移除。
如果强行升级集群版本，Helm 部署 `kube-state-metrics` 会报错：

```shell
Error: UPGRADE FAILED: current release manifest 
contains removed kubernetes api(s) for this kubernetes
version and it is therefore unable to build the
kubernetes objects for performing the diff. error from
kubernetes: unable to recognize "": no matches for kind
"PodSecurityPolicy" in version "policy/v1beta1"
```

#### 备份 Helm values {#get-values}

```shell
helm get values -n datakit datakit -o yaml > values.yaml
```

#### 清空 Helm 信息 {#delete-values}

删除 Datakit namespace 的 secrets Helm 信息。

- 获取 secrets

  ```shell
  $ kubectl get secrets -n datakit
  NAME                            TYPE                 DATA   AGE
  sh.helm.release.v1.datakit.v1   helm.sh/release.v1   1      4h17m
  sh.helm.release.v1.datakit.v2   helm.sh/release.v1   1      4h17m
  sh.helm.release.v1.datakit.v3   helm.sh/release.v1   1      4h16m
  ```

- 删除带有 `sh.helm.release.v1.datakit` 的 secrets

  ```shell
  kubectl delete  secrets sh.helm.release.v1.datakit.v1 sh.helm.release.v1.datakit.v2 sh.helm.release.v1.datakit.v3   -n datakit
  ```

#### 重新升级或安装 {#reinstall}

```shell
helm upgrade -i -n datakit datakit  --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit  -f values.yaml
```

