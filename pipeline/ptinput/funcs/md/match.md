### `match()` {#fn-match}

函数原型：`fn match(pattern: str, s: str) bool`

函数说明：使用指定的正则表达式匹配字符串，匹配成功返回 true，否则返回 false

参数：

- `pattern`: 正则表达式
- `s`: 待匹配的字符串

示例：

```python
# 脚本
test_1 = "pattern 1,a"
test_2 = "pattern -1,"

add_key(match_1, match('''\w+\s[,\w]+''', test_1)) 

add_key(match_2, match('''\w+\s[,\w]+''', test_2)) 

# 处理结果
{
    "match_1": true,
    "match_2": false
}
```
