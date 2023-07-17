
# 性能基准和优化

---

## 测试环境 {#env}

- 操作系统：Ubuntu 22.04.1 LTS
- CPU：12th Gen Intel(R) Core(TM) i7-12700H
- 内存：40GB DDR5-4800
- DataKit： v1.9.1

## JSON 基准 {#benchmark-json}

测试数据：

```json
{
    "tcpSeq": "71234234923",
    "language": "C",
    "channel": "26",
    "check_bit": "dSa-aoHjw7b42-dcCE2Sc-aULcaeZav",
    "cid": "2139-02102-213-122341-1190",
    "address": "application/a.out_func_a_64475d386d3:0xfe21023",
    "time": 1681508212100,
    "sub_source": "N",
    "id": 508,
    "status": "ok",
    "cost": 101020 
}
```

分别使用函数 `json()` 和 `load_json()` 函数进行提取

- `load_json()`

```python
data = load_json(_)
add_key(tcpSeq, data["tcpSeq"])
add_key(language, data["language"])
add_key(channel, data["channel"])
add_key(check_bit, data["check_bit"])
add_key(cid, data["cid"])
add_key(address, data["address"])
add_key(time, data["time"])
add_key(sub_source, data["sub_source"])
add_key(id, data["id"])
add_key(status, data["status"])
add_key(cost, data["cost"])
```

- `json()`

```python
json(_, tcpSeq, tcpSeq)
json(_, language, language)
json(_, channel, channel)
json(_, check_bit, check_bit)
json(_, cid, cid)
json(_, address, address)
json(_, time, time)
json(_, sub_source, sub_source)
json(_, id, id)
json(_, status, status)
json(_, cost, cost)
```

基准测试结果：

```not-set
BenchmarkScript/load_json()-20            202762          5674 ns/op        2865 B/op         61 allocs/op
BenchmarkScript/json()-20                  31024         41463 ns/op       21144 B/op        455 allocs/op
```

结果上，使用 `load_json()` 函数相较于 `json()` 函数在测试数据的处理上的耗时大幅减少，单个脚本的运行时间由 41.46us 减少至 5.67us。

## Grok 基准 {#benchmark-grok}

以 nginx 访问日志为测试数据：

```not-set
192.168.158.20 - - [19/Jun/2021:04:04:58 +0000] "POST /baxrrrrqc.php?daxd=a%20&d=1 HTTP/1.1" 404 118 "-" "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.6; fr; rv:1.9.2.8) Gecko/20100722 Firefox/3.6.8"
```

使用以下两个脚本进行测试，测试内容为对比不同 Grok 模式对 grok 函数的性能影响，主要优化为将 `IPORHOST` 替换为 `NOTSPACE`：

- 优化前的脚本

```python
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
```


- 优化后的脚本

```python
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
```

基准测试结果：

```not-set
BenchmarkScript/grok_nginx-20            19292       67006 ns/op        3828 B/op         42 allocs/op
BenchmarkScript/grok_p1-20              139440        7860 ns/op        3665 B/op         42 allocs/op
```

结果上，将 Grok 模式 `IPORHOST` 替换为 `NOTSPACE` 后，在测试数据的处理上，单个脚本执行耗时由 67us 减少至 7.86 us。
