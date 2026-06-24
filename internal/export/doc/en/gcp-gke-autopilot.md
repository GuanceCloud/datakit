# GCP GKE Autopilot Integration
---

GKE Autopilot does not allow DataKit to use the traditional DaemonSet model with host `hostPath` mounts, container runtime sockets, host log directories, or privileged containers. In GKE Autopilot, DataKit uses the separately published `datakit-gke-autopilot` Helm chart, runs as a single-replica Deployment by default, and collects container data through GCP Cloud APIs.

The regular `datakit` chart targets standard Kubernetes environments and installs a DaemonSet by default. The `datakit-gke-autopilot` chart targets GKE Autopilot and installs a Deployment by default. Both charts use the same templates, but `datakit-gke-autopilot` has dedicated default values that disable host access and enable Cloud API collection.

## Capabilities {#capabilities}

<!-- markdownlint-disable MD046 -->
???+ note "Version Requirement"

    Collecting container metrics, objects, and logs through Cloud APIs requires DataKit 2.3.0 or later. Earlier versions do not support GKE Autopilot mode.
<!-- markdownlint-enable MD046 -->

`datakit-gke-autopilot` enables `dk,container` and election by default. A single leader DataKit instance collects cluster-level data, so multiple replicas do not call the same Cloud APIs repeatedly.

- Container metrics: collected from Cloud Monitoring and written to `docker_containers`.
- Container stdout/stderr logs: collected from Cloud Logging.
- Kubernetes resource metrics and objects: collected from the Kubernetes API, including Pod, Deployment, Service, and Node resources.
- GCP authentication: uses Workload Identity and does not require a Service Account key file.

This mode does not support host metrics, container runtime sockets, local container file logs, or eBPF. Cloud Monitoring metrics are typically delayed by several minutes. Cloud Logging uses an overlapping time window and `timestamp + insertId` deduplication, so regular polling is deduplicated. After Pod replacement or leader changes, logs in the overlap window may be read again, so delivery is at least once.

## Prerequisites {#prerequisites}

- A GKE Autopilot cluster that your local `kubectl` can access.
- A DataWay URL, for example `https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>`.
- GCP Cloud Monitoring API and Cloud Logging API are enabled.
- A GCP Service Account with `roles/monitoring.viewer` and `roles/logging.viewer`.
- Workload Identity binding for the Kubernetes ServiceAccount.

The following example uses `my-project`, the `datakit` namespace, the `my-datakit` release name, and the `datakit-cloud-monitor` GCP Service Account. Replace them with your actual values.

```shell
PROJECT_ID="my-project"
NAMESPACE="datakit"
RELEASE_NAME="my-datakit"
GSA_NAME="datakit-cloud-monitor"
KSA_NAME="datakit"
```

Create the GCP Service Account and grant permissions:

```shell
gcloud iam service-accounts create "${GSA_NAME}" \
  --project "${PROJECT_ID}"

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member "serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/monitoring.viewer

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member "serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/logging.viewer
```

Allow the Kubernetes ServiceAccount created by the Helm chart to use the GCP Service Account:

```shell
gcloud iam service-accounts add-iam-policy-binding \
  "${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --project "${PROJECT_ID}" \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/${KSA_NAME}]"
```

<!-- markdownlint-disable MD046 -->
???+ note

    The `datakit-gke-autopilot` chart sets `fullnameOverride=datakit` by default, so the example command creates a Kubernetes ServiceAccount named `datakit`. If `fullnameOverride` is overridden, adjust the Workload Identity binding to match the actual ServiceAccount name.
<!-- markdownlint-enable MD046 -->

## Helm Installation {#helm-install}

Add the Helm repository:

```shell
helm repo add datakit-gke https://pubrepo.truewatch.com/chartrepo/truewatch
```

Install DataKit:

```shell
helm install my-datakit datakit-gke/datakit-gke-autopilot \
         --namespace datakit \
         --create-namespace \
         --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>" \
         --set serviceAccountAnnotations."iam\\.gke\\.io/gcp-service-account"="datakit-cloud-monitor@my-project.iam.gserviceaccount.com"
```

Check the status:

```shell
helm -n datakit list
kubectl -n datakit get pod -l app.kubernetes.io/instance=my-datakit
kubectl -n datakit logs deploy/datakit
```

Back up the current values before upgrading:

```shell
helm -n datakit get values my-datakit -o yaml > values-current.yaml
```

Upgrade:

```shell
helm upgrade my-datakit datakit-gke/datakit-gke-autopilot \
         --namespace datakit \
         -f values-current.yaml
```

<!-- markdownlint-disable MD046 -->
???+ note "Upgrading from the old GKE Autopilot chart"

    The `datakit-gke-autopilot` chart published from the old `helm-gke-autopilot` branch created a DaemonSet and ServiceAccount named like `my-datakit-datakit-gke-autopilot`. The current chart creates a Deployment and ServiceAccount named `datakit` by default. Before upgrading, update the Workload Identity binding to `datakit/datakit` and make sure the target namespace has no other `datakit` resources with the same names.
<!-- markdownlint-enable MD046 -->

## YAML Installation {#yaml-install}

If Helm is not used, download the GKE Autopilot dedicated Deployment YAML:

```shell
curl -o datakit-gke-autopilot.yaml \
  https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-gke-autopilot.yaml
```

Before applying it, update `ENV_DATAWAY` and add the Workload Identity GCP Service Account annotation to the ServiceAccount:

```yaml
metadata:
  annotations:
    iam.gke.io/gcp-service-account: "datakit-cloud-monitor@my-project.iam.gserviceaccount.com"
```

Apply the YAML:

```shell
kubectl apply -f datakit-gke-autopilot.yaml
```

## Key Configuration {#configuration}

Default settings:

- `workload.kind=Deployment`
- `workload.replicas=1`
- `gkeAutopilot.enabled=true`
- `datakit.default_enabled_inputs=dk,container`
- `datakit.enabled_election=true`
- `ENV_INPUT_CONTAINER_GCP_CLOUD_API_ENABLED=true`
- `ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL=false`
- CPU/memory requests are `500m/500Mi`
- Only `/usr/local/datakit/cache` is mounted by default

The default replica count is `1`. For high availability, you can increase `workload.replicas`, but keep election enabled. Otherwise, multiple replicas may collect duplicate Cloud API data.

The GCP project ID, GKE cluster name, and cluster location are discovered from the GKE metadata server by default. For cross-project collection or when metadata discovery does not match the target cluster, override them as needed:

```yaml
extraEnvs:
  - name: ENV_INPUT_CONTAINER_GCP_PROJECT_ID
    value: "my-project"
  - name: ENV_INPUT_CONTAINER_GCP_CLUSTER_NAME
    value: "my-cluster"
  - name: ENV_INPUT_CONTAINER_GCP_CLUSTER_LOCATION
    value: "asia-southeast1"
```

`ENV_INPUT_CONTAINER_ENABLE_GCP_CLOUD_MONITORING` and `ENV_INPUT_CONTAINER_ENABLE_GCP_CLOUD_LOGGING` both default to `true`; set them only when one service must be disabled.

## Container Host Tag {#container-host-tag}

In the traditional DaemonSet mode, each DataKit instance collects only containers on its own node, so the `host` tag on container metrics, objects, and logs can be appended uniformly during Feed.

In GKE Autopilot Cloud API mode, a single leader DataKit collects cluster-wide container data through Cloud APIs, so the node running the DataKit Pod cannot be used as the `host` for all containers. This mode extracts the original GKE Node name of each container from Kubernetes Pod information and writes it to the `host` tag on container metrics, objects, and logs.

This mode does not support `ENV_K8S_CLUSTER_NODE_NAME` for renaming the `host` tag on container data. To distinguish multiple clusters, use `cluster_name_k8s`, `gcp_project_id`, `gcp_location`, or custom global tags.

## Running as Non-Root {#run-as-non-root}

`datakit-gke-autopilot` keeps the container running as root by default to match the DataKit image default. The current Cloud API mode does not require host permissions. To run as non-root:

```shell
helm upgrade my-datakit datakit-gke/datakit-gke-autopilot \
         --namespace datakit \
         --reuse-values \
         --set gkeAutopilot.runAsNonRoot=true
```

The Pod then uses:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 10001
  runAsGroup: 10001
  fsGroup: 10001
```

This mode does not need an initContainer by default. The chart only mounts `/usr/local/datakit/cache` for WAL, Cloud Logging state, and local cache, and `fsGroup=10001` keeps it writable. If extra directories such as `conf.d`, `data`, `pipeline`, or `python.d` are mounted, make sure they are writable by UID/GID `10001`. Add an initContainer only when an extra volume does not honor `fsGroup` or when pre-created file permissions must be fixed.

## Differences From Regular Helm Installation {#difference}

| chart | Default workload | Target environment | Host access | Container metrics and logs source |
| --- | --- | --- | --- | --- |
| `datakit` | DaemonSet | Standard Kubernetes | Uses `hostPath`, runtime sockets, and host log directories by default | Container runtime and host files |
| `datakit-gke-autopilot` | Deployment | GKE Autopilot | Does not use `hostPath`, privileged containers, or host networking by default | Cloud Monitoring and Cloud Logging |

If the target environment allows mounting host directories and the container runtime socket, use the regular `datakit` chart. For GKE Autopilot or similar Serverless Kubernetes environments that do not allow those host capabilities, use `datakit-gke-autopilot`.

## Troubleshooting {#troubleshooting}

- Pod rejected by GKE Autopilot: check whether extra `hostPath`, privileged container, `hostNetwork`, `hostPID`, or `hostIPC` settings were enabled.
- No container metrics or logs: check that the GCP Service Account has `roles/monitoring.viewer` and `roles/logging.viewer`, the Kubernetes ServiceAccount has the `iam.gke.io/gcp-service-account` annotation, and the Workload Identity binding matches the actual namespace/name.
- No Kubernetes objects or resource metrics: check that the DataKit Pod can access the Kubernetes API and that `dk,container` is still included in `datakit.default_enabled_inputs`.
- Duplicate data with multiple replicas: check that `datakit.enabled_election=true` and all replicas use the election configuration in the same namespace.
- Non-root write failures: check that extra mounted directories are writable by UID/GID `10001`; add volume permission handling only when needed.
