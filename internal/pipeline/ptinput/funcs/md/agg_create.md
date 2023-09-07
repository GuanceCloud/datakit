### `agg_create()` {#fn-agg-create}

函数原型：`fn agg_create(bucket: str, on_interval: str = "60s", on_count: int = 0, keep_value: bool = false, const_tags: map[string]string = nil, category: str = "M")`

函数说明：创建一个用于聚合的指标集，通过 `on_interval` 或 `on_count` 设置时间或次数作为聚合周期，聚合结束后将上传聚合数据，可以选择是否保留上一次聚合的数据

函数参数：

- `bucket`: 字符串类型，作为聚合出的指标的指标集名，如果该 bucket 已经创建，则函数不执行任何操作
- `on_interval`：默认值 `60s`, 以时间作为聚合周期，单位 `s`，值大于 `0` 时参数生效；不能同时与 `on_count` 小于等于 0；
- `on_count`: 默认值 `0`，以处理的点数作为聚合周期，值大于 `0` 时参数生效
- `keep_value`: 默认值 `false`
- `const_tags`: 自定义的 tags，默认为空
- `category`: 聚合数据的数据类别，可选参数，默认值为 "M"，表示指标类别数据。

示例：

```python
agg_create("cpu_agg_info", on_interval = "30s")
```
