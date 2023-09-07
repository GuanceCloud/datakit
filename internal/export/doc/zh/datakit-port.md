# Datakit 占用端口

在 Datakit 运行过程中，根据不同的功能，需要开启不同的本地端口。目前占
用的端口情况如下：

| 默认端口(可能多个) | 协议(L4/L7) | 功能名                        | 默认路由或者 domain socket(可能多个)      |
| ---                | ---         | ---                          | ---                                      |
| 2055               | UDP         | NetFlow netflow9 默认端口     | N/A                                       |
| 2056               | UDP         | NetFlow netflow5 默认端口     | N/A                                       |
| 2280               | TCP         | Cat Trace 数据接入            | N/A                                       |
| 4040               | HTTP        | Pyroscope Profile 数据接入    | `/ingest`                                 |
| 4317               | gRPC        | OpenTelemetry 数据接入        | `otel/v1/trace`,`otel/v1/metric`          |
| 4739               | UDP         | NetFlow ipfix 默认端口        | N/A                                       |
| 5044               | TCP         | Beats 数据接入                | N/A                                       |
| 6343               | UDP         | NetFlow sflow5 默认端口       | N/A                                       |
| 8125               | UDP         | StatsD 数据接入               | N/A                                       |
| 9529               | HTTP        | Datakit HTTP 服务            |                                           |
| 9530               | TCP         | Socket（TCP）日志接入          | N/A                                       |
| 9531               | TCP         | DCA Server                   | N/A                                       |
| 9531               | UDP         | Socket（UDP）日志接入          | N/A                                       |
| 9533               | WebSocket   | SideCar logfwdserver 数据接入 | N/A                                       |
| 9542               | HTTP        | 远程升级                      | `/v1/datakit/version,/v1/datakit/upgrade` |
