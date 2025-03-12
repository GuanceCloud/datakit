### `window_hit()` {fn-window-hit}

函数原型： `fn window_hit()`

函数说明： 触发上下文被丢弃的数据的恢复事件，从 `point_window` 函数记录的数据中进行恢复

函数参数： 无

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
