# DataKit Operator

---

:material-kubernetes:

---

DataKit Operator is a project integrated with DataKit in Kubernetes orchestration, aimed at facilitating easier deployment of DataKit, as well as providing other functions such as verification and injection.

## Overview {#overview}

DataKit Operator provides automated injection capabilities for Kubernetes clusters through the Kubernetes Admission Controller mechanism, helping users integrate observability more easily. Key features include:

- **DDTrace Injection**: Automatically inject APM tracing agents for Java applications
- **Log Collection**: Automatically collect container logs via logfwd sidecar
- **Performance Profiling**: Inject Flameshot or Profiler components for application performance monitoring
- **Configuration Management**: Support both global configuration and declarative configuration injection methods

**Core Advantages**:

- **Automated Deployment**: No need to manually modify application YAML, reducing configuration errors
- **Batch Management**: Implement batch injection through namespace and label selectors
- **Flexible Configuration**: Support JSON configuration and fine-grained control via Annotations
- **Version Compatibility**: Maintain backward compatibility and support smooth upgrades

## Prerequisites {#prerequisites}

- Kubernetes v1.24.1 or higher is recommended, with internet access (to download yaml files and pull corresponding images)
- Ensure `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` [controllers](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites){:target="_blank"} are enabled
- Ensure `admissionregistration.k8s.io/v1` API is enabled

## Installation {#install}

<!-- markdownlint-disable MD046 -->
=== "Deployment"

    Download [*datakit-operator.yaml*](https://static.<<<custom_key.brand_main_domain>>>/datakit-operator/datakit-operator.yaml){:target="_blank"}, follow these steps:
    
    ``` shell
    $ kubectl create namespace datakit
    $ wget https://static.<<<custom_key.brand_main_domain>>>/datakit-operator/datakit-operator.yaml
    $ kubectl apply -f datakit-operator.yaml
    $ kubectl get pod -n datakit
    
    NAME                               READY   STATUS    RESTARTS   AGE
    datakit-operator-f948897fb-5w5nm   1/1     Running   0          15s
    ```

=== "Helm"

    Prerequisites

    * Kubernetes >= 1.14
    * Helm >= 3.0+

    ```shell
    $ helm install datakit-operator datakit-operator \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
        <<<% else -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit --create-namespace
    ```

    Check deployment status:

    ```shell
    $ helm -n datakit list
    ```

    Upgrade using the following command:

    ```shell
    $ helm -n datakit get values datakit-operator -a -o yaml > values.yaml
    $ helm upgrade datakit-operator datakit-operator \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
        <<<% else -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit \
        -f values.yaml
    ```

    Uninstall using the following command:

    ```shell
    $ helm uninstall datakit-operator -n datakit
    ```

???+ attention

    - DataKit Operator has a strict correspondence between the program and yaml. If an outdated yaml is used, the new version of DataKit-Operator may not be installed. Please download the latest yaml.
    - If `InvalidImageName` error occurs, you can manually pull the image.
<!-- markdownlint-enable -->

### Configuration Explanation {#jsonconfig}

DataKit Operator configuration is in JSON format, stored separately as a ConfigMap in Kubernetes, and loaded into the container as environment variables.

<!-- markdownlint-disable MD046 -->
=== "DataKit Operator >= v1.7.0"

    Starting from DataKit-Operator v1.7.0, it is recommended to use the `admission_inject_v2` configuration item. The new configuration uses an array structure, supporting more flexible configuration methods.

    ```json
    {
        "server_listen": "0.0.0.0:9543", // Operator service listening address
        "log_level": "info",             // Operator log level
        "admission_inject_v2": {         // Injection configuration v2
            "ddtraces": [...],           // DDTrace configuration array
            "logfwds": [...],            // Log forwarding configuration array
            "flameshots": [...]          // Profiling configuration array
        },
        "admission_mutate": {            // Configuration mutation
            "loggings": [...]            // Log configuration mutation
        }
    }
    ```

=== "DataKit Operator < v1.7.0"

    ```json
    {
        "server_listen": "0.0.0.0:9543",
        "log_level":     "info",
        "admission_inject": {
            "ddtrace": {...},
            "profiler": {...},
            "logfwd": {...}
        },
        "admission_mutate": {
            "loggings": [...]
        }
    }
    ```
<!-- markdownlint-enable -->

## Injection Methods {#datakit-operator-inject}

DataKit Operator supports two resource input methods:

1. Selector Configuration Injection (Imperative)

    Specify the Namespace and Selector of the target Pod by modifying the DataKit-Operator config. If a Pod meets the conditions, injection is performed.

    **Advantages**: No need to add Annotations to the target Pod (but the target Pod needs to be restarted)

    **Disadvantages**: Scope is not precise enough, potentially leading to invalid injections

2. Annotation Configuration Injection (Declarative)

    Add Annotations to the target Pod to enable its own injection.

    **Advantages**: Injection rejection can be precisely controlled via Annotation

    **Disadvantages**: Injection cannot be triggered solely by Annotation; matching rules still need to be configured, meaning besides enabling injection in the target Pod annotation, other fields in the Operator configuration are also required.

### Selector Configuration Injection {#selectors-injection}

Batch injection can be achieved by configuring `namespace_selectors` and `label_selectors`.

In the `admission_inject_v2` configuration, `namespace_selectors` and `label_selectors` are configured directly in the array item. Taking DDTrace injection as an example:

```json
{
    "admission_inject_v2": {
        "ddtraces": [
            {
                "namespace_selectors": ["testns"],
                "label_selectors":     ["app=log-output"],
                ...
            }
        ]
    }
}
```

- `namespace_selectors`: Namespace selector array, supports regular expression matching. For exact matching, surround the pattern with `^` and `$`, e.g., `^testns$`
- `label_selectors`: Label selector array, uses Kubernetes Label Selector syntax

If both selectors are configured, the target Pod must satisfy both conditions simultaneously. For guidelines on writing label selectors, refer to this [official documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors){:target="_blank"}.

### Annotation Configuration Injection {#annotation-injection}

Adding specific Annotations to Deployment can control whether injection is allowed. Note that Annotations should be added in the template.

Supported Annotations are as follows:

| Annotation | Description | Values | Priority |
|------------|-------------|--------|----------|
| `admission.datakit/ddtrace.enabled` | Controls ddtrace injection | `"true"/"false"` | Medium |
| `admission.datakit/logfwd.enabled` | Controls logfwd injection | `"true"/"false"` | Medium |
| `admission.datakit/flameshot.enabled` | Controls flameshot injection | `"true"/"false"` | Medium |
| `admission.datakit/enabled` | Controls all injection functions | `"true"/"false"` | **Highest** |

Example:

```yaml
    annotations:
    admission.datakit/ddtrace.enabled: "true"
    admission.datakit/logfwd.enabled: "true"
```

<!-- markdownlint-disable MD046 -->
???+ tip

    Annotations can be used to reject injection (set to `"false"`), but for active injection, the following configuration is required:

    1. Set matching rules (`namespace_selectors`/`label_selectors`) and corresponding configuration fields in DataKit-Operator configuration
    2. Pod matches configured selectors
<!-- markdownlint-enable -->

Injection Method Summary

- **Global Configuration**: Suitable for batch scenarios, controlling injection scope via Operator configuration
- **Annotation Configuration**: Suitable for fine-grained control, deciding whether to inject via Pod annotations
- **Priority**: Annotation configuration takes precedence over global configuration, useful for rejecting injection
- **Compatibility**: Supported features vary slightly across versions, please refer to specific version notes

## Supported Injection Function List {#supported-operator}

| Function | Brief Description |
|--- |--- |
| DDtrace Agent | Inject DDTrace component, see [DDTrace](operator-ddtrace.md) |
| logfwd | Inject logfwd component to collect logs inside containers, see [logfwd](operator-logfwd.md) |
| Flameshot | Inject Flameshot component for dynamic application Profiling, see [Flameshot](operator-flameshot.md) |
| async-profiler | Inject async-profiler for periodic Profiling of Java applications, see [async-profile](operator-asyncprofile.md) |
| py-spy | Inject py-spy for Profiling of Python applications, see [py-spy](operator-pyspy.md) |
| logging | Inject log collection configuration, see [Logging](operator-logging.md) |

## Downward API {#downwardapi}

In DataKit Operator [:octicons-tag-24: v1.4.2](operator-changelog.md#cl-1.4.2) and later versions, `envs` supports Kubernetes Downward API [environment variable value fields](https://kubernetes.io/docs/concepts/workloads/pods/downward-api/#downwardapi-fieldRef). The following are currently supported:

| Field | Description | Example |
| ------: | :------ | :------ |
| `metadata.name` | The name of the Pod | nginx-123 |
| `metadata.namespace` | The namespace of the Pod | middleware |
| `metadata.uid` | The unique ID of the Pod | 12345678-1234-1234-1234-123456789abc |
| `metadata.annotations['<KEY>']` | The value of the Pod's annotation `<KEY>` | metadata.annotations['myannotation'] |
| `metadata.labels['<KEY>']` | The value of the Pod's label `<KEY>` | metadata.labels['app'] |
| `spec.serviceAccountName` | The name of the Pod's service account | default |
| `spec.nodeName` | The name of the node where the Pod is running | node-01 |
| `status.hostIP` | The primary IP address of the node where the Pod is located | 192.168.1.1 |
| `status.hostIPs` | Dual-stack version of status.hostIP | ["192.168.1.1", "2001:db8::1"] |
| `status.podIP` | The primary IP address of the Pod | 10.0.0.1 |
| `status.podIPs` | Dual-stack version of status.podIP | ["10.0.0.1", "2001:db8::2"] |

For example, if there is a Pod named `nginx-123` in the `middleware` namespace, and you want to inject the environment variables `POD_NAME` and `POD_NAMESPACE`, refer to the following:

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

Ultimately, in that Pod you can see:

``` shell
$ env | grep POD
POD_NAME=nginx-123
POD_NAMESPACE=middleware
```

<!-- markdownlint-disable MD046 -->
???+ note

    If the Value placeholder is unrecognizable, it will be added to the environment variable as a plain string. For example, `"POD_NAME": "{fieldRef:metadata.PODNAME}"` is an incorrect syntax; the environment variable will be `POD_NAME={fieldRef:metadata.PODNAME}`.
<!-- markdownlint-enable -->

## FAQ {#faq}

### How to disable injection for a specific Pod? {#disable-inject}

Add Annotation `"admission.datakit/enabled": "false"` to that Pod, and no operations will be performed for it. This has the highest priority.

### How does it work? {#principles}

DataKit-Operator uses Kubernetes Admission Controller function for resource injection. For detailed mechanisms, please check the [official documentation](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/){:target="_blank"}

### What to note in AWS EKS environment? {#aws-eks}

Deploying in an AWS EKS environment may cause DataKit-Operator not to take effect; you need to open port `9543` in the security group.

### Troubleshooting Guide {#debug}

| Issue | Possible Cause | Solution |
|------|----------|----------|
| Injection not taking effect | Webhook not configured correctly | Check `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` |
| Image pull failed | Image address or permission issue | Verify image address, check image repository access permissions |
| Port unreachable | Network or security group configuration | Open port `9543`, check network policies |
