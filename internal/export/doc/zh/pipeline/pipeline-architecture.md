# Pipeline 基础和原理
---

以下为 DataKit Pipeline 的模块设计和工作原理，能帮你更加了解 Pipeline 功能，但您可以选择跳过以下内容直接开始使用。

## 数据在 DataKit 的流转 {#data-flow}

在 DataKit 各种采集器插件或 DataKit API 等在采集或接收到数据后，数据将在上传前将经由 Pipeline 功能进行数据操作。

DataKit Pipeline 包含一个可编程数据处理器(Pipeline)和一个可编程[数据过滤器](../../datakit/datakit-filter.md)(Filter)，数据处理器用于数据的加工、过滤等，而数据过滤器专注于数据过滤功能。

简化的 DataKit 中的数据流转如下图所示：

![data-flow](img/pipeline-data-flow.drawio.png)

## 数据处理器工作流程 {#data-processor}

Pipeline 数据处理器工作流程见数据流程图：

![data-processor](img/pipeline-data-processor.drawio.png)
