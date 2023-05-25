### `len()` {#fn-len}

函数原型：`fn len(val: str|map|list) int`

函数说明：计算 string 字节数，map 和 list 的元素个数。

参数：

- `val`: 可以是 map、list 或 string

示例：

```python
# 例 1
add_key(abc, len("abc"))
# 输出
{
 "abc": 3,
}

# 例 2
add_key(abc, len(["abc"]))
#处理结果
{
  "abc": 1,
}
```
