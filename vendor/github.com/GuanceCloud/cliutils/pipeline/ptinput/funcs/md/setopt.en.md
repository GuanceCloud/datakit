### `setopt()` {#fn-setopt}

Function prototype: `fn setopt(status_mapping: bool = true)`

Function description: Modify Pipeline settings, parameters must be in the form of `key=value`

Function parameters:

- `status_mapping`: Set the mapping function of the `status` field of log data, enabled by default

Example:

```py
# Disable the mapping function for the status field
setopt(status_mapping=false)

add_key("status", "w")

# Processing result
{
"status": "w",
}
```

```py
# Enable the mapping function for the status field by default
setopt(status_mapping=true)

add_key("status", "w")

# Processing result
{
"status": "warning",
}
```
