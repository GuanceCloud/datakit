### `pt_kvs_del()` {#fn_pt_kvs_del}

Function prototype: `fn pt_kvs_del(name: str)`

Function description: Delete the key specified in Point

Function parameters:

- `name`: Key to be deleted

Example:

```python
key_blacklist = ["k1", "k2", "k3"]
for k in pt_kvs_keys() {
    if k in key_blacklist {
        pt_kvs_del(k)
    }
}
```
