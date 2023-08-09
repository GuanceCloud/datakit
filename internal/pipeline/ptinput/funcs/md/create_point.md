### `create_point()` {#fn-create-point}

函数原型：`fn create_point(name, tags, fields, ts = 0, category = "M", after_use = "")`

函数说明：创建新的数据并输出

函数参数：

- `name`: point name，视为指标集的名、日志 source 等
- `tags`: 数据标签
- `fields`:  数据字段
- `ts`:   可选参数，unix 纳秒时间戳，默认为当前时间
- `category`: 可选参数，数据类别，支持类别名称和名称简写，如指标类别可以填写 `M` 或 `metric`，日志则是 `L` 或 `logging`
- `after_use`: 可选参数，创建 point 后，对创建的 point 执行指定的 pl 脚本；如果原始数据类型是 L，被创建的数据的类别为 M，此时执行的还是 L 类别下的脚本


示例：

```py
# input
'''
{"a": "b"}
'''
fields = load_json(_)
create_point("name_pt", {"a": "b"}, fields)
```
