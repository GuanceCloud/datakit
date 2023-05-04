# Datakit Operator User Guide

---

:material-kubernetes:

---

## Overview and Installation {#datakit-operator-overview-and-install}

Datakit Operator is a collaborative project between Datakit and Kubernetes orchestration. It aims to assist the deployment of Datakit as well as other functions such as verification and injection.

Currently, Datakit Operator provides the following functions:

- Injection DDTrace SDK(Java/Python/JavaScript) and related environments. See [documentation](datakit-operator.md#datakit-operator-inject-lib).
- Injection Sidecar logfwd to collect Pod logging. See [documentation](datakit-operator.md#datakit-operator-inject-logfwd).

Prerequisites:

- Recommended Kubernetes version 1.24.1 or above and internet access (to download yaml file and pull images).
- Ensure `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` [controllers](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites) are enabled.
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

???+ attention

    - There is a strict correspondence between Datakit-Operator's program and yaml files. If an outdated yaml file is used, it may not be possible to install the new version of Datakit-Operator. Please download the latest yaml file.
    - If you encounter `InvalidImageName` error, you can manually pull the image.

## Using Datakit-Operator to Inject Files and Programs {#datakit-operator-inject-sidecar}

In large Kubernetes clusters, it can be quite difficult to make bulk configuration changes. Datakit-Operator will determine whether or not to modify or inject data based on Annotation configuration.

The following functions are currently supported:

- Injection of `dd-lib` files and environment
- Mounting of `logfwd` sidecar and enabling log collection

???+ info

    Only version v1 of `deployments/daemonsets/cronjobs/jobs/statefulsets` Kind is supported, and because Datakit-Operator actually operates on the PodTemplate, Pod is not supported. In this article, we will use `Deployment` to describe these five kinds of Kind.

### Injection of dd-lib Files and Relevant Environment Variables {#datakit-operator-inject-lib}

#### Usage {#datakit-operator-inject-lib-usage}

1. On the target Kubernetes cluster, [download and install Datakit-Operator](datakit-operator.md#datakit-operator-inject-lib).
2. Add a specified Annotation in deployment, indicating the need to inject dd-lib files. Note that the Annotation needs to be added in the template.
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

#### Related Configuration {#datakit-operator-inject-lib-configurations}

Datakit-Operator supports the following environment variable configurations (modified in datakit-operator.yaml):

| Environment Variable Name   | Default Value                                                         | Configuration Meaning    |
| :----                       | :----                                                                 | :----                    |
| `ENV_DD_JAVA_AGENT_IMAGE`   | `pubrepo.jiagouyun.com/datakit-operator/dd-lib-java-init:v1.8.4-guance` | Java lib image path      |
| `ENV_DD_PYTHON_AGENT_IMAGE` | `pubrepo.jiagouyun.com/datakit-operator/dd-lib-python-init:v1.6.2`      | Python lib image path    |
| `ENV_DD_JS_AGENT_IMAGE`     | `pubrepo.jiagouyun.com/datakit-operator/dd-lib-js-init:v3.9.2`          | Js lib image path        |
| `ENV_DD_AGENT_HOST`         | `datakit-service.datakit.svc`                                         | Specify the Datakit host |
| `ENV_DD_TRACE_AGENT_PORT`   | `"9529"`                                                              | Specify the Datakit port |

**Datakit-Operator does not check images. If the image path is incorrect, Kubernetes will report an error when creating it.**

The dd-lib image of Datakit-Operator is stored in `pubrepo.jiagouyun.com/datakit-operator`. However, it may not be convenient to access this image repository for some special environments. In this case, you can modify the environment variable and specify the image path as follows:

1. In an environment where `pubrepo.jiagouyun.com` can be accessed, pull the image `pubrepo.jiagouyun.com/datakit-operator/dd-lib-java-init:v1.8.4-guance` and save it to your own image repository, such as `inside.image.hub/datakit-operator/dd-lib-java-init:v1.8.4-guance`.
2. Modify datakit-operator.yaml and change the environment variable `ENV_DD_JAVA_AGENT_IMAGE` to `inside.image.hub/datakit-operator/dd-lib-java-init:v1.8.4-guance`, then apply this yaml.
3. After that, Datakit-Operator will use the new Java lib image path.

> If a version has already been specified in the `admission.datakit/java-lib.version` annotation, such as `admission.datakit/java-lib.version:v2.0.1-guance` or `admission.datakit/java-lib.version:latest`, it will use that version.

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

#### Related Configuration {#datakit-operator-inject-logfwd-configurations}

Datakit-Operator supports the following environment variables (modify them in `datakit-operator.yaml`):

| Environment Variable Name | Default Value                                  | Configuration Meaning |
| :------------------------ | :-------------------------------------------- | :-------------------- |
| `ENV_LOGFWD_IMAGE`        | `pubrepo.jiagouyun.com/datakit/logfwd:1.5.8` | logfwd image path      |
