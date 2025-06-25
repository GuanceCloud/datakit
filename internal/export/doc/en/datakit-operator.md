# DataKit Operator User Guide

---

:material-kubernetes:

---

## Overview and Installation {#datakit-operator-overview-and-install}

DataKit Operator is a collaborative project between DataKit and Kubernetes orchestration. It aims to assist the deployment of DataKit as well as other functions such as verification and injection.

Currently, DataKit Operator provides the following functions:

- Injection DDTrace Java SDK and related environments. See [documentation](datakit-operator.md#datakit-operator-inject-lib).
- Injection Sidecar logfwd to collect Pod logging. See [documentation](datakit-operator.md#datakit-operator-inject-logfwd).
- Support task distribution for DataKit plugins. See [documentation](election.md#plugins-election).

Prerequisites:

- Recommended Kubernetes version 1.24.1 or above and internet access (to download yaml file and pull images).
- Ensure `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` [controllers](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites){:target="_blank"} are enabled.
- Ensure `admissionregistration.k8s.io/v1` API is enabled.

### Installation Steps {#datakit-operator-install}

<!-- markdownlint-disable MD046 -->
=== "Deployment"

    Download [*datakit-operator.yaml*](https://static.<<<custom_key.brand_main_domain>>>/datakit-operator/datakit-operator.yaml){:target="_blank"}, and follow these steps:
    
    ``` shell
    $ kubectl create namespace datakit
    $ wget https://static.<<<custom_key.brand_main_domain>>>/datakit-operator/datakit-operator.yaml
    $ kubectl apply -f datakit-operator.yaml
    $ kubectl get pod -n datakit
    
    NAME                               READY   STATUS    RESTARTS   AGE
    datakit-operator-f948897fb-5w5nm   1/1     Running   0          15s
    ```

=== "Helm"

    Precondition:

    * Kubernetes >= 1.14
    * Helm >= 3.0+

    ```shell
    $ helm install datakit-operator datakit-operator \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
         -n datakit --create-namespace
    ```

    View deployment status:

    ```shell
    $ helm -n datakit list
    ```

    Upgrade with the following command:

    ```shell
    $ helm -n datakit get values datakit-operator -a -o yaml > values.yaml
    $ helm upgrade datakit-operator datakit-operator \
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
        -n datakit \
        -f values.yaml
    ```

    Uninstall with the following command:

    ```shell
    $ helm uninstall datakit-operator -n datakit
    ```
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ attention

    - There is a strict correspondence between DataKit-Operator's program and yaml files. If an outdated yaml file is used, it may not be possible to install the new version of DataKit-Operator. Please download the latest yaml file.
    - If you encounter `InvalidImageName` error, you can manually pull the image.
<!-- markdownlint-enable -->

## Configuration Explanation {#datakit-operator-jsonconfig}

[:octicons-tag-24: Version-1.4.2](changelog.md#cl-1.4.2)

The DataKit Operator configuration is in JSON format and is stored in Kubernetes as a separate ConfigMap. It is loaded into the container as environment variables.

The default configuration is as follows:

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        "ddtrace": {
            "enabled_namespaces":     [],
            "enabled_labelselectors": [],
            "images": {
                "java_agent_image":   "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:v1.30.1-guance",
            },
            "envs": {
                "DD_AGENT_HOST":           "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT":     "9529",
                "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc",
                "DD_JMXFETCH_STATSD_PORT": "8125",
                "DD_SERVICE":              "{fieldRef:metadata.labels['service']}",
                "POD_NAME":                "{fieldRef:metadata.name}",
                "POD_NAMESPACE":           "{fieldRef:metadata.namespace}",
                "NODE_NAME":               "{fieldRef:spec.nodeName}",
                "DD_TAGS":                 "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
            }
        },
        "logfwd": {
            "images": {
                "logfwd_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:1.28.1"
            }
        },
        "profiler": {
            "images": {
                "java_profiler_image":   "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/async-profiler:0.1.0",
                "python_profiler_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/py-spy:0.1.0",
                "golang_profiler_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/go-pprof:0.1.0"
            },
            "envs": {
                "DK_AGENT_HOST":  "datakit-service.datakit.svc",
                "DK_AGENT_PORT":  "9529",
                "DK_PROFILE_VERSION": "1.2.333",
                "DK_PROFILE_ENV": "prod",
                "DK_PROFILE_DURATION": "240",
                "DK_PROFILE_SCHEDULE": "0 * * * *"
            }
        }
    },
    "admission_mutate": {
        "loggings": [
            {
                "namespace_selectors": ["test01"],
                "label_selectors":     ["app=logging"],
                "config":"[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/**/*.log\",\"storage_index\":\"logging-index\"\"source\":\"logging-tmp\"},{\"disable\":true,\"type\":\"file\",\"path\":\"/var/log/opt/**/*.log\",\"source\":\"logging-var\"}]"
            }
        ]
    }
}
```

The main configuration items are `ddtrace`, `logfwd`, and `profiler`, which specify the injected images and environment variables. In addition, `ddtrace` also supports batch injection based on `enabled_namespaces` and `enabled_selectors`, as detailed in the section below.

### Configuration of Images {#datakit-operator-config-images}

The primary function of the DataKit Operator is to inject images and environment variables, using the `images` configuration to specify the image addresses. The `images` configuration consists of multiple Key/Value pairs, where the Key is fixed, and the Value is modified to specify the image address.

Under normal circumstances, images are stored in `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator`. However, for some special environments where accessing this image repository is not convenient, you can use the following method (taking the `dd-lib-java-init` image as an example):

1. In an environment where `pubrepo.<<<custom_key.brand_main_domain>>>` is accessible, pull the image `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:v1.30.1-guance`, and then re-store it in your own image repository, for example, `inside.image.hub/datakit-operator/dd-lib-java-init:v1.30.1-guance`.
1. Modify the JSON configuration, changing `admission_inject`->`ddtrace`->`images`->`java_agent_image` to `inside.image.hub/datakit-operator/dd-lib-java-init:v1.30.1-guance`, and apply this YAML file.
1. After this, the DataKit Operator will use the new Java Agent image path.

<!-- markdownlint-disable MD046 -->
???+ attention

    The DataKit Operator does not validate the image. If the image path is incorrect, Kubernetes will throw an error when creating the Pod.**
<!-- markdownlint-enable -->

### Adding Environment Variables {#datakit-operator-config-envs}

All environment variables that need to be injected must be specified in the configuration file, as the DataKit Operator does not add any environment variables by default.

The environment variable configuration is called `envs`, `envs` consists of multiple Key/Value pairs: the Key is a fixed value, and the Value can either be a fixed value or a placeholder, depending on the actual situation.

For example, to add an environment variable `testing-env` in `envs`:

```json
    "admission_inject": {
        "ddtrace": {
            # other..
            "envs": {
                "DD_AGENT_HOST":           "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT":     "9529",
                "FAKE_ENV":                "ok"
            }
        }
    }
```

All containers that have `ddtrace` agent injected into them will have five environment variables added to their `envs`.

In DataKit Operator v1.4.2 and later versions, `envs` `envs` support for the Kubernetes Downward API [environment variable fetch field](https://kubernetes.io/docs/concepts/workloads/pods/downward-api/#available-fields). The following are now supported:

- `metadata.name`: The pod's name.
- `metadata.namespace`:  The pod's namespace.
- `metadata.uid`:  The pod's unique ID.
- `metadata.annotations['<KEY>']`:  The value of the pod's annotation named `<KEY>` (for example, metadata.annotations['myannotation']).
- `metadata.labels['<KEY>']`:  The text value of the pod's label named `<KEY>` (for example, metadata.labels['mylabel']).
- `spec.serviceAccountName`:  The name of the pod's service account.
- `spec.nodeName`:  The name of the node where the Pod is executing.
- `status.hostIP`:  The primary IP address of the node to which the Pod is assigned.
- `status.hostIPs`:  The IP addresses is a dual-stack version of status.hostIP, the first is always the same as status.hostIP. The field is available if you enable the PodHostIPs feature gate.
- `status.podIP`:  The pod's primary IP address (usually, its IPv4 address).
- `status.podIPs`:  The IP addresses is a dual-stack version of status.podIP, the first is always the same as status.podIP.

For example, if there is a Pod with the name `nginx-123` and the namespace `middleware`, and you want to inject the environment variables `POD_NAME` and `POD_NAMESPACE`, refer to the following configuration:

```json
{
    "admission_inject": {
        "ddtrace": {
            "envs": {
                "POD_NAME":      "{fieldRef:metadata.name}",
                "POD_NAMESPACE": "{fieldRef:metadata.namespace}"
            }
        }
    }
}
```

Eventually, the environment variables can be seen in the Pod:

``` shell
$ env | grep POD
POD_NAME=nginx-123
POD_NAMESPACE=middleware
```

<!-- markdownlint-disable MD046 -->
???+ attention

    If the placeholder in the Value is unrecognized, it will be added to the environment variable as a plain string. For example, `"POD_NAME": "{fieldRef:metadata.PODNAME}"` is incorrect, and the environment variable will be set as `POD_NAME={fieldRef:metadata.PODNAME}`.
<!-- markdownlint-enable -->

## Injection Methods {#datakit-operator-inject}

DataKit-Operator supports two methods for resource injection `global configuration namespaces and selectors`, and adding specific annotations to target Pods. The differences between them are as follows:

- Global Configuration: Namespace and Selector: By modifying the DataKit-Operator configuration, you specify the target Pod's Namespace and Selector. If a Pod matches the criteria, the resource injection will occur.
    - Advantages: No need to add annotations to the target Pod (but the target Pod must be restarted).
    - Disadvantages: The scope is not precise enough, which may lead to unnecessary injections.

- Adding Annotations to Target Pods: Add annotations to the target Pod, and DataKit-Operator will check the Pod's annotations to decide whether to perform the injection based on the conditions.
    - Advantages: Precise scope, preventing unnecessary injections.
    - Disadvantages: You must manually add annotations to the target Pod, and the Pod needs to be restarted.

<!-- markdownlint-disable MD046 -->
???+ attention

    As of DataKit-Operator v1.5.8, the `global configuration namespaces and selectors` method only applies to `DDtrace injection`. It does not apply to `logfwd` and `profiler`, for which annotations are still required.
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### Global Configuration: Namespaces and Selectors {#datakit-operator-config-ddtrace-enabled}
<!-- markdownlint-enable -->

The `enabled_namespaces` and `enabled_labelselectors` fields are specific to `ddtrace`. They are object arrays that require the specification of `namespace` and `language`. The relationships between the arrays are "OR" (i.e., any match in the array will trigger injection). The configuration is written as follows (refer to the configuration details later):

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        "ddtrace": {
            "enabled_namespaces": [
                {
                    "namespace": "testns",  # Specify the namespace
                    "language": "java"      # Specify the language for the agent to inject
                }
            ],
            "enabled_labelselectors": [
                {
                    "labelselector": "app=log-output",  # Specify the label selector
                    "language": "java"                  # Specify the language for the agent to inject
                }
            ]
            # other..
        }
    }
}
```

If a Pod satisfies both the `enabled_namespaces` rule and the `enabled_labelselectors` rule, the configuration in `enabled_labelselectors` will take precedence (usually applied when the `language` value is used).

For guidelines on how to write `labelselector`, please refer to the [official documentation](https://kubernetes.io/en/docs/concepts/overview/working-with-objects/labels/#label-selectors).

<!-- markdownlint-disable MD046 -->
???+ note

    - In Kubernetes versions 1.16.9 or earlier, Admission does not record the Pod Namespace, so the `enabled_namespaces` feature cannot be used.
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### Adding Annotation Configuration for Injection {#datakit-operator-config-annotation}
<!-- markdownlint-enable -->

To inject the `ddtrace` file into a Pod, add the specified annotation to the Deployment. Make sure to add the annotation to the `template` section.

The annotation format is as follows:

- The key is `admission.datakit/%s-lib.version`, where `%s` should be replaced with the desired language. Currently, it supports `java`.
- The value is the specified version number. By default, it uses the version specified by the DataKit-Operator configuration in the `java_agent_image` setting. For more details, see the configuration explanation below.

For example, to add an annotation:

```yaml
      annotations:
        admission.datakit/java-lib.version: "v1.36.2-guance"
```

This indicates that the image version to be injected for this Pod is `v1.36.2-guance`. The image address is taken from the configuration `admission_inject` -> `ddtrace` -> `images` -> `java_agent_image`, where the image version is replaced with `"v1.36.2-guance"`, similar to `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:v1.36.2-guance`.

<!-- markdownlint-disable MD013 -->
## Using DataKit-Operator to Inject Files and Programs {#datakit-operator-inject-sidecar}
<!-- markdownlint-enable -->

In large Kubernetes clusters, it can be quite difficult to make bulk configuration changes. DataKit-Operator will determine whether or not to modify or inject data based on Annotation configuration.

The following functions are currently supported:

- Injection of `ddtrace` agent and environment
- Mounting of `logfwd` sidecar and enabling log collection
- Inject [`async-profiler`](https://github.com/async-profiler/async-profiler){:target="_blank"} for JVM profiling [:octicons-beaker-24: Experimental](index.md#experimental)
- Inject [`py-spy`](https://github.com/benfred/py-spy){:target="_blank"} for Python profiling [:octicons-beaker-24: Experimental](index.md#experimental)

<!-- markdownlint-disable MD046 -->
???+ info

    Only version v1 of `deployments/daemonsets/cronjobs/jobs/statefulsets` Kind is supported, and because DataKit-Operator actually operates on the PodTemplate, Pod is not supported. In this article, we will use `Deployment` to describe these five kinds of Kind.
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### DDTrace Agent {#datakit-operator-inject-lib}
<!-- markdownlint-enable -->

#### Usage {#datakit-operator-inject-lib-usage}

1. On the target Kubernetes cluster, [download and install DataKit-Operator](datakit-operator.md#datakit-operator-overview-and-install).
1. Add a Annotation `admission.datakit/java-lib.version: ""` in deployment.

#### Example {#datakit-operator-inject-lib-example}

The following is an example of Deployment that injects `dd-java-lib` into all Pods created by Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
      annotations:
        admission.datakit/java-lib.version: ""
    spec:
      containers:
      - name: nginx
        image: nginx:1.22
        ports:
        - containerPort: 80
```

Create a resource using yaml file:

```shell
kubectl apply -f nginx.yaml
```

Verify as follows:

```shell
$ kubectl get pod
NAME                                   READY   STATUS    RESTARTS      AGE
nginx-deployment-7bd8dd85f-fzmt2       1/1     Running   0             4s

$ kubectl get pod nginx-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.initContainers\[\*\].name}
datakit-lib-init
```

<!-- markdownlint-disable MD013 -->
### logfwd {#datakit-operator-inject-logfwd}
<!-- markdownlint-enable -->

#### Prerequisites {#datakit-operator-inject-logfwd-prerequisites}

[logfwd](../integrations/logfwd.md#using) is a proprietary log collection application for DataKit. To use it, you need to first deploy DataKit in the same Kubernetes cluster and satisfy the following two conditions:

1. The DataKit `logfwdserver` collector is enabled, for example, listening on port `9533`.
2. The DataKit service needs to open port `9533` to allow other Pods to access `datakit-service.datakit.svc:9533`.

#### Instructions {#datakit-operator-inject-logfwd-instructions}

1. On the target Kubernetes cluster, [download and install DataKit-Operator](datakit-operator.md#datakit-operator-overview-and-install).
1. In the deployment, add the specified Annotation to indicate that a logfwd sidecar needs to be mounted. Note that the Annotation should be added in the template.
    - The key is uniformly `admission.datakit/logfwd.instances`.
    - The value is a JSON string of specific logfwd configuration, as shown below:

```json
[
    {
        "datakit_addr": "datakit-service.datakit.svc:9533",
        "loggings": [
            {
                "logfiles":      ["<your-logfile-path>"],
                "ignore":        [],
                "storage_index": "<your-storage-index>",
                "source":        "<your-source>",
                "service":       "<your-service>",
                "pipeline":      "<your-pipeline.p>",
                "character_encoding": "",
                "multiline_match": "<your-match>",
                "tags": {}
            },
            {
                "logfiles": ["<your-logfile-path-2>"],
                "source": "<your-source-2>"
            }
        ]
    }
]
```

Parameter explanation can refer to [logfwd configuration](../integrations/logfwd.md#config):

- `datakit_addr` is the DataKit logfwdserver address.
- `loggings` is the main configuration and is an array that can refer to [DataKit logging collector](../integrations/logging.md).
    - `logfiles` is a list of log files, which can specify absolute paths and support batch specification using glob rules. Absolute paths are recommended.
    - `ignore` filters file paths using glob rules. If it meets any filtering condition, the file will not be collected.
    - `storage_index` set storage index.
    - `source` is the data source. If it is empty, `'default'` will be used by default.
    - `service` adds a new tag. If it is empty, `$source` will be used by default.
    - `pipeline` is the Pipeline script path. If it is empty, `$source.p` will be used. If `$source.p` does not exist, the Pipeline will not be used. (This script file exists on the DataKit side.)
    - `character_encoding` selects an encoding. If the encoding is incorrect, the data cannot be viewed. It is recommended to leave it blank. Supported encodings include `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030`, or "".
    - `multiline_match` is for multiline matching, as described in [DataKit Log Multiline Configuration](../integrations/logging.md#multiline). Note that since it is in the JSON format, it does not support the "unescaped writing method" of three single quotes. The regex `^\d{4}` needs to be written as `^\\d{4}` with an escape character.
    - `tags` adds additional tags in JSON map format, such as `{ "key1":"value1", "key2":"value2" }`.

<!-- markdownlint-disable MD046 -->
???+ attention

    That there is a difference between paths with and without a trailing slash. `/var/log` and `/var/log/` are considered different paths and cannot be reused.
<!-- markdownlint-enable -->

#### Example {#datakit-operator-inject-logfwd-example}

Here is an example Deployment that continuously writes data to a file using shell and configures the collection of that file:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: logging-deployment
  labels:
    app: logging
spec:
  replicas: 1
  selector:
    matchLabels:
      app: logging
  template:
    metadata:
      labels:
        app: logging
      annotations:
        admission.datakit/logfwd.instances: '[{"datakit_addr":"datakit-service.datakit.svc:9533","loggings":[{"logfiles":["/var/log/log-test/*.log"],"source":"deployment-logging","tags":{"key01":"value01"}}]}]'
    spec:
      containers:
      - name: log-container
        image: busybox
        args: [/bin/sh, -c, 'mkdir -p /var/log/log-test; i=0; while true; do printf "$(date "+%F %H:%M:%S") [%-8d] Bash For Loop Examples.\\n" $i >> /var/log/log-test/1.log; i=$((i+1)); sleep 1; done']
```

Creating Resources Using yaml File:

```shell
$ kubectl apply -f logging.yaml
...
```

Verify as follows:

```shell
$ kubectl get pod
NAME                                   READY   STATUS    RESTARTS      AGE
logging-deployment-5d48bf9995-vt6bb       1/1     Running   0             4s

$ kubectl get pod logging-deployment-5d48bf9995-vt6bb -o=jsonpath={.spec.containers\[\*\].name}
log-container datakit-logfwd
```

Finally, you can check whether the logs have been collected on the <<<custom_key.brand_name>>> Log Platform.

## DataKit Operator Resource Changes {#datakit-operator-mutate-resource}

### Adding Configuration for DataKit Logging {#add-logging-configs}

The DataKit Operator can automatically add the configuration required for DataKit Logging collection to the specified Pods, including the `datakit/logs` annotation and the corresponding file path volume/volumeMount. This simplifies the tedious manual configuration steps. As a result, users do not need to manually intervene in each Pod's configuration to enable log collection functionality automatically.

Below is an example of a configuration that shows how to implement the automatic injection of log collection configuration through the DataKit Operator's `admission_mutate` configuration:

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        # Other configurations...
    },
    "admission_mutate": {
        "loggings": [
            {
                "namespace_selectors": ["middleware"],
                "label_selectors":     ["app=logging"],
                "config": "[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/**/*.log\",\"source\":\"logging-tmp\"}]"
            }
        ]
    }
}
```

`admission_mutate.loggings`: This is an array of objects that contains multiple log collection configurations. Each log configuration includes the following fields:

- `namespace_selectors`: Specifies the namespaces where Pods must be located to meet the criteria. Multiple namespaces can be set, and a Pod must match at least one namespace to be selected. It operates as an "OR" relation with `label_selectors`.
- `label_selectors`: Specifies the labels of Pods that must meet the criteria. A Pod must match at least one label selector to be selected. It operates as an "OR" relation with `namespace_selectors`.
- `config`: This is a JSON string that will be added to the Pod's annotation under the key `datakit/logs`. If the key already exists, it will not be overwritten or added again. This configuration tells DataKit how to collect logs.

The DataKit Operator will automatically parse the `config` configuration and create corresponding volumes and volumeMounts for the Pod based on the paths (`path`) specified within.

Taking the above DataKit Operator configuration as an example, if a Pod's Namespace is middleware or its Labels match app=logging, an annotation and mount will be added to the Pod. For example:

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    datakit/logs: '[{"disable":false,"type":"file","path":"/tmp/opt/**/*.log","source":"logging-tmp"}]'
  labels:
    app: logging
  name: logging-test
  namespace: default
spec:
  containers:
  - args:
    - |
      mkdir -p /tmp/opt/log1;
      i=1;
      while true; do
        echo "Writing logs to file ${i}.log";
        for ((j=1;j<=10000000;j++)); do
          echo "$(date +'%F %H:%M:%S')  [$j]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log1/file_${i}.log;
          sleep 1;
        done;
        echo "Finished writing 5000000 lines to file_${i}.log";
        i=$((i+1));
      done
    command:
    - /bin/bash
    - -c
    - --
    image: pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04
    imagePullPolicy: IfNotPresent
    name: demo
    volumeMounts:
    - mountPath: /tmp/opt
      name: datakit-logs-volume-0
  volumes:
  - emptyDir: {}
    name: datakit-logs-volume-0
```

This Pod has the label `app=logging`, which allows it to match the selector. As a result, DataKit Operator adds the `datakit/logs` annotation and mounts an EmptyDir volume at the path `/tmp/opt`.

Once DataKit Log Collection detects the Pod, it will customize the log collection according to the contents of the `datakit/logs` annotation.

### FAQ {#datakit-operator-faq}

- How to specify that a certain Pod should not be injected? Add the annotation `"admission.datakit/enabled": "false"` to the Pod. This will prevent any actions from being performed on it, with the highest priority.

- DataKit-Operator utilizes Kubernetes Admission Controller functionality for resource injection. For detailed mechanisms, please refer to the [official documentation](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/){:target="_blank"}
