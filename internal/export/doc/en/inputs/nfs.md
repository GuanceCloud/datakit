---
title     : 'NFS'
summary   : 'Collect metrics of NFS'
tags:
  - 'HOST'
__int_icon      : 'icon/nfs'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor:
  - desc: 'N/A'
    path: '-'
---

{{.AvailableArchs}}

---

NFS metrics collector that collects the following data:

- RPC throughput metrics
- NFS mount point metrics (NFSv3 and v4 only)
- NFSd throughput metrics

## Configuration {#config}

### Preconditions {#requirements}

- The NFS client environment is properly configured.
- The NFS client is properly mounted to the server's shared directory.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

### NFSd Start {#nfsd}

NFSd is the daemon of the NFS service, a key component on the server side, responsible for handling NFS requests sent by clients. If the local machine is also used as an NFS server, you can enable this metric to view statistics such as network, disk I/O, and threads on which users process NFS requests.

If you want to enable it, you need to modify the configuration file.

```toml
[[inputs.nfs]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
  ## Whether to enable NFSd metric collection
  nfsd = true

...

```

### NFS mount point stats Start {#nfs-mountstats}

The default set of nfs_mountstats metrics only displays statistics about the disk usage and NFS running time of the mount point, and you need to modify the configuration file to view the R/W, Transport, Event, and Operations information of the NFS mount point.

```toml
[[inputs.nfs]]

  ...

  ## NFS mount point metric configuration
  [inputs.nfs.mountstats]
    ## Enable r/w statistics
    rw = true
    ## Enable transport statistics
    transport = true
    ## Enable event statistics
    event = true
    ## Enable operation statistics
    operations = true

...

```

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
