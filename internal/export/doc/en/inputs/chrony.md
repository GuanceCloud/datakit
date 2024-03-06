---
title     : 'Chrony'
summary   : 'Collect metrics related to Chrony server'
__int_icon: 'icon/chrony'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Chrony
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

The Chrony collector is used to collect metrics related to the Chrony server.
The Chrony collector supports remote collection, and the collector Datakit can run on multiple operating systems.

## Configuration {#config}

### Precondition {#requirements}

- Install Chrony service

```shell
$ yum -y install chrony    # [On CentOS/RHEL]
...

$ apt install chrony       # [On Debian/Ubuntu]
...

$ dnf -y install chrony    # [On Fedora 22+]
...

```

- Verify if the installation is correct, execute the following command on the command line, and obtain similar results:

```shell
$ chronyc -n tracking
Reference ID    : CA760151 (202.118.1.81)
Stratum         : 2
Ref time (UTC)  : Thu Jun 08 07:28:42 2023
System time     : 0.000000000 seconds slow of NTP time
Last offset     : -1.841502666 seconds
RMS offset      : 1.841502666 seconds
Frequency       : 1.606 ppm slow
Residual freq   : +651.673 ppm
Skew            : 0.360 ppm
Root delay      : 0.058808800 seconds
Root dispersion : 0.011350543 seconds
Update interval : 0.0 seconds
Leap status     : Normal
```

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):

{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
