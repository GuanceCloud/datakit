---
title     : 'Kubernetes CRD Extended Collection'
summary   : 'Create Datakit CRD to collect'
tags      :
  - 'PROMETHEUS'
  - 'KUBERNETES'
__int_icon: 'icon/kubernetes'
---

:material-kubernetes:

---

[:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

## Introduction {#intro}

**This feature is deprecated in Datakit 1.63.0.**

This document describes how to create a DataKit resource in a Kubernetes cluster and configure an extension collector.

### Add Authentication {#authorization}

If it is an upgraded version of DataKit, you need to add authentication in the `apiVersion: rbac.authorization.k8s.io/v1` entry of `datakit.yaml`, that is, copy the following lines and add them to the end:

```yaml
- apiGroups:
  - <<<custom_key.brand_main_domain>>>
  resources:
  - datakits
  verbs:
  - get
  - list
```

<!-- markdownlint-disable MD013 -->
### Create v1beta1 DataKit Instance, Create DataKit Object {#create}
<!-- markdownlint-enable -->

Write the following to the yaml configuration, such as `datakit-crd.yaml`, where each field has the following meaning:

- `k8sNamespace`: Specify namespace, locates a collection's Pod with deployment, required
- `k8sDaemonSet`: Specify the DaemonSet name to locate a collection's Pod with namespace
- `k8sDeployment`: Specify the deployment name, and locates the Pod of a collection with namespace
- `inputConf`: Collector configuration file, find the corresponding Pod according to namespace and deployment, replace the wildcard information of Pod, and then run the collector according to inputConf content. The following wildcard characters are supported.
    - `$IP`: Pod's intranet IP
    - `$NAMESPACE`: Pod Namespace
    - `$PODNAME`: Pod Name
    - `$NODENAME`: The name of the current node

Execute the `kubectl apply -f datakit-crd.yaml` command.

<!-- markdownlint-disable MD046 -->
???+ attention

    - DaemonSet and Deployment are two different Kubernetes resources, but here `k8s DaemonSet` and `k8s Deployment` can exist at the same time. That is, under the same Namespace, the Pod created by DaemonSet and the Pod created by Deployment share the same CRD configuration. This is not recommended, however, because fields like `source` are used to identify data sources in specific configurations, and mixing them leads to unclear data boundaries. It is recommended that only one `k8s DaemonSet` and `k8s Deployment` exist in the same CRD configuration.

    - Datakit only collects Pod in the same node as it, which belongs to nearby collection and will not be collected across nodes.
<!-- markdownlint-enable -->

## Example {#example}

A complete example is as follows, including:

- Create CRD Datakit
- Namespace and Datakit instance objects used for testing
- Configure the Prom collector (`inputConf`)

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: datakits.<<<custom_key.brand_main_domain>>>
spec:
  group: <<<custom_key.brand_main_domain>>>
  versions:
  - name: v1beta1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              instances:
                type: array
                items:
                  type: object
                  properties:
                    k8sNamespace:
                      type: string
                    k8sDaemonSet:
                      type: string
                    k8sDeployment:
                      type: string
                    datakit/logs:
                      type: string
                    inputConf:
                      type: string
  scope: Namespaced
  names:
    plural: datakits
    singular: datakit
    kind: Datakit
    shortNames:
    - dk
---
apiVersion: v1
kind: Namespace
metadata:
  name: datakit-crd
---
apiVersion: "<<<custom_key.brand_main_domain>>>/v1beta1"
kind: Datakit
metadata:
  name: my-test-crd-object
  namespace: datakit-crd
spec:
  instances:
    - k8sNamespace: "testing-namespace"
      k8sDaemonSet: "testing-daemonset"
      inputConf: |
        [inputs.prom]
          url="http://prom"
```

### Nginx Ingress Configuration Sample {#example-nginx}

Here, we use DataKit CRD extension to collect Ingress metrics, that is, we collect Ingress metrics through prom collector.

#### Requirements {#nginx-requirements}

- Deployed [DaemonSet DataKit](../datakit/datakit-daemonset-deploy.md)
- If the `Deployment` is called `ingress-nginx-controller`, the yaml configuration over there is as follows:

  ``` yaml
  ...
  spec:
    selector:
      matchLabels:
        app.kubernetes.io/component: controller
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: ingress-nginx-controller  # 这里只是一个示例名称
  ...
  ```

#### Configuration Step {#nginx-steps}

- Create Datakit CustomResourceDefinition

Execute the following create command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: datakits.<<<custom_key.brand_main_domain>>>
spec:
  group: <<<custom_key.brand_main_domain>>>
  versions:
    - name: v1beta1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                instances:
                  type: array
                  items:
                    type: object
                    properties:
                      k8sNamespace:
                        type: string
                      k8sDeployment:
                        type: string
                      datakit/logs:
                        type: string
                      inputConf:
                        type: string
  scope: Namespaced
  names:
    plural: datakits
    singular: datakit
    kind: Datakit
    shortNames:
    - dk
EOF
```

View deployment:

```bash
$ kubectl get crds | grep <<<custom_key.brand_main_domain>>>
datakits.<<<custom_key.brand_main_domain>>>   2022-08-18T10:44:09Z
```

- Create a Datakit resource

Prometheus configuration can be found in [link](kubernetes-prom.md)

Execute the following `yaml`:

```yaml
apiVersion: <<<custom_key.brand_main_domain>>>/v1beta1
kind: DataKit
metadata:
  name: prom-ingress
  namespace: datakit
spec:
  instances:
    - k8sNamespace: ingress-nginx
      k8sDeployment: ingress-nginx-controller
      inputConf: |-
        [[inputs.prom]]
          url = "http://$IP:10254/metrics"
          source = "prom-ingress"
          metric_types = ["counter", "gauge", "histogram"]
          measurement_name = "prom_ingress"
          interval = "60s"
          tags_ignore = ["build","le","method","release","repository"]
          metric_name_filter = ["nginx_process_cpu_seconds_total","nginx_process_resident_memory_bytes","request_size_sum","response_size_sum","requests","success","config_last_reload_successful"]
        [[inputs.prom.measurements]]
          prefix = "nginx_ingress_controller_"
          name = "prom_ingress"
        [inputs.prom.tags]
          namespace = "$NAMESPACE"
```

> !!! Note that `namespace` is customizable, while `k8sDeployment` and `k8sNamespace` must be accurate.

View deployment:

```bash
$ kubectl get dk -n datakit
NAME           AGE
prom-ingress   18m
```

- View Metric Collection

Log in to `Datakit pod` and execute the following command:

```bash
datakit monitor
```

<figure markdown>
  ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/datakit-crd-ingress.png){ width="800" }
  <figcaption> Ingress 数据采集 </figcaption>
</figure>

You can also log in to [<<<custom_key.brand_name>>> Platform](https://www.<<<custom_key.brand_main_domain>>>/){:target="_blank"}, "Indicator"-"Viewer" to view metric data
