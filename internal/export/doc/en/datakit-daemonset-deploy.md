# Kubernetes
---

This document describes how to install DataKit in K8s via DaemonSet.

## Installation {#install}

<!-- markdownlint-disable MD046 -->
=== "DaemonSet"

    Download [`datakit.yaml`](https://static.guance.com/datakit/datakit.yaml){:target="_blank"}, in which many [default collectors](datakit-input-conf.md#default-enabled-inputs) are turned on without configuration.
    
    ???+ attention
    
        If you want to modify the default configuration of these collectors, you can configure them by [mounting a separate conf in `ConfigMap` mode](k8s-config-how-to.md#via-configmap-conf). Some collectors can be adjusted directly by means of environment variables. See the documents of specific collectors for details. All in all, configuring the collector through [`ConfigMap`](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/){:target="_blank"} is always effective when deploying the DataKit in DaemonSet mode, whether it is a collector turned on by default or other collectors.
    
    Modify the Dataway configuration in `datakit.yaml`
    
    ```yaml
        - name: ENV_DATAWAY
            value: https://openway.guance.com?token=<your-token> # Fill in the real address of DataWay here
    ```
    
    If you choose another node, change the corresponding DataWay address here, such as AWS node:
    
    ```yaml
        - name: ENV_DATAWAY
            value: https://aws-openway.guance.com?token=<your-token> 
    ```
    
    Install yaml
    
    ```shell
    $ kubectl apply -f datakit.yaml
    ```
    
    After installation, a DaemonSet deployment of DataKit is created:
    
    ```shell
    $ kubectl get pod -n datakit
    ```

=== "Helm"

    Precondition:
    
    * Kubernetes >= 1.14
    * Helm >= 3.0+
    
    Helm installs Datakit (note modifying the `datakit.dataway_url` parameter)，in which many [default collectors](datakit-input-conf.md#default-enabled-inputs) are turned on without configuration.
    
    ```shell
    $ helm install datakit datakit \
               --repo  https://pubrepo.guance.com/chartrepo/datakit \
               -n datakit --create-namespace \
               --set datakit.dataway_url="https://openway.guance.com?token=<your-token>" 
    ```
    
    View deployment status:
    
    ```shell
    $ helm -n datakit list
    ```
    
    You can upgrade with the following command:
    
    ```shell
    $ helm -n datakit get  values datakit -o yaml > values.yaml
    $ helm upgrade datakit datakit \
        --repo  https://pubrepo.guance.com/chartrepo/datakit \
        -n datakit \
        -f values.yaml
    ```
    
    You can uninstall it with the following command:
    
    ```shell
    $ helm uninstall datakit -n datakit
    ```
<!-- markdownlint-enable -->

## Kubernetes Tolerance Configuration {#toleration}

DataKit is deployed on all nodes in the Kubernetes cluster by default (that is, all stains are ignored). If some node nodes in Kubernetes have added stain scheduling and do not want to deploy DataKit on them, you can modify `datakit.yaml` to adjust the stain tolerance:

```yaml
      tolerations:
      - operator: Exists    <--- Modify the stain tolerance here
```

For specific bypass strategies, see [official doc](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration){:target="_blank"}。

## ConfigMap Settings {#configmap-setting}

The opening of some collectors needs to be injected through ConfigMap. The following is an injection example of MySQL and Redis collectors:

```yaml
# datakit.yaml

volumeMounts: #  this configuration have existed in datakit.yaml, and you can locate it by searching directly
- mountPath: /usr/local/datakit/conf.d/db/mysql.conf
  name: datakit-conf
  subPath: mysql.conf
    readOnly: true
- mountPath: /usr/local/datakit/conf.d/db/redis.conf
  name: datakit-conf
  subPath: redis.conf
    readOnly: true

# append directly to the bottom of datakit.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:
    mysql.conf: |-
          [[inputs.mysql]]
               ...
    redis.conf: |-
          [[inputs.redis]]
               ...
```

## Environment Variables {#using-k8-env}

> Note: If ENV_LOG is configured to `stdout`, do not set ENV_LOG_LEVEL to `debug`, otherwise looping logs may result in large amounts of log data.

In DaemonSet mode, DataKit supports multiple environment variable configurations.

- The approximate format in `datakit.yaml` is

```yaml
spec:
  containers:
    - env
    - name: ENV_XXX
      value: YYY
    - name: ENV_OTHER_XXX
      value: YYY
```

- The approximate format in Helm values.yaml is

```yaml
  extraEnvs: 
    - name: "ENV_XXX"
      value: "YYY"
    - name: "ENV_OTHER_XXX"
      value: "YYY"    
```

## ENV Set Collectors: {#env-setting}

The opening of some collectors can also be injected through ENV_DATAKIT_INPUTS.
The following is an injection example of MySQL and Redis collectors:

- The approximate format in `datakit.yaml` is

```yaml
spec:
  containers:
    - env
    - name: ENV_XXX
      value: YYY
    - name: ENV_DATAKIT_INPUTS
      value: |
        [[inputs.mysql]]
          interval = "10s"
          ...
        [inputs.mysql.tags]
          some_tag = "some_value"

        [[inputs.redis]]
          interval = "10s"
          ...
        [inputs.redis.tags]
          some_tag = "some_value"
```

- The approximate format in Helm values.yaml is

```yaml
  extraEnvs: 
    - name: "ENV_XXX"
      value: "YYY"
    - name: "ENV_DATAKIT_INPUTS"
      value: |
        [[inputs.mysql]]
          interval = "10s"
          ...
        [inputs.mysql.tags]
          some_tag = "some_value"

        [[inputs.redis]]
          interval = "10s"
          ...
        [inputs.redis.tags]
          some_tag = "some_value"
```

The injected content will be stored in the conf.d/env_datakit_inputs.conf file of the container.

### Description of Environment Variable Type {#env-types}

The values of the following environment variables are divided into the following data types:

- string: string type
- JSON: some of the more complex configurations that require setting environment variables in the form of a JSON string
- bool: switch type. Given **any non-empty string** , this function is turned on. It is recommended to use `"on"` as its value when turned on. If it is not opened, it must be deleted or commented out.
- string-list: a string separated by an English comma, commonly used to represent a list
- duration: a string representation of the length of time, such as `10s` for 10 seconds, where the unit supports h/m/s/ms/us/ns. **Don't give a negative value**.
- int: integer type
- float: floating point type

For string/bool/string-list/duration, it is recommended to use double quotation marks to avoid possible problems caused by k8s parsing yaml.

### Most Commonly Used Environment Variables {#env-common}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envCommon 0}}
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ note "Distinguish between *global host tag*  and *global election tag*"

    `ENV_GLOBAL_HOST_TAGS` is used to specify host class global tags whose values generally follow host transitions, such as host name and host IP. Of course, other tags that do not follow the host changes can also be added. All collectors of non-elective classes are taken by default with the tag specified in `ENV_GLOBAL_HOST_TAGS`.
    
    And `ENV_GLOBAL_ELECTION_TAGS` recommends adding only tags that do not change with host switching, such as cluster name, project name, etc. For [election collector](election.md#inputs), only the tag specified in `ENV_GLOBAL_ELECTION_TAGS` will be added, not the tag specified in `ENV_GLOBAL_HOST_TAGS`.
    
    Whether it is a host class global tag or an environment class global tag, if there is already a corresponding tag in the original data, the existing tag will not be appended, and we think that the tag in the original data should be used.

???+ attention "About Protect Mode(ENV_DISABLE_PROTECT_MODE)"

    Once protected mode is disabled, some dangerous configuration parameters can be set, and Datakit will accept any configuration parameters. These parameters may cause some Datakit functions to be abnormal or affect the collection function of the collector. For example, if the HTTP sending body is too small, the data upload function will be affected. And the collection frequency of some collectors set too high, which may affect the entities(for example MySQL) to be collected.
<!-- markdownlint-enable -->

### Point Pool Environments {#env-pointpool}

[:octicons-tag-24: Version-1.28.0](changelog.md#cl-1.28.0) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envPointPool 0}}
<!-- markdownlint-enable -->


### Dataway Configuration Environments {#env-dataway}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envDataway 0}}
<!-- markdownlint-enable -->

### Log Configuration Environments {#env-log}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envLog 0}}
<!-- markdownlint-enable -->

### Something about DataKit pprof {#env-pprof}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envPprof 0}}
<!-- markdownlint-enable -->

> `ENV_ENABLE_PPROF`: [:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2) enabled pprof by default.

### Election-related Environmental Variables {#env-elect}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envElect 0}}
<!-- markdownlint-enable -->

### HTTP/API Environment Variables {#env-http-api}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envHTTPAPI 0}}
<!-- markdownlint-enable -->

### Confd Environment Variables {#env-confd}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envConfd 0}}
<!-- markdownlint-enable -->

### Git Environment Variable {#env-git}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envGit 0}}
<!-- markdownlint-enable -->

### Sinker {#env-sinker}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envSinker 0}}
<!-- markdownlint-enable -->

### IO Module Environment Variables {#env-io}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envIO 0}}
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ note "description on buffer and queue"

    `ENV_IO_MAX_CACHE_COUNT` is used to control the data sending policy, that is, when the number of (row protocol) points of the cache in memory exceeds this value, an attempt is made to send the number of points of the current cache in memory to the center. If the threshold of the cache is set too high, the data will accumulate in memory, causing memory to soar, but will improve the compression effect of GZip. If it is too small, it may affect the transmission throughput.
<!-- markdownlint-enable -->

`ENV_IO_FILTERS` is a JSON string, as shown below:

```json
{
  "logging":[
      "{ source = 'datakit' and ( host in ['ubt-dev-01', 'tanb-ubt-dev-test'] )}",
      "{ source = 'abc' and ( host in ['ubt-dev-02', 'tanb-ubt-dev-test-1'] )}"
  ],

  "metric":[
      "{ measurement in in ['datakit', 'redis_client'] )}"
  ],
}
```

### DCA {#env-dca}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envDca 0}}
<!-- markdownlint-enable -->

### Refer Table About Environment Variables {#env-reftab}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envRefta 0}}
<!-- markdownlint-enable -->

### Recorder Environment Variables {#env-recorder}

[:octicons-tag-24: Version-1.22.0](changelog.md#1.22.0)

For more info about recorder, see [here](datakit-tools-how-to.md#record-and-replay).

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envRecorder 0}}
<!-- markdownlint-enable -->

### Others {#env-others}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envOthers 0}}
<!-- markdownlint-enable -->

### Special Environment Variable {#env-special}

#### ENV_K8S_NODE_NAME {#env_k8s_node_name}

When the k8s node name is different from its corresponding host name, the k8s node name can be replaced by the default collected host name, and the environment variable can be added in *datakit.yaml*:

> This configuration is included by default in `datakit.yaml` version  [1.2.19](changelog.md#cl-1.2.19). If you upgrade directly from the old version of yaml, you need to make the following manual changes to *datakit.yaml*.

```yaml
- env:
    - name: ENV_K8S_NODE_NAME
        valueFrom:
            fieldRef:
                apiVersion: v1
                fieldPath: spec.nodeName
```
<!-- markdownlint-disable MD013 -->
### Individual Collector-specific Environment Variable {#inputs-envs}
<!-- markdownlint-enable -->

Some collectors support external injection of environment variables to adjust the default configuration of the collector itself. See each specific collector document for details.

## Extended Readings {#more-readings}

- [DataKit election](election.md)
- [Several Configuration Methods of DataKit](k8s-config-how-to.md)
