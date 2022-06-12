# DataKit

## 概述

DataKit 是一款开源、一体式的数据采集 Agent，它提供全平台操作系统（Linux/Windows/macOS）支持，拥有全面数据采集能力，涵盖主机、容器、中间件、Tracing、日志以及安全巡检等各种场景。

## 主要功能

## 主要功能点

- 支持主机、中间件、日志、APM 等领域的指标、日志以及 Tracing 几大类数据采集
- 完整支持 Kubernetes 云原生生态
- [Pipeline](pipeline)：简便的结构化数据提取
- 支持接入其它第三方数据采集
	- [Telegraf](telegraf)
	- [Prometheus](prom)
	- [Statsd](statsd)
	- [Fluentd](logstreaming)
	- [Filebeats](beats_output)
	- [Function](https://www.yuque.com/dataflux/func/write-data-via-datakit)
	- Tracing 相关（[OpenTelemetry]()/[DDTrace]()/[Zipkin]()/[Jaeger]()/[Skywalking]()）

---

观测云是一款面向开发、运维、测试及业务团队的实时数据监测平台，能够统一满足云、云原生、应用及业务上的监测需求，快速实现系统可观测。**立即前往观测云，开启一站式可观测之旅：**[www.guance.com](https://www.guance.com)
![](img/logo_2.png)
