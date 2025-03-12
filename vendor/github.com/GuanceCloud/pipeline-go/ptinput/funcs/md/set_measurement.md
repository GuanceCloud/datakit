### `set_measurement()` {#fn-set-measurement}

函数原型：`fn set_measurement(name: str, delete_key: bool = false)`

函数说明：改变行协议的 name
函数参数：

- `name`: 值作为 measurement name，可传入字符串常量或变量
- `delete_key`: 如果在 point 中存在与变量同名的 tag 或 field 则删除它

行协议 name 与各个类型数据存储时的字段映射关系或其他用途：

| 类别          | 字段名 | 其他用途 |
| -             | -      | -        |
| custom_object | class  | -        |
| keyevent      | -      | -        |
| logging       | source | -        |
| metric        | -      | 指标集名 |
| network       | source | -        |
| object        | class  | -        |
| profiling     | source | -        |
| rum           | source | -        |
| security      | rule   | -        |
| tracing       | source | -        |
