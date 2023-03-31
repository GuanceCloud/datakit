# Dataway
---

If you want to upload data into a different workspace, you can use the Dataway Sinker:

1. In a Datakit, you can configure multiple dataway sinker addresses. There are filters(conditions) to applied on different categories of data.
1. For the data that meets the filters(checking on data's tags and fields), upload them to corresponding workspace(dataway URL with the workspace's token).
1. If the data does not meet filters, the data will continue to be upload to the default workspace(dataway).

???+ attention

    Among multiple sinkers, their filters may have intersections(one data meet multiple sinker's filter), the data will be written into multiple sinker(and maybe multiple workspace), and cause data duplication.

## Supported Categories {#categories}

Dataway Sinker supports all categories(M/O/CO/L/T/N/R/P/S/E).

## Supported Categories {#categories}

Dataway Sinker support [all categories](apis.md#category) in Datakit.

## Sink Configuration {#config}

=== "datakit.conf"

    In *datakit.conf*, add following parts under key `dataway`(See [sample here](datakit-conf.md#maincfg-example))
    
    ```toml
    [[dataway.sinkers]]
      categories = [ "L/M/O/..." ]
      filters = [
        "{ cpu = 'cpu-total' }",
        "{ source = 'some-logging-source'}",
      ]
      url = "https//openway.guance.com?token=<YOUR-TOKEN>"
    
    [[dataway.sinkers]]
      another sinker...
      ...
    ```
    
    Sinker Dataway support following configures: 
    
    - `url` (required): Dataway address(with token)
    - `filters` (optional): Filter rules. See [here](datakit-filter.md)
    - `proxy` (optional): sinker proxy address, such as `127.0.0.1:1080`.

=== "Kubernetes"

    In Kubernetes, you can configure the database sink through environment variables, see [here](datakit-daemonset-deploy.md#env-sinker).

- Step 3: [restart DataKit](datakit-service-how-to.md#manage-service)

???+ attention

    If Datakit upload Dataway failed, we can setup [disk cache](datakit-conf.md#io-disk-cache) to hold these failed data points, but for dataway sinker, disk cache not support for now. If upload to the sinker failed, these data points dropped.

## Extend Readings {#more-readings}

- [Filter](datakit-filter.md#howto)
