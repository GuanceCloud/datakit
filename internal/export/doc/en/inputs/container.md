---
title: 'Kubernetes'
summary: 'Collect metrics, objects, and log data for Container and Kubernetes, and report them to the guance cloud.'
tags:
  - 'KUBERNETES'
  - 'CONTAINER'
__int_icon:    'icon/kubernetes/'  
dashboard:
  - desc: 'Kubernetes Cluster Overview'
    path: 'dashboard/zh/kubernetes'
  - desc: 'Kubernetes Nodes Overview'
    path: 'dashboard/zh/kubernetes_nodes_overview'
  - desc: 'Kubernetes Services'
    path: 'dashboard/zh/kubernetes_services'
  - desc: 'Kubernetes Deployments'
    path: 'dashboard/zh/kubernetes_deployment'
  - desc: 'Kubernetes DaemonSets'
    path: 'dashboard/zh/kubernetes_daemonset'
  - desc: 'Kubernetes StatefulSets'
    path: 'dashboard/zh/kubernetes_statefulset'
  - desc: 'Kubernetes Pods Overview'
    path: 'dashboard/zh/kubernetes_pods_overview'
  - desc: 'Kubernetes Pods Detail'
    path: 'dashboard/zh/kubernetes_pod_detail'
  - desc: 'Kubernetes Events'
    path: 'dashboard/zh/kubernetes_events'
 
monitor:
  - desc: 'Kubernetes'
    path: 'monitor/en/kubernetes'
---

{{.AvailableArchs}}

---

Collect indicators, objects and log data of container and Kubernetes and report them to <<<custom_key.brand_name>>>.

## Configuration {#config}

### Preconditions {#requrements}

- At present, container supported Docker/Containerd/CRI-O runtime
    - Docker v17.04 and above should be installed, Container v15.1 and above should be installed, CRI-O 1.20.1 and above should be installed.
- Collecting Kubernetes data requires the DataKit to [be deployed as a DaemonSet](../datakit/datakit-daemonset-deploy.md).

<!-- markdownlint-disable MD046 -->
???+ info

    - Container collection supports both Docker and Containerd runtime[:octicons-tag-24: Version-1.5.7](../datakit/changelog.md#cl-1.5.7), and both are enabled by default.


=== "host installation"

    In the case of a pure Docker or Containerd environment, the DataKit can only be installed on the host machine.
    
    Go to the *conf.d/{{.Catalog}}* directory under the DataKit installation directory, copy *{{.InputName}}.conf.sample* and name it *{{.InputName}}.conf*. Examples are as follows:
    
    ``` toml
    {{ CodeBlock .InputSample 4 }}
    ```



=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

    Additional description of environment variables:
    
    - ENV_INPUT_CONTAINER_TAGS: If there is a tag with the same name in the configuration file (*container.conf*), it will be overwritten by the configuration here.
    
    - ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP: Specifying the replacement source with the argument format `regular expression=new_source`, which is replaced by new_source when a source matches the regular expression. If the replacement is successful, the source（[:octicons-tag-24: Version-1.4.7](../datakit/changelog.md#cl-1.4.7)）configured in `annotations/labels` is no longer used. If you want to make an exact match, you need to use `^` and `$` to enclose the content. For example, if a regular expression is written as `datakit`, it can not only match the word `datakit` , but also match `datakit123`; Written as `^datakit$` , you can only match `datakit`.
    
    - ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON: Used to specify the mapping of source to multi-row configuration. If a log is not configured with `multiline_match`, the corresponding `multiline_match` is found and used here based on its source. Because the `multiline_match` value is a regular expression, it is more complex, so the value format is a JSON string that can be coded and compressed into a single line using [json.cn](https://www.json.cn/){:target="_blank"}.


???+ note

    - Object data collection interval is 5 minutes and metric data collection interval is 20 seconds. Configuration is not supported for the time being.
    - Acquired log has a maximum length of ~800KB per line (including after `multiline_match` processing), the excess will split into new logging lines.

### Docker and Containerd Sock File Configuration {#sock-config}

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
<!-- markdownlint-enable -->

Environment Variables `ENV_INPUT_CONTAINER_ENDPOINTS` is added to the existing endpoints configuration, and the actual endpoints configuration may have many items. The collector will remove duplicates and connect and collect them one by one.

The default endpoints configuration is:

```yaml
  endpoints = [
    "unix:///var/run/docker.sock",
    "unix:///var/run/containerd/containerd.sock",
    "unix:///var/run/crio/crio.sock",
  ] 
```

Using Environment Variables `ENV_INPUT_CONTAINER_ENDPOINTS` is`["unix:///path/to/new//run/containerd.sock"]`,The final endpoints configuration is as follows:

```yaml
  endpoints = [
    "unix:///var/run/docker.sock",
    "unix:///var/run/containerd/containerd.sock",
    "unix:///var/run/crio/crio.sock",
    "unix:///path/to/new//run/containerd.sock",
  ] 
```

The collector will connect and collect these containers during runtime. If the sock file does not exist, an error log will be output when the first connection fails, which does not affect subsequent collection.

### Prometheus Exporter Metrics Collection {#k8s-prom-exporter}

<!-- markdownlint-disable MD024 -->
If the Pod/container has exposed Prometheus metrics, there are two ways to collect them, see [here](kubernetes-prom.md).


### Log Collection {#logging-config}

See [here](container-log.md) for the relevant configuration of log collection.

---

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.container.tags]`:

```toml
 [inputs.container.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```


{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}{{end}}

{{ end }}

## Object {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## Logs {#logging}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}{{end}}

{{ end }}
<!-- markdownlint-enable -->

## Link Dataway Sink Function {#link-dataway-sink}

Dataway Sink [see documentation](../deployment/dataway-sink.md).

All collected Kubernetes resources will have a Label that matches the CustomerKey. For example, if the CustomerKey is `name`, DaemonSets, Deployments, Pods, and other resources will search for `name` in their own current Labels and add it to tags.

Containers will add Customer Labels of the Pods they belong to.

## FAQ {#faq}

### Filtering Metrics Collection by Pod Namespace {#config-metric-on-pod-namespace}

When Kubernetes Pod metrics collection is enabled (`enable_pod_metric = true`), DataKit will collect metrics data for all Pods in the cluster. Since this can result in a large volume of data, you can filter the metrics collection by the Pod's `namespace` field to only collect metrics from Pods in specific namespaces.

You can control which Pods are included or excluded from metrics collection by configuring `pod_include_metric` and `pod_exclude_metric`.

<!-- markdownlint-disable md046 -->
=== "Host Installation"

    ``` toml
      ## Collect metrics for Pods whose namespace matches `datakit`
      pod_include_metric = ["namespace:datakit"]
    
      ## Exclude all Pods with namespace `kodo`
      pod_exclude_metric = ["namespace:kodo"]
    ```
    
    - Both `include` and `exclude` settings must start with a field name, formatted as a [glob pattern](https://en.wikipedia.org/wiki/glob_(programming)): `"<field name>:<glob rule>"`.
    - Currently, only the `namespace` field is supported for filtering, e.g., `namespace:datakit-ns`.
    
    If both `include` and `exclude` are configured, the Pod must meet the following conditions:
    
    - It must satisfy the `include` rule
    - It must not satisfy the `exclude` rule
    
    For example, the following configuration will filter out all Pods:
    
    ```toml
      ## Collect only Pods in `namespace:datakit` and exclude all namespaces
      pod_include_metric = ["namespace:datakit"]
      pod_exclude_metric = ["namespace:*"]
    ```

=== "Kubernetes"

    In a Kubernetes environment, you can configure the following environment variables:
    
    - `ENV_INPUT_CONTAINER_POD_INCLUDE_METRIC`
    - `ENV_INPUT_CONTAINER_POD_EXCLUDE_METRIC`
    
    For example, if you want to collect metrics only for Pods in the `kube-system` namespace, you can set the `ENV_INPUT_CONTAINER_POD_INCLUDE_METRIC` environment variable as shown below:
    
    ```yaml
      - env:
          - name: ENV_INPUT_CONTAINER_POD_INCLUDE_METRIC
            value: namespace:kube-system  # Specify the namespace to collect metrics from
    ```
    
    By using this method, you can flexibly control the range of Pods from which DataKit collects metrics, reducing the collection of unnecessary data and optimizing system performance and resource usage.

<!-- markdownlint-disable MD013 -->
### NODE_LOCAL Mode Requires New RBAC Permissions {#rbac-nodes-stats}
<!-- markdownlint-enable -->

The `ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL` mode is only recommended for DaemonSet deployment and requires access to kubelet, so the `nodes/stats` permission needs to be added to RBAC. For example:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
- apiGroups: [""]
  resources: ["nodes", "nodes/stats"]
  verbs: ["get", "list", "watch"]
```

In addition, the DataKit Pod needs to have the `hostNetwork: true` configuration item enabled.

<!-- markdownlint-disable MD013 -->
### Collect PersistentVolumes and PersistentVolumeClaims Requires New Permissions {#rbac-pv-pvc}
<!-- markdownlint-enable -->

DataKit version 1.25.0[:octicons-tag-24: Version-1.25.0](../datakit/changelog.md#cl-1.25.0) supported the collection of object data for Kubernetes PersistentVolume and PersistentVolumeClaim, which require new RBAC permissions, as described below:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
- apiGroups: [""]
  resources: ["persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
```

<!-- markdownlint-disable MD013 -->
### Kubernetes YAML Sensitive Field Mask {#yaml-secret}
<!-- markdownlint-enable -->

DataKit collects yaml configurations for resources such as Kubernetes Pod or Service and stores them in the `yaml` field of the object data. If the yaml contains sensitive data (such as passwords), DataKit does not support manually configuring and shielding sensitive fields for the time being. It is recommended to use Kubernetes' official practice, that is, to use ConfigMap or Secret to hide sensitive fields.

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
- [Proper use of regular expressions to configure](../datakit/datakit-input-conf.md#debug-regex)
- [Several configurations of DataKit under Kubernetes](../datakit/k8s-config-how-to.md)
