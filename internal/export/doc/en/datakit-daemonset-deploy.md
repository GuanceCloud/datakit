# Kubernetes
---

This document describes how to install DataKit in K8s via DaemonSet.

## Installation {#install}

<!-- markdownlint-disable MD046 -->
=== "DaemonSet"

    Download [`datakit.yaml`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit.yaml){:target="_blank"}, in which many [default collectors](datakit-input-conf.md#default-enabled-inputs) are turned on without configuration.
    
    ???+ note
    
        If you want to modify the default configuration of these collectors, you can configure them by [mounting a separate conf in `ConfigMap` mode](k8s-config-how-to.md#via-configmap-conf). Some collectors can be adjusted directly by means of environment variables. See the documents of specific collectors for details. All in all, configuring the collector through [`ConfigMap`](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/){:target="_blank"} is always effective when deploying the DataKit in DaemonSet mode, whether it is a collector turned on by default or other collectors.
    
    Modify the Dataway configuration in `datakit.yaml`
    
    ```yaml
        - name: ENV_DATAWAY
            value: https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN> # Fill in the real address of DataWay here
    ```
    
    If you choose another node, change the corresponding DataWay address here, such as AWS node:
    
    ```yaml
        - name: ENV_DATAWAY
            value: https://aws-openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>
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
    
    Helm installs DataKit (note modifying the `datakit.dataway_url` parameter)，in which many [default collectors](datakit-input-conf.md#default-enabled-inputs) are turned on without configuration.
    
    ```shell
    helm install datakit datakit \
    <<<% if custom_key.brand_key == 'guance' %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
    <<<% else %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
    <<<% endif %>>>
        -n datakit --create-namespace \
        --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>"
    ```
    
    View deployment status:
    
    ```shell
    helm -n datakit list
    ```
    
    You can upgrade with the following command:
    
    ```shell
    helm -n datakit get  values datakit -o yaml > values.yaml
    helm upgrade datakit datakit \
    <<<% if custom_key.brand_key == 'guance' %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
    <<<% else %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
    <<<% endif %>>>
        -n datakit \
        -f values.yaml
    ```
    
    You can uninstall it with the following command:
    
    ```shell
    $ helm uninstall datakit -n datakit
    ```

    ### More Helm Examples {#helm-examples}

    In addition to manually editing the *values.yaml* to adjust DataKit configurations (it is still recommended to use *values.yaml* directly for complex escape operations), you can also specify these parameters during the Helm installation phase. Note that these parameters must adhere to Helm's command-line syntax.

    **Setting Default Collector List**

    ```shell
    helm install datakit datakit \
    <<<% if custom_key.brand_key == 'guance' %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
    <<<% else %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
    <<<% endif %>>>
         -n datakit --create-namespace \
         --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>" \
         --set datakit.default_enabled_inputs="statsd\,dk\,cpu\,mem"
    ```

    **Note**: The comma `,` must be escaped here; otherwise, Helm will throw an error.
    
    **Setting Environment Variables**

    DataKit supports numerous [environment variable configurations](datakit-daemonset-install.md#env-setting), which can be appended using the following method:

    ```shell
    helm install datakit datakit \
    <<<% if custom_key.brand_key == 'guance' %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
    <<<% else %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
    <<<% endif %>>>
        -n datakit --create-namespace \
        --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=tkn_xxx" \
        --set "extraEnvs[0].name=ENV_INPUT_OTEL_GRPC" \
        --set 'extraEnvs[0].value=\{"trace_enable":true\,"metric_enable":true\,"addr":"0.0.0.0:4317"\}' \
        --set "extraEnvs[1].name=ENV_INPUT_CPU_PERCPU" \
        --set 'extraEnvs[1].value=true'
    ```

    Here, `extraEnvs` is an entry defined in the DataKit Helm chart for setting environment variables. Since environment variables are in array format, we use array indices (starting from 0) to append multiple variables. `name` represents the environment variable name, and `value` is the corresponding value. Notably, if an environment variable's value is a JSON string, characters like `{},` must be escaped.
    
    **Installing a Specific Version**

    You can specify the DataKit image version using `image.tag`:

    ```shell
    helm install datakit datakit \
    <<<% if custom_key.brand_key == 'guance' %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
    <<<% else %>>>
        --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
    <<<% endif %>>>
        -n datakit --create-namespace \
        --set image.tag="1.70.0" \
        ...
    ```
<!-- markdownlint-enable -->

### Resource Limits {#requests-limits}

DataKit has set default Requests and Limits. If the DataKit container status changes to OOMKilled, you can customize and modify the configuration.

<!-- markdownlint-disable MD046 -->
=== "Yaml"

    The approximate format in *datakit.yaml* is as follows:

    ```yaml
    ...
                resources:
                  requests:
                    cpu: "200m"
                    memory: "128Mi"
                  limits:
                    cpu: "2000m"
                    memory: "4Gi"
    ...
    ```

=== "Helm"

    The approximate format in Helm values.yaml is as follows:

    ```yaml
    ...
        resources:
          requests:
            cpu: "200m"
            memory: "128Mi"
          limits:
            cpu: "2000m"
            memory: "4Gi"
    ...
    ```
<!-- markdownlint-enable -->

For specific configurations, refer to the [official document](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits){:target="_blank"}.

### Security Restrictions {#security-context}

DataKit is recommended to run as the root user and in privileged mode. However, it also supports running as a non-root user with permission controls, though this may affect the collection of some data. The following steps can be used to reduce the privilege level of the DataKit container while ensuring it can collect network data, container data, etc.

**Prerequisites:**

- Kubernetes cluster
- Node access permissions (to perform host-level configuration)
- The DataKit image includes a `datakit` user with UID 10001 (DataKit images version 1.83.0 and later have the `datakit` user created; the default user remains `root`)

**Configuration Steps:**

1. **Create user groups and set permissions on the host.** Execute the following commands on each Kubernetes node:

    ```bash
    # Create a dedicated user group
    groupadd datakit-reader

    # Set log directory permissions
    chgrp -R datakit-reader /var/log/pods
    chmod -R g+rx /var/log/pods

    # Set Docker socket permissions (if used)
    chgrp datakit-reader /var/run/docker.sock
    chmod g+r /var/run/docker.sock

    # Set Containerd socket permissions
    chgrp datakit-reader /var/run/containerd/containerd.sock
    chmod g+r /var/run/containerd/containerd.sock

    # Set CRI-O socket permissions (if used)
    chgrp datakit-reader /var/run/crio/crio.sock
    chmod g+r /var/run/crio/crio.sock

    # Set Kubelet directory permissions
    chgrp -R datakit-reader /var/lib/kubelet/pods
    chmod -R g+rx /var/lib/kubelet/pods
    ```

1. **Obtain the user group GID.** Execute the following command on each node to get the GID of the `datakit-reader` group:

    ```bash
    getent group datakit-reader | cut -d: -f3
    ```

    Note the output GID value (e.g., `12345`), which will be needed in the next step.

1. **Configure the Kubernetes Deployment/DaemonSet.** Update your DataKit Kubernetes configuration file:

    ```yaml
    apiVersion: apps/v1
    kind: DaemonSet  # Or Deployment
    metadata:
      name: datakit
      namespace: monitoring
    spec:
      template:
        spec:
          # Security context configuration
          securityContext:
            runAsUser: 10001  # UID of the datakit user
            runAsGroup: 10001 # GID of the datakit user
            fsGroup: 10001    # Filesystem group
            supplementalGroups: [12345]  # GID of the datakit-reader group obtained in the previous step
          containers:
          - name: datakit
            image: your-datakit-image:tag
            # Container security context
            securityContext:
              privileged: false             # Disable privileged mode
              allowPrivilegeEscalation: false
              readOnlyRootFilesystem: true  # Optional: Set root filesystem as read-only
              capabilities:
                drop: ["ALL"]  # Drop all capabilities
                add: ["SYS_ADMIN", "SYS_PTRACE", "DAC_READ_SEARCH", "NET_RAW"] # Add necessary capabilities
        # Other content...
    ```

**Notes:**

- When running as a non-root user, the following functionalities of DataKit may be restricted:
    - Collection of some system metrics: Certain system files and directories requiring root permissions may be inaccessible.
    - Limited container runtime data: If the container socket and container log directories are not at the default paths, remounting and re-authorization are required.
    - Missing kernel-level metrics: Some system calls requiring privileged capabilities cannot be executed.
- This configuration requires host-level permission settings on each Kubernetes node.
- When nodes are scaled up, the permission setting steps must be repeated on the new nodes.
- Consider using configuration management tools (Ansible, Chef, Puppet) to automate node configuration.

**Reverting to Root Mode:**

If the non-root mode cannot meet your monitoring requirements, you can revert to root mode at any time:

1. Remove the `runAsUser`, `runAsGroup`, and `supplementalGroups` configurations from the YAML.
1. Set `privileged` to `true`.
1. Redeploy DataKit.

### Kubernetes Tolerance Configuration {#toleration}

DataKit is deployed on all nodes in the Kubernetes cluster by default (that is, all stains are ignored). If some node nodes in Kubernetes have added stain scheduling and do not want to deploy DataKit on them, you can modify `datakit.yaml` to adjust the stain tolerance:

```yaml
      tolerations:
      - operator: Exists    <--- Modify the stain tolerance here
```

For specific bypass strategies, see [official doc](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration){:target="_blank"}。

## Collector Configuration {#input-config}

There are two ways to configure collectors in DataKit for Kubernetes:

1. ConfigMap: Append collector configurations by injecting a ConfigMap.
1. Environment Variables: Inject a complete TOML collection configuration through a specific environment variable.

### ConfigMap Settings {#configmap-setting}

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

### ENV Set Collectors {#env-setting}

The opening of some collectors can also be injected through `ENV_DATAKIT_INPUTS`. The following is an injection example of MySQL and Redis collectors:

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

## Environments about DataKit main configure {#using-k8-env}

In Kubernetes, DataKit no longer uses *datkait.conf* for configuration and can only use environment variables. In the DaemonSet mode, DataKit supports multiple environment variable configurations.

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

???+ note "About Protect Mode(`ENV_DISABLE_PROTECT_MODE`)"

    Once protected mode is disabled, some dangerous configuration parameters can be set, and DataKit will accept any configuration parameters. These parameters may cause some DataKit functions to be abnormal or affect the collection function of the collector. For example, if the HTTP sending body is too small, the data upload function will be affected. And the collection frequency of some collectors set too high, which may affect the entities(for example MySQL) to be collected.
<!-- markdownlint-enable -->

<!--
### Point Pool Environments {#env-pointpool}

[:octicons-tag-24: Version-1.28.0](changelog.md#cl-1.28.0) ·
[:octicons-beaker-24: Experimental](index.md#experimental)
-->


### Dataway {#env-dataway}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envDataway 0}}
<!-- markdownlint-enable -->

### Logging {#env-log}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envLog 0}}
<!-- markdownlint-enable -->

### Election {#env-elect}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envElect 0}}
<!-- markdownlint-enable -->

### HTTP/API {#env-http-api}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envHTTPAPI 0}}
<!-- markdownlint-enable -->

### Confd {#env-confd}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envConfd 0}}
<!-- markdownlint-enable -->

### Git {#env-git}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envGit 0}}
<!-- markdownlint-enable -->

### Sinker {#env-sinker}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envSinker 0}}
<!-- markdownlint-enable -->

### IO {#env-io}

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

### Refer Table {#env-reftab}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envRefta 0}}
<!-- markdownlint-enable -->

### Recorder {#env-recorder}

[:octicons-tag-24: Version-1.22.0](changelog.md#1.22.0)

For more info about recorder, see [here](datakit-tools-how-to.md#record-and-replay).

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envRecorder 0}}
<!-- markdownlint-enable -->

### Remote Job {#remote_job}

[:octicons-tag-24: Version-1.63.0](changelog.md#cl-1.63.0)

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.remote_job 0}}
<!-- markdownlint-enable -->

### Others {#env-others}

<!-- markdownlint-disable MD046 -->
{{ CodeBlock .NonInputENVSample.envOthers 0}}
<!-- markdownlint-enable -->

### Special Environments {#env-special}

#### `ENV_K8S_NODE_NAME` {#env_k8s_node_name}

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

#### ENV_K8s_CLUSTER_NODE_NAME {#env-rename-node}

[:octicons-tag-24: Version-1.36.0](changelog.md#1.36.0)

When multiple clusters share a workspace and contain nodes with identical names, the `ENV_K8S_CLUSTER_NODE_NAME` environment variable can be used to manually customize the collected node name. During deployment, add a new configuration section **after** the `ENV_K8S_NODE_NAME` section in your `datakit.yaml` file:

```yaml
- name: ENV_K8S_CLUSTER_NODE_NAME
  value: cluster_a_$(ENV_K8S_NODE_NAME) # Ensure that ENV_K8S_NODE_NAME is defined beforehand
```

This configuration appends `cluster_a_` to the original hostname, effectively creating a unique identifier for nodes in this cluster. As a result, the `host` tag associated with metrics such as logs, processes, CPU usage, and memory will also be prefixed with `cluster_a_`, enabling better data organization and filtering.

<!-- markdownlint-disable MD013 -->
### Collector-specific Environment Variable {#inputs-envs}
<!-- markdownlint-enable -->

Some collectors support external injection of environment variables to adjust the default configuration of the collector itself. See each specific collector document for details.

## Extended Readings {#more-readings}

<font size=3>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit election</u></font>](election.md)
</div>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Various configurations for DataKit</u></font>](k8s-config-how-to.md)
</div>
</font>
