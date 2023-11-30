# Datakit 资源导出

随着 Datakit 导出的内容越来越多，Datakit 自身的文档架构做了一些调整，主要包含如下几个方面：

- 文档目录结构
- 文档内容组成

## 目录结构

Datakit 文档目录结构遵循如下约定：

- 中英文文档分别放到对应目录下（*doc/{zh,en}*）

- 在中英目录下，分别有 *inputs/*、*pipeline/* 两个目录以及一堆 *xxx.md* 文件

  - *inputs* 存放采集器文档，包括跟采集器关联的文档，以 ddtrace 为例，除了采集器文档，各种语言的示例文档，都应该归入 *inputs/* 目录下，这些文档会发布到[集成](https://docs.guance.com/integrations/integration-index/)中。
  - *pipeline* 存放 Pipeline 有关的文档，其文档会发布到[自定义开发](https://docs.guance.com/developers/pipeline/)中。
  - *xxx.md* 这些才是 Datakit 自身的文档，其不含任何采集器有关的文档，比如 Datakit 安装、工具链使用等文档。其文档会发布到 [Datakit 文档](https://docs.guance.com/datakit/) 中。 

由于 Datakit 的文档目前分出了 3 批导出，在 docs.guance.com 站点，它们分别分流到了不同的顶级目录，如果要引用的话，需加上对应的目录跳转，比如 *prom.md* 中如果要引用 *datakit-tools-how-to.md* 中的某一个章节，需这样写：

```markdown
这里引用一下[工具的使用](../datakit/datakit-tools-how-to.md#some-section)方式...
```

而 Pipeline 可能要跳跃两层，因为在自定义开发的文档中，它是放在 *developers/pipeline/* 下，所以，如果要引用上面 Datakit 的文档，需这样写：

```markdown
这里引用一下[工具的使用](../../datakit/datakit-tools-how-to.md#some-section)方式...
```

即，先需要跳出 *developers/pipeline/* 两层目录。

### 采集器关联的资源导出

与采集器关联的资源有两种，即 dashboard 和 monitor，所有采集器的相关模板位于这两个目录下的以**采集器名称**命名的子目录下。例如，文件 `dashboard/dk/dk.json`，表示 dk 采集器的 dashboard 模板文件。

如果模板文件命名后缀包含 `__zh` 和 `__en`，则分别表示该文件为中文和英文模板，无需进行翻译。反之，如果文件后缀不包含这两个字符串，则该文件模板将分别翻译为中文和英文。

若文件模板需要翻译，相关的采集器需要实现 Dashboard 和 Monitor 两个接口，并返回相应的中英文翻译数据。

## 文档内容结构

### 采集器文档结构

Datakit 采集器内容导出，以文档为中心，同时，文档关联了采集器对应的 dashboard 和 monitor，其在文档头部标记，以 CPU 采集器为例：

```markdown
---
title     : 'CPU'
summary   : '采集 CPU 指标数据'
__int_icon: 'icon/cpu'
dashboard :
  - desc  : 'CPU'
    path  : 'dashboard/zh/cpu'
monitor   :
  - desc  : '主机检测库'
    path  : 'monitor/zh/host'
---
```

字段解释：

- `title`：文档标题
- `summary`: 一句话总结文档的内容
- `__int_icon`：集成文档显示时所用的图标目录，参见[这个目录](https://gitee.com/dataflux/dataflux-template/tree/dev/icon)
- `dashboard`：采集器对应的 dashboard 信息，参见[这个目录](https://gitee.com/dataflux/dataflux-template/tree/dev/dashboard)
- `monitor`：采集器对应的 monitor 信息，参见[这个目录](https://gitee.com/dataflux/dataflux-template/tree/dev/monitor)

原则上，每个采集器文档都应该有这些 meta 信息。

## 验证

运行 `make lint` 后，在 *dist/export* 目录下有最终的导出效果

``` shell
▾ dist/
  ▾ export/
    ▾ guance-doc/docs/  # 最终导出到 docs.guance.com 的文档（含 datakit 文档和采集器文档）
      ▸ en/
      ▸ zh/
    ▾ integration/      # 最终导出到集成的内容，含文档/dashboard/monitor
      ▸ dashboard/
      ▸ datakit/        # 这是 datakit 之前导出到 oss 的内容，比如 pipeline 有关的一些文档和示例，暂时无用
      ▸ integration/
      ▸ monitor/
```

Datakit 的资源分发示意图：

```mermaid
graph TD  %% set direction top down

%% define various node
dk[Datakit];
doc[doc];
dashboard[dashboard];
monitor[monitor];
other["其它资源（PL有关）"];
oss[(OSS)];
integration[dataflux-template];
guance_doc[docs.guance.com];

guance_doc_dk["docs.guance.com/datakit"];
guance_doc_integration["docs.guance.com/integrations"];
guance_doc_dev["docs.guance.com/developers/"];

guance_doc --> |xxx.md|guance_doc_dk;
guance_doc --> |inputs/*.md|guance_doc_integration;
guance_doc --> |pipeline/*.md|guance_doc_dev;

dk --> doc;
dk --> dashboard;
dk --> monitor;
dk --> other;
doc --> guance_doc;
doc --> integration;
dashboard --> integration;
monitor --> integration;
other --> integration;
other --> oss;
```
