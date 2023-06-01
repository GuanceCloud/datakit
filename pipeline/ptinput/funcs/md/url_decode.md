### `url_decode()` {#fn-url-decode}

函数原型：`fn url_decode(key: str)`

函数说明：将已提取 `key` 中的 URL 解析成明文

参数：

- `key`: 已经提取的某个 `key`

示例：

```python
# 待处理数据：{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}

# 处理脚本
json(_, url) url_decode(url)

# 处理结果
{
  "message": "{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}",
  "url": "http://www.baidu.com/s?wd=测试"
}
```
