# DataKit Sink 使用文档

## DataKit Sinker

本文将讲述什么是 DataKit 的 Sinker 模块(以下简称 Sinker 模块、Sinker)、以及如何使用 Sinker 模块。

## 什么是 Sinker

Sinker 是 DataKit 中数据存储定义模块。默认情况下，DataKit 采集到的数据是上报给[观测云](https://console.guance.com/)，但通过配置不同的 Sinker 配置，我们可以将数据发送给不同的自定义存储。

### 目前支持的 Sinker 实例

- [InfluxDB](datakit-sink-influxdb)：目前支持将 DataKit 采集的时序数据（M）发送到本地的 InfluxDB 存储
- [M3DB](datakit-sink-m3db)：目前支持将 DataKit 采集的时序数据（M）发送到本地的 InfluxDB 存储（同 InfluxDB）
- [Logstash](datakit-sink-logstash)：目前支持将 DataKit 采集的日志数据（L）发送到本地 Logstash 服务

当让，同一定的开发，也能将现有 DataKit 采集到的各种其它数据发送到任何其它存储，参见[Sinker 开发文档](datakit-sink-dev)。

## Sinker 的配置

只需要以下简单三步:

- 搭建后端存储，目前支持 [InfluxDB](datakit-sink-influxdb)、[Logstash](datakit-sink-logstash) 以及 [M3DB](datakit-sink-m3db)

- 增加 Sinker 配置：在 `datakit.conf` 配置中增加 Sinker 实例的相关参数，也能在 DataKit 安装阶段即指定 Sinker 配置。具体参见各个已有 Sinker 的安装文档。

  - [InfluxDB 安装](datakit-sink-influxdb#dc8b9023)
  - [M3DB 安装](datakit-sink-m3db#3ab48619)
  - [Logstash 安装](datakit-sink-logstash#dc8b9023)

- 重启 DataKit

```shell
$ sudo datakit --restart
```

## 通用参数的说明

无论哪种 Sinker 实例, 都必须支持以下参数:

- `target`: Sinker 实例目标, 即要写入的存储是什么，如 `influxdb`
- `categories`: 汇报数据的类型。如 `["M", "N", "K", "O", "CO", "L", "T", "R", "S"]`

`categories` 中各字符串对应的上报指标集如下:

| `categories` 字符串 | 对应数据类型 |
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

> 注：对于未指定 Sinker 的 categories，默认仍然发送给观测云。

## 扩展阅读

- [Sinker 之 InfluxDB](datakit-sink-influxdb)
- [Sinker 之 M3DB](datakit-sink-m3db)
- [Sinker 之 Logstash](datakit-sink-logstash)
