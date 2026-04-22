# Managing Configuration with Helm

This document describes how to install and upgrade DataKit with Helm, and how to manage DataKit environment variables and collection configurations. In Kubernetes, DataKit is mainly configured through environment variables and mounted configuration files. Helm configuration is centralized in *values.yaml*.

## Installation and Configuration {#install-config}

### Prerequisites {#helm-prerequisites}

- Kubernetes >= 1.14
- Helm >= 3.0
- DataWay URL and token

### Download DataKit Charts Package with Helm {#download-config}

```shell
helm pull datakit --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit --untar
```

### Modify values.yaml {#values-configuration}

<!-- markdownlint-disable MD046 -->
???+ info

     `values.yaml` is located in the `datakit` directory.
<!-- markdownlint-enable -->

#### Modify `dataway url`  {#helm-dataway}

```yaml
...
datakit:
  # DataKit will send the indicator data to dataway. Please be sure to change the parameters
  # @param dataway_url - string - optional - default: 'https://<<<custom_key.brand_main_domain>>>'
  # The host of the DataKit intake server to send Agent data to, only set this option
  dataway_url: https://openway.<<<custom_key.brand_main_domain>>>?token=tkn_xxxxxxxxxx
...
```

#### Add Default Collectors  {#helm-default-config}
  
Add `rum` by appending the parameter to the end of `default_enabled_inputs`.

```yaml
..
datakit:
  ...
  # @param default_enabled_inputs - string
  # The default open collector list, format example: input1, input2, input3
  default_enabled_inputs: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,rum
....
```

#### Add Global Tags {#helm-tag}

Add `cluster_name_k8s` global tag.

```yaml
datakit:
  ...
  # @param global_tags - string - optional - default: 'host=__datakit_hostname,host_ip=__datakit_ip'
  # It supports filling in global tags in the installation phase. The format example is: Project = ABC, owner = Zhang San (multiple tags are separated by English commas)
  global_tags: host=__datakit_hostname,host_ip=__datakit_ip,cluster_name_k8s=prod  
```

#### Add DataKit Environment Variables {#helm-env}

For more environment variables, refer to [Container Environment Variables](datakit-daemonset-deploy.md#using-k8-env)

```yaml
# @param extraEnvs - array - optional
# extra env Add env for customization
# more, see: https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env
# You can add more than one parameter  
extraEnvs:
 - name: ENV_NAMESPACE
   value: government-prod
 - name: ENV_GLOBAL_ELECTION_TAGS
   value: cluster_name_k8s=government-prod
```

#### Mount Collector Configurations {#helm-config}
  
Taking container host system log collection as an example, `path` is the container path and must be under `/usr/local/datakit/conf.d/`. `name` is the configuration name. `value` is the collection configuration content. You can obtain the collector's sample files by entering the `/usr/local/datakit/conf.d/` directory in the container.

```yaml
dkconfig:   
 - path: "/usr/local/datakit/conf.d/logging.conf"
   name: logging.conf
   value: |-
     [[inputs.logging]]
       logfiles = [
         "/var/log/syslog",
         "/var/log/message",
       ]
       ignore = [""]
       source = ""
       service = ""
       pipeline = ""
       ignore_status = []
       character_encoding = ""
       auto_multiline_detection = true
       auto_multiline_extra_patterns = []
       remove_ansi_escape_codes = true
       blocking_mode = true
       ignore_dead_log = "1h"
       [inputs.logging.tags]
```

#### Mount Pipeline  {#helm-pipeline}

Taking `test.p` as an example, `path` is the absolute path of the configuration file and must be under */usr/local/datakit/pipeline/*. `name` is the Pipeline name. `value` is the Pipeline content.

```yaml
dkconfig:
 - path: "/usr/local/datakit/pipeline/test.p"
   name: test.p
   value: |-
     # access log
     grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{NUMBER:http_code} ")
     grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"-\" %{NUMBER:http_code} ")
     default_time(time)
     cast(http_code,"int")

     # error log
     grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{GREEDYDATA:pid}:tid %{GREEDYDATA:tid}\\] ")
     grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{INT:pid}\\] ")
     default_time(time)
```

### Install DataKit {#datakit-install}

You can install DataKit directly from the remote chart repository:

```shell
helm install datakit datakit \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit --create-namespace \
         -f values.yaml
```

Output:

```shell
NAME: datakit
LAST DEPLOYED: Tue Apr  4 19:13:29 2023
NAMESPACE: datakit
STATUS: deployed
REVISION: 1
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace datakit -l "app.kubernetes.io/name=datakit,app.kubernetes.io/instance=datakit" -o jsonpath="{.items[0].metadata.name}")
  export CONTAINER_PORT=$(kubectl get pod --namespace datakit $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
  echo "Visit http://127.0.0.1:9527 to use your application"
  kubectl --namespace datakit port-forward $POD_NAME 9527:$CONTAINER_PORT
```

After installation, check the Helm release and Pod status:

```shell
helm -n datakit list
kubectl -n datakit get ds,pod -l app.kubernetes.io/instance=datakit
```

## Install Specific Version {#version-install}

There are two common version concepts in Helm installation:

- `--version`: specifies the Helm chart version.
- `image.tag`: specifies the DataKit container image version. If not set, the chart `appVersion` is used by default.

Specify the chart version:

```shell
helm install datakit datakit \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit --create-namespace \
         -f values.yaml \
         --version 1.5.x
```

Specify the DataKit image version:

```shell
helm install datakit datakit \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit --create-namespace \
         -f values.yaml \
         --set image.tag="<DATAKIT-IMAGE-TAG>"
```

## Upgrade {#datakit-upgrade}

<!-- markdownlint-disable MD046 -->
???+ info

     If *values.yaml* is lost, you can execute `helm -n datakit get values datakit -o yaml > values.yaml` to retrieve it.
<!-- markdownlint-enable -->

```shell
helm upgrade datakit datakit \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit \
         -f values.yaml
```

To pin both chart and image versions:

```shell
helm upgrade datakit datakit \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit \
         -f values.yaml \
         --version <CHART-VERSION> \
         --set image.tag="<DATAKIT-IMAGE-TAG>"
```

## Uninstall {#datakit-uninstall}

```shell
helm uninstall datakit -n datakit 
```

## GKE Autopilot {#gke-autopilot}

GKE Autopilot has additional restrictions on workload privileges and host access. The regular DataKit chart may fail Autopilot admission checks because it uses settings such as `hostNetwork`, `hostPID`, `hostIPC`, `hostPath`, and privileged containers. Use the separately released Helm chart instead: `datakit-gke-autopilot`.

The GKE Autopilot chart is not released in sync with the main DataKit version. You do not need to specify an image version during installation; the image version declared by this chart is used by default.

Main differences from the regular DataKit chart:

- The default collector list is reduced to `dk,cpu,mem,container,kubernetesprometheus`.
- The DataKit container runs as a non-root user, with UID/GID `10001` by default, and privileged mode and privilege escalation are disabled.
- `hostNetwork`, `hostPID`, and `hostIPC` are disabled, and `emptyDir` is used instead of host `hostPath` mounts.
- Host filesystem, container runtime socket, eBPF, and similar host-level collection capabilities are limited. If these capabilities are required, use GKE Standard or a regular Kubernetes cluster with the DataKit chart.

### Install {#gke-autopilot-install}

```shell
helm install datakit datakit-gke-autopilot \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit --create-namespace \
         --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>"
```

### Upgrade {#gke-autopilot-upgrade}

Back up the current values before upgrading:

```shell
helm -n datakit get values datakit -o yaml > values-gke-autopilot.yaml
```

```shell
helm upgrade datakit datakit-gke-autopilot \
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit \
         -n datakit \
         -f values-gke-autopilot.yaml
```

Check the status:

```shell
helm -n datakit list
kubectl -n datakit get pod -l app.kubernetes.io/instance=datakit
```

If the Pod is rejected by GKE Autopilot, first check whether the regular `datakit` chart was used by mistake, or whether extra `hostPath`, privileged container, host network, or other Autopilot-disallowed settings were enabled in values.

## Configuration File Reference {#config-reference}

<!-- markdownlint-disable MD046 -->
???- info "values.yaml"

    ```yaml
    # Default values for datakit.
    # This is a YAML-formatted file.
    # Declare variables to be passed into your templates.

    datakit:
      # DataKit will send the indicator data to dataway. Please be sure to change the parameters
      # @param dataway_url - string - optional - default: 'https://<<<custom_key.brand_main_domain>>>'
      # The host of the DataKit intake server to send Agent data to, only set this option
      dataway_url: https://openway.<<<custom_key.brand_main_domain>>>?token=tkn_xxxxxxxxxx

      # @param global_tags - string - optional - default: 'host=__datakit_hostname,host_ip=__datakit_ip'
      # It supports filling in global tags in the installation phase. The format example is: Project = ABC, owner = Zhang San (multiple tags are separated by English commas)
      global_tags: host=__datakit_hostname,host_ip=__datakit_ip,cluster_name_k8s=government-prod

      # @param default_enabled_inputs - string
      # The default open collector list, format example: input1, input2, input3
      default_enabled_inputs: cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,rum

      # @param enabled_election - boolean
      # When the election is enabled, it is enabled by default. If it needs to be enabled, you can give any non empty string value to the environment variable. (e.g. true / false)
      enabled_election: true

      # @param log - string
      # Set logging verbosity, valid log levels are:
      # info, debug, stdout, warn, error, critical, and off
      log_level: info

      # @param http_listen - string
      # It supports specifying the network card bound to the DataKit HTTP service in the installation phase (default localhost)
      http_listen: 0.0.0.0:9529

    image:
      # @param repository - string - required
      # Define the repository to use:
      #
      repository:  pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit

      # @param tag - string - required
      # Define the Cluster-Agent version to use.
      #
      tag: ""

      # @param pullPolicy - string - optional
      # The Kubernetes [imagePullPolicy][] value
      #
      pullPolicy: Always

    # https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/

    git_repos:
      # use git management DataKit input
      enable: false

      # @param git_url - string - required
      # You Can Set git@github.com:path/to/repository.git or http://username:password@github.com/path/to/repository.git.
      # see https://docs.<<<custom_key.brand_main_domain>>>/best-practices/insight/datakit-daemonset/#git
      git_url: "-"

      # @param git_key_path - string - optional
      # The Git Ssh Key Content,
      # For details,
      # -----BEGIN OPENSSH PRIVATE KEY--
      # ---xxxxx---
      #--END OPENSSH PRIVATE KEY-----
      git_key_path: "-"

      # @param git_key_pw - string - optional
      # The ssh Key Password
      git_key_pw: "-"

      # @param git_url - string - required
      # Specifies the branch to pull. If it is blank, it is the default. The default is the main branch specified remotely, usually the master.
      git_branch: "master"

      # @param git_url - string - required
      # Timed pull interval. (e.g. 1m)
      git_interval: "1m"
      is_use_key: false

    # If true, DataKit install ipdb.
    # ref: https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-tools-how-to/#install-ipdb
    iploc:
      enable: true
      image:
        # @param repository - string - required
        # Define the repository to use:
        #
        repository: "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/iploc"

        # @param tag - string - required
        # Define the Cluster-Agent version to use.
        #
        tag: "1.0"

    # @param extraEnvs - array - optional
    # extra env Add env for customization
    # more, see: https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env
    # You can add more than one parameter
    extraEnvs:
     - name: ENV_NAMESPACE
       value: government-prod
     - name: ENV_GLOBAL_ELECTION_TAGS
       value: cluster_name_k8s=government-prod
     # - name: ENV_NAMESPACE # electoral
     #   value: k8s
     # - name: "NODE_OPTIONS"
     #   value: "--max-old-space-size=1800"


    resources:
      requests:
        cpu: "200m"
        memory: "128Mi" 
      limits:
        cpu: "2000m"
        memory: "4Gi"

    # @param nameOverride - string - optional
    # Override name of app.
    #
    nameOverride: ""

    # @param fullnameOverride - string - optional
    # Override name of app.
    #
    fullnameOverride: ""

    podAnnotations:
      datakit/logs: |
        [{"disable": true}]

    # @param tolerations - array - optional
    # Allow the DaemonSet to schedule on tainted nodes (requires Kubernetes >= 1.6)
    #
    tolerations:
      - operator: Exists

    service:
      type: ClusterIP
      port: 9529

    # @param dkconfig - array - optional
    # Configure DataKit custom input
    #
    dkconfig: 
     - path: "/usr/local/datakit/conf.d/logging.conf"
       name: logging.conf
       value: |-
         [[inputs.logging]]
           logfiles = [
             "/var/log/syslog",
             "/var/log/message",
           ]
           ignore = [""]
           source = ""
           service = ""
           pipeline = ""
           ignore_status = []
           character_encoding = ""
           auto_multiline_detection = true
           auto_multiline_extra_patterns = []
           remove_ansi_escape_codes = true
           blocking_mode = true
           ignore_dead_log = "1h"
           [inputs.logging.tags]
     - path: "/usr/local/datakit/pipeline/test.p"
       name: test.p
       value: |-
         # access log
         grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{NUMBER:http_code} ")
         grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"-\" %{NUMBER:http_code} ")
         default_time(time)
         cast(http_code,"int")

         # error log
         grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{GREEDYDATA:pid}:tid %{GREEDYDATA:tid}\\] ")
         grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{INT:pid}\\] ")

    # If true, deploys the kube-state-metrics deployment.
    # ref: https://github.com/kubernetes/charts/tree/master/stable/kube-state-metrics
    kubeStateMetricsEnabled: true

    # If true, deploys the metrics-server deployment.
    # ref: https://github.com/kubernetes-sigs/metrics-server/tree/master/charts/metrics-server
    MetricsServerEnabled: false
    ```
<!-- markdownlint-enable -->

## FAQ {#faq}

### Securing Dataway Token with Kubernetes Secret {#secure-dataway}

DataKit supports two methods to secure `dataway_token` in Kubernetes configuration.

- When installing DataKit with Helm, you can hide `dataway_token` by configuring Secret:

    **Install with Helm command, enable Secret mode**

    ```bash
    helm install datakit charts/datakit \
      --set datakit.dataway_url="https://openway.example.com?token=tkn_xxxxxxxxxxxx" \
      --set datakit.dataway_secret_enabled=true
    ```

    With this approach:
    - Helm automatically creates a Kubernetes Secret to store the encrypted `dataway_url`
    - The `ENV_DATAWAY` environment variable in the Pod references the Secret

- When installing with native YAML files, you need to manually create the Secret and modify environment variable references:

    1. **Create Secret**
        Create a Secret containing ENV_DATAWAY:

        ```yaml
        apiVersion: v1
        kind: Secret
        metadata:
          name: datakit-dataway-secret
          namespace: datakit
        type: Opaque
        data:
          ENV_DATAWAY: <base64-encoded-dataway-url>
        ```

        Base64 encode your `dataway_url`:

        ```bash
        echo -n "https://openway.example.com?token=tkn_xxxxxxxxxxxx" | base64
        ```

    2. **Modify Environment Variable Reference**
        In `datakit.template.yaml` or `datakit-deployment.template.yaml`, change the `ENV_DATAWAY` environment variable definition from:

        ```yaml
        - name: ENV_DATAWAY
          value: "https://openway.example.com?token=tkn_xxxxxxxxxxxx"
        ```

        to:

        ```yaml
        - name: ENV_DATAWAY
          valueFrom:
            secretKeyRef:
              name: datakit-dataway-secret
              key: ENV_DATAWAY
        ```

    3. **Apply Configuration**

       ```bash
       kubectl apply -f datakit.yaml
       ```

### PodSecurityPolicy Issue {#pod-security-policy}

`PodSecurityPolicy` was deprecated in [Kubernetes`1.21`](https://kubernetes.io/blog/2021/04/06/podsecuritypolicy-deprecation-past-present-and-future/){:target="_blank"} and removed in Kubernetes`1.25`.
If you forcibly upgrade the cluster version, Helm deployment of `kube-state-metrics` will report an error:

```shell
Error: UPGRADE FAILED: current release manifest 
contains removed kubernetes api(s) for this kubernetes
version and it is therefore unable to build the
kubernetes objects for performing the diff. error from
kubernetes: unable to recognize "": no matches for kind
"PodSecurityPolicy" in version "policy/v1beta1"
```

#### Backup Helm Values {#get-values}

```shell
helm get values -n datakit datakit -o yaml > values.yaml
```

#### Clear Helm Information {#delete-values}

Delete Helm information secrets in the DataKit namespace.

- Get secrets

  ```shell
  $ kubectl get secrets -n datakit
  NAME                            TYPE                 DATA   AGE
  sh.helm.release.v1.datakit.v1   helm.sh/release.v1   1      4h17m
  sh.helm.release.v1.datakit.v2   helm.sh/release.v1   1      4h17m
  sh.helm.release.v1.datakit.v3   helm.sh/release.v1   1      4h16m
  ```

- Delete secrets with `sh.helm.release.v1.datakit`

  ```shell
  kubectl delete  secrets sh.helm.release.v1.datakit.v1 sh.helm.release.v1.datakit.v2 sh.helm.release.v1.datakit.v3   -n datakit
  ```

#### Re-upgrade or Install {#reinstall}

```shell
helm upgrade -i -n datakit datakit  --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit  -f values.yaml
```
