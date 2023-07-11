
# NodeJS Example
---

## Install Dependency {#dependence}

Install the NodeJS extension for ddtrace.

**NodeJS v12+**

```shell
npm install dd-trace --save
```

**NodeJS v10 v8**

```shell
npm install dd-trace@latest-node10
```

**Note:** You need to import and initialize the ddtracer lib before any NodeJS code or loading any Module. If the ddtracer lib is not properly initialized, it may not receive the detection data.

## Example {#example}

In an environment that simply runs JavaScript:

```js
// This line must come before importing any instrumented module.
const tracer = require("dd-trace").init();
```

For environments that use TypeScript and bundlers and support EcmaScript Module syntax, you need to initialize ddtracer in a different file:

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

In addition, if the default configuration is valid enough or ddtracer is successfully configured through environment variables, you can directly introduce module into your code:

```js
import "dd-trace/init";
```

## Run {#run}

Run Node Code

```shell
DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 node server
```

## Environment Variable Support {#envs}

- DD_ENV: Set environment variables for the service.
- DD_VERSION: APP version number.
- DD_SERVICE: Used to set the service name of the application, using the name field in package. json by default.
- DD_SERVICE_MAPPING: Define service name mappings for renaming services in Tracing.
- DD_TAGS: Add default Tags for each Span.
- DD_TRACE_AGENT_HOSTNAME: The name of the address where Datakit listens, default localhost.
- DD_TRACE_AGENT_PORT: The port number on which Datakit listens, the default is 9529.
- DD_TRACE_SAMPLE_RATE: Set the sampling rate from 0.0 (0%) to 1.0 (100%).
