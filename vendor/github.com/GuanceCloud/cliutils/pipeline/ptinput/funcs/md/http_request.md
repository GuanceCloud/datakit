### `http_request()` {#fn-http-request}

函数原型： `fn http_request(method: str, url: str, headers: map, body: any) map`

函数说明： 发送 HTTP 请求，接收响应并封装成 map

参数：

- `method`：GET|POST
- `url`: 请求路径
- `headers`：附加的 header，类型为 map[string]string
- `body`：请求体

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
