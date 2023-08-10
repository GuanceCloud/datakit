
# Container Data Collection
---

{{.AvailableArchs}}

---

Collect indicators, objects and log data of container and Kubernetes and report them to Guance Cloud.

## Preconditions {#requrements}

- At present, container supported Docker/Containerd/CRI-O runtime
    - Docker v17.04 and above should be installed, Container v15.1 and above should be installed, CRI-O 1.20.1 and above should be installed.
- Collecting Kubernetes data requires the DataKit to [be deployed as a DaemonSet](datakit-daemonset-deploy.md).
- Collecting Kubernetes Pod metric data [requires Kubernetes to install the Metrics-Server component](https://github.com/kubernetes-sigs/metrics-server#installation){:target="_blank"}.

???+ info

    - Container collection supports both Docker and Containerd runtimes[:octicons-tag-24: Version-1.5.7](../datakit/changelog.md#cl-1.5.7), and both are enabled by default.

## Configuration {#config}

=== "host installation"

    In the case of a pure Docker or Containerd environment, the DataKit can only be installed on the host machine.
    
    Go to the *conf.d/{{.Catalog}}* directory under the DataKit installation directory, copy *{{.InputName}}.conf.sample* and name it *{{.InputName}}.conf*. Examples are as follows:
    
    ``` toml
    {{ CodeBlock .InputSample 4 }}
    ```



=== "Kubernetes"

    Container collectors in Kubernetes generally turn on automatically by default and do not need to be configured through *container.conf*. However, the configuration parameters can be adjusted by the following environment variables:
    
    | Environment Variable Name                                                     | Descrition                                                                                                                                                                          | Default Value                                                                                                   | Parameter example (need to be enclosed in double quotation marks when configuring yaml)               |
    | ----:                                                                         | ----:                                                                                                                                                                               | ----:                                                                                                           | ----                                                                                                  |
    | `ENV_INPUT_CONTAINER_ENDPOINTS`                                               | Append to container endpoints                                                                                                                                                       | ["unix:///var/run/docker.sock", "unix:///var/run/containerd/containerd.sock", "unix:///var/run/crio/crio.sock"] | `["unix:///<new_path>/run/containerd.sock"]`                                                          |
    | `ENV_INPUT_CONTAINER_DOCKER_ENDPOINT`                                         | Deprecated, specify the enpoint of Docker Engine                                                                                                                                    | "unix:///var/run/docker.sock"                                                                                   | `"unix:///var/run/docker.sock"`                                                                       |
    | `ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS`                                      | Deprecated, Specify the enpoint of Containerd                                                                                                                                       | "/var/run/containerd/containerd.sock"                                                                           | `"/var/run/containerd/containerd.sock"`                                                               |
    | `ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC`                                 | Start container index collection                                                                                                                                                    | true                                                                                                            | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC`                                       | Start k8s index collection                                                                                                                                                          | true                                                                                                            | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_ENABLE_POD_METRIC`                                       | Turn on Pod index collection                                                                                                                                                        | true                                                                                                            | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS`                               | Whether to append pod label to the collected indicator tag                                                                                                                          | false                                                                                                           | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS`     | Whether to turn on Prometheuse Pod Annotations and collect metrics automatically                                                                                                    | false                                                                                                           | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS` | Whether to turn on Prometheuse Service Annotations and collect metrics automatically                                                                                                | false                                                                                                           | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS`        | Whether to turn on automatic discovery of Prometheuse PodMonitor CRD and collection of metrics, see [Prometheus-Operator CRD doc](kubernetes-prometheus-operator-crd.md#config)     | false                                                                                                           | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_MONITORS`    | Whether to turn on automatic discovery of Prometheuse ServiceMonitor CRD and collection of metrics, see [Prometheus-Operator CRD doc](kubernetes-prometheus-operator-crd.md#config) | false                                                                                                           | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG`                                   | include condition of container log, filtering with image                                                                                                                            | None                                                                                                            | `"image:pubrepo.jiagouyun.com/datakit/logfwd*"`                                                       |
    | `ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG`                                   | exclude condition of container log, filtering with image                                                                                                                            | None                                                                                                            | `"image:pubrepo.jiagouyun.com/datakit/logfwd*"`                                                       |
    | `ENV_INPUT_CONTAINER_KUBERNETES_URL`                                          | k8s api-server access address                                                                                                                                                       | "https://kubernetes.default:443"                                                                                | `"https://kubernetes.default:443"`                                                                    |
    | `ENV_INPUT_CONTAINER_BEARER_TOKEN`                                            | The path to the token file required to access k8s api-server                                                                                                                        | "/run/secrets/kubernetes.io/serviceaccount/token"                                                               | `"/run/secrets/kubernetes.io/serviceaccount/token"`                                                   |
    | `ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING`                                     | Token string required to access k8s api-server                                                                                                                                      | None                                                                                                            | `"<your-token-string>"`                                                                               |
    | `ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL`                                 | The time interval of log discovery, that is, how often logs are retrieved. If the interval is too long, some logs with short survival will be ignored                               | "60s"                                                                                                           | `"30s"`                                                                                               |
    | `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES`                        | Log collection deletes included color characters.                                                                                                                                   | false                                                                                                           | `"true"`/`"false"`                                                                                    |
    | `ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP`                                | Log collection configures additional source matching, and the regular source will be renamed.                                                                                       | None                                                                                                            | `"source_regex*=new_source,regex*=new_source2"`  multiple "key=value" separated by English commas     |
    | `ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON`                       | For multi-row configuration of source, log collection can automatically select multiple rows using source.                                                                          | None                                                                                                            | `'{"source_nginx":"^\\d{4}", "source_redis":"^[A-Za-z_]"}'` JSON 格式的 map                           |
    | `ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION`                        | Whether the automatic multi-line mode is turned on for log collection; the applicable multi-line rules will be matched in the patterns list after it is turned on.                  | true                                                                                                            | `"true"/"false"`                                                                                      |
    | `ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON`              | Automatic multi-line pattern pattens list for log collection, supporting manual configuration of multiple multi-line rules.                                                         | For more default rules, see [doc](logging.md#auto-multiline)                                                    | `'["^\\d{4}-\\d{2}", "^[A-Za-z_]"]'`an array of strings in JSON format                                |
    | `ENV_INPUT_CONTAINER_LOGGING_MIN_FLUSH_INTERVAL`                              | Minimum upload interval for log collection. If there is no new data during this period, the cached data will be emptied and uploaded to avoid accumulation.                         | "5s"                                                                                                            | `"10s"`                                                                                               |
    | `ENV_INPUT_CONTAINER_LOGGING_MAX_MULTILINE_LIFE_DURATION`                     | Maximum single multi-row life cycle of log collection. At the end of this cycle, existing multi-row data will be emptied and uploaded to avoid accumulation.                        | "3s"                                                                                                            | `"5s"`                                                                                                |
    | `ENV_INPUT_CONTAINER_TAGS`                                                    | add extra tags                                                                                                                                                                      | None                                                                                                            | `"tag1=value1,tag2=value2"`       multiple "key=value" separated by English commas                    |
    | `ENV_INPUT_CONTAINER_PROMETHEUS_MONITORING_MATCHES_CONFIG`                    | Add additional config for Prometheus-Operator CRD                                                                                                                                   | None                                                                                                            | For more JSON format，see [Prometheus-Operator CRD doc](kubernetes-prometheus-operator-crd.md#config) |
    
    Additional description of environment variables:
    
    - ENV_INPUT_CONTAINER_TAGS: If there is a tag with the same name in the configuration file (*container.conf*), it will be overwritten by the configuration here.
    
    - ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP: Specifying the replacement source with the argument format `regular expression=new_source`, which is replaced by new_source when a source matches the regular expression. If the replacement is successful, the source（[:octicons-tag-24: Version-1.4.7](../datakit/changelog.md#cl-1.4.7)）configured in `annotations/labels` is no longer used. If you want to make an exact match, you need to use `^` and `$` to enclose the content. For example, if a regular expression is written as `datakit`, it can not only match the word `datakit` , but also match `datakit123`; Written as `^datakit$` , you can only match `datakit`.
    
    - ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON: Used to specify the mapping of source to multi-row configuration. If a log is not configured with `multiline_match`, the corresponding `multiline_match` is found and used here based on its source. Because the `multiline_match` value is a regular expression, it is more complex, so the value format is a JSON string that can be coded and compressed into a single line using [json.cn](https://www.json.cn/){:target="_blank"}.


???+ attention

    - Object data collection interval is 5 minutes and metric data collection interval is 20 seconds. Configuration is not supported for the time being.
    - Acquired log has a maximum length of 32MB per line (including after `multiline_match` processing), the excess will be truncated and discarded.

#### Docker and Containerd Sock File Configuration {#docker-containerd-sock}

If the sock path of Docker or Containerd is not the default, you need to specify the sock file path. According to different deployment methods of DataKit, the methods are different. Take Containerd as an example:

=== "Host deployment"

    Modify the `containerd_address` configuration entry of container.conf to set it to the corresponding sock path.

=== "Kubernetes"

    Change the volumes `containerd-socket` of DataKit.yaml, mount the new path into the DataKit, and configure the environment variables`ENV_INPUT_CONTAINER_ENDPOINTS`：
    
    ``` yaml hl_lines="3 4 7 14"
    # add envs
    - env:
      - name: ENV_INPUT_CONTAINER_ENDPOINTS
        value: ["unix:///path/to/new/containerd/containerd.sock"]
    
    # modify mountPath
      - mountPath: /path/to/new/containerd/containerd.sock
        name: containerd-socket
        readOnly: true
    
    # modify volumes
    volumes:
    - hostPath:
        path: /path/to/new/containerd/containerd.sock
      name: containerd-socket
    ```
---

## Log Collection {#logging-config}

See [here](container-log.md) for the relevant configuration of log collection.

### Prometheuse Exporter Metrics Collection {#k8s-prom-exporter}

If the Pod/container has exposed Prometheuse metrics, there are two ways to collect them, see [here](kubernetes-prom.md).

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.container.tags]`:

```toml
 [inputs.container.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

### Metrics {#metrics}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}

### Objects {#objects}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

### Logs {#logging}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}

## FAQ {#faq}

### Kubernetes YAML Sensitive Field Mask {#yaml-secret}

Datakit collects yaml configurations for resources such as Kubernetes Pod or Service and stores them in the `yaml` field of the object data. If the yaml contains sensitive data (such as passwords), Datakit does not support manually configuring and shielding sensitive fields for the time being. It is recommended to use Kubernetes' official practice, that is, to use ConfigMap or Secret to hide sensitive fields.

For example, you now need to add a password to the env, which would normally be like this:

```yaml
    containers:
    - name: mycontainer
      image: redis
      env:
        - name: SECRET_PASSWORD
	  value: password123
```

When orchestrating yaml configuration, passwords will be stored in clear text, which is very unsafe. You can use Kubernetes Secret to implement hiding as follows:

Create a Secret：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  username: username123
  password: password123
```

Execute:

```shell
kubectl apply -f mysecret.yaml
```

Using Secret in env:

```yaml
    containers:
    - name: mycontainer
      image: redis
      env:
        - name: SECRET_PASSWORD
	  valueFrom:
          secretKeyRef:
            name: mysecret
            key: password
            optional: false
```

See [doc](https://kubernetes.io/zh-cn/docs/concepts/configuration/secret/#using-secrets-as-environment-variables){:target="_blank"}.

## More Readings {#more-reading}

- [eBPF Collector: Support flow collection in container environment](ebpf.md)
- [Proper use of regular expressions to configure](datakit-input-conf.md#debug-regex) 
- [Several configurations of DataKit under Kubernetes](k8s-config-how-to.md)
