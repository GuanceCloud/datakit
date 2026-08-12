### `http_request()` {#fn-http-request}

Function prototype: `fn http_request(method: str, url: str, headers: map = nil, body: any = nil, prefix: str = "") map`

Function description: Send an HTTP request, receive the response, and encapsulate it into a map

Function parameters:

- `method`: HTTP request method, such as GET, POST, or PUT
- `url`: Request path
- `headers`: Optional. Additional headers of type map[string]string
- `body`: Optional. Request body
- `prefix`: Optional. Adds a custom prefix to the fixed keys (`status_code`, `body`) of the returned map to avoid conflicts with existing fields. It can also be passed as a named parameter, e.g. `http_request(method, url, prefix="hr_")`.

Return type: map

key contains status code (status_code) and result body (body)

- `status_code`: Status code
- `body`: Response body

Example:

```python
resp = http_request("GET", "http://localhost:8080/testResp")
resp_body = load_json(resp["body"])

add_key(abc, resp["status_code"])
add_key(abc, resp_body["a"])
```

Example with prefix:

```python
resp = http_request("GET", "http://localhost:8080/testResp", prefix="hr_")
resp_body = load_json(resp["hr_body"])

add_key(abc, resp["hr_status_code"])
add_key(abc, resp_body["a"])
```
