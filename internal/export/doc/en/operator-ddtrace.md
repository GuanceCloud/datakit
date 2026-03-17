# Injecting DDTrace via DataKit Operator

## Usage Instructions {#datakit-operator-inject-lib-usage}

1. [Download and install DataKit-Operator](datakit-operator.md#install) in the target Kubernetes cluster.
1. Add the following ConfigMap configuration to the Operator:

    ```json
    {
        "server_listen": "0.0.0.0:9543", // Service listening address
        "log_level": "info", // Log level

        "admission_inject_v2": { // Injection config v2 (operator version >= v1.8.0)
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

<!-- markdownlint-disable MD013 -->
### `check_annotation` Configuration Item Explanation {#check-annotation-config}
<!-- markdownlint-enable -->

`check_annotation` is an important configuration field used to control how DataKit Operator handles **version annotations** on Pods (e.g., `admission.datakit/java-lib.version`). The values and behaviors of this field are as follows:

| Value | Behavior Description |
|----------|--------------------------------------------------------------------------|
| `false`  | **(Default)** Ignores **version annotation** checks on Pods, injects directly based on selector rules |
| `true`   | Enables **version annotation** checks, only injects if version annotations exist on matching Pods |

#### Important Logic Explanation {#important-logic}

1. **Function-specific annotations always effective**:
   - `admission.datakit/ddtrace.enabled` is **not affected** by `check_annotation` configuration
   - Regardless of whether `check_annotation` is `true` or `false`, `admission.datakit/ddtrace.enabled` will be checked
   - If `admission.datakit/ddtrace.enabled: "false"`, injection will be directly rejected

2. **Version annotations controlled by `check_annotation`**:
   - `admission.datakit/java-lib.version` **is affected** by `check_annotation` configuration
   - When `check_annotation: true`, version annotations must exist to inject
   - When `check_annotation: false`, version annotation checks are ignored

3. **Global annotations always effective**:
   - `admission.datakit/enabled` is **not affected** by `check_annotation` configuration
   - If `admission.datakit/enabled: "false"`, all injections will be completely rejected (highest priority)

Supported DDTrace-related Annotations:

| Annotation | Description | Values | Affected by `check_annotation` | Explanation |
|--------------------------------------|----------------------------------------|------------------|---------------------------|----------------------------------------------------------------------|
| `admission.datakit/ddtrace.enabled` | Controls DDTrace injection | `"true"`/`"false"` | **No** | `"true"`: allows injection; `"false"`: rejects injection; not set: determined by rule matching |
| `admission.datakit/java-lib.version` | Specifies DDTrace Java Agent version | Version string | **Yes** | e.g., `"1.12.0"`, used to override default image version in configuration |
| `admission.datakit/enabled` | Controls all injection functions (highest priority) | `"true"`/`"false"` | **No** | `"false"`: completely rejects any injection, highest priority |

#### When `check_annotation: true` {#when-check-annotation-true}

The following conditions must be simultaneously met to perform injection:

1. **Configuration matches**: Pod must match `namespace_selectors` and `label_selectors` rules
2. **Function annotation allows**: `admission.datakit/ddtrace.enabled` is not `"false"` (if exists)
3. **Version annotation exists**: Version annotations must exist on Pod (e.g., `admission.datakit/java-lib.version`)

#### When `check_annotation: false` {#when-check-annotation-false}

The following conditions must be met to perform injection:

1. **Configuration matches**: Pod must match `namespace_selectors` and `label_selectors` rules
2. **Function annotation allows**: `admission.datakit/ddtrace.enabled` is not `"false"` (if exists)
3. **Ignore version annotations**: Injection occurs even without version annotations

#### Use Case Examples {#use-case-examples}

1. **Strict version control scenario** (`check_annotation: true`):

   ```json
   {
       "namespace_selectors": ["prod"],
       "label_selectors": ["app=backend"],
       "check_annotation": true,
       "image": "internal-registry/dd-lib-java:{{.DDTraceJavaExtVersion}}",
       "language": "java"
   }
   ```

   **Injection conditions**:
   - Pod is in `prod` namespace with `app=backend` label
   - Pod **does not have** `admission.datakit/ddtrace.enabled: "false"` (if exists)
   - Pod **must have** `admission.datakit/java-lib.version` annotation

2. **Batch injection scenario** (`check_annotation: false`):

   ```json
   {
       "namespace_selectors": ["staging"],
       "label_selectors": ["env=test"],
       "check_annotation": false,
       "image": "internal-registry/dd-lib-java:latest",
       "language": "java"
   }
   ```

   **Injection conditions**:
   - Pod is in `staging` namespace with `env=test` label
   - Pod **does not have** `admission.datakit/ddtrace.enabled: "false"` (if exists)
   - **Ignore** `admission.datakit/java-lib.version` annotation check

3. **Selective rejection scenario**:

   ```json
   {
       "namespace_selectors": ["prod"],
       "label_selectors": ["app=java-app"],
       "check_annotation": false,
       "image": "internal-registry/dd-lib-java:latest",
       "language": "java"
   }
   ```

   **Injection logic**:
   - All matching Pods will be injected
   - If a Pod has `admission.datakit/ddtrace.enabled: "false"`, that Pod will be excluded
   - Version annotation `admission.datakit/java-lib.version: "1.15.0"` can be used to override image version, but won't affect injection decision

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

> **Annotation Usage Instructions**: For how `check_annotation` configuration affects version annotation behavior, please refer to [Annotation Configuration Injection](datakit-operator.md#annotation-injection) and [`check_annotation` Configuration Item Explanation](datakit-operator.md#check-annotation-config).

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
