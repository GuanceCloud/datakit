# NodeJS Profiling

[:octicons-tag-24: Version-1.9.0](../datakit/changelog.md#cl-1.9.0)

At present, DataKit supports one way to collect NodeJS profiling data, namely [Pyroscope](https://pyroscope.io/){:target="_blank"}.

## Pyroscope {#pyroscope}

[Pyroscope](https://pyroscope.io/){:target="_blank"} is an open source continuous profiling platform, and DataKit already supports displaying its reported profiling data in [Guance](https://www.guance.com/){:target="_blank"}。

Pyroscope uses C/S architecture, and its running modes are divided into [Pyroscope Agent](https://pyroscope.io/docs/agent-overview/){:target="_blank"} and [Pyroscope Server](https://pyroscope.io/docs/server-overview/){:target="_blank"}, which are integrated in a binary file and displayed by different command line commands.

What you need here is the Pyroscope Agent pattern. DataKit has integrated the Pyroscope Server functionality and can receive profiling data reported by the Pyroscope Agent by exposing the HTTP interface to the outside world.

profiling data flow: `Pyroscope Agent collects profiling data -> Datakit -> Guance`。

Here, your NodeJS application could be treated as a Pyroscope Agent.

### Preconditions {#pyroscope-requirement}

- According to Pyroscope official document [NodeJS](https://pyroscope.io/docs/nodejs/){:target="_blank"}  Pyroscope supported following platforms:

|  Linux   | macOS  | Windows  | Docker  |
|  ----  | ----  | ----  | ----  |
| :white_check_mark:  | :white_check_mark: | :x: | :white_check_mark: |

- Profiling NodeJS application

To start profiling a NodeJS application, you need to include the npm module in your app:

```sh
npm install @pyroscope/nodejs

# or
yarn add @pyroscope/nodejs
```

Then add the following code to your application:

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

- [DataKit](https://www.guance.com/){:target="_blank"} is installed and the [profile](profile.md#config) collector is turned on with the following configuration references:

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

Restart Datakit and your NodeJS application.

## View Profile {#pyroscope-view}

After running the above profiling command, your NodeJS application starts collecting the specified profiling data and reports the data to Datakit, the Datakit would turns these data to Guance. After a few minutes, you can view the corresponding data in Guance hosting [application performance monitoring -> Profile](https://console.guance.com/tracing/profile){:target="_blank"}.

## Pull Mode (Optional) {#pyroscope-pull}

NodeJS integration also supports pull mode. For that to work you will need to make sure you have profiling routes (`/debug/pprof/profile` and `/debug/pprof/heap`) enabled in your http server. For that you may use our `expressMiddleware` or create endpoints yourself

```js
const Pyroscope, { expressMiddleware } = require('@pyroscope/nodejs');

Pyroscope.init({...})

app.use(expressMiddleware())
```

>Note: you don't need to `.start()` but you'll need to `.init()`

## FAQ {#pyroscope-faq}

### Pyroscope troubleshooting {#pyroscope-troubleshooting}

You may set `DEBUG` env to `pyroscope` and see debugging information which can help you understand if everything is OK.

```sh
DEBUG=pyroscope node index.js
```
