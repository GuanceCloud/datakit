### `agg_create()` {#fn-agg-create}

Function prototype: `fn agg_create(bucket: str, on_interval: str = "60s", on_count: int = 0, keep_value: bool = false, const_tags: map[string]string = nil, category: str = "M")`

Function description: Create an aggregation measurement, set the time or number of times through `on_interval` or `on_count` as the aggregation period, upload the aggregated data after the aggregation is completed, and choose whether to keep the last aggregated data

Function parameters:

- `bucket`: String type, as an aggregated field, if the bucket has already been created, the function will not perform any operations.
- `on_interval`：The default value is `60s`, which takes time as the aggregation period, and the unit is `s`, and the parameter takes effect when the value is greater than `0`; it cannot be combined with `on_count` less than or equal to 0.
- `on_count`: The default value is `0`, the number of processed points is used as the aggregation period, and the parameter takes effect when the value is greater than `0`
- `keep_value`: The default value is `false`
- `const_tags`: Custom tags, empty by default
- `category`: Data category for aggregated data, optional parameter, the default value is "M", indicating the indicator category data.

示例：

```python
agg_create("cpu_agg_info", interval = 60)
```
