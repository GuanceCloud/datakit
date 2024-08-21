### `hash()` {#fn_hash}

Function prototype: `fn hash(text: str, method: str) -> str`

Function description: Calculate the hash of the text

Function parameters:

- `text`: input text
- `method`: Hash algorithm, allowing values including `md5`, `sha1`, `sha256`, `sha512`

Example:

```python
pt_kvs_set("md5sum", hash("abc", "sha1"))
```
