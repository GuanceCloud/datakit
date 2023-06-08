
# httpapi 模块设计

http 模块主要负责服务外部的 API 请求. 部分采集器的数据处理,也是寄生在当前的 HTTP 服务上(:9529)


## Prometheus Metrics

http 模块暴露如下 metrics：

| 指标                               | 类型      | 说明                              | labels            |
| ---------------------------------- | --------- | --------------------------------- | ----------------- |
| datakit_http_api_total             | count     | API request counter               | api,method,status |
| datakit_http_api_elapsed           | summary   | API request cost(in ms)           | api,method,status |
| datakit_http_api_elapsed_histogram | histogram | API request cost(in ms) histogram | api,method,status |
| datakit_http_api_req_size_count    | count     | API request count                 | api,method,status |
| datakit_http_api_req_size_sum      | summary   | API request body total size       | api,method,status |

