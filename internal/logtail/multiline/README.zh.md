# 多行日志匹配策略

这个 package 用来判断一行日志是否是一条多行日志的新起始行。

## 优先级

多行匹配有两种模式：

1. 手动模式：配置了 `multiline_match`。
   - 调用路径：`tailer.WithMultilinePattern(pattern)` -> `multiline.New([]string{pattern})`。
   - 只使用这个用户配置的规则。
   - 不使用内置自动规则，也不使用 `auto_multiline_extra_patterns`。

2. 自动模式：启用了多行功能，并且 `multiline_match` 为空。
   - 调用路径：`tailer.WithAutoMultilineExtraPatterns(extraPatterns)` -> `multiline.NewAuto(extraPatterns)`。
   - 先匹配内置自动规则。
   - `auto_multiline_extra_patterns` 是补充规则，只在选中的内置规则组未命中后才匹配。

## 内置分组

自动模式下，内置规则按输入行的第一个字节分为三组：

- `GlobalDigitPatterns`：首字节是 `0-9`。
- `GlobalLetterPatterns`：首字节是 `A-Z` 或 `a-z`。
- `GlobalSymbolPatterns`：非字母数字首字节，例如 `[`, `(`, `<`, `#`, `=`, `{`。

当前没有 `GlobalPatterns` 这种全量兜底规则表。不要为了自动匹配重新引入它。

## 空白字符

空格和 tab 有单独的快速路径，因为它们通常是调用栈或异常栈的续行前缀。

自动模式运行流程：

```text
line
  -> 如果首字节是空格或 tab：
       只匹配 extraPatterns，然后返回
  -> 否则：
       根据首字节选择 digit / letter / symbol 其中一组
       匹配选中的内置规则组
       如果内置规则组未命中，再匹配 extraPatterns
```

这样可以避免常见调用栈反复扫描 symbol 组。如果部署环境确实存在空格或 tab 开头的新日志，
需要通过 `auto_multiline_extra_patterns` 显式补充。

## 添加规则

- 新增内置起始行规则时，直接加到 `patterns.go` 的一个或多个数组中。
- 正则必须使用 `^` 锚定行首。
- 规则必须放在其首字节可能命中的分组里。只能匹配字母开头日志的规则，不应该放到
  `GlobalSymbolPatterns`。
- 这些规则只负责识别日志起始行，不是完整日志解析器。
- 对调用栈和异常关键字要保守。`Traceback`、`Exception`、`Caused by` 和缩进栈帧通常是续行。
