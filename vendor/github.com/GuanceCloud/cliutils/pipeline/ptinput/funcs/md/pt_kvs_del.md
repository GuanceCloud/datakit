### `pt_kvs_del()` {#fn_pt_kvs_del}

函数原型：`fn pt_kvs_del(name: str)`

函数说明：删除 Point 中指定的 key

函数参数：

- `name`: 待删除的 key

示例：

```python
key_blacklist = ["k1", "k2", "k3"]
for k in pt_kvs_keys() {
    if k in key_blacklist {
        pt_kvs_del(k)
    }
}
```
