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

先确认 Node.js 运行时与 `dd-trace` 主版本兼容，再安装 SDK。新项目建议使用仍受维护的 Node.js 版本；SDK 与运行时不兼容时，应用可能在启动阶段失败。完整兼容性和接入方式见 [Datadog Node.js 接入文档](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/nodejs/){:target="_blank"}。

## 当前受支持的 Node.js 版本 {#node-12}

```shell
npm install dd-trace --save
```

当前主线 `dd-trace` 适用于现代 Node.js 运行时；若使用容器，请将 SDK 主版本、Node.js 主版本和镜像中的版本一起锁定并在发布前验证。

## Node.js 10 / 8（仅遗留维护） {#node-10-8}

```shell
npm install dd-trace@latest-node10
```

Node.js 10 和 8 均已停止维护。仅在无法升级的遗留系统中使用该分支，并先在预发布环境完成兼容性和安全评估。

> 必须在加载任何待自动插桩的模块之前加载并初始化 DDTrace。初始化太晚不会补录已经加载模块的调用，也不会产生相应的 trace。

## 示例 {#example}

在 CommonJS 应用中，将初始化放在入口文件的第一行：

```nodejs
// This line must come before importing any instrumented module.
const tracer = require("dd-trace").init();
```

使用 TypeScript、bundler 或 ECMAScript Module 时，将初始化放在独立文件中，并确保它是应用的第一个 import：

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

如果全部配置均通过环境变量提供，也可以使用预加载模块：

```typescript
import "dd-trace/init";
```

## 运行 {#run}

以下示例将 trace 发送到本机的 DataKit。跨主机或 Kubernetes 部署时，把主机名替换为 DataKit Service/DNS 名称，并确保 DataKit 的 HTTP 服务允许远程连接：

```shell
DD_SERVICE=my-node-service \
DD_ENV=production \
DD_VERSION=1.0.0 \
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
node server.js
```

应用启动后访问一个已插桩的路由，再在 DataKit 的 monitor 中确认 `/v0.4/traces`（或 SDK 使用的其他兼容端点）有请求。排障时可临时设置 `DD_TRACE_DEBUG=true`；确认后立即关闭。

## 环境变量支持 {#envs}

下列变量均应在 Node.js 进程启动前设置。完整列表和 SDK 版本差异见 [Datadog 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/nodejs/){:target="_blank"}。

- **DD_ENV**

    设置服务运行环境，例如 `production`、`staging`。

- **DD_VERSION**

    设置应用版本。

- **DD_SERVICE**

    设置服务名称；未设置时通常使用 *package.json* 的 `name`。生产环境建议显式设置。

- **DD_SERVICE_MAPPING**

    定义依赖服务名映射，例如 `postgres:orders-db`。它不会改变本服务的 `DD_SERVICE`。

- **DD_TAGS**

    为每个 span 添加默认标签，格式为 `key:value,key:value`。避免写入用户标识、令牌或请求内容。

- **DD_AGENT_HOST**

    DataKit 的主机名或 IP 地址；默认通常是 `localhost`。若设置 `DD_TRACE_AGENT_URL`，URL 配置优先。

- **DD_TRACE_AGENT_PORT**

    Trace 接收端端口。上游默认通常是 `8126`；接入 DataKit 时显式设为 `9529`。

- **DD_TRACE_SAMPLE_RATE**

    设置 SDK 侧采样率，范围为 `0.0`（0%）到 `1.0`（100%）。它与 DataKit 接收端采样独立。

- **DD_TRACE_ENABLED**

    控制自动插桩和 trace 生成。排障时确认它没有被设置为 `false`。
