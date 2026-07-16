---
title     : 'DDTrace NodeJS'
summary   : 'Tracing NodeJS applications with DDTrace'
tags      :
  - 'DDTRACE'
  - 'NODEJS'
  - 'APM'
  - 'TRACING'
__int_icon: 'icon/ddtrace'
---


## Install Dependencies {#dependence}

Confirm that the Node.js runtime is compatible with the `dd-trace` major version before installing the SDK. Use a maintained Node.js release for new applications; an incompatible SDK can fail during startup. See the [Datadog Node.js setup guide](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/nodejs/){:target="_blank"} for compatibility and complete setup.

## Currently Supported Node.js Versions {#node-12}

```shell
npm install dd-trace --save
```

The current `dd-trace` line targets modern Node.js runtimes. For containers, pin and validate the SDK major version, the Node.js major version, and the image together before release.

## Node.js 10 / 8 (Legacy Maintenance Only) {#node-10-8}

```shell
npm install dd-trace@latest-node10
```

Node.js 10 and 8 are end of life. Use this branch only for applications that cannot yet be upgraded, and validate compatibility and security in a pre-production environment.

> Load and initialize DDTrace before any module that should be auto-instrumented. Initialization cannot retroactively instrument modules that have already been loaded, so their calls will not produce the expected traces.

## Example {#example}

In a CommonJS application, put initialization on the first line of the entry file:

```nodejs
// This line must come before importing any instrumented module.
const tracer = require("dd-trace").init();
```

For TypeScript, bundlers, or ECMAScript Modules, use a dedicated initialization file and make it the first import of the application:

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

If all configuration is provided through environment variables, you can preload the module instead:

```typescript
import "dd-trace/init";
```

## Run {#run}

This example sends traces to a local DataKit. For another host or Kubernetes, replace the host with the DataKit Service/DNS name and make sure DataKit's HTTP service accepts remote connections:

```shell
DD_SERVICE=my-node-service \
DD_ENV=production \
DD_VERSION=1.0.0 \
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
node server.js
```

After startup, call an instrumented route and confirm requests to `/v0.4/traces` (or another SDK-compatible endpoint) in the DataKit monitor. For troubleshooting, temporarily set `DD_TRACE_DEBUG=true` and disable it after verification.

## Environment Variable Support {#envs}

Set these variables before the Node.js process starts. For the complete list and version-specific behavior, see the [Datadog configuration guide](https://docs.datadoghq.com/tracing/trace_collection/library_config/nodejs/){:target="_blank"}.

- **DD_ENV**

    Sets the deployment environment, for example `production` or `staging`.

- **DD_VERSION**

    Sets the application version.

- **DD_SERVICE**

    Sets the service name. It usually falls back to `name` in *package.json*, but production deployments should set it explicitly.

- **DD_SERVICE_MAPPING**

    Defines dependency-service mappings, for example `postgres:orders-db`. It does not change this service's `DD_SERVICE`.

- **DD_TAGS**

    Adds default tags to each span in `key:value,key:value` form. Do not include user identifiers, tokens, or request contents.

- **DD_AGENT_HOST**

    The DataKit host name or IP address. It normally defaults to `localhost`; `DD_TRACE_AGENT_URL`, when set, takes precedence.

- **DD_TRACE_AGENT_PORT**

    The trace receiver port. The common upstream default is `8126`; explicitly set `9529` for DataKit.

- **DD_TRACE_SAMPLE_RATE**

    Sets the SDK-side sampling rate from `0.0` (0%) to `1.0` (100%). It is independent of DataKit receiver-side sampling.

- **DD_TRACE_ENABLED**

    Controls automatic instrumentation and trace generation. During troubleshooting, make sure it is not set to `false`.
