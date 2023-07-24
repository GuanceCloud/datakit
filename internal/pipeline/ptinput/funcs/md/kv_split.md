### `kv_split()` {#fn-kv_split}

函数原型：`fn kv_split(key, field_split_pattern = " ", value_split_pattern = "=", trim_key = "", trim_value = "", include_keys = [], prefix = "") -> bool`

函数说明：从字符串中提取出所有的键值对

参数：

- `key`: key 名称
- `include_keys`: 包含的 key 名称列表，仅提取在该列表内的 key；**默认值为 []，不提取任何 key**
- `field_split_pattern`: 字符串分割，用于提取出所有键值对的正则表达式；默认值为 `" "`
- `value_split_pattern`: 用于从键值对字符串分割出键和值，非递归；默认值为 `"="`
- `trim_key`: 删除提取出的 key 的前导和尾随的所有指定的字符；默认值为 `""`
- `trim_value`: 删除提取出的 value 的前导和尾随的所有指定的字符；默认值为 `""`
- `prefix`: 给所有的 key 添加前缀字符串

示例：


```python
# input: "a=1, b=2 c=3"
kv_split(_)
 
'''output:
{
  "message": "a=1, b=2 c=3",
  "status": "unknown",
  "time": 1679558730846377132
}
'''
```

```python
# input: "a=1, b=2 c=3"
kv_split(_, include_keys=["a", "c", "b"])
 
'''output:
{
  "a": "1,",
  "b": "2",
  "c": "3",
  "message": "a=1 b=2 c=3",
  "status": "unknown",
  "time": 1678087119072769560
}
'''
```

```python
# input: "a=1, b=2 c=3"
kv_split(_, trim_value=",", include_keys=["a", "c", "b"])

'''output:
{
  "a": "1",
  "b": "2",
  "c": "3",
  "message": "a=1, b=2 c=3",
  "status": "unknown",
  "time": 1678087173651846101
}
'''
```


```python
# input: "a=1, b=2 c=3"
kv_split(_, trim_value=",", include_keys=["a", "c"])

'''output:
{
  "a": "1",
  "c": "3",
  "message": "a=1, b=2 c=3",
  "status": "unknown",
  "time": 1678087514906492912
}
'''
```

```python
# input: "a::1,+b::2+c::3" 
kv_split(_, field_split_pattern="\\+", value_split_pattern="[:]{2}",
    prefix="with_prefix_",trim_value=",", trim_key="a", include_keys=["a", "b", "c"])

'''output:
{
  "message": "a::1,+b::2+c::3",
  "status": "unknown",
  "time": 1678087473255241547,
  "with_prefix_b": "2",
  "with_prefix_c": "3"
}
'''
```
