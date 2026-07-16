# Injecting DDTrace via DataKit Operator

At **Pod creation time**, DataKit Operator mutates the Pod template through an admission webhook. It adds a `datakit-lib-init` initContainer, copies language libraries to the shared `/datadog-lib` volume, then injects that volume and required startup environment into application containers. It does not modify running Pods. After changing an Operator ConfigMap, image version, or Deployment annotation, create new Pods through a rollout/restart for the change to take effect.

Establish these facts before enabling injection:

1. Selectors decide which workloads are injected. Empty selectors can broaden impact, so pilot in a dedicated namespace first.
1. The library version in `image` must match the **language runtime in the application container**, not the initContainer runtime.
1. Library injection does not guarantee that an application starts or emits traces. Verify the application's startup settings, environment, and actual trace requests.

## Usage Instructions {#datakit-operator-inject-lib-usage}

1. [Download and install DataKit-Operator](datakit-operator.md#install) in the target Kubernetes cluster.
1. Add the following ConfigMap configuration to the Operator:

    ```json
    {
        "server_listen": "0.0.0.0:9543",
        "log_level": "info",
        "admission_inject_v2": {
            "ddtraces": [
                {
                    "namespace_selectors": ["staging"],
                    "label_selectors": ["app=example"],
                    "check_annotation": false,
                    "image": "<ddtrace-library-image>",
                    "language": "java",
                    "envs": {
                        "DD_AGENT_HOST": "datakit-service.datakit.svc.cluster.local",
                        "DD_TRACE_AGENT_PORT": "9529"
                    }
                }
            ]
        },
        "admission_inject": {
            "ddtrace": {}
        }
    }
    ```

    This is valid JSON. `admission_inject_v2` (Operator `v1.8.0+`) supports multiple DDTrace configurations and is preferred. `admission_inject` is a legacy compatibility configuration that normally expresses only one DDTrace rule. Do not paste JSON containing `//` comments directly into a ConfigMap.

    DDTrace injection has the following configurable fields:

    | Field                        | Type    | Description                                                     | Required     | Example Value                    |
    | ------:                      | :-----: | :------                                                         | :---:        | :--------                        |
    | `envs`                       | object  | Environment variable map                                        | Y[^envs]     | See example below                |
    | `image`                      | string  | DDTrace image address                                           | Y[^image]    | See example below                |
    | `label_selectors`            | array   | Label selector array                                            | Y[^selector] | `["app=nginx", "tier=frontend"]` |
    | `language`                   | string  | Supported language type (`java`/`python`/`php`/`nodejs`)         | Y[^lang]     | `"nodejs"`                       |
    | `namespace_selectors`        | array   | Namespace selector using regular expressions                    | Y[^selector] | `["^prod-.*$", "^test$"]`       |
    | `resources`                  | object  | Resource limit configuration                                    | N            | See example below                |
    | ~~`enabled_namespaces`~~     | object  | Select target Kubernetes namespace and set development language | Y            | Deprecated in 1.7.0 `admission_inject_v2` |
    | ~~`enabled_labelselectors`~~ | object  | Select target using Kubernetes label                            | Y            | Deprecated in 1.7.0 `admission_inject_v2` |

    [^selector]: The field itself must be present or Operator rejects the injection. An empty array broadens selection; verify that scope before production.
    [^image]: The installation templates provide default image addresses. For offline environments, users generally need to copy the images to an internal registry and then use the internal image addresses.
    [^lang]: The language selected here must match the content of the corresponding DDTrace image. If it does not match, the injection will fail.
    [^envs]: These environment variable settings are crucial and directly affect the final data outcome. For the supported `fieldRef` list here, see [here](datakit-operator.md#downwardapi).

    An allowed `language` value does not mean that every version has an available image. This page includes verified version mappings for Node.js and Python. For Java, use the example below and verify that the final JVM command loaded the Agent. For PHP, use only an image explicitly matched to the PHP execution mode in installation templates or release notes. Never use one language's image to inject another runtime.

### Pilot a Narrow Scope First {#pilot}

Start with a test namespace and an explicit `app=<name>` label selector, then expand gradually. Pin an exact image tag instead of using `latest` in a production rule. After each change, at minimum verify:

```shell
kubectl get pod <pod-name> -o jsonpath='{.spec.initContainers[*].name}'
kubectl describe pod <pod-name>
```

The first command should include `datakit-lib-init`; the second helps inspect mounts, environment variables, and webhook events. Then check language startup settings in the application container and trace requests in the DataKit monitor.

### DDTrace Lib Injection and Image Selection {#ddtrace-lib-image-selection}

When DataKit Operator injects DDTrace, it adds an initContainer named `datakit-lib-init`, copies DDTrace libraries into the shared `/datadog-lib` volume, and mounts that directory into application containers. Select the image version according to the **language runtime version in the application container**, not according to the initContainer runtime.

Image repositories are brand-specific. The examples below use `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator`, which is replaced with the corresponding repository when the document is published for each brand.

#### Node.js Injection {#ddtrace-nodejs-injection}

Node.js injection sets or appends `NODE_OPTIONS` in the application container:

```shell
--require=/datadog-lib/node_modules/dd-trace/init
```

If `NODE_OPTIONS` already exists in the application container, Operator appends the option above. The Node.js image must match the major Node.js version in the application container.

If an application image or startup script overwrites `NODE_OPTIONS`, the injected option is lost and no traces are produced. After deployment, use `kubectl exec` to check that `NODE_OPTIONS` still contains `--require=/datadog-lib/node_modules/dd-trace/init`.

| Application Node.js version | Recommended image | Version requirement |
| --- | --- | --- |
| Node.js 18 to 25 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-js-init:5.102.0` | Default recommended version. It includes `dd-trace@5.102.0`, which requires `node >=18 <26` |
| Node.js 16 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-js-init:4.55.0` | Use the `dd-trace` 4.x series. Do not use `5.102.0` for Node.js 16 |
| Node.js 14 and earlier | Not in the default support scope | Requires an older `dd-trace` major version and the corresponding image. Upgrading Node.js is recommended |

Node.js DDTrace configuration example:

```json
{
    "namespace_selectors": ["default"],
    "label_selectors": [],
    "check_annotation": false,
    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-js-init:5.102.0",
    "language": "nodejs",
    "envs": {
        "DD_AGENT_HOST": "datakit-service.datakit.svc.cluster.local",
        "DD_TRACE_AGENT_PORT": "9529",
        "DD_SERVICE": "{fieldRef:metadata.labels['app']}",
        "POD_NAME": "{fieldRef:metadata.name}",
        "POD_NAMESPACE": "{fieldRef:metadata.namespace}",
        "NODE_NAME": "{fieldRef:spec.nodeName}",
        "DD_TAGS": "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
    }
}
```

#### Python Injection {#ddtrace-python-injection}

Python injection sets or appends `PYTHONPATH=/datadog-lib/` in the application container, so the Python process can load DDTrace libraries and the injection bootstrap from `/datadog-lib`. Python DDTrace packages include CPython ABI-specific wheels, so the image version must match the Python minor version in the application container. If the version does not match, the application may fail with `ModuleNotFoundError`, native extension loading errors, or startup errors.

If an application image or startup script overwrites `PYTHONPATH`, it must preserve `/datadog-lib/` or the injected library cannot load. After startup, `python -c 'import ddtrace; print(ddtrace.__version__)'` provides a minimal load check.

| Application Python version | Recommended image | Version requirement |
| --- | --- | --- |
| Python 3.7 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-python-init:v2.21.12` | `ddtrace` 2.x supports Python 3.7. Python 3.7 is not supported by `ddtrace` 3.x/4.x |
| Python 3.8 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-python-init:v3.19.7` | `ddtrace` 3.x supports Python 3.8. `ddtrace` 4.x requires Python 3.9 or later |
| Python 3.9 to 3.14 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-python-init:v4.8.3` | Current `ddtrace` 4.x requires `python >=3.9,<3.15` |

Python DDTrace configuration example:

```json
{
    "namespace_selectors": ["default"],
    "label_selectors": [],
    "check_annotation": false,
    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-python-init:v4.8.3",
    "language": "python",
    "envs": {
        "DD_AGENT_HOST": "datakit-service.datakit.svc.cluster.local",
        "DD_TRACE_AGENT_PORT": "9529",
        "DD_SERVICE": "{fieldRef:metadata.labels['app']}",
        "POD_NAME": "{fieldRef:metadata.name}",
        "POD_NAMESPACE": "{fieldRef:metadata.namespace}",
        "NODE_NAME": "{fieldRef:spec.nodeName}",
        "DD_TAGS": "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
    }
}
```

> Version selection is based on upstream DDTrace package runtime requirements: Node.js uses the npm `dd-trace` `engines.node` field, and Python uses the PyPI `ddtrace` `Requires-Python` field and wheel support.

#### Verify Java Injection {#ddtrace-java-injection}

Java injection needs more than a library volume: the JVM must actually load `-javaagent`. After startup, inspect the final Pod's `JAVA_TOOL_OPTIONS`, container command/args, or process command line to confirm that it contains the Operator-injected Agent path; also confirm `DD_AGENT_HOST` and `DD_TRACE_AGENT_PORT=9529`. A successful `datakit-lib-init` alone does not prove that the JVM loaded an Agent.

<!-- markdownlint-disable MD013 -->
### `check_annotation` Configuration Item Explanation {#check-annotation-config}
<!-- markdownlint-enable MD013 -->

`check_annotation` is an important configuration field used to control how DataKit Operator handles **version annotations** on Pods (for example, `admission.datakit/java-lib.version`, `admission.datakit/python-lib.version`, and `admission.datakit/nodejs-lib.version`). The values and behaviors of this field are as follows:

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
   - `admission.datakit/<language>-lib.version` **is affected** by `check_annotation` configuration
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
| `admission.datakit/python-lib.version` | Specifies DDTrace Python Lib version | Version string | **Yes** | e.g., `"v3.19.7"`, used to override default image version in configuration |
| `admission.datakit/nodejs-lib.version` | Specifies DDTrace Node.js Lib version | Version string | **Yes** | e.g., `"5.102.0"`, used to override default image version in configuration |
| `admission.datakit/enabled` | Controls all injection functions (highest priority) | `"true"`/`"false"` | **No** | `"false"`: completely rejects any injection, highest priority |

#### When `check_annotation: true` {#when-check-annotation-true}

The following conditions must be simultaneously met to perform injection:

1. **Configuration matches**: Pod must match `namespace_selectors` and `label_selectors` rules
2. **Function annotation allows**: `admission.datakit/ddtrace.enabled` is not `"false"` (if exists)
3. **Version annotation exists**: Version annotations must exist on Pod (for example, `admission.datakit/java-lib.version`, `admission.datakit/python-lib.version`, or `admission.datakit/nodejs-lib.version`)

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
       "image": "internal-registry/dd-lib-java:<pinned-version>",
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
       "image": "internal-registry/dd-lib-java:<pinned-version>",
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

The Operator can recognize the following Annotations:

- `admission.datakit/ddtrace.enabled`: Marks whether to enable injection in a single Deployment standard. Fill in `"true"` to enable injection, `"false"` to block injection. If blocked, the Operator will ignore injecting this Deployment.
- `admission.datakit/java-lib.version`: Specifies a specific DDTrace Java Agent version.
- `admission.datakit/python-lib.version`: Specifies a specific DDTrace Python Lib version.
- `admission.datakit/nodejs-lib.version`: Specifies a specific DDTrace Node.js Lib version.

> **Annotation Usage Instructions**: For how `check_annotation` affects version annotations, see [Annotation Configuration Injection](datakit-operator.md#annotation-injection) and [the `check_annotation` explanation on this page](operator-ddtrace.md#check-annotation-config).

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
