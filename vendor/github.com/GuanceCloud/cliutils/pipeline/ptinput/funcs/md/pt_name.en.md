### `pt_name()` {#fn-pt-name}

Function prototype: `fn pt_name(name: str = "") -> str`

Function description: Get the name of point; if the parameter is not empty, set the new name.

Function parameters:

- `name`: Value as point name; defaults to empty string.

The field mapping relationship between point name and various types of data storage:

| category      | field name |
| ------------- | ---------- |
| custom_object | class      |
| keyevent      | -          |
| logging       | source     |
| metric        | -          |
| network       | source     |
| object        | class      |
| profiling     | source     |
| rum           | source     |
| security      | rule       |
| tracing       | source     |
