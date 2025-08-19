### `strlen()` {#fn-strlen}

函数原型：`fn strlen(val: str) int`

函数说明：计算字符串的字符数量，而不是字节数。

参数：

- `val`: 输入字符串

示例：

```python
add_key("len_char", strlen("hello 你好"))
add_key("len_byte", len("hello 你好"))
```

输出：

```json
{
 "len_char": 8,
 "len_byte": 12
}
```
