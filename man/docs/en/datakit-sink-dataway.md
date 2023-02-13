# Dataway
---

If you want to type data into a different workspace, you can use the Dataway Sink feature:

1. In a DataKit, you can configure multiple dataway sink addresses, which are generally just token different. In addition, one or more data decision conditions may be attached to each sink address.
1. For the data that meets the conditions (generally by judging the value on tag/field), call the corresponding data.
1. If the data does not meet all the judgment conditions, the data will continue to be called to the default data.

Currently the data sink supports all kinds of data (M/O/L/T/...).

???+ attention

    If there is an intersection between the decision conditions of multiple data sinks, the data meeting multiple decision conditions will be written into the corresponding workspace respectively, which may cause some data duplication.

## Dataway Sink Configuration {#config}

- Step 1: Build back-end storage

Use Dataway from [Guance Cloud](https://console.guance.com/), or build your own Dataway server environment.

- Step 2: Add configuration

=== "datakit.conf"

    Add the following fragment to `datakit.conf`:
    
    ```toml
    ...
    [sinks]
       [[sinks.sink]]
         categories = ["M"]
         filters = ["{host='user-ubuntu'}", "{cpu='cpu-total'}"]
         target = "dataway"
         token = <YOUR-TOKEN1>
         url = "https://openway.guance.com"
     
       [[sinks.sink]]
         categories = ["M"]
         filters = ["{cpu='cpu-total'}"]
         target = "dataway"
         token = <YOUR-TOKEN2>
         url = "https://openway.guance.com"
    ...
    ```
    
    In addition to the fact that the Sink must be configured with the [generic parameter](datakit-sink-guide.md), the Sink instance of Dataway currently supports the following parameters:
    
    - `url` (required): Fill in the full address of the data here (with token).
    - `token` (optional): The token of the workspace. If you write here in `url`, you don't have to fill it out.
    - `filters` (optional): Filter rules. Similar to io's `filters`, but the function is completely opposite. The filters match in sink is satisfied before writing data; If the filters match in. io is satisfied, the data is discarded. The former is `include` and the latter `exclude`.
    - `proxy` (optional): proxy address, such as `127.0.0.1:1080`.

=== "Kubernetes"

    In Kubernetes, you can configure the database sink through environment variables, see [here](datakit-daemonset-deploy.md#env-sinker).

- Step 3: [restart DataKit](datakit-service-how-to.md#manage-service)

## Setup Phase Settings {#dw-setup}

Dataway Sink supports setting through environment variables during installation:

```shell
DK_SINK_M="dataway://?url=https://openway.guance.com&token=<YOUR-TOKEN>&filters={host='user-ubuntu'}&filters={cpu='cpu-total'}" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

### Sink Multi-instance Configuration {#multi-dw-sink}

For a single data type, if you want to specify multiple data sink configurations, you can split them by `||`:

```shell
DK_SINK_M="dataway://?url=https://openway.guance.com&token=<TOKEN-1>&filters={host='user-ubuntu'}||dataway://?url=https://openway.guance.com&token=<TOKEN-2>&filters={host='user-centos'}" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

What this means is:

> All time series data (`M`), call the space with token `TOKEN-1` if the hostname（`host`) is `user-ubuntu` , and call the workspace with `TOKEN-2` if the hostname is `user-centos`.

By analogy, other data types (such as L/O/T/...) can be set accordingly.

## Extend Readings {#more-readings}

- [Filter](datakit-filter.md#howto)
