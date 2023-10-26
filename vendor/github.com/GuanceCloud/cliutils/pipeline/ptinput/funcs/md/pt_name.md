### `pt_name()` {#fn-pt-name}

函数原型：`fn pt_name(name: str = "") -> str`

函数说明：获取 point 的 name；如果参数不为空则设置新的 name。

函数参数：

- `name`: 值作为 point name；默认值为空字符串

Point Name 与各个类型数据存储时的字段映射关系：

| 类别          | 字段名 |
| ------------- | ------ |
| custom_object | class  |
| keyevent      | -      |
| logging       | source |
| metric        | -      |
| network       | source |
| object        | class  |
| profiling     | source |
| rum           | source |
| security      | rule   |
| tracing       | source |
