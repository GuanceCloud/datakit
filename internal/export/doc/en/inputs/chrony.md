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

    Kubernetes supports modifying configuration parameters in the form of environment variables(Only effective when Datakit runs in K8s DaemonSet mode, not supported by host deployed Datakit):

    | Environment Variable Name                             | Corresponding Configuration Parameter Item         | Parameter                                                     |
    | :-----------------------------          | ---                      | ---                                                    |
    | `ENV_INPUT_CHRONY_INTERVAL`             | `interval`               | `"30s"` (`"10s"` ~ `"60s"`)                            |
    | `ENV_INPUT_CHRONY_TIMEOUT`              | `timeout`                | `"8s"`  (`"5s"` ~ `"30s"`)                             |
    | `ENV_INPUT_CHRONY_BIN_PATH`             | `bin_path`               | `"chronyc"`                                            |
    | `ENV_INPUT_CHRONY_REMOTE_ADDRS`         | `remote_addrs`           | `["192.168.1.1:22"]`                                   |
    | `ENV_INPUT_CHRONY_REMOTE_USERS`         | `remote_users`           | `["remote_login_name"]`                                |
    | `ENV_INPUT_CHRONY_REMOTE_RSA_PATHS`     | `remote_rsa_paths`       | `["/home/<your_name>/.ssh/id_rsa"]`                    |
    | `ENV_INPUT_CHRONY_REMOTE_COMMAND`       | `remote_command`         | `"chronyc -n tracking"`                                |
    | `ENV_INPUT_CHRONY_TAGS`                 | `tags`                   | `tag1=value1,tag2=value2` If there is a tag with the same name in the configuration file, it will be overwritten. |
    | `ENV_INPUT_CHRONY_ELECTION`             | `election`               | `"true"` or `"false"`                                   |

<!-- markdownlint-enable -->

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
