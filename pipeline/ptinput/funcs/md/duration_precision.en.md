### `duration_precision()` {#fn-duration-precision}

Function prototype: `fn duration_precision(key, old_precision: str, new_precision: str)`

Function description: Perform duration precision conversion, and specify the current precision and target precision through Function parameters:. Support conversion between s, ms, us, ns.

Example:

```python
# in << {"ts":12345}
json(_, ts)
cast(ts, "int")
duration_precision(ts, "ms", "ns")

# Extracted data(drop: false, cost: 33.279µs):
# {
#   "message": "{\"ts\":12345}",
#   "ts": 12345000000
# }
```
