### `create_point()` {#fn-create-point}

Function prototype: `fn create_point(name, tags, fields, ts = 0, category = "M", after_use = "")`

Function description: Create new data and output

Function parameters:

- `name`: point name, which is regarded as the name of the metric set, log source, etc.
- `tags`: data tags
- `fields`: data fields
- `ts`: optional parameter, unix nanosecond timestamp, defaults to current time
- `category`: optional parameter, data category, supports category name and name abbreviation, such as metric category can be filled with `M` or `metric`, log is `L` or `logging`
- `after_use`: optional parameter, after the point is created, execute the specified pl script on the created point; if the original data type is L, the created data category is M, and the script under the L category is executed at this time

Example:

```py
# input
'''
{"a": "b"}
'''
fields = load_json(_)
create_point("name_pt", {"a": "b"}, fields)
```
