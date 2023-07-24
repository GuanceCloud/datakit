# Datakit Ports

Datakit need to open several local ports to serve it's features, We may open these ports when you enabled related features:

| Port(default) | Protocol(L4/L7) | Related Feature        | Default route                             |
| ---           | ---             | ---                    | ---                                       |
| 2280          | TCP             | Cat Tracing            | N/A                                       |
| 4040          | HTTP            | Pyroscope Profile      | `/ingest`                                 |
| 4317          | gRPC            | OpenTelemetry          | `otel/v1/trace`,`otel/v1/metric`          |
| 5044          | TCP             | Beats                  | N/A                                       |
| 8125          | UDP             | StatsD                 | N/A                                       |
| 9529          | HTTP            | Datakit HTTP           |                                           |
| 9530          | TCP             | Logging on socket(TCP) | N/A                                       |
| 9531          | TCP             | DCA Server             | N/A                                       |
| 9531          | UDP             | Logging on Socket(UDP) | N/A                                       |
| 9533          | WebSocket       | SideCar logfwdserver   | N/A                                       |
| 9542          | HTTP            | Remote upgrading       | `/v1/datakit/version,/v1/datakit/upgrade` |
