### 关于采集器 nginx 调研

#### 指标相关
   - `telegraf` 目前 支持 五个采集器 `nginx` 、 `NginxUpstreamCheck` 、 `NginxPlusApi` 、`NginxPlus`、 `NginxVTS`
   - `datadog` 目前 采集的指标数据 在 `vhost_traffic_status` 和 `plus api` 基本上都能找到
   - 目前先支持 `nginx`  `ngx_http_stub_status_module` 和 `vhost_traffic_status` 模块的数据
    
 
#### conf

```
    [[inputs.nginx]]
      url = "http://localhost/server_status"
      use_vts = false
      ## Optional TLS Config
      # tls_ca = "/etc/telegraf/ca.pem"
      # tls_cert = "/etc/telegraf/cert.cer"
      # tls_key = "/etc/telegraf/key.key"
      ## Use TLS but skip chain & host verification
      insecure_skip_verify = false
    
      # HTTP response timeout (default: 5s)
      response_timeout = "5s"
```

**注意**

 - 默认开启的是`http ngx_http_stub_status_module` 模块，开启了此模块 会产生 `nginx` 指标集 详细开启介绍 [参见](http://nginx.org/en/docs/http/ngx_http_stub_status_module.html)
 
 - 如果使用的是 `vhost_traffic_status` 模块，在 config文件 中修改 `use_vts=true`, 开启 `vhost_traffic_status` 模块 会产生  `nginx`, `nginx_server_zone`, `nginx_upstream_zone`, `nginx_cache_zone` 等指标集  [参见](https://github.com/vozlt/nginx-module-vts#synopsis)
<!--  - 开启 `ngx_http_api_module` 模块 [参见](http://nginx.org/en/docs/http/ngx_http_api_module.html) --->
 


#### 采集的指标集跟指标及其介绍如下

  - nginx(metric_name)
  
  | fields        | type   | unit     |description |  
  | :----:        | :----: | :----:   |  :----: | 
  |   load_timestamp      |   int  |   milliseconds  | Loaded process time in milliseconds, when exist by open vts | 
  |   connection_active      |   int  |   count  | The current number of active client connections | 
  |   connection_reading     |   int  |   count  | The total  number of reading client connections | 
  |   connection_writing     |   int  |   count  | The total  number of writing client connections | 
  |   connection_waiting     |   int  |   count  | The total  number of waiting client connections | 
  |   connection_accepted    |   int  |   count  | The total  number of accepted client connections | 
  |   connection_handled     |   int  |   count  | The total  number of handled client connections | 
  |   connection_requests    |   int  |   count  | The total  number of requested client connections | 

    
  
  - nginx_server_zone

  | fields        | type   | unit     |description |  
  | :----:        | :----: | :----:   |  :----: |
  | request_count | int| count        | The total number of client requests received from clients. | 
  | received      | int| bytes        | The total amount of data received from clients. | 
  | sent          | int| bytes        | The total amount of data sent to clients. | 
  | response_1xx  | int| count        |The number of responses with status codes 1xx | 
  | response_2xx  | int| count        |The number of responses with status codes 2xx | 
  | response_3xx  | int| count        |The number of responses with status codes 3xx | 
  | response_4xx  | int| count        |The number of responses with status codes 4xx | 
  | response_5xx  | int| count        |  The number of responses with status codes 5xx | 

  -  nginx_upstream_zone
   
  | fields          | type   | unit     |description |   
  | :----:          | :----: | :----:   |  :----: |
  |   request_count | int| count       | The total number of client requests received from server. | 
  |   received      | int| bytes   | The total number of bytes received from this server. | 
  |   sent          | int| bytes   | The total number of bytes sent to clients. | 
  |   response_1xx  | int| count   | The number of responses with status codes 1xx | 
  |   response_2xx  | int| count   | The number of responses with status codes 2xx | 
  |   response_3xx  | int| count   | The number of responses with status codes 3xx | 
  |   response_4xx  | int| count   | The number of responses with status codes 4xx | 
  |   response_5xx  | int| count   | The number of responses with status codes 5xx | 
  
   -  nginx_cache_zone
         
 | fields          | type   | unit     |description |   
 | :----:          | :----: | :----:   |  :----: |
 |   max_size              | int| bytes | The limit on the maximum size of the cache specified in the configuration | 
 |   used_size             | int| bytes | The current size of the cache. | 
 |   receive               | int| bytes | The total number of bytes received from the cache. | 
 |   sent                  | int| bytes | The total number of bytes sent from the cache. | 
 |   responses_miss        | int| count | The number of cache miss | 
 |   responses_bypass      | int| count | The number of cache bypass | 
 |   responses_expired     | int| count | The number of cache expired | 
 |   responses_stale       | int| count | The number of cache stale | 
 |   responses_updating    | int| count | The number of cache updating | 
 |   responses_revalidated | int| count | The number of cache revalidated | 
 |   responses_hit         | int| count | The number of cache hit | 
 |   responses_scarce      | int| count | The number of cache scarce | 
 
 
 
<!-- 
  
     
     
   -  nginx_plus_api_processes
   
  | fields          | type   | unit     |description |   
  | :----:          | :----: | :----:   |  :----: |
  |   respawned     | int| count |    The total number of abnormally terminated and respawned child processes | 
       
   - nginx_plus_api_connections
   
  | fields          | type   | unit     |description |   
  | :----:          | :----: | :----:   |  :----: | 
  |   accepted      | int| count |  The total number of accepted client connections. | 
  |   dropped       | int| count |  The total number of dropped client connections. | 
  |   active        | int| count |  The current number of active client connections. | 
  |   idle          | int| count |  The current number of idle client connections. | 
          
   -  nginx_plus_api_ssl
   
  | fields          | type   | unit     |description |    
  | :----:          | :----: | :----:   |  :----: |  
  |   handshakes        | int| count | The total number of successful SSL handshakes | 
  |   handshakes_failed | int| count | The total number of failed SSL handshakes. | 
  |   session_reuses    | int| count | The total number of session reuses during SSL handshake. | 
     
   - nginx_plus_api_http_requests
      
  | fields          | type   | unit     |description |   
  | :----:          | :----: | :----:   |  :----: |  
  |   total     | int| count |    The total number of client requests. | 
  |   current   | int| count |    The current number of client requests. | 
   
   
   - nginx_plus_api_http_server_zones
         
  | fields          | type   | unit     |description |   
  | :----:          | :----: | :----:   |  :----: | 
  |   processing         | int| count         | The number of client requests that are currently being processed. | 
  |   requests           | int| count         | The total number of client requests received from clients. | 
  |   received           | int| byte          | The total amount of data received from clients. | 
  |   sent               | int| byte          | The total amount of data sent to clients. | 
  |   discarded          | int| count          | The total number of requests completed without sending a response. | 
  |   responses_1xx      | int| count         | The number of responses with 1xx status code. | 
  |   responses_2xx      | int| count         | The number of responses with 2xx status code.| 
  |   responses_3xx      | int| count         | The number of responses with 3xx status code. | 
  |   responses_4xx      | int| count         | The number of responses with 4xx status code. | 
  |   responses_5xx      | int| count         | The number of responses with 5xx status code. | 
  |   responses_totle    | int| count         | The total number of responses sent to clients. | 

   - nginx_plus_api_http_upstream_peers
   
  | fields          | type   | unit     |description |   
  | :----:          | :----: | :----:   |  :----: |  
  |   received      | int| bytes  | The total amount of data received from this server. | 
  |   requests      | int| count | The total number of client requests forwarded to this server. | 
  |   unavail       | int| count | How many times the server became unavailable for client requests (state "unavail") due to the number of unsuccessful attempts reaching the max_fails threshold. | 
  |   active        | int| count| The current number of active connections. | 
  |   backup        | int| count| A boolean value indicating whether the server is a backup server. | 
  |   downstart     | int| seconds   | The time (since Epoch) when the server became "unavail" or "unhealthy". | 
  |   downtime      | int| seconds  | Total time the server was in the "unavail" and "unhealthy" states. | 
  |   fails         | int| count   | The total number of unsuccessful attempts to communicate with the server. | 
  |   weight        | float | _  | Weight of the server. | 
  |   healthchecks_checks   | int| count     | The total number of health check requests made. | 
  |   healthchecks_fails    | int| count    | The number of failed health checks. | 
  |   healthchecks_last_passed  |bool|  _    | Boolean indicating if the last health check request was successful and passed tests. | 
  |   healthchecks_unhealthy    |int| time    | ow many times the server became unhealthy (state "unhealthy"). | 
  |   responses_1xx       | int| count  | The number of responses with 1xx status code. | 
  |   responses_2xx       | int| count  | The number of responses with 2xx status code.| 
  |   responses_3xx       | int| count  | The number of responses with 3xx status code. | 
  |   responses_4xx       | int| count  | The number of responses with 4xx status code. | 
  |   responses_5xx        | int| count | The number of responses with 5xx status code. | 
  |   responses_totle      | int| count   | The total number of responses obtained from this server. | 
  
-->