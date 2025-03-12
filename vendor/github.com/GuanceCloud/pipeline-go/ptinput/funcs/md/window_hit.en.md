### `window_hit()` {fn-window-hit}

Function prototype: `fn window_hit()`

Function description: Trigger the recovery event of the context discarded data, and recover from the data recorded by the `point_window` function。

Function parameters: None

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
