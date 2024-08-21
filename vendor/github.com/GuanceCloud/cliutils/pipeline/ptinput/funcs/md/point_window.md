### `point_window()` {fn-point-window}

函数原型： `fn point_window(before: int, after: int, stream_tags = ["filepath", "host"])`

函数说明： 记录被丢弃的数据，配合 `window_hit` 函数使用，上传被丢弃的上下文 `Point` 数据。

函数参数：

- `before`: 函数 `window_hit` 执行之前的最大可暂存的 Point 个数，未丢弃的数据参与计数。
- `after`: 函数 `window_hit` 执行之后保留的 Point 个数，未丢弃的数据参与计数。
- `stream_tags`: 通过数据上的标签区分日志（指标，链路等）流，默认数使用 `filepath` 和 `host` 可用于区分来自同一文件的日志。

示例：

```python
# 建议放置在脚本首行
#
point_window(8, 8)

# 如果是 panic 日志，保留前 8 条，以及后 8 条（包含当前一条）
if grok(_, "abc.go:25 panic: xxxxxx") {
    # 只有此次运行过程中 point_window() 被执行，这个函数才会生效
    # 触发窗口内的数据恢复行为
    #
    window_hit()
}

# 默认丢弃全部的 service 为 test_app 的日志；
# 若包含 panic 的日志，则保留相邻的 15 条以及当前这条
#
if service == "test_app" {
    drop()
}
```
