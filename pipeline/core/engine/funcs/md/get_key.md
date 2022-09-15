### `get_key()` {#fn-get-key}

函数原型：`get_key(key_name=required)`

函数说明：从 point 中读取 key 的值，而不是堆栈上的变量的值

函数参数

- `key_name`: key 的名称

示例一:

```python
# scipt 1
key = "shanghai"
add_key(key)
key = "tokyo" 
add_key(add_new_key, key)

# 处理结果
{
  "add_new_key": "tokyo",
  "key": "shanghai",
}

```

示例二:

```python
# scipt 2
key = "shanghai"
add_key(key)
key = "tokyo" 
add_key(add_new_key, get_key(key))

#处理结果
{
  "add_new_key": "shanghai",
  "key": "shanghai",
}
```
