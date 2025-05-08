---
title     : 'Profiling NodeJS'
summary   : 'NodeJS Profiling 集成'
tags:
  - 'NODEJS'
  - 'PROFILE'
__int_icon: 'icon/profiling'
---

[:octicons-tag-24: Version-1.9.0](../datakit/changelog.md#cl-1.9.0)

目前 DataKit 支持 1 种方式来采集 NodeJS profiling 数据，即 [Pyroscope](https://pyroscope.io/){:target="_blank"}。

## Pyroscope {#pyroscope}

[Pyroscope](https://pyroscope.io/){:target="_blank"} 是一款开源的持续 profiling 平台，DataKit 已经支持将其上报的 profiling 数据展示在[<<<custom_key.brand_name>>>](https://www.<<<custom_key.brand_main_domain>>>/){:target="_blank"}。

Pyroscope 采用 C/S 架构，运行模式分为 [Pyroscope Agent](https://pyroscope.io/docs/agent-overview/){:target="_blank"} 和 [Pyroscope Server](https://pyroscope.io/docs/server-overview/){:target="_blank"}，这两个模式均集成在一个二进制文件中，通过不同的命令行命令来展现。

这里需要的是 Pyroscope Agent 模式。DataKit 已经集成了 Pyroscope Server 功能，通过对外暴露 HTTP 接口的方式，可以接收 Pyroscope Agent 上报的 profiling 数据。

Profiling 数据流向：

```mermaid
flowchart LR
subgraph App
app(App Process)
pyro(Pyroscope Agent)
end
dk(Datakit)
brand_name("<<<custom_key.brand_name>>>")

app --> pyro --> |profiling data|dk --> brand_name
```

在这里，你的 NodeJS 程序就相当于是一个 Pyroscope Agent。

### 前置条件 {#pyroscope-requirement}

- 根据 Pyroscope 官方文档 [NodeJS](https://pyroscope.io/docs/nodejs/){:target="_blank"}, 支持以下平台：

|  Linux   | macOS  | Windows  | Docker  |
|  ----  | ----  | ----  | ----  |
| :white_check_mark:  | :white_check_mark: | :x: | :white_check_mark: |

- Profiling NodeJS 程序

Profiling NodeJS 程序需要把 [npm](https://www.npmjs.com/){:target="_blank"} 模块引入到程序中：

```sh
npm install @pyroscope/nodejs

# or
yarn add @pyroscope/nodejs
```

把以下代码加入到你的 NodeJS 程序代码中：

```js
const Pyroscope = require('@pyroscope/nodejs');

Pyroscope.init({
  serverAddress: 'http://pyroscope:4040',
  appName: 'myNodeService',
  tags: {
    region: 'cn'
  }
});

Pyroscope.start()
```

- 已安装 [DataKit](https://www.<<<custom_key.brand_main_domain>>>/){:target="_blank"} 并且已开启 [profile](profile.md#config) 采集器，配置参考如下：

```toml
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]

  #  config
  [[inputs.profile.pyroscope]]
    # listen url
    url = "0.0.0.0:4040"

    # service name
    service = "pyroscope-demo"

    # app env
    env = "dev"

    # app version
    version = "0.0.0"

  [inputs.profile.pyroscope.tags]
    tag1 = "val1"
```

启动 DataKit, 然后启动你的 NodeJS 程序。

## 查看 Profile {#pyroscope-view}

执行上述操作后，你的 NodeJS 程序会开始采集 profiling 数据并将数据报给给 Datakit，Datakit 会将这些数据上报给<<<custom_key.brand_name>>>。稍等几分钟后就可以在<<<custom_key.brand_name>>>空间[应用性能监测 -> Profile](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} 查看相应数据。

## Pull 模式 (可选) {#pyroscope-pull}

集成 NodeJS 程序的方式也支持 Pull 模式。你必须确保你的 NodeJS 程序有 profiling 路由 (`/debug/pprof/profile` 和 `/debug/pprof/heap`) 且是启用状态。为此你可以使用 `expressMiddleware` 模块或者自己创建路由接入点：

```js
const Pyroscope, { expressMiddleware } = require('@pyroscope/nodejs');

Pyroscope.init({...})

app.use(expressMiddleware())
```

>注意：你不必再使用 `.start()` 但必须使用 `.init()`。

## FAQ {#pyroscope-faq}

### 如何排查 Pyroscope 问题 {#pyroscope-troubleshooting}

可以设置环境变量 `DEBUG` 到 `pyroscope`, 然后查看调试信息：

```sh
DEBUG=pyroscope node index.js
```
