# Datakit Operator User Guide

---

:material-kubernetes:

---

## Overview and Installation {#datakit-operator-overview-and-install}

Datakit Operator is a collaborative project between Datakit and Kubernetes orchestration. It aims to assist the deployment of Datakit as well as other functions such as verification and injection.

Currently, Datakit Operator provides the following functions:

- Injection DDTrace SDK(Java/Python/JavaScript) and related environments. See [documentation](datakit-operator.md#datakit-operator-inject-lib).
- Injection Sidecar logfwd to collect Pod logging. See [documentation](datakit-operator.md#datakit-operator-inject-logfwd).
- Support task distribution for Datakit plugins. See [documentation](election.md#plugins-election).

Prerequisites:

- Recommended Kubernetes version 1.24.1 or above and internet access (to download yaml file and pull images).
- Ensure `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` [controllers](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites){:target="_blank"} are enabled.
- Ensure admissionregistration.k8s.io/v1 API is enabled.

### Installation Steps {#datakit-operator-install}

Download [*datakit-operator.yaml*](https://static.guance.com/datakit-operator/datakit-operator.yaml), and follow these steps:

``` shell
kubectl create namespace datakit

wget https://static.guance.com/datakit-operator/datakit-operator.yaml

kubectl apply -f datakit-operator.yaml

kubectl get pod -n datakit
NAME                               READY   STATUS    RESTARTS   AGE
datakit-operator-f948897fb-5w5nm   1/1     Running   0          15s
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - There is a strict correspondence between Datakit-Operator's program and yaml files. If an outdated yaml file is used, it may not be possible to install the new version of Datakit-Operator. Please download the latest yaml file.
    - If you encounter `InvalidImageName` error, you can manually pull the image.
<!-- markdownlint-enable -->

### Relevant Configuration {#datakit-operator-jsonconfig}

[:octicons-tag-24: Datakit Operator v1.2.1]

The configuration for the Datakit Operator is in JSON format and is stored as a separate ConfigMap in Kubernetes, which is loaded into the container as an environment variable.

The default configuration is as follows:

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        "ddtrace": {
            "images": {
                "java_agent_image":   "pubrepo.guance.com/datakit-operator/dd-lib-java-init:v1.8.4-guance",
                "python_agent_image": "pubrepo.guance.com/datakit-operator/dd-lib-python-init:v1.6.2",
                "js_agent_image":     "pubrepo.guance.com/datakit-operator/dd-lib-js-init:v3.9.2"
            },
            "envs": {
                "DD_AGENT_HOST":           "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT":     "9529",
                "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc",
                "DD_JMXFETCH_STATSD_PORT": "8125"
            }
        },
        "logfwd": {
            "images": {
                "logfwd_image": "pubrepo.guance.com/datakit/logfwd:1.5.8"
            }
        }
    }
}
```

In `admission_inject`, you can configure `ddtrace` and `logfwd` more finely:

- `images` is a collection of Key/Value pairs with fixed keys, where modifying the Value allows for customisation of image paths.

<!-- markdownlint-disable MD046 -->
???+ info

    The Datakit Operator's ddtrace agent image is stored centrally at pubrepo.guance.com/datakit-operator. For certain special environments that may not have access to this image repository, it is possible to modify the environment variables and specify an image path, as follows:
    
    1. In an environment that can access pubrepo.guance.com, pull the image pubrepo.guance.com/datakit-operator/dd-lib-java-init:v1.8.4-guance and save it to your own image repository, for example inside.image.hub/datakit-operator/dd-lib-java-init:v1.8.4-guance.
    1. Modify the JSON configuration by changing admission_inject->ddtrace->images->java_agent_image to inside.image.hub/datakit-operator/dd-lib-java-init:v1.8.4-guance, and apply this YAML.
    1. Thereafter, the Datakit Operator will use the new Java Agent image path.
    
    **The Datakit Operator does not check images. If the image path is incorrect, Kubernetes will throw an error when creating the image.**
    
    If a version has already been specified in the admission.datakit/java-lib.version annotation, for example admission.datakit/java-lib.version:v2.0.1-guance or admission.datakit/java-lib.version:latest, the v2.0.1-guance version will be used.
<!-- markdownlint-enable -->

- `envs` is also a collection of Key/Value pairs, but with variable keys and values. The Datakit Operator injects all Key/Value environment variables into the target container. For example, add FAKE_ENV to envs:

```json
    "admission_inject": {
        "ddtrace": {
            "images": {
                "java_agent_image":   "pubrepo.guance.com/datakit-operator/dd-lib-java-init:v1.8.4-guance",
                "python_agent_image": "pubrepo.guance.com/datakit-operator/dd-lib-python-init:v1.6.2",
                "js_agent_image":     "pubrepo.guance.com/datakit-operator/dd-lib-js-init:v3.9.2"
            },
            "envs": {
                "DD_AGENT_HOST":           "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT":     "9529",
                "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc",
                "DD_JMXFETCH_STATSD_PORT": "8125",
                "FAKE_ENV":                "ok"
            }
        }
    }
```

All containers that have `ddtrace` agent injected into them will have five environment variables added to their `envs`.

## Using Datakit-Operator to Inject Files and Programs {#datakit-operator-inject-sidecar}

In large Kubernetes clusters, it can be quite difficult to make bulk configuration changes. Datakit-Operator will determine whether or not to modify or inject data based on Annotation configuration.

The following functions are currently supported:

- Injection of `ddtrace` agent and environment
- Mounting of `logfwd` sidecar and enabling log collection

???+ info

    Only version v1 of `deployments/daemonsets/cronjobs/jobs/statefulsets` Kind is supported, and because Datakit-Operator actually operates on the PodTemplate, Pod is not supported. In this article, we will use `Deployment` to describe these five kinds of Kind.

### Injection of ddtrace Agent and Relevant Environment Variables {#datakit-operator-inject-lib}

#### Usage {#datakit-operator-inject-lib-usage}

1. On the target Kubernetes cluster, [download and install Datakit-Operator](datakit-operator.md#datakit-operator-inject-lib).
2. Add a specified Annotation in deployment, indicating the need to inject ddtrace files. Note that the Annotation needs to be added in the template.
    - The key is `admission.datakit/%s-lib.version`, where %s needs to be replaced with the specified language. Currently supports `java`, `python` and `js`.
    - The value is the specified version number. If left blank, the default image version of the environment variable will be used.

#### Example {#datakit-operator-inject-lib-example}

The following is an example of Deployment that injects `dd-js-lib` into all Pods created by Deployment:

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
        admission.datakit/js-lib.version: ""
    spec:
      containers:
      - name: nginx
        image: nginx:1.22
        ports:
        - containerPort: 80
```

Create a resource using yaml file:

```shell
$ kubectl apply -f nginx.yaml
```

Verify as follows:

```shell
$ kubectl get pod
$ NAME                                   READY   STATUS    RESTARTS      AGE
nginx-deployment-7bd8dd85f-fzmt2       1/1     Running   0             4s
$ kubectl get pod nginx-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.initContainers\[\*\].name}
$ datakit-lib-init
```

### Injecting Logfwd Program and Enabling Log Collection {#datakit-operator-inject-logfwd}

#### Prerequisites {#datakit-operator-inject-logfwd-prerequisites}

[logfwd](logfwd.md#using) is a proprietary log collection application for Datakit. To use it, you need to first deploy Datakit in the same Kubernetes cluster and satisfy the following two conditions:

1. The Datakit `logfwdserver` collector is enabled, for example, listening on port `9533`.
2. The Datakit service needs to open port `9533` to allow other Pods to access `datakit-service.datakit.svc:9533`.

#### Instructions {#datakit-operator-inject-logfwd-instructions}

1. On the target Kubernetes cluster, [download and install Datakit-Operator](datakit-operator.md#datakit-operator-inject-lib).
2. In the deployment, add the specified Annotation to indicate that a logfwd sidecar needs to be mounted. Note that the Annotation should be added in the template.
    - The key is uniformly `admission.datakit/logfwd.instances`.
    - The value is a JSON string of specific logfwd configuration, as shown below:

```json
[
    {
        "datakit_addr": "datakit-service.datakit.svc:9533",
        "loggings": [
            {
                "logfiles": ["<your-logfile-path>"],
                "ignore": [],
                "source": "<your-source>",
                "service": "<your-service>",
                "pipeline": "<your-pipeline.p>",
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

Parameter explanation can refer to [logfwd configuration](logfwd.md#config):

- `datakit_addr` is the Datakit logfwdserver address.
- `loggings` is the main configuration and is an array that can refer to [Datakit logging collector](logging.md).
    - `logfiles` is a list of log files, which can specify absolute paths and support batch specification using glob rules. Absolute paths are recommended.
    - `ignore` filters file paths using glob rules. If it meets any filtering condition, the file will not be collected.
    - `source` is the data source. If it is empty, `'default'` will be used by default.
    - `service` adds a new tag. If it is empty, `$source` will be used by default.
    - `pipeline` is the pipeline script path. If it is empty, `$source.p` will be used. If `$source.p` does not exist, the pipeline will not be used. (This script file exists on the DataKit side.)
    - `character_encoding` selects an encoding. If the encoding is incorrect, the data cannot be viewed. It is recommended to leave it blank. Supported encodings include `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030`, or "".
    - `multiline_match` is for multiline matching, as described in [Datakit Log Multiline Configuration](logging.md#multiline). Note that since it is in the JSON format, it does not support the "unescaped writing method" of three single quotes. The regex `^\d{4}` needs to be written as `^\\d{4}` with an escape character.
    - `tags` adds additional tags in JSON map format, such as `{ "key1":"value1", "key2":"value2" }`.

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
```

Verify as follows:

```shell
$ kubectl get pod
$ NAME                                   READY   STATUS    RESTARTS      AGE
logging-deployment-5d48bf9995-vt6bb       1/1     Running   0             4s
$ kubectl get pod logging-deployment-5d48bf9995-vt6bb -o=jsonpath={.spec.containers\[\*\].name}
$ log-container datakit-logfwd
```

Finally, you can check whether the logs have been collected on the Observability Cloud Log Platform.

---

References：

- Kubernetes [Admission Controlle](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/){:target="_blank"}
