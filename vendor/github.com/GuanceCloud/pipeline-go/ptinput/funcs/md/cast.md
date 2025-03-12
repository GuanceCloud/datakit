### `cast()` {#fn-cast}

函数原型：`fn cast(key, dst_type: str)`

函数说明：将 key 值转换成指定类型

函数参数

- `key`: 已提取的某字段
- `type`：转换的目标类型，支持 `\"str\", \"float\", \"int\", \"bool\"` 这几种，目标类型需要用英文状态双引号括起来

示例：

```python
# 待处理数据：{"first": 1,"second":2,"third":"aBC","forth":true}

# 处理脚本
json(_, first) 
cast(first, "str")

# 处理结果
{
  "first": "1"
}
```
