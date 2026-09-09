
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
| datakit_election_provider_info | gauge | Selected election provider for the process                                                          | provider,namespace             |
| datakit_election_last_success_timestamp_seconds | gauge | Last successful leader response time                                                   | provider,namespace             |
| datakit_election_lease_remaining_seconds | gauge | Time remaining before the local safe lease deadline                                         | provider,namespace             |
| datakit_election_epoch        | gauge | Last epoch observed from the selected provider                                                       | provider,namespace             |
| datakit_election_request_errors_total | count | Election request failures grouped by a bounded reason                                          | provider,namespace,operation,reason |
| datakit_election_transitions_total | count | Leader lifecycle transitions grouped by a bounded reason                                         | provider,namespace,from,to,reason |
