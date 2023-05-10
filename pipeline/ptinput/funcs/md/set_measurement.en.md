### `set_measurement()` {#fn-set-measurement}

Function prototype: `fn set_measurement(name: str, delete_key: bool = false)`

Function description: change the name of the line protocol

Function parameters:

- `name`: The value is used as the measurement name, which can be passed in as a string constant or variable
- `delete_key`: If there is a tag or field with the same name as the variable in point, delete it

The field mapping relationship between row protocol name and various types of data storage or other purposes:

| category      | field name | other usage     |
| -             | -          | -               |
| custom_object | class      | -               |
| keyevent      | -          | -               |
| logging       | source     | -               |
| metric        | -          | metric set name |
| network       | source     | -               |
| object        | class      | -               |
| profiling     | source     | -               |
| rum           | source     | -               |
| security      | rule       | -               |
| tracing       | source     | -               |
