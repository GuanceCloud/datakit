# DataKit Sink Use
---

## DataKit Sinker {#intro}

This article describes what the Sinker module of DataKit is (hereinafter referred to as the Sinker module, Sinker) and how to use the Sinker module.

## What is Sinker {#what}

Sinker is the data store definition module in DataKit. By default, the data collected by DataKit is reported to [Guance Cloud](https://console.guance.com/){:target="_blank"}, but by configuring different Sinker configurations, we can send the data to different custom stores.

### Currently Supported Sinker Instances {#list}

- [InfluxDB](datakit-sink-influxdb.md): Currently, it supports sending time series data (M) collected by DataKit to local InfuxDB storage.
- [Logstash](datakit-sink-logstash.md): Currently, it supports sending log data (L) collected by DataKit to the local Logstash service.
- [M3DB](datakit-sink-m3db.md): Currently, it supports sending time series data (M) collected by DataKit to local M3DB storage (same as InfluxDB).
- [OpenTelemetry and Jaeger](datakit-sink-otel-jaeger.md): OpenTelemetry (OTEL) provides a variety of Export to send link data (T) to multiple acquisition terminals, such as Jaeger, otlp, zipkin, prometheus.
- [Dataway](datakit-sink-dataway.md): It currently supports sending all types of data collected by the DataKit to the Dataway store.

With a certain amount of development, various other data collected by existing DataKit can also be sent to any other store, as shown in [Sinker Development Documentation](datakit-sink-dev.md)。

## Configuration of Sinker {#config}

All you need is the following three simple steps:

- Build back-end storage, which currently supports [InfluxDB](datakit-sink-influxdb.md), [Logstash](datakit-sink-logstash.md)、[M3DB](datakit-sink-m3db.md), [OpenTelemetry and Jaeger](datakit-sink-otel-jaeger.md) and [Dataway](datakit-sink-dataway.md).

- Add Sinker configuration: Add the Sinker instance parameters to the `datakit.conf` configuration, or specify the Sinker configuration during the DataKit installation phase. See the installation documentation of each existing Sinker for details.

  - [InfluxDB installation](datakit-sink-influxdb.md)
  - [Logstash installation](datakit-sink-logstash.md)
  - [M3DB installation](datakit-sink-m3db.md)
  - [OpenTelemetry and Jaeger installation](datakit-sink-otel-jaeger.md)
  - [Dataway installation](datakit-sink-dataway.md)

- Restart DataKit

```shell
$ sudo datakit --restart
```

## Description of Generic Parameters {#args}

No matter what kind of Sinker instance, the following parameters must be supported:

- `target`: The Sinker instance target, that is, what storage is to be written to, such as `influxdb`
- `categories`: The type of data that is reported. Such as `["M", "N", "K", "O", "CO", "L", "T", "R", "S"]`

The reporting measurement corresponding to each string in `categories` is as follows:

| `categories` String | Corresponding Data Type |
| ----                | ----         |
| `M`                 | Metric       |
| `N`                 | Network      |
| `K`                 | KeyEvent     |
| `O`                 | Object       |
| `CO`                | CustomObject |
| `L`                 | Logging      |
| `T`                 | Tracing      |
| `R`                 | RUM          |
| `S`                 | Security     |

> Note: categories that do not specify Sinker are still sent to Guance Cloud by default.

## Extended Readings {#more-readings}

- [Sinker's InfluxDB](datakit-sink-influxdb.md)
- [Sinker's Logstash](datakit-sink-logstash.md)
- [Sinker's M3DB](datakit-sink-m3db.md)
- [Sinker's OpenTelemetry and Jaeger](datakit-sink-otel-jaeger.md)
- [Sinker's Dataway](datakit-sink-dataway.md)
