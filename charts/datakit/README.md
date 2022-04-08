# DataKit Helm Chart

This Helm chart installs [Datakit](https://github.com/GuanceCloud/datakit) with configurable TLS, RBAC and much more configurations. This chart caters a number of different use cases and setups.

- [Requirements](#requirements)
- [Installing](#installing)
- [Uninstalling](#uninstalling)
- [Configuration](#configuration)

## Requirements

* Kubernetes >= 1.14

* Helm >= 2.17.0

  

## Installing

- default configuration

​	Once you've added this Helm repository as per the repository-level [README](../../README.md#installing) then you can install the chart as follows:

 ```shell
 helm repo add dataflux https://pubrepo.guance.com/chartrepo/datakit
 
 helm install my-datakit dataflux/datakit -n datakit --set dataway_url="https://openway.guance.com?token=<your-token>" --create-namespace 
 ```

​	The command deploys DataKit on the Kubernetes cluster in the default configuration.

​	**NOTE:** If using Helm 2 then you'll need to add the [`--name`](https://v2.helm.sh/docs/helm/#options-21) command line argument. If unspecified, Helm 2 will autogenerate a name for you.

- use git management datakit input
  - git passwd
  
    ```
    helm repo add dataflux https://pubrepo.guance.com/chartrepo/datakit
    
    helm install my-datakit dataflux/datakit -n datakit --set git_repos.enable=true  --set dataway_url="https://openway.guance.com?token=<your-token>" \
    --set git_repos.git_url="http://username:password@github.com/path/to/repository.git" \
    --create-namespace 
    ```
  
  - git key
  
    ```
    helm repo add dataflux https://pubrepo.guance.com/chartrepo/datakit
    
    helm install my-datakit dataflux/datakit -n datakit --set git_repos.enable=true  --set dataway_url="https://openway.guance.com?token=<your-token>"  \
    --set git_repos.git_url="git@github.com:path/to/repository.git" \
    --set-file git_repos.git_key_path="/Users/buleleaf/.ssh/id_rsa" \
    --create-namespace 
    ```
    
    

## Uninstalling
To delete/uninstall the chart with the release name `my-release`:

```shell
helm uninstall my-datakit -n datakit
```

## Configuration

| Parameter                | Description                                                  | Default                                                      | Required |
| ------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ | -------- |
| `image.repository`       | The DataKit Docker image                                     | `pubrepo.guance.com/chartrepo/datakit`                      | `true`   |
| `image.pullPolicy`       | The Kubernetes [imagePullPolicy][] value                     | `IfNotPresent`                                               |          |
| `image.tag`              | The DataKit Docker image tag                                 | `""`                                                         |          |
| `dataway_url`            | The DataWay url, contain`TOKEN`                              | `https://openway.guance.com?token=<your-token>`              | `true`   |
| `global_tags`            | It supports filling in global tags in the installation phase. The format example is: Project = ABC, owner = Zhang San (multiple tags are separated by English commas) | `host=__datakit_hostname,host_ip=__datakit_ip`               |          |
| `default_enabled_inputs` | The default open collector list, format example: input1, input2, input3 | cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container |          |
| `enabled_election`       | When the election is enabled, it is not enabled by default. If it needs to be enabled, you can give any non empty string value to the environment variable. (e.g. true / false) | `enable`                                                     |          |
| `log`                    | Optional value info / debug / stdout                         | `stdout`                                                     |          |
| `http_listen`            | It supports specifying the network card bound to the Datakit HTTP service in the installation phase (default localhost) | `0.0.0.0:9529`                                               |          |
| `git_repos.enable`       | use git management DataKit input                             | `false`                                                      |          |
| `git_repos.git_url`      | The remote git repo address of the management profile. (e.g http://username:password @github. com/username/repository. git） | `-`                                                          |          |
| `git_repos.git_key_path` | The full path of the local privatekey. (e.g. / users / username /. SSH / id_rsa) | `-`                                                          |          |
| `git_repos.git_key_pw`   | The password used by the local privatekey. (e.g. passwd)     | `-`                                                          |          |
| `git_repos.git_branch`   | Specifies the branch to pull. If it is blank, it is the default. The default is the main branch specified remotely, usually the master. | `master`                                                     |          |
| `git_repos.git_interval` | Timed pull interval. (e.g. 1m)                               | `1m`                                                         |          |
| `extraEnvs`              | extra env Add env for customization,[more](https://www.yuque.com/dataflux/datakit/datakit-install#f9858758) | `[]`                                                         |          |
| `nameOverride`           | Overrides the `clusterName` when used in the naming of resources | ""                                                           |          |
| `fullnameOverride`       | Overrides the `clusterName` and `nodeGroup` when used in the naming of resources. This should only be used when using a single `nodeGroup`, otherwise you will have name conflicts | ""                                                           |          |
| `podAnnotations`         | Configurable [annotations][] applied to all OpenSearch pods  | `  datakit/logs: | [{"disable": true}]`                      |          |
| `tolerations`            | Configurable [tolerations][]                                 | `- operator: Exists`                                         |          |
| `service.type`           | DataKit [Service Types][]                                    | `ClusterIP`                                                  |          |
| `service.port`           | DataKit service port                                         | `9529`                                                       |          |
| `dkconfig.path`          | DataKit input path                                           | `nil`                                                        |          |
| `dkconfig.name`          | DataKit input name                                           | `nil`                                                        |          |
| `dkconfig.value`         | DataKit input value                                          | `nil`                                                        |          |



[environment from variables]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#configure-all-key-value-pairs-in-a-configmap-as-container-environment-variables

[hostAliases]: https://kubernetes.io/docs/concepts/services-networking/add-entries-to-pod-etc-hosts-with-host-aliases/

[image.pullPolicy]: https://kubernetes.io/docs/concepts/containers/images/#updating-images


[annotations]: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/

[tolerations]: https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/

[service types]: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
