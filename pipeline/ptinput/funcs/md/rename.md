### `rename()` {#fn-rename}

函数原型：`fn rename(new_key, old_key)`

函数说明：将已提取的字段重新命名

参数：

- `new_key`: 新字段名
- `old_key`: 已提取的字段名

示例：

```python
# 把已提取的 abc 字段重新命名为 abc1
rename('abc1', abc)

# or 

rename(abc1, abc)
```

```python
# 待处理数据：{"info": {"age": 17, "name": "zhangsan", "height": 180}}

# 处理脚本
json(_, info.name, "姓名")

# 处理结果
{
  "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}",
  "zhangsan": {
    "age": 17,
    "height": 180,
    "姓名": "zhangsan"
  }
}
```

