
# Performance Benchmarks and Optimizations

---

## Test Environment {#env}

- OS: Ubuntu 22.04.1 LTS
- CPU: 12th Gen Intel(R) Core(TM) i7-12700H
- Memory: 40GB DDR5-4800
- DataKit: v1.9.1

## JSON Benchmark {#benchmark-json}

Test data:

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

Use the functions `json()` and `load_json()` functions to extract:

- load_json

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

- json

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


Benchmark results:

```not-set
BenchmarkScript/load_json()-20            202762          5674 ns/op        2865 B/op         61 allocs/op
BenchmarkScript/json()-20                  31024         41463 ns/op       21144 B/op        455 allocs/op
```

As a result, using the `load_json()` function compared to the `json()` function greatly reduces the time spent on processing test data, and the running time of a single script is reduced from 41.46us to 5.67us.

## Grok Benchmark {#benchmark-grok}

Take nginx access logs as test data:

```not-set
192.168.158.20 - - [19/Jun/2021:04:04:58 +0000] "POST /baxrrrrqc.php?daxd=a%20&d=1 HTTP/1.1" 404 118 "-" "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.6; fr; rv:1.9.2.8) Gecko/20100722 Firefox/3.6.8"
```

Use the following two scripts to test. The test content is to compare the performance impact of different Grok modes on the grok function. The main optimization is to replace `IPORHOST` with `NOTSPACE`:

- Script before optimization

```python
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
```


- Script after optimization

```python
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
```

Benchmark results:

```not-set
BenchmarkScript/grok_nginx-20            19292       67006 ns/op        3828 B/op         42 allocs/op
BenchmarkScript/grok_p1-20              139440        7860 ns/op        3665 B/op         42 allocs/op
```

As a result, after replacing the Grok mode `IPORHOST` with `NOTSPACE`, in the processing of test data, the execution time of a single script was reduced from 67us to 7.86us.
