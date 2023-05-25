### `cover()` {#fn-cover}

Function prototype: `fn cover(key: str, range: list)`

Function description: Perform data desensitization by range on the string data obtained on the specified field

Function parameters:

- `key`: Key name
- `range`: The index range of the desensitized string (`[start,end]`) Both start and end support negative subscripts, which are used to express the semantics of tracing back from the end. The interval is reasonable. If end is greater than the maximum length of the string, it will default to the maximum length

Example:

```python
# input data {"str": "13789123014"}
json(_, `str`)
cover(`str`, [8, 9])

# input data {"abc": "13789123014"}
json(_, abc)
cover(abc, [2, 4])
```
