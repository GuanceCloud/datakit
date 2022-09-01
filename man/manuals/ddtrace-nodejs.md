{{.CSS}}
# NodeJS 示例
---

## 安装依赖 {#dependence}

安装 ddtrace 的 NodeJS 扩展

**NodeJS v12+**

```shell
npm install dd-trace --save
```

**NodeJS v10 v8**

```shell
npm install dd-trace@latest-node10
```

**Note:** 你需要在任何 NodeJS 代码或载入任何 Module 前 import 并 initialize ddtracer lib，如果 ddtrace lib 没有被适当的初始化可能无法接收检测数据。

## 示例 {#example}

在只单纯运行 JavaScript 的环境下：

```js
// This line must come before importing any instrumented module.
const tracer = require("dd-trace").init();
```

对于使用了 TypeScript 和 bundlers 并支持 EcmaScript Module 语法的环境需要在不同的文件中初始化 ddtracer：

**server.ts**

```ts
import "./tracer"; // must come before importing any instrumented module.
```

**tracer.ts**

```ts
import tracer from "dd-trace";
tracer.init(); // initialized in a different file to avoid hoisting.
export default tracer;
```

另外如果默认配置足够有效或者通过环境变量以成功配置了 ddtracer 可以直接在代码中引入 module：

```js
import "dd-trace/init";
```

## 运行 {#run}

运行 Node Code

```shell
DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 node server
```

## 环境变量支持 {#envs}

- DD_ENV: 为服务设置环境变量。
- DD_VERSION: APP 版本号。
- DD_SERVICE: 用于设置应用程序的服务名称，默认使用 package.json 中的 name 字段。
- DD_SERVICE_MAPPING: 定义服务名映射用于在 Tracing 里重命名服务。
- DD_TAGS: 为每个 Span 添加默认 Tags。
- DD_TRACE_AGENT_HOSTNAME: Datakit 监听的地址名，默认 localhost。
- DD_AGENT_PORT: Datakit 监听的端口号，默认 9529。
- DD_TRACE_SAMPLE_RATE: 设置采样率从 0.0(0%) ~ 1.0(100%)。
