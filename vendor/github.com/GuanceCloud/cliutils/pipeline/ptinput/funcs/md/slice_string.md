### `slice_string()` {#fn_slice_string}

函数原型：`fn slice_string(name: str, start: int, end: int) -> str`

函数说明：返回字符串从索引 start 到 end 的子字符串。

函数参数：

- `name`: 要截取的字符串
- `start`: 子字符串的起始索引（包含）
- `end`: 子字符串的结束索引（不包含）

示例：

```python
substring = slice_string("15384073392", 0, 3)
# substring 的值为 "153"
```