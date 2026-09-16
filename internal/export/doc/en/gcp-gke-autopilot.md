# GCP GKE Autopilot Integration
---

Choose one collection mode, then install it using either Helm or YAML to avoid duplicate collection.

## Choose a mode {#choose-mode}

| Comparison | Partner | Cloud API |
| --- | --- | --- |
| Workload | DaemonSet, with DataKit on each node | Deployment, with one replica by default |
| Helm chart | `datakit` with its bundled Partner values | `datakit-gke-autopilot` |
| Container metrics | Node containerd | Cloud Monitoring |
| Container logs | Kubernetes Pod stdout/stderr logs on each node | Cloud Logging |
| Prerequisites | Synchronize the formal Google V2 allowlist | GCP APIs, IAM permissions, and Workload Identity |
| YAML deployment | Render a complete manifest locally from the chart | Download the dedicated Deployment YAML |

Both modes collect Kubernetes resources through the Kubernetes API and enable `dk,container` and election by default. Cloud API mode requires DataKit 2.3.0 or later.

## Shared preparation {#prerequisites}

- A GKE Autopilot cluster, with the local `kubectl` context pointing to it and permissions to deploy workloads and RBAC resources.
- DataKit can reach your DataWay URL containing a valid workspace token, and nodes can pull the required images.
- Examples use `datakit` for the namespace and resource names. Check that no other installation owns the same workloads, Services, ClusterRole, or ClusterRoleBinding.
- Helm installations and Partner YAML rendering require Helm 3 locally. Cloud API IAM configuration requires `gcloud`.

In a new working directory, prepare these Helm user values and replace the DataWay URL and cluster name. Protect this file because it contains your token. **For Cloud API YAML installation, skip the Helm setup below and continue with [Cloud API deployment](gcp-gke-autopilot.md#cloud-api).**

```shell
umask 077
cat > datakit-user-values.yaml <<'VALUES'
datakit:
  dataway_url: "https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>"
  cluster_name_k8s: "my-autopilot-cluster"
VALUES
```

Add the repository, list available versions, and replace `CHART_VERSION` with your chosen chart version. Partner requires the two bundled files checked below.

```shell
helm repo add truewatch https://pubrepo.truewatch.com/chartrepo/truewatch
helm repo update truewatch
helm search repo truewatch/datakit --versions
CHART_VERSION="<released chart version>"
```

## Partner deployment {#partner}

**Partner V2 supports TrueWatch only and requires GKE `1.35.6-gke.1258000` or later.** Use the TrueWatch `datakit` chart and `pubrepo.truewatch.com/truewatch/datakit` image. Confirm that both have been published before deployment.

Leave `image.tag` empty to use the chart's `appVersion`. After downloading the chart and synchronizing the allowlist, choose either Helm or YAML installation.

### Download Partner files {#partner-download}

Download the regular `datakit` chart and check its Partner files:

```shell
helm pull truewatch/datakit --version "$CHART_VERSION" --untar
test -f datakit/values-gke-autopilot-partner.yaml
test -f datakit/gke-autopilot-partner-allowlist.yaml
```

### Sync the V2 allowlist {#allowlist}

Your account needs permission to manage AllowlistSynchronizers. The cluster must permit `gke://TrueWatch/datakit/truewatch-datakit-autopilot-v2.yaml`; GKE allows `gke://*` by default. If an administrator restricted the sources, add this path while preserving existing paths. See [Google's allowlist installation guide](https://docs.cloud.google.com/kubernetes-engine/docs/how-to/run-autopilot-partner-workloads){:target="_blank"}.

Install the synchronizer once per cluster. The commands below create a new synchronizer. If an existing synchronizer manages this path, skip `apply` and replace `truewatch-datakit` in the inspection commands with its name. **Complete synchronization before either Partner Helm or YAML installation.**

```shell
kubectl apply -f datakit/gke-autopilot-partner-allowlist.yaml
kubectl wait --for=condition=Ready allowlistsynchronizer/truewatch-datakit --timeout=10m
kubectl get allowlistsynchronizer truewatch-datakit -o yaml
kubectl get workloadallowlist truewatch-datakit-autopilot-v2
```

Before installing DataKit, confirm `Ready=True`, the V2 file is `Installed` under `status.managedAllowlistStatus`, and the WorkloadAllowlist exists.

### Partner Helm installation {#partner-helm}

```shell
helm upgrade --install datakit ./datakit \
  --namespace datakit --create-namespace \
  -f datakit/values-gke-autopilot-partner.yaml \
  -f datakit-user-values.yaml --wait --timeout 15m
```

The chart creates the `datakit-dataway` Secret and passes the DataWay URL to DataKit through `secretKeyRef`.

### Partner YAML installation {#partner-yaml}

Choose this or Helm installation. This method requires Helm 3 locally to render YAML, then manages resources with `kubectl` without creating a Helm release. `--no-hooks` excludes the chart's Helm test Pod.

```shell
umask 077
helm template datakit ./datakit --namespace datakit --no-hooks \
  -f datakit/values-gke-autopilot-partner.yaml \
  -f datakit-user-values.yaml > datakit-partner.yaml
kubectl create namespace datakit --dry-run=client -o yaml | kubectl apply -f -
kubectl apply --namespace datakit -f datakit-partner.yaml
```

Protect the rendered manifest: it contains the complete workload and its Secret. The static `datakit.yaml` is for regular nodes, and `datakit-gke-autopilot.yaml` is for Cloud API mode. Use the manifest rendered here for Partner.

### Partner configuration {#partner-configuration}

The Partner preset fixes `fullnameOverride` to `datakit`; changing only the Helm release name does not change resource names. To distinguish the installation from existing resources with the same names, set another `fullnameOverride` in user values (for example, `datakit-autopilot`) and use a separate namespace. Update the namespace and DaemonSet name in this guide's commands accordingly. The compatibility Service remains named `datakit-service` within each namespace.

V2 mounts only `/var/run/containerd`, `/proc`, and `/var/log/pods` read-only. It does not support eBPF, arbitrary host file collection, or full host monitoring. Cache uses `emptyDir` and is lost when the Pod is replaced.

Add `extraEnvs` to `datakit-user-values.yaml` for extra settings. For example, send runtime logs to stdout so they are available through `kubectl logs`:

```yaml
extraEnvs:
  - name: ENV_LOG
    value: "stdout"
```

Names must match `^ENV_[A-Z0-9_]+$`. Use `valueFrom` for existing Secrets/ConfigMaps in the same namespace. Append entries if `extraEnvs` already exists; use `datakit.*` for DataWay and cluster settings. Preserve the preset image repository, allowlist, mounts, and security settings. Do not enable `iploc`, `dkconfig`, Git SSH key mounts, or additional collector charts.

## Cloud API deployment {#cloud-api}

Cloud API collects through Cloud Monitoring/Logging without host access. Local file logs, host monitoring, and eBPF are unsupported. Metrics can lag by several minutes, and Pod or leader changes can duplicate recent logs.

### Configure GCP permissions {#cloud-permissions}

Use an account authorized to enable APIs, create service accounts, and configure IAM. Set the cluster project, enable the APIs, and bind the service account through Workload Identity. If the service account already exists, reuse its name and skip the create command.

```shell
PROJECT_ID="my-project"
GSA_NAME="datakit-cloud-monitor"
gcloud services enable monitoring.googleapis.com logging.googleapis.com \
  iam.googleapis.com iamcredentials.googleapis.com --project "$PROJECT_ID"
gcloud iam service-accounts create "$GSA_NAME" --project "$PROJECT_ID"
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member "serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/monitoring.viewer
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member "serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/logging.viewer
gcloud iam service-accounts add-iam-policy-binding \
  "${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" --project "$PROJECT_ID" \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[datakit/datakit]"
```

This binds the `datakit` Kubernetes ServiceAccount in the `datakit` namespace. Update the binding if you override the namespace or `fullnameOverride`.

### Cloud API Helm installation {#cloud-helm}

Append the actual GCP Service Account email to `datakit-user-values.yaml`:

```yaml
serviceAccountAnnotations:
  iam.gke.io/gcp-service-account: "datakit-cloud-monitor@my-project.iam.gserviceaccount.com"
```

```shell
helm upgrade --install datakit truewatch/datakit-gke-autopilot \
  --version "$CHART_VERSION" --namespace datakit --create-namespace \
  -f datakit-user-values.yaml --wait --timeout 15m
```

### Cloud API YAML installation {#cloud-yaml}

Choose this or Cloud API Helm installation. Download the dedicated Deployment YAML:

```shell
umask 077
curl -f -o datakit-gke-autopilot.yaml \
  https://static.<<<custom_key.brand_main_domain>>>/datakit-v2/datakit-gke-autopilot.yaml
```

Set `ENV_DATAWAY` (including the token) and `ENV_CLUSTER_NAME_K8S`, and replace the existing annotation on the **ServiceAccount** with the actual email. This method does not read Helm user values:

```yaml
metadata:
  annotations:
    iam.gke.io/gcp-service-account: "datakit-cloud-monitor@my-project.iam.gserviceaccount.com"
```

```shell
kubectl apply -f datakit-gke-autopilot.yaml
```

### Cloud API configuration {#cloud-configuration}

The GCP project, cluster, and location are discovered from the GKE metadata server by default. To override them, add the complete list below to Helm user values. **`extraEnvs` replaces the chart's default list. Keep the first three entries or Cloud API collection will be disabled.**

```yaml
extraEnvs:
  - name: ENV_NAMESPACE
    value: "datakit"
  - name: ENV_INPUT_CONTAINER_GCP_CLOUD_API_ENABLED
    value: "true"
  - name: ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL
    value: "false"
  - name: ENV_INPUT_CONTAINER_GCP_PROJECT_ID
    value: "my-project"
  - name: ENV_INPUT_CONTAINER_GCP_CLUSTER_NAME
    value: "my-cluster"
  - name: ENV_INPUT_CONTAINER_GCP_CLUSTER_LOCATION
    value: "asia-southeast1"
```

For YAML, edit the container env directly. The default is one replica; keep election enabled with the same election settings on every replica when scaling. For optional non-root operation, set `gkeAutopilot.runAsNonRoot=true` in Helm, or follow the YAML file's `securityContext` comments. Mounted directories must be writable by UID/GID `10001`.

## Verify operation and data {#verify}

Check rollout for the selected mode:

```shell
# Partner
kubectl -n datakit rollout status daemonset/datakit --timeout=15m
# Cloud API
kubectl -n datakit rollout status deployment/datakit --timeout=15m
```

Then check Pods and the health endpoint:

```shell
kubectl -n datakit get daemonset,deployment,pods -o wide
kubectl get --raw '/api/v1/namespaces/datakit/services/http:datakit-service:9529/proxy/v1/health'
```

Confirm desired is greater than zero, ready equals desired, actual Pods remain healthy, and the health endpoint returns `live=true`. In a cluster without nodes, the Partner DaemonSet may show `0/0`; a successful Helm installation does not mean DataKit is running. Deploy an application workload, wait for Autopilot to provision nodes, then check the DataKit Pods and health endpoint.

In the platform, filter by `cluster_name_k8s` and deployment time to check container metrics, Kubernetes objects, and logs from application Pods with collection enabled. Confirm data belongs to the target cluster and timestamps continue advancing. The DataWay URL must include a valid token; a hostname alone is insufficient to verify data delivery. Pod readiness only confirms that the workload is running.

Partner V2 disables `exec`, `attach`, and `port-forward`. Use Pod status, Events, `kubectl logs`, and the Service proxy for diagnosis. Remove tokens and complete DataWay URLs before sharing logs.

## Lifecycle {#lifecycle}

**Helm installations**: retain the current chart version and user values, select the target version, repeat the installation command for your mode, and verify again. For Partner, first download the target chart into a separate directory, explicitly supply both values files, and leave `image.tag` empty so the image follows the target `appVersion`. Cloud API user values must include the Workload Identity annotation.

Changes to the Partner chart-managed DataWay Secret trigger a DaemonSet rolling update. After changing an external Secret/ConfigMap referenced by `extraEnvs.valueFrom`, run `kubectl -n datakit rollout restart daemonset/datakit` to reload environment variables.

Use these commands as needed, replacing `<REVISION>` with a number from the history. After rollback, verify the image and external configuration:

```shell
helm history datakit --namespace datakit
helm rollback datakit <REVISION> --namespace datakit --wait --timeout 15m
helm uninstall datakit --namespace datakit
```

**YAML installations**: retain the previous manifest, render the new Partner chart or download the new Cloud API manifest, restore your configuration, then apply it. Apply the previous manifest to roll back. Identify and delete resources removed by an upgrade separately. Uninstall with `kubectl delete -n datakit -f <installation-manifest>`. The Cloud API manifest includes a Namespace; remove that object from the uninstall manifest first if the namespace contains other resources. Treat Secrets and DataWay URLs in YAML as sensitive information.

Only remove the Partner synchronizer after confirming no workloads in the cluster still use this allowlist. Preserve namespaces and configuration shared by other deployments:

```shell
kubectl delete -f datakit/gke-autopilot-partner-allowlist.yaml
```

## Troubleshooting {#troubleshooting}

- Partner preset missing: select a released chart containing this feature, update the repository index, and download again.
- Allowlist not synchronized: inspect errors under `status.managedAllowlistStatus`, the full GKE version, and allowed cluster paths. Require `Ready=True` and the file to be `Installed`.
- Partner rejected: inspect DaemonSet Events, the Pod label `cloud.google.com/matching-allowlist=truewatch-datakit-autopilot-v2`, and any injected sidecars, extra mounts, or replaced images. See [Google's admission troubleshooting guide](https://docs.cloud.google.com/kubernetes-engine/docs/troubleshooting/autopilot-privileged-workloads){:target="_blank"}.
- Partner reports `Image Mismatch`: restore `image.repository` in user values to `pubrepo.truewatch.com/truewatch/datakit` and leave `image.tag` empty. For YAML installations, render the same chart version again and apply it. For `ErrImagePull` or `ImagePullBackOff`, check that the corresponding TrueWatch image has been published, and check node connectivity to the registry and image pull permissions.
- Logs report `dataway.emptyToken` or `token missing`: add a valid workspace token to the DataWay URL and update using the original installation method. An empty token causes rejected uploads, dropped data, and failed elections, even when Pods are Ready and health checks pass.
- Cloud API rejected or missing data: check for added host privileges, GCP IAM permissions, the ServiceAccount annotation, and matching namespace/name in the Workload Identity binding.
- Healthy Pods with missing or duplicate data: check DataWay connectivity, settings, Kubernetes API permissions, application Pod log collection settings, and election. For non-root write failures, check directory permissions.
- `kubectl logs` shows only startup logs: set `ENV_LOG=stdout` in Partner user values (see the [configuration example](gcp-gke-autopilot.md#partner-configuration)), then upgrade using the original installation method. For YAML, render the manifest again before applying it.
