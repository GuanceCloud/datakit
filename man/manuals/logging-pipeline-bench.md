## DataKit 日志采集器性能测试

### 环境和工具

- 操作系统：Ubuntu 20.04.2 LTS
- CPU：Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
- 内存：16GB  Speed 2133 MT/s
- DataKit 版本：1.1.8-rc1
- 日志文本（nginx access log）： 
```
172.17.0.1 - - [06/Jan/2017:16:16:37 +0000] "GET /datadoghq/company?test=var1%20Pl HTTP/1.1" 401 612 "http://www.perdu.com/" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36" "-"
```
- 日志数量：10w 行
- 使用 pipeline：见后文

### 测试结果

| 测试条件                                                                                                   | 耗时   |
| :--                                                                                                        | ---    |
| 不使用 pipeline，纯日志文本处理（包括编码、多行和字段检测等）                                              | 3秒63  |
| 使用完整版的 pipeline（见附录一）                                                                          | 43秒70 |
| 使用单一匹配的 pipeline，与完整版相比舍弃多种匹配格式，例如 nginx error log，只针对 access log（见附录二） | 16秒91 |
| 使用优化过的单一匹配 pipeline，替换性能消耗多的 pattern（见附录三）                                        | 4秒40  |


注：
> pipeline 耗时期间，CPU 单核心满负载运行，使用率持续在 100% 左右，当 10w 条日志处理结束时 CPU 回落

> 测试期间内存消耗稳定，没有明显的使用率增加

> 耗时为 DataKit 程序计算，不同环境下可能会有偏差

### 对比

使用 Fluentd 对同样 10w 行日志进行采集，CPU 使用率在 3秒内从 43% 升至 77% 然后回落，可以预见此时已经处理结束。

因 Fluentd 存在 metadate 缓存机制，分批次输出结果，所以无法确切计算究竟耗时多少。

Fluentd 的 pipeline 匹配模式单一，没有进行同数据源多格式的 pipeline（例如 nginx 只支持 access log 而不支持 error log）。

### 结论

DataKit 日志采集，在 pipeline 单一匹配模式下，处理耗时和 Fluentd 相差 30% 左右。

但是如果使用完整版全量匹配 pipeline，耗时剧增。

### 附录（pipeline）

#### 一, 完整版/全量匹配 pipeline

```
add_pattern("date2", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")

# access log
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

# access log
add_pattern("access_common", "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}')
user_agent(agent)

# error log
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{IPORHOST:client_ip}, server: %{IPORHOST:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{IPORHOST:ip_or_host}\"")
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{IPORHOST:client_ip}, server: %{IPORHOST:server}, request: \"%{GREEDYDATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", host: \"%{IPORHOST:ip_or_host}\"")
grok(_,"%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

group_in(status, ["warn", "notice"], "warning")
group_in(status, ["error", "crit", "alert", "emerg"], "error")

cast(status_code, "int")
cast(bytes, "int")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)


nullif(http_ident, "-")
nullif(http_auth, "-")
nullif(upstream, "")
default_time(time)
```

#### 二, 单一匹配 pipeline

```
# access log
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
```

#### 三, 优化过的单一匹配 pipeline（将性能消耗极大的 IPORHOST 改为 NOTSPACE）

```
# access log
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
```
