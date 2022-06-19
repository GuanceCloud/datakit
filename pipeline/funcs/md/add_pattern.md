### `add_pattern()` {#fn-add-pattern}

函数原型：`add_pattern(name=required, pattern=required)`

函数说明：创建自定义 grok 模式。grok 模式有作用域限制, 如在 if else 语句内将产生新的作用域, 该 pattern 仅在此作用域内有效。该函数不可覆盖同一作用域或者上一作用域已经存在的 grok 模式

参数:

- `name`：模式命名
- `pattern`: 自定义模式内容

示例:

```python
# 待处理数据: "11,abc,end1", "22,abc,end1", "33,abc,end3"

# pipline脚本
add_pattern("aa", "\\d{2}")
grok(_, "%{aa:aa}")
if false {

} else {
    add_pattern("bb", "[a-z]{3}")
    if aa == "11" {
        add_pattern("cc", "end1")
        grok(_, "%{aa:aa},%{bb:bb},%{cc:cc}")
    } elif aa == "22" {
        # 此处使用 pattern cc 将导致编译失败: no pattern found for %{cc}
        grok(_, "%{aa:aa},%{bb:bb},%{INT:cc}")
    } elif aa == "33" {
        add_pattern("bb", "[\\d]{5}")	# 此处覆盖 bb 失败
        add_pattern("cc", "end3")
        grok(_, "%{aa:aa},%{bb:bb},%{cc:cc}")
    }
}

# 处理结果
{
    "aa":      "11"
    "bb":      "abc"
    "cc":      "end1"
    "message": "11,abc,end1"
}
{
    "aa":      "22"
	 "message": "22,abc,end1"
}
{
    "aa":      "33"
    "bb":      "abc"
    "cc":      "end3"
    "message": "33,abc,end3"
}
```
