---
icon: zy/datakit
---

# DataKit
---


## 概述

DataKit 是一款开源、一体式的数据采集 Agent，它提供全平台操作系统（Linux/Windows/macOS）支持，拥有全面数据采集能力，涵盖主机、容器、中间件、Tracing、日志以及安全巡检等各种场景。

## 主要功能

## 主要功能点

- 支持主机、中间件、日志、APM 等领域的指标、日志以及 Tracing 几大类数据采集
- 完整支持 Kubernetes 云原生生态
- [Pipeline](pipeline.md)：简便的结构化数据提取
- 支持接入其它第三方数据采集
	- [Telegraf](telegraf.md)
	- [Prometheus](prom.md)
	- [Statsd](statsd.md)
	- [Fluentd](logstreaming.md)
	- [Filebeats](beats_output.md)
	- [Function](../dataflux-func/write-data-via-datakit.md)
	- Tracing 相关（[OpenTelemetry](opentelemetry.md)/[DDTrace](ddtrace.md)/[Zipkin](zipkin.md)/[Jaeger](jaeger.md)/[Skywalking](skywalking.md)）
