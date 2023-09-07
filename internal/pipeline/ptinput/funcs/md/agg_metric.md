### `agg_metric()` {#fn-agg-metric}

函数原型：`fn agg_metric(bucket: str, new_field: str, agg_fn: str, agg_by: []string, agg_field: str, category: str = "M")`

函数说明：根据输入的数据中的字段的名，自动取值后作为聚合数据的 tag，并将这些聚合数据存储在对应的 bucket 中

函数参数：

- `bucket`: 字符串类型，函数 `agg_create` 创建出的对应指标集合的 bucket，如果该 bucket 未被创建，则函数不执行任何操作
- `new_field`： 聚合出的数据中的指标名，其值的数据类型为 `float`
- `agg_fn`: 聚合函数，可以是 `"avg"`,`"sum"`,`"min"`,`"max"`,`"set"` 中的一种
- `agg_by`: 输入的数据中的字段的名，将作为聚合出的数据的 tag，这些字段的值只能是字符串类型的数据
- `agg_field`: 输入的数据中的字段名，自动获取字段值进行聚合
- `category`: 聚合数据的数据类别，可选参数，默认值为 "M"，表示指标类别数据。

示例：

以日志类别数据为例：

多个输入日志：

``` not-set
1
```

``` not-set
2
```

``` not-set
3
```

脚本：

```python
agg_create("cpu_agg_info", interval=10, const_tags={"tag1":"value_user_define_tag"})

set_tag("tag1", "value1")

field1 = _

cast(field1, "int")

agg_metric("cpu_agg_info", "agg_field_1", "sum", ["tag1", "host"], "field1")
```

指标输出：

``` not-set
{
    "host": "your_hostname",
    "tag1": "value1",
    "agg_field_1": 6,
}
```
