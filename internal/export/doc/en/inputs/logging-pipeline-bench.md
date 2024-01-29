# DataKit Performance Test of Log Collector
---

## Environment and Tools {#env-tools}

- Operating system: Ubuntu 20.04.2 LTS
- CPU：Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
- Memory: 16GB  Speed 2133 MT/s
- DataKit：1.1.8-rc1
- Log text (nginx access log):

``` not-set
172.17.0.1 - - [06/Jan/2017:16:16:37 +0000] "GET /datadoghq/company?test=var1%20Pl HTTP/1.1" 401 612 "http://www.perdu.com/" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36" "-"
```

- Number of logs: 10w lines
- Using Pipeline: See below

## Test Result {#result}

| Test Conditions                                                                                                                                                      | Time Consuming   |
|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------| ---    |
| No Pipeline, pure log text processing (including encoding, multi-line, field detection and so on)                                                                    | 3.63 seconds  |
| Use the full Pipeline (see Appendix I)                                                                                                                               | 43.70 seconds|
| Use a single matching Pipeline instead of multiple matching formats, such as nginx error log, as compared to the full version, for access log only (see Appendix II) | 16.91 seconds |
| Replace the performance-intensive pattern with an optimized single matching Pipeline (see Appendix III)                                                              | 4.40 seconds |


Note:

> Pipeline time-consuming period, the CPU single core runs at full load, the utilization rate continues at about 100%, and the CPU falls back when the 10w log processing is finished
>
> Memory consumption was stable during the test, with no significant increase in usage
>
> Time-consuming computation for DataKit programs, which may be biased in different environments

## Comparison {#compare}

Using Fluentd to capture the same 10w row log, CPU utilization increased from 43% to 77% in 3 seconds and then dropped, which can be predicted to be the end of processing.

As Fluentd has a metadata caching mechanism, which outputs results in batches, it is impossible to calculate exactly how much time it takes.

Fluentd's Pipeline matching pattern is single, and there is no Pipeline with multiple formats of the same data source (for example, nginx only supports access log but not error log).

## Conclusion {#conclusion}

In Pipeline single matching mode, DataKit log collection processing time is about 30% different from Fluentd.

However, if you use the full version of the full matching Pipeline, the time consumption increases dramatically.

## Appendix {#appendix}

### 1, full version/full match Pipeline {#full-match}

```python
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

### 2, a single matching Pipeline {#single-match}

```python
# access log
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
```

### 3, the optimized single matching Pipeline {#optimized-pl}

```python
# access log
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
```
