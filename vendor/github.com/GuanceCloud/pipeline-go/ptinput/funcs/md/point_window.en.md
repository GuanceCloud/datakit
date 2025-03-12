### `point_window()` {fn-point-window}

Function prototype: `fn point_window(before: int, after: int, stream_tags = ["filepath", "host"])`

Function description: Record the discarded data and use it with the `window_hit` function to upload the discarded context `Point` data.

Function parameters:

- `before`: The maximum number of points that can be temporarily stored before the function `window_hit`  is executed, and the data that has not been discarded is included in the count.
- `after`: The number of points retained after the `window_hit` function is executed, and the data that has not been discarded is included in the count.
- `stream_tags`: Differentiate log (metrics, tracing, etc.) streams by labels on the data, the default number using `filepath` and `host` can be used to distinguish logs from the same file.

Example:

```python
# It is recommended to place it in the first line of the script
#
point_window(8, 8)

# If it is a panic log, keep the first 8 entries 
# and the last 8 entries (including the current one)
#
if grok(_, "abc.go:25 panic: xxxxxx") {
    # This function will only take effect if point_window() is executed during this run.
    # Trigger data recovery behavior within the window
    #
    window_hit()
}

# By default, all logs whose service is test_app are discarded;
# If it contains panic logs, keep the 15 adjacent ones and the current one.
#
if service == "test_app" {
    drop()
}
```
