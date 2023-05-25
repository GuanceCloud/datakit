### `trim()` {#fn-trim}

函数原型：`fn trim(key, cutset: str = "")`

函数说明：删除 `key` 中首尾中指定的字符，`cutset` 为空字符串时默认删除所有空白符

函数参数：

- `key`: 已提取的某字段，字符串类型
- `cutset`: 删除 `key` 中出现在 `cutset` 字符串的中首尾字符

示例：

```python
# 待处理数据："trim(key, cutset)"

# 处理脚本
add_key(test_data, "ACCAA_test_DataA_ACBA")
trim(test_data, "ABC_")

# 处理结果
{
  "test_data": "test_Data"
}
```
