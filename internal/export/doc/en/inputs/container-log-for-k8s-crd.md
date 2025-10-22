---
title: 'Configuring Container Log Collection via Kubernetes CRD'
summary: 'Configure container log collection based on Kubernetes CRD'
tags:
  - 'Logs'
  - 'Container'
  - 'KUBERNETES'
__int_icon:    'icon/kubernetes/'
---

DataKit provides a declarative approach for container log collection configuration through Kubernetes Custom Resource Definitions (CRDs). Users can automatically configure DataKit's log collection by creating `ClusterLoggingConfig` resources, eliminating the need to manually modify DataKit configuration files or restart DataKit.

## Prerequisites {#prerequisites}

- Kubernetes cluster version 1.16+
- DataKit [:octicons-tag-24: Version-1.84.0](../datakit/changelog-2025.md#cl-1.84.0) or later
- Cluster administrator permissions (for registering the CRD)

## Usage Workflow {#usage-workflow}

1. Register the Kubernetes CRD
1. Create CRD resources to automatically apply collection configurations
1. Configure CRD-related RBAC permissions for DataKit and start the DataKit service

### Register Kubernetes CRD {#register-kubernetes-crd}

Use the following YAML to register the `ClusterLoggingConfig` CRD:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: clusterloggingconfigs.logging.datakits.io
  labels:
    app: datakit-logging-config
    version: v1alpha1
spec:
  group: logging.datakits.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            apiVersion:
              type: string
            kind:
              type: string
            metadata:
              type: object
            spec:
              type: object
              required:
                - selector
              properties:
                selector:
                  type: object
                  properties:
                    namespaceRegex:
                      type: string
                    podRegex:
                      type: string
                    podLabelSelector:
                      type: string
                    containerRegex:
                      type: string
                podTargetLabels:
                  type: array
                  items:
                    type: string
                configs:
                  type: array
                  items:
                    type: object
                    required:
                      - source
                      - type
                    properties:
                      source:
                        type: string
                      type:
                        type: string
                      disable:
                        type: boolean
                      path:
                        type: string
                      multiline_match:
                        type: string
                      pipeline:
                        type: string
                      storage_index:
                        type: string
                      tags:
                        type: object
                        additionalProperties:
                          type: string
  scope: Cluster
  names:
    plural: clusterloggingconfigs
    singular: clusterloggingconfig
    kind: ClusterLoggingConfig
    shortNames:
      - logging
```

Apply the CRD:

```bash
kubectl apply -f clusterloggingconfig-crd.yaml
```

Verify the CRD registration:

```bash
kubectl get crd clusterloggingconfigs.logging.datakits.io
```

### Create CRD Configuration Resource {#create-crd-configuration-resource}

The following example configures the collection of log files from all Pods in the `test01` namespace whose names start with `logging`:

```yaml
apiVersion: logging.datakits.io/v1alpha1
kind: ClusterLoggingConfig
metadata:
  name: nginx-logs
spec:
  selector:
    namespaceRegex: "^(test01)$"
    podRegex: "^(logging.*)$"

  podTargetLabels:
    - app
    - version

  configs:
    - source: "nginx-access"
      type: "file"
      path: "/var/log/nginx/access.log"
      pipeline: "nginx-access.p"
      tags:
        log_type: "access"
        component: "nginx"
        
    - source: "nginx-error"
      type: "file"
      path: "/var/log/nginx/error.log"
      pipeline: "nginx-error.p"
      tags:
        log_type: "error"
        component: "nginx"
```

Apply the configuration:

```bash
kubectl apply -f logging-config.yaml
```

#### Configuration Details {#configuration-details}

- `selector` Configuration

  The selector is used to match target Pods and containers. All conditions have an **AND** relationship.

  | Field               | Type   | Required | Description                                           | Example                                 |
  | ------------------- | ------ | -------- | ----------------------------------------------------- | --------------------------------------- |
  | `namespaceRegex`    | string | No       | Regular expression to match namespace names           | `"^(default\                            | nginx)$"`         |
  | `podRegex`          | string | No       | Regular expression to match Pod names                 | `"^(nginx-log-demo.*)$"`                |
  | `podLabelSelector`  | string | No       | Pod label selector (comma-separated key=value pairs)  | `"app=nginx,environment=production"`    |
  | `containerRegex`    | string | No       | Regular expression to match container names           | `"^(nginx\                              | app-container)$"` |

  Selector example combination:

  ```yaml
  selector:
    namespaceRegex: "^(production|staging)$" # Match production or staging namespaces
    podLabelSelector: "app=web-server"       # Match Pods with the app=web-server label
    containerRegex: "^(app|web)$"            # Match containers named app or web
  ```

- `podTargetLabels` Pod Label Propagation

  | Field              | Type       | Required | Description                                           | Example                                |
  | ------------------ | ---------- | -------- | ----------------------------------------------------- | -------------------------------------- |
  | `podTargetLabels`  | []string   | No       | List of keys to copy from Pod Labels to log tags      | `["app", "version", "environment"]`    |

- `configs` Collection Configuration

  | Field              | Type              | Required               | Description                                                           | Example                                           |
  | ------------------ | ----------------- | ---------              | --------------------------------------------------------------------- | ------------------------------------------------- |
  | `type`             | string            | Yes                    | Collection type: `file` - file logs, `stdout` - standard output       | `"file"`                                          |
  | `source`           | string            | Yes                    | Log source identifier, used to distinguish different log streams      | `"nginx-access"`                                  |
  | `path`             | string            | Conditionally Required | Log file path (supports glob patterns), required when type=file       | `"/var/log/nginx/*.log"`                          |
  | `disable`          | boolean           | No                     | Whether to disable this collection configuration                      | `false`                                           |
  | `multiline_match`  | string            | No                     | Regular expression for the starting line of multi-line logs           | `"^\\d{4}-\\d{2}-\\d{2}"`                         |
  | `pipeline`         | string            | No                     | Name of the log parsing pipeline configuration file                   | `"nginx-access.p"`                                |
  | `storage_index`    | string            | No                     | Index name for log storage                                            | `"app-logs"`                                      |
  | `tags`             | map[string]string | No                     | Key-value pairs of tags attached to the logs                          | `{"log_type": "access", "component": "nginx"}`    |

### Add Relevant RBAC Configuration {#add-rbac-configuration}

Add the following permissions to DataKit's ClusterRole:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
  # Other existing permissions
  - apiGroups: ["logging.datakits.io"]
    resources: ["clusterloggingconfigs"]
    verbs: ["get", "list", "watch"]
```

Complete RBAC configuration example:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["clusterroles"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["nodes", "nodes/stats", "nodes/metrics", "namespaces", "pods", "pods/log", "events", "services", "endpoints", "persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets", "statefulsets", "replicasets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: [ "get", "list", "watch"]
- apiGroups: ["monitoring.coreos.com"]
  resources: ["podmonitors", "servicemonitors"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["logging.datakits.io"]
  resources: ["clusterloggingconfigs"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
```

## Example Application {#example-application}

The following is a complete CRD test application example:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: test01

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: logging-deployment
  namespace: test01
  labels:
    app: logging
    version: v1.0
    environment: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: logging
  template:
    metadata:
      labels:
        app: logging
        version: v1.0
        environment: test
    spec:
      containers:
      - name: demo
        image: ubuntu:22.04
        env:
        resources:
          limits:
            cpu: "200m"
            memory: "100Mi"
          requests:
            cpu: "100m"
            memory: "50Mi"
        command: ["/bin/bash", "-c", "--"]
        args:
        - |
          mkdir -p /tmp/opt/abc;
          i=1;
          while true; do
            echo "Writing logs to file ${i}.log";
            for ((j=1;j<=10000;j++)); do
              echo "$(date +'%F %H:%M:%S')  [$j]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/abc/file_${i}.log;
              sleep 1;
            done;
            echo "Finished writing 10000 lines to file_${i}.log";
            i=$((i+1));
          done
```

Apply the deployment:

```bash
kubectl apply -f test-application.yaml
kubectl apply -f logging-config.yaml
```

## FAQ {#faq}

- Does it support dynamic creation, modification, or deletion of CRDs?

Yes. DataKit dynamically adjusts log and field collection based on the status of CRDs. When a CRD is created or updated, the configuration is automatically applied to all matching containers. If a CRD is deleted, any ongoing log collection using that configuration will be terminated, though container stdout will continue to be collected using the default configuration.

- Which has higher priority: CRD configuration or Pod Annotations configuration?

Pod Annotations configuration has higher priority. If a container matches both a CRD configuration and its Pod contains a `datakit/logs` annotation configuration, the Pod Annotations configuration will take effect, and the CRD configuration will be ignored.

- How long does it take for CRD configuration changes to take effect?

The maximum window for configuration changes to take effect is 1 minute.

- What happens when multiple ClusterLoggingConfigs match the same Pod?

In theory, the ClusterLoggingConfig that was created first (with the smallest ResourceVersion) will be applied. It is best to avoid such situations.

- Do I need to add a mount to collect log files inside containers?

Starting from [:octicons-tag-24: Version-1.84.0](../datakit/changelog-2025.md#cl-1.84.0), for standard Docker mode or Containerd runtime (excluding CRI-O), log files inside containers can be collected without mounting. For the CRI-O runtime, Docker uses a tmpfs mount for the path, requiring an `emptyDir` mount to be added.
