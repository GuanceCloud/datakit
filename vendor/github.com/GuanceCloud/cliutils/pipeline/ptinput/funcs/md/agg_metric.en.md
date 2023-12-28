### `agg_metric()` {#fn-agg-metric}

[:octicons-tag-24: Version-1.5.10](../datakit/changelog.md#cl-1.5.10)

Function prototype: `fn agg_metric(bucket: str, new_field: str, agg_fn: str, agg_by: []string, agg_field: str, category: str = "M")`

Function description: According to the field name in the input data, the value is automatically taken as the label of the aggregated data, and the aggregated data is stored in the corresponding bucket. This function does not work with central Pipeline.

Function parameters:

- `bucket`: String type, the bucket created by the agg_create function, if the bucket has not been created, the function will not perform any operations.
- `new_field`： The name of the field in the aggregated data, the data type of its value is `float`.
- `agg_fn`: Aggregation function, can be one of `"avg"`, `"sum"`, `"min"`, `"max"`, `"set"`.
- `agg_by`: The name of the field in the input data will be used as the tag of the aggregated data, and the value of these fields can only be string type data.
- `agg_field`: The field name in the input data, automatically obtain the field value for aggregation.
- `category`: Data category for aggregated data, optional parameter, the default value is "M", indicating the indicator category data.

Example:

Take `logging` category data as an example:

Multiple inputs in a row:

- Sample log one: `{"a": 1}`
- Sample log two: `{"a": 2}`

script:

```python
agg_create("cpu_agg_info", on_interval="10s", const_tags={"tag1":"value_user_define_tag"})

set_tag("tag1", "value1")

field1 = load_json(_)

field1 = field1["a"]

agg_metric("cpu_agg_info", "agg_field_1", "sum", ["tag1", "host"], "field1")
```

metric output:

```json
{
    "host": "your_hostname",
    "tag1": "value1",
    "agg_field_1": 3
}
```
