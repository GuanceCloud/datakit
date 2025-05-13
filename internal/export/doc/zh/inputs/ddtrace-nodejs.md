---
title     : 'DDTrace NodeJS'
summary   : 'DDTrace NodeJS 集成'
tags      :
  - 'DDTRACE'
  - 'NODEJS'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---


## 安装依赖 {#dependence}

安装 DDTrace 的 NodeJS 扩展，完整的 APM 接入步骤，参见 [Datadog NodeJS 接入文档](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/nodejs/){:target="_blank"}。

## NodeJS v12+ {#node-12}

```shell
npm install dd-trace --save
```

## NodeJS v10 v8 {#node-10-8}

```shell
npm install dd-trace@latest-node10
```

> 注意：你需要在任何 NodeJS 代码或载入任何 Module 前 import 并初始化 DDTrace lib，如果 DDTrace lib 没有被适当的初始化可能无法接收检测数据。

## 示例 {#example}

在只单纯运行 JavaScript 的环境下：

```nodejs
// This line must come before importing any instrumented module.
const tracer = require("dd-trace").init();
```

对于使用了 TypeScript 和 bundlers 并支持 ECMAScript Module 语法的环境需要在不同的文件中初始化 DDTrace：

```nodejs
//
// server.ts
//
import "./tracer"; // must come before importing any instrumented module.
```

```typescript
//
// tracer.ts
//
import tracer from "dd-trace";
tracer.init(); // initialized in a different file to avoid hoisting.
export default tracer;
```

另外如果默认配置足够有效或者通过环境变量以成功配置了 DDTrace 可以直接在代码中引入 module：

```typescript
import "dd-trace/init";
```

## 运行 {#run}

运行 Node Code

```shell
DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 node server
```

## 环境变量支持 {#envs}

下面列举了常见的 ENV 支持，完整的 ENV 支持列表，参见 [Datadog 文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/nodejs/){:target="_blank"}

- **DD_ENV**

    为服务设置环境变量

- **DD_VERSION**

    APP 版本号

- **DD_SERVICE**

    用于设置应用程序的服务名称，默认使用 *package.json* 中的 `name` 字段

- **DD_SERVICE_MAPPING**

    定义服务名映射用于在 Tracing 里重命名服务。

- **DD_TAGS**

    为每个 Span 添加默认 Tags

- **DD_TRACE_AGENT_HOSTNAME**

    DataKit 监听的地址名，默认 localhost

- **DD_TRACE_AGENT_PORT**

    DataKit 监听的端口号，默认 9529

- **DD_TRACE_SAMPLE_RATE**

    设置采样率从 0.0(0%) ~ 1.0(100%)
