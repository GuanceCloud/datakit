### `http_request()` {#fn-http-request}

函数原型： `fn http_request(method: str, url: str, headers: map = nil, body: any = nil, prefix: str = "") map`

函数说明： 发送 HTTP 请求，接收响应并封装成 map

参数：

- `method`：HTTP 请求方法，如 GET、POST、PUT
- `url`: 请求路径
- `headers`：可选参数，附加的 header，类型为 map[string]string
- `body`：可选参数，请求体
- `prefix`：可选参数，为返回 map 中固定的 key（`status_code`、`body`）添加自定义前缀，避免与已有字段产生冲突。支持以命名参数形式传入，如 `http_request(method, url, prefix="hr_")`

返回值类型：map

key 包含了状态码（status_code）和返回体（body）

- `status_code`: 状态码
- `body`: 返回体

示例：

```python
resp = http_request("GET", "http://localhost:8080/testResp")
resp_body = load_json(resp["body"])

add_key(abc, resp["status_code"])
add_key(abc, resp_body["a"])
```

添加前缀示例：

```python
resp = http_request("GET", "http://localhost:8080/testResp", prefix="hr_")
resp_body = load_json(resp["hr_body"])

add_key(abc, resp["hr_status_code"])
add_key(abc, resp_body["a"])
```
