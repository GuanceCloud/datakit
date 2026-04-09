### `pt_kvs_get()` {#fn_pt_kvs_get}

函数原型：`fn pt_kvs_get(name: str, raw: bool = false) -> any`

函数说明：返回 Point 中指定 key 的值

说明：

- 默认按普通值返回；数组或字典默认会转换成字符串
- 传入 `raw=true` 时，数组或字典会保留为 `list` / `map`，可继续参与 `append()`、`len()`、`in` 等脚本运算
- tag 始终按字符串返回
- 如果该值已在更早的处理链路中被存成 JSON 字符串，则返回的仍然是字符串

函数参数：

- `name`: Key 名
- `raw`: 是否按 Point 原生值返回；为 `false` 时，数组/字典会按普通字符串返回

示例：

```python
host = pt_kvs_get("host")

nums = pt_kvs_get("nums", raw=true)
nums = append(nums, 4)
pt_kvs_set("nums", nums, raw=true)

obj = pt_kvs_get("obj", raw=true)
obj_str = pt_kvs_get("obj", raw=false)
```
