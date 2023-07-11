# Pipeline 手册

## 简介

自 DataKit v1.4.0 起，可通过内置的 Pipeline 的数据处理器功能对 DataKit 采集或接收到数据进行修改、提取、过滤、聚合等操作，支持目前所有的数据类别（如 Logging、Metric、Tracing、Network 和 Object 等）。

DataKit Pipeline 可编程数据处理器支持观测云开发定制的领域特定语言 Platypus 提供的运行时。后续将为 Pipeline 添加更多的编程语言和运行时。


## 目录

1. [快速开始](pipeline-quick-start.md)
2. [基础和原理](pipeline-architecture.md)
3. [Platypus 语法](pipeline-platypus-grammar.md)
4. [内置函数](pipeline-built-in-function.md)
5. 附加功能
6. [性能基准和优化](pipeline-benchmark.md)
