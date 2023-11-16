# Kubernetes
---

This document describes how to install DataKit in K8s via DaemonSet.

## Installation {#install}

=== "Daemonset"

    Download [datakit.yaml](https://static.guance.com/datakit/datakit.yaml){:target="_blank"}, in which many [default collectors](datakit-input-conf.md#default-enabled-inputs) are turned on without configuration.
    
    ???+ attention
    
        If you want to modify the default configuration of these collectors, you can configure them by [mounting a separate conf in Configmap mode](k8s-config-how-to.md#via-configmap-conf). Some collectors can be adjusted directly by means of environment variables. See the documents of specific collectors for details. All in all, configuring the collector through [Configmap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/){:target="_blank"} is always effective when deploying the DataKit in DaemonSet mode, whether it is a collector turned on by default or other collectors.
    
    Modify the dataway configuration in `datakit.yaml`
    
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
    
    After installation, a DaemonSet deployment of datakit is created:
    
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

## Kubernetes Tolerance Configuration {#toleration}

DataKit is deployed on all nodes in the Kubernetes cluster by default (that is, all stains are ignored). If some node nodes in Kubernetes have added stain scheduling and do not want to deploy DataKit on them, you can modify datakit.yaml to adjust the stain tolerance:

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

## Settings of Other Environment Variables in DataKit {#using-k8-env}

> Note: If ENV_LOG is configured to `stdout`, do not set ENV_LOG_LEVEL to `debug`, otherwise looping logs may result in large amounts of log data.

In DaemonSet mode, DataKit supports multiple environment variable configurations.

- The approximate format in datakit.yaml is

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

### Description of Environment Variable Type {#env-types}

The values of the following environment variables are divided into the following data types:

- string: string type
- json: some of the more complex configurations that require setting environment variables in the form of a json string
- bool: switch type. Given **any non-empty string** , this function is turned on. It is recommended to use `"on"` as its value when turned on. If it is not opened, it must be deleted or commented out.
- string-list: a string separated by an English comma, commonly used to represent a list
- duration: a string representation of the length of time, such as `10s` for 10 seconds, where the unit supports h/m/s/ms/us/ns. **Don't give a negative value**.
- int: integer type
- float: floating point type

For string/bool/string-list/duration, it is recommended to use double quotation marks to avoid possible problems caused by k8s parsing yaml.

### Most Commonly Used Environment Variables {#env-common}

| Environment Variable Name    | Type        | Default Value | Required | Description                                                                                                                                                                                         |
| :---                         | :---:       | :---          | :---     | :---                                                                                                                                                                                                |
| `ENV_DISABLE_PROTECT_MODE`   | bool        | -             | No       | Disable protect mode                                                                                                                                                                                |
| `ENV_DATAWAY`                | string      | None          | Yes      | Configure the DataWay address, such as `https://openway.guance.com?token=xxx`                                                                                                                       |
| `ENV_DEFAULT_ENABLED_INPUTS` | string-list | None          | No       | [The list of collectors](datakit-input-conf.md#default-enabled-inputs) is opened by default, divided by English commas, such as `cpu,mem,disk`, and the old  `ENV_ENABLE_INPUTS` will be discarded. |
| `ENV_GLOBAL_HOST_TAGS`       | string-list | None          | No       | Global tag, multiple tags are divided by English commas, such as `tag1=val,tag2=val2`. The old `ENV_GLOBAL_TAGS` will be discarded.                                                                 |

???+ note "Distinguish between *global host tag*  and *global election tag*"

    `ENV_GLOBAL_HOST_TAGS` is used to specify host class global tags whose values generally follow host transitions, such as host name and host IP. Of course, other tags that do not follow the host changes can also be added. All collectors of non-elective classes are taken by default with the tag specified in `ENV_GLOBAL_HOST_TAGS`.
    
    And `ENV_GLOBAL_ELECTION_TAGS` recommends adding only tags that do not change with host switching, such as cluster name, project name, etc. For [election collector](election.md#inputs), only the tag specified in `ENV_GLOBAL_ELECTION_TAGS` will be added, not the tag specified in `ENV_GLOBAL_HOST_TAGS`.
    
    Whether it is a host class global tag or an environment class global tag, if there is already a corresponding tag in the original data, the existing tag will not be appended, and we think that the tag in the original data should be used.

???+ attention "About Protect Mode(ENV_DISABLE_PROTECT_MODE)"

    Once protected mode is disabled, some dangerous configuration parameters can be set, and Datakit will accept any configuration parameters. These parameters may cause some Datakit functions to be abnormal or affect the collection function of the collector. For example, if the HTTP sending body is too small, the data upload function will be affected. And the collection frequency of some collectors set too high, which may affect the entities(for example MySQL) to be collected.

### Dataway Configuration Related Environments {#env-dataway}

| Environment Variable Name       | Type     | Default Value | Required | Description                                                                                                                                                             |
| ---:                            | ---:     | ---:          | ---:     | ---:                                                                                                                                                                    |
| `ENV_DATAWAY_MAX_RAW_BODY_SIZE` | int      | 10MB          | No       | Set upload package size(before gzip)                                                                                                                                    |
| `ENV_DATAWAY`                   | string   | No            | Yes      | Set DataWay address, such as `https://openway.guance.com?token=xxx`                                                                                                     |
| `ENV_DATAWAY_TIMEOUT`           | duration | "30s"         | No       | Set DataWay request timeout                                                                                                                                             |
| `ENV_DATAWAY_ENABLE_HTTPTRACE`  | bool     | -             | No       | Enable metrics on DataWay HTTP request                                                                                                                                  |
| `ENV_DATAWAY_HTTP_PROXY`        | string   | No            | No       | Set DataWay HTTP Proxy                                                                                                                                                  |
| `ENV_DATAWAY_MAX_IDLE_CONNS`    | int      | 100           | No       | Set DataWay HTTP connection pool size([:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0))                                                                         |
| `ENV_DATAWAY_IDLE_TIMEOUT`      | duration | "90s"         | No       | Set DataWay HTTP Keep-Alive timeout([:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0))                                                                           |
| `ENV_DATAWAY_MAX_RETRY_COUNT`   | int      | 4             | No       | Specify at most how many times the data sending operation will be performed when encounter failures([:octicons-tag-24: Version-1.18.0](changelog.md#cl-1.18.0))         |
| `ENV_DATAWAY_RETRY_DELAY`       | duration | "200ms"       | No       | The interval between two data sending retry, valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"([:octicons-tag-24: Version-1.18.0](changelog.md#cl-1.18.0)) |
| `ENV_DATAWAY_CONTENT_ENCODING`  | string   | `v1`          | No       | Set the encoding of the point data at upload time (optional list: 'v1' is the line protocol, 'v2' is Protobuf)                                                          |

### Log Configuration Related Environments {#env-log}

| Environment Variable Name | Type   | Default Value              | Required | Description                                                                                                                              |
|:---                       |:---    |:---                        |:---      |:---                                                                                                                                      |
| `ENV_GIN_LOG`             | string | */var/log/datakit/gin.log* | No       | If it is changed to `stdout`, the DataKit's own gin log will not be written to the file, but will be output by the terminal.             |
| `ENV_LOG`                 | string | */var/log/datakit/log*     | No       | If it is changed to `stdout`, DataKit's own log will not be written to the file, but will be output by the terminal.                     |
| `ENV_LOG_LEVEL`           | string | info                       | No       | Set DataKit's own log level, optional `info/debug`(case insensitive).                                                                                      |
| `ENV_DISABLE_LOG_COLOR`   | bool   | -                          | No       | Turn off log colors                                                                                                                      |
| `ENV_LOG_ROTATE_BACKUP`   | int    | 5                          | No       | The upper limit count for log files to be reserve.                                                                                       |
| `ENV_LOG_ROTATE_SIZE_MB`  | int    | 32                         | No       | The threshold for automatic log rotating in MB, which automatically switches to a new file when the log file size reaches the threshold. |

###  Something about DataKit pprof {#env-pprof}

| Environment Variable Name | Type   | Default Value | Required | Description                       |
| :---                      | :---   | :---          | :---     | :---                              |
| `ENV_ENABLE_PPROF`        | bool   | -             | No       | Whether to start `pprof`          |
| `ENV_PPROF_LISTEN`        | string | None          | No       | `pprof` service listening address |

> `ENV_ENABLE_PPROF`: [:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2) enabled pprof by default.

### Election-related Environmental Variables {#env-elect}

| Environment Variable Name           | Type        | Default Value | Required | Description                                                                                                                                                                                                                                                                                        |
| :---                                | :---        | :---          | :---     | :---                                                                                                                                                                                                                                                                                               |
| `ENV_ENABLE_ELECTION`               | bool        | -             | No       | If you want to open the [election](election.md), it will not be opened by default. If you want to open it, you can give any non-empty string value to the environment variable.                                                                                                                    |
| `ENV_NAMESPACE`                     | string      | `default`     | No       | The namespace in which the DataKit resides, which defaults to null to indicate that it is namespace-insensitive and accepts any non-null string, such as `dk-namespace-example`. If the election is turned on, you can specify the workspace through this environment variable.                    |
| `ENV_ENABLE_ELECTION_NAMESPACE_TAG` | bool        | -             | No       | When this option is turned on, all election classes are collected with an extra tag of `election_namespace=<your-election-namespace>`, which may result in some timeline growth. ([:octicons-tag-24: Version-1.4.7](changelog.md#cl-1.4.7))                                                        |
| `ENV_GLOBAL_ELECTION_TAGS`          | string-list |               | No       | Tags are elected globally, and multiple tags are divided by English commas, such as `tag1=val,tag2=val2`. ENV_GLOBAL_ENV_TAGS will be discarded.                                                                                                                                                   |
| `ENV_CLUSTER_NAME_K8S`              | string      | -             | No       | The cluster name in which the Datakit residers, if the cluster is not empty, a specified tag will be added to [global election tags](election.md#global-tags), the key is `cluster_name_k8s` and the value is the environment variable. ([:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8)) |

### HTTP/API Related Environment Variables {#env-http-api}

| Environment Variable Name        | Type        | Default Value     | Required | Description                                                                                                                                                                                                                         |
| :---                             | :---        | :---              | :---     | :---                                                                                                                                                                                                                                |
| `ENV_DISABLE_404PAGE`            | bool        | -                 | No       | Disable the DataKit 404 page (commonly used when deploying DataKit RUM on the public network).                                                                                                                                      |
| `ENV_HTTP_LISTEN`                | string      | localhost:9529    | No       | The address can be modified so that the [DataKit interface](apis) can be called externally.                                                                                                                                         |
| `ENV_HTTP_PUBLIC_APIS`           | string-list | None              | No       | [API list](apis) that allow external access, separated by English commas between multiple APIs. When DataKit is deployed on the public network, it is used to disable some APIs.                                                    |
| `ENV_HTTP_TIMEOUT`               | duration    | 30s               | No       | Setting the 9529 HTTP API Server Timeout [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental).                                                                     |
| `ENV_HTTP_CLOSE_IDLE_CONNECTION` | bool        | -                 | No       | If turned on, the 9529 HTTP server actively closes idle connections (idle time equal to `ENV_HTTP_TIMEOUT`） [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental). |
| `ENV_REQUEST_RATE_LIMIT`         | float       | None              | No       | Limit 9529 [API requests per second](datakit-conf.md#set-http-api-limit).                                                                                                                                                           |
| `ENV_RUM_ORIGIN_IP_HEADER`       | string      | `X-Forwarded-For` | No       | RUM dedicated                                                                                                                                                                                                                       |
| `ENV_RUM_APP_ID_WHITE_LIST`      | string      | None              | No       | RUM app-id white list, split by `,`,  such as `appid-1,appid-2`.                                                                                                                                                                    |

### Confd Configures Related Environment Variables {#env-confd}

| Environment Variable Name  | Type   | Applicable Scenario             | Description            | Sample Value                                   |
| ---                        | ---    | ---                             | ---                    | ---                                            |
| `ENV_CONFD_BACKEND`        | string | All                             | Backend source type    | `etcdv3` or `zookeeper` or `redis` or `consul` |
| `ENV_CONFD_BASIC_AUTH`     | string | `etcdv3` or `consul`            | Optional               |                                                |
| `ENV_CONFD_CLIENT_CA_KEYS` | string | `etcdv3` or `consul`            | Optional               |                                                |
| `ENV_CONFD_CLIENT_CERT`    | string | `etcdv3` or `consul`            | Optional               |                                                |
| `ENV_CONFD_CLIENT_KEY`     | string | `etcdv3` or `consul` or `redis` | Optional               |                                                |
| `ENV_CONFD_BACKEND_NODES`  | string | All                             | Backend source address | `[IP address:2379,IP address2:2379]`           |
| `ENV_CONFD_PASSWORD`       | string | `etcdv3` or `consul`            | Optional               |                                                |
| `ENV_CONFD_SCHEME`         | string | `etcdv3` or `consul`            | Optional               |                                                |
| `ENV_CONFD_SEPARATOR`      | string | `redis`                         | Optional default 0     |                                                |
| `ENV_CONFD_USERNAME`       | string | `etcdv3` or `consul`            | Optional               |                                                |

### Git Configuration Related Environment Variable {#env-git}

| Environment Variable Name | Type     | Default Value | Required | Description                                                                                                                                                        |
| :---                      | :---     | :---          | :---     | :---                                                                                                                                                               |
| `ENV_GIT_BRANCH`          | string   | None          | No       | Specifies the branch to pull. <stong>If it is empty, it is the default.</strong> And the default is the remotely specified main branch, which is usually `master`. |
| `ENV_GIT_INTERVAL`        | duration | None          | No       | The interval of timed pull. (e.g. `1m`)                                                                                                                            |
| `ENV_GIT_KEY_PATH`        | string   | None          | No       | The full path of the local PrivateKey. (e.g. `/Users/username/.ssh/id_rsa`）                                                                                       |
| `ENV_GIT_KEY_PW`          | string   | None          | No       | Use password of local PrivateKey. (e.g. `passwd`）                                                                                                                 |
| `ENV_GIT_URL`             | string   | None          | No       | Manage the remote git repo address of the configuration file. (e.g. `http://username:password@github.com/username/repository.git`）                                |

### Sinker {#env-sinker}

| Environment Variable Name         | Type   | Default Value | Required | Description                                                                         |
| :---                              | :---   | :---          | :---     | :---                                                                                |
| `ENV_SINKER_GLOBAL_CUSTOMER_KEYS` | string | None          | No       | Sinker Global Customer Key list, keys are splited with `,`                          |
| `ENV_DATAWAY_ENABLE_SINKER`       | bool   | None          | No       | Enable DataWay Sinker ([:octicons-tag-24: Version-1.14.0](changelog.md#cl-1.14.0)). |

### IO Module Configuring Related Environment Variables {#env-io}

| Environment Variable Name     | Type     | Default Value      | Required | Description                                                                        |
| `ENV_IO_CONTENT_ENCODING`     | string   | `line-protocol`    | No       | Set uploading content encoding of point(candidates: `json/protobuf/line-protocol`) |
| `ENV_IO_FILTERS`              | json     | None               | No       | Add [row protocol filter](datakit-filter)                                          |
| `ENV_IO_FLUSH_INTERVAL`       | duration | 10s                | No       | IO transmission time frequency                                                     |
| `ENV_IO_FLUSH_WORKERS`        | int      | `cpu_core * 2 + 1` | No       | IO flush workers([:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9))         |
| `ENV_IO_MAX_CACHE_COUNT`      | int      | 1000               | No       | Send buffer size                                                                   |
| `ENV_IO_ENABLE_CACHE`         | bool     | false              | No       | Whether to open the disk cache that failed to send                                 |
| `ENV_IO_CACHE_ALL`            | bool     | false              | No       | cache failed data points of all categories                                         |
| `ENV_IO_CACHE_MAX_SIZE_GB`    | int      | 10                 | No       | Disk size of send failure cache (in GB)                                            |
| `ENV_IO_CACHE_CLEAN_INTERVAL` | duration | 5s                 | No       | Periodically send failed tasks cached on disk                                      |

???+ note "description on buffer and queue"

    `ENV_IO_MAX_CACHE_COUNT` is used to control the data sending policy, that is, when the number of (row protocol) points of the cache in memory exceeds this value, an attempt is made to send the number of points of the current cache in memory to the center. If the threshold of the cache is set too high, the data will accumulate in memory, causing memory to soar, but will improve the compression effect of GZip. If it is too small, it may affect the transmission throughput.

`ENV_IO_FILTERS` is a json string, as shown below:

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

| Environment Variable Name | Type   | Default Value  | Required | Description                                                                                                                                                     |
| :---                      | :---   | :---           | :---     | :---                                                                                                                                                            |
| `ENV_DCA_LISTEN`          | string | localhost:9531 | No       | The address can be modified so that the [DCA](dca.md) client can manage the DataKit. Once `ENV_DCA_LISTEN` is turned on, the DCA function is enabled by default |
| `ENV_DCA_WHITE_LIST`      | string | None           | No       | Configure DCA white list, separated by English commas                                                                                                           |

### Refer Table About Environment Variables {#env-reftab}

| Environment Variable Name                    | Type   | Default Value | Required   | Description                          |
| :---                            | :---   | :---   | :---   | :---                          |
| `ENV_REFER_TABLE_URL`           | string | None     | No     | Set the data source URL                |
| `ENV_REFER_TABLE_PULL_INTERVAL` | string | 5m     | No     | Set the request interval for the data source URL |

### Others {#env-others}

| Environment Variable Name        | Type   | Default Value  | Required | Description                                                                                                  |
| :---                             | :---   | :---           | :---     | :---                                                                                                         |
| `ENV_CLOUD_PROVIDER`             | string | None           | No       | Support filling in cloud suppliers during installation(`aliyun/aws/tencent/hwcloud/azure`)                   |
| `ENV_HOSTNAME`                   | string | None           | No       | The default is the local host name, which can be specified at installation time, such as, `dk-your-hostname` |
| `ENV_IPDB`                       | string | None           | No       | Specify the IP repository type, currently only supports `iploc/geolite2`                                     |
| `ENV_ULIMIT`                     | int    | None           | No       | Specify the maximum number of open files for Datakit                                                         |
| `ENV_PIPELINE_OFFLOAD_RECEIVER`  | string | `datakit-http` | false    | Set offload receiver                                                                                         |
| `ENV_PIPELINE_OFFLOAD_ADDRESSES` | string | None           | false    | Set offload addresses                                                                                        |

### Special Environment Variable {#env-special}

#### ENV_K8S_NODE_NAME {#env_k8s_node_name}

When the k8s node name is different from its corresponding host name, the k8s node name can be replaced by the default collected host name, and the environment variable can be added in *datakit.yaml*:

> This configuration is included by default in datakit.yaml version  [1.2.19](changelog.md#cl-1.2.19). If you upgrade directly from the old version of yaml, you need to make the following manual changes to *datakit.yaml*.

```yaml
- env:
	- name: ENV_K8S_NODE_NAME
		valueFrom:
			fieldRef:
				apiVersion: v1
				fieldPath: spec.nodeName
```

### Individual Collector-specific Environment Variable {#inputs-envs}

Some collectors support external injection of environment variables to adjust the default configuration of the collector itself. See each specific collector document for details.

## Extended Readings {#more-readings}

- [DataKit election](election.md)
- [Several Configuration Methods of DataKit](k8s-config-how-to.md)
