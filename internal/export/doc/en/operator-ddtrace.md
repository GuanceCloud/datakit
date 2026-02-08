# Injecting DDTrace via DataKit Operator

## Usage Instructions {#datakit-operator-inject-lib-usage}

1. [Download and install DataKit-Operator](datakit-operator.md#install) in the target Kubernetes cluster.
1. Add the following ConfigMap configuration to the Operator:

    ```json
    {
        "server_listen": "0.0.0.0:9543", // Service listening address
        "log_level": "info", // Log level

        "admission_inject_v2": { // Injection config v2 (operator version >= v1.7.0)
            "ddtraces": [ // DDTrace configuration array
                {
                  // Fill in DDTrace injection config here...
                },
                {
                  // Can inject another DDTrace config...
                }
            ]                   
        },
    
        "admission_inject": { // Injection config v1
            "ddtrace": {
                // Fill in DDTrace injection config here...
                // For v1 config, operator only supports injecting one DDTrace config. v2 is recommended for flexibility.
            }
        },
    }
    ```

    <!-- 1. Add the specified Annotation `admission.datakit/java-lib.version: "{{.DDTraceJavaExtVersion}}"` to the deployment, indicating the need to inject the default version of DDtrace Java Agent. -->

    DDTrace injection has the following configurable fields:

    | Field                        | Type    | Description                                                     | Required     | Example Value                    |
    | ------:                      | :-----: | :------                                                         | :---:        | :--------                        |
    | `envs`                       | object  | Environment variable map                                        | Y[^envs]     | See example below                |
    | `image`                      | string  | DDTrace image address                                           | Y[^image]    | See example below                |
    | `label_selectors`            | array   | Label selector array                                            | Y[^selector] | `["app=nginx", "tier=frontend"]` |
    | `language`                   | string  | Supported language type (optional `java`)                       | Y[^lang]     | `"java"`                         |
    | `namespace_selectors`        | array   | Namespace selector, supports regex                              | Y[^selector] | `["prod-*", "test"]`             |
    | `resources`                  | object  | Resource limit configuration                                    | N            | See example below                |
    | ~~`enabled_namespaces`~~     | object  | Select target Kubernetes namespace and set development language | Y            | Deprecated in 1.7.0 `admission_inject_v2` |
    | ~~`enabled_labelselectors`~~ | object  | Select target using Kubernetes label                            | Y            | Deprecated in 1.7.0 `admission_inject_v2` |

    [^selector]: This field must be filled; otherwise, the Operator will reject the injection.
    [^image]: Although the Operator has a built-in default image address, users generally need to copy the image to the internal network for offline environments and then use the internal image address.
    [^lang]: The language selected here must match the content of the corresponding DDTrace image. If it does not match, the injection will fail.
    [^envs]: These environment variable settings are crucial and directly affect the final data outcome. For the supported `fieldRef` list here, see [here](datakit-operator.md#downwardapi).

    Here is an example:

    ```json
    {
        "namespace_selectors": [],
        "label_selectors": [],
        "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:{{.DDTraceJavaExtVersion}}",
        "language": "java",
        "envs": {
             "DD_AGENT_HOST":           "datakit-service.datakit.svc.cluster.local",
             "DD_TRACE_AGENT_PORT":     "9529",
             "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc.cluster.local",
             "DD_JMXFETCH_STATSD_PORT": "8125",
             "DD_SERVICE":              "{fieldRef:metadata.labels['service']}",
             "POD_NAME":                "{fieldRef:metadata.name}",
             "POD_NAMESPACE":           "{fieldRef:metadata.namespace}",
             "NODE_NAME":               "{fieldRef:spec.nodeName}",
             "DD_TAGS":                 "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
         },
        "resources": {
            "requests": {
                "cpu":    "100m",
                "memory": "64Mi"
            },
            "limits": {
                "cpu":    "500m",
                "memory": "512Mi"
            }
        }
    }
    ```

## Handling Special Deployments {#special-deployment}

The Operator configuration above is for DDTrace injection across the entire cluster. Sometimes this one-size-fits-all approach is not suitable for certain specific Deployments. Therefore, we can make some Annotation markings for these Deployments individually:

The Operator can recognize the following two Annotations:

- `admission.datakit/ddtrace.enabled`: Marks whether to enable injection in a single Deployment standard. Fill in `"true"` to enable injection, `"false"` to block injection. If blocked, the Operator will ignore injecting this Deployment.
- `admission.datakit/java-lib.version`: Specifies a specific DDTrace version.

### Annotation Example {#anno-demo}

Mark whether to inject in Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-deployment
  labels:
    app: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
      annotations:
        admission.datakit/ddtrace.enabled: "true"
    spec:
      containers:
      - name: my-app
        image: my-app:1.2.3
        ports:
        - containerPort: 80
```

Inject specific version of `dd-java-lib` into Deployment [^replace-ddtrace-version]:

[^replace-ddtrace-version]: What is replaced here is a different version of the same image address in the Operator ConfigMap; different image addresses cannot be switched here.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-deployment
  labels:
    app: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
      annotations:
        admission.datakit/java-lib.version: "{{.DDTraceJavaExtVersion}}"
    spec:
      containers:
      - name: my-app
        image: my-app:1.2.3
        ports:
        - containerPort: 80
```

Create resources using the yaml file:

```shell
$ kubectl apply -f my-app.yaml
...
```

Verify as follows:

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
my-app-deployment-7bd8dd85f-fzmt2       1/1     Running   0             4s

$ kubectl get pod my-app-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.initContainers\[\*\].name}

datakit-lib-init
```
