# DataKit Helm Chart

This Helm chart installs DataKit with configurable TLS, RBAC, and collector settings.

- [Requirements](#requirements)
- [Installing](#installing)
- [Uninstalling](#uninstalling)
- [Configuration](#configuration)

## Requirements

- Kubernetes 1.14+

- Helm 3.0+

## Installing

### Default Configuration

Add the Helm repository and install DataKit:

```shell
helm repo add datakit https://pubrepo.{{.BrandChartRepo}}
helm install datakit datakit/datakit -n datakit --set datakit.dataway_url="https://openway.{{.BrandDomain}}?token=<YOUR-TOKEN>" --create-namespace
```

The command deploys DataKit on the Kubernetes cluster with the default configuration.

### Manage Collector Configuration with Git

Using a Git password:

```shell
helm repo add datakit https://pubrepo.{{.BrandChartRepo}}
helm install datakit datakit/datakit -n datakit \
    --set git_repos.enable=true \
    --set datakit.dataway_url="https://openway.{{.BrandDomain}}?token=<YOUR-TOKEN>" \
    --set git_repos.git_url="http://username:password@github.com/path/to/repository.git" \
    --create-namespace
```

Using an SSH private key:

```shell
helm repo add datakit https://pubrepo.{{.BrandChartRepo}}
helm install datakit datakit/datakit -n datakit \
    --set git_repos.enable=true \
    --set datakit.dataway_url="https://openway.{{.BrandDomain}}?token=<YOUR-TOKEN>" \
    --set git_repos.git_url="git@github.com:path/to/repository.git" \
    --set-file git_repos.git_key_path="$HOME/.ssh/id_rsa" \
    --create-namespace
```

## GKE Autopilot Partner

**Partner V2 supports TrueWatch only and requires GKE `1.35.6-gke.1258000` or later.**
The regular `datakit` chart bundles `values-gke-autopilot-partner.yaml` and
`gke-autopilot-partner-allowlist.yaml`. Use the TrueWatch chart and
`pubrepo.truewatch.com/truewatch/datakit` image; leave `image.tag` empty to follow
`appVersion`.

Follow the [Partner deployment guide](https://docs.truewatch.com/datakit/gcp-gke-autopilot/#partner)
to synchronize the allowlist, configure a valid DataWay token, and install with
Helm or rendered YAML. The same guide covers upgrades and cleanup, plus the
alternative `datakit-gke-autopilot` Cloud API chart.

## Uninstalling
To uninstall the chart with the release name `datakit`:

```shell
helm uninstall datakit -n datakit
```

## Configuration

| Parameter                        | Description                                                                                                                                                                        | Default                                                                   | Required             |
| ---                              | ---                                                                                                                                                                                | ---                                                                       | ---                  |
| `image.repository`               | The DataKit Docker image                                                                                                                                                           | `pubrepo.{{.BrandChartRepo}}`                                             | `true`               |
| `image.pullPolicy`               | The Kubernetes [imagePullPolicy][] value                                                                                                                                           | `IfNotPresent`                                                            |                      |
| `image.tag`                      | The DataKit Docker image tag                                                                                                                                                       | `""`                                                                      |                      |
| `datakit.dataway_url`            | The DataWay url, contain`TOKEN`                                                                                                                                                    | `https://openway.{{.BrandDomain}}?token=<YOUR-TOKEN>`                     | `true`               |
| `datakit.global_tags`            | It supports filling in global tags in the installation phase. The format example is: `project=abc,owner=def` (multiple tags are separated by commas)                               | `host=__datakit_hostname,host_ip=__datakit_ip`                            |                      |
| `datakit.default_enabled_inputs` | The default open collector list, format example: input1, input2, input3                                                                                                            | `cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container` |                      |
| `datakit.enabled_election`       | When the election is enabled, it is not enabled by default. If it needs to be enabled, you can give any non empty string value to the environment variable. (e.g. true / false)    | `enable`                                                                  |                      |
| `datakit.election_operator_url`  | DataKit Operator URL used for central election. When empty, DataKit continues to use DataWay/Kodo.                                                                                | `""`                                                                      |                      |
| `datakit.log_level`              | Optional value info / debug                                                                                                                                                        | `info`                                                                    |                      |
| `datakit.http_listen`            | It supports specifying the network card bound to the Datakit HTTP service in the installation phase (default localhost)                                                            | `0.0.0.0:9529`                                                            |                      |
| `workload.kind`                  | Workload type, either `DaemonSet` or `Deployment`                                                                                                                                  | `DaemonSet`                                                               |                      |
| `workload.replicas`              | Replica count when `workload.kind` is `Deployment`                                                                                                                                | `1`                                                                       |                      |
| `gkeAutopilot.enabled`           | Disable host access and privileged mode for GKE Autopilot; requires `workload.kind=Deployment`                                                                                    | `false`                                                                   |                      |
| `serviceAccountAnnotations`      | Annotations added to the DataKit Kubernetes ServiceAccount, such as the GKE Workload Identity annotation                                                                            | `{}`                                                                      |                      |
| `git_repos.enable`               | use git management DataKit input                                                                                                                                                   | `false`                                                                   |                      |
| `git_repos.git_url`              | The remote git repo address of the management profile. (e.g. `http://username:password@github.com/username/repository.git`)                                                       | `-`                                                                       |                      |
| `git_repos.git_key_path`         | The full path of the local privatekey. (e.g. `~/.ssh/id_rsa`)                                                                                                                  | `-`                                                                       |                      |
| `git_repos.git_key_pw`           | The password used by the local privatekey. (e.g. passwd)                                                                                                                           | `-`                                                                       |                      |
| `git_repos.git_branch`           | Specifies the branch to pull. If it is blank, it is the default. The default is the main branch specified remotely, usually the master.                                            | `master`                                                                  |                      |
| `git_repos.git_interval`         | Timed pull interval. (e.g. 1m)                                                                                                                                                     | `1m`                                                                      |                      |
| `extraEnvs`                      | extra env Add env for customization,[more](https://www.yuque.com/dataflux/datakit/datakit-install#f9858758)                                                                        | `[]`                                                                      |                      |
| `componentHealthLivenessProbe.enabled` | Restart DataKit when `/v1/health` reports that a component cannot recover. Requires DataKit >= 2.5.0 | `false` | |
| `nameOverride`                   | Overrides the `clusterName` when used in the naming of resources                                                                                                                   | ""                                                                        |                      |
| `fullnameOverride`               | Overrides the `clusterName` and `nodeGroup` when used in the naming of resources. This should only be used when using a single `nodeGroup`, otherwise you will have name conflicts | ""                                                                        |                      |
| `podAnnotations`                 | Configurable [annotations][] applied to DataKit pods                                                                                                                        | `datakit/logs: '[{"disable": true}]'` | |
| `tolerations`                    | Configurable [tolerations][]                                                                                                                                                       | `- operator: Exists`                                                      |                      |
| `service.type`                   | DataKit [Service Types][]                                                                                                                                                          | `ClusterIP`                                                               |                      |
| `service.port`                   | DataKit service port                                                                                                                                                               | `9529`                                                                    |                      |
| `dkconfig.path`                  | DataKit input path                                                                                                                                                                 | `nil`                                                                     |                      |
| `dkconfig.name`                  | DataKit input name                                                                                                                                                                 | `nil`                                                                     |                      |
| `dkconfig.value`                 | DataKit input value                                                                                                                                                                | `nil`                                                                     |                      |
| `kubeStateMetricsEnabled`        | For large clusters where the Kubernetes State Metrics Check Core needs to be distributed on dedicated workers.                                                                     | `true`                                                                    |                      |
| `MetricsServerEnabled`           | Kubernetes Metrics Server                                                                                                                                                          | `true`                                                                    |                      |
| `iploc.enable`                   | Datakit install ipdb                                                                                                                                                               | `false`                                                                   |                      |
| `iploc.image`                    | Iploc image repository                                                                                                                                                             | `pubrepo.{{.BrandChartRepo}}/iploc`                                       |                      |
| `iploc.tag`                      | Iploc image tag                                                                                                                                                                    | `1.0`                                                                     |                      |

[imagePullPolicy]: https://kubernetes.io/docs/concepts/containers/images/#updating-images

[annotations]: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/

[tolerations]: https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

[service types]: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
