
# 选举模块设计

选举模块主要用来控制采集器在集群部署模式下的采集行为,目前是通过调用中心的选举接口来实现.

## Prometheus Metrics

选举模块暴露如下 metrics：

| 指标                          | 类型  | 说明                                                                                                 | labels                         |
| ---                           | ---   | ---                                                                                                  | ---                            |
| datakit_election_pause_total  | count | Input paused count when election failed                                                              | id,namespace                   |
| datakit_election_resume_total | count | Input resume count when election OK                                                                  | id,namespace                   |
| datakit_election_status       | gauge | Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second) | elected_id,id,namespace,status |
| datakit_election_inputs       | gauge | Datakit election input count                                                                         | namespace                      |
| datakit_election              | gauge | Election latency(in millisecond)                                                                     | namespace,status               |
