### `setopt()` {#fn-setopt}

函数原型：`fn setopt(status_mapping: bool = true)`

函数说明：修改 Pipeline 的设置，参数必须采用 `key=value` 的形式

函数参数：

- `status_mapping`: 设置日志类数据的 `status` 字段的映射功能，默认开启

示例：

```py
# 关闭对 status 字段的映射功能
setopt(status_mapping=false)

add_key("status", "w")

# 处理结果
{
  "status": "w",
}
```

```py
# 对 status 字段的映射功能是默认开启的
setopt(status_mapping=true)

add_key("status", "w")

# 处理结果
{
  "status": "warning",
}
```
