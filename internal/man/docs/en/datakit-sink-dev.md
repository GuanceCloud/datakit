# DataKit Sinker Development
---

This article describes how to develop a new instance of DataKit's Sinker module (hereinafter referred to as Sinker Module, Sinker). Suitable for people who want to develop new instances of Sinker, or who want to learn more about the principles of Sinker modules.

## How to Develop a Sinker Instance {#dev-sink}

At present, the community version only implements a limited number of Sinkers. If you want to support other storage, you can do corresponding development (in Go language), which can be roughly divided into the following steps (taking `influxdb` as an example):

- Clone [DataKit code](https://jihulab.com/guance-cloud/datakit){:target="_blank"}, and create a new package under *io/sink/* called `sinkinfluxdb` (it is recommended to start with `sink`).

- Create a new source file `sink_influxdb.go`, under the above package, and create a new constant `creatorID`, which cannot have the same name as `creatorID` in other packages; Implement the `interface` of `ISink` by implementing the following functions:

    - `GetInfo() *SinkInfo`: Return information about the sink instance. There are currently `ID` (instance internal identity, generated internally from the program for internal use, unique in configuration), `CreateID` (instance creation identity, unique in code), and shorthand for supported types (for example, `Metrics` returns `M`).
    - `LoadConfig(mConf map[string]interface{}) error`: Load external configuration to internal.
    - `Write(category string, pts []ISinkPoint) error`: Write data.

The general code is as follows:

```golang
const creatorID = "influxdb"

type SinkInfluxDB struct {
  // Here, write connection, write and other internal operations need to use some parameters, such as saving the parameters used in the connection.
  ...
}

func (s *SinkInfluxDB) GetInfo() *SinkInfo {
  // Return information about the sink instance
  ...
}

func (s *SinkInfluxDB) LoadConfig(mConf map[string]interface{}) error {
  // Load external configuration into internal configuration
  ...
}

func (s *SinkInfluxDB) Write(category string, pts []ISinkPoint) error {
  // Write data
  // Here you may want to familiarize yourself with the ISinkPoint interface, which has two methods, ToPoint and ToJSON, for use.
  //   ToPoint returns the point of influxdb;
  //   ToJSON returns the structure, which you can use if you don't want to use infuxdb.
  ...
}
```

Finally, the new sink is introduced in `io/sink/sink.go`:

```go
package sink

import (
  ...
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkinfluxdb"
)
```

> Generally speaking, you can refer to the code implementation of `influxdb`, which is still very simple. Everything is simple as the first design principle, and you don't want to maintain it if it is complicated. Welcome to submit the code to github community and maintain it together.

- Step 3: Add configuration to `datakit.conf` and `target` with custom instance name, that is, `creatorID` is unique. For example:

```conf
...
[sinks]

  [[sinks.sink]]
    categories = ["M", "N", "K", "O", "CO", "L", "T", "R", "S"]
    target = "influxdb"
    host = "10.200.7.21:8086"
    protocol = "http"
    database = "db0"
    precision = "ns"
    timeout = "15s"
...
```

## Notice {#notice}

1. The new instance needs to customize a `createID`, that is, the "identity" of the instance, such as `influxdb`, `elasticsearch` , etc. This cannot be duplicated with the existing `createID`. The `target` in the configuration corresponds to this `createID`.
