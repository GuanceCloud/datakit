### `trim()` {#fn-trim}

Function prototype: `fn trim(key, cutset: str = "")`

Function description: delete the characters specified at the beginning and end of the key, and delete all blank characters by default when the `cutset` is an empty string

Function parameters:

- `key`: a field that has been extracted, string type
- `cutset`: Delete the first and last characters in the `cutset` string in the key

Example:

```python
# Data to be processed: "trim(key, cutset)"

# process script
add_key(test_data, "ACCAA_test_DataA_ACBA")
trim(test_data, "ABC_")

# process result
{
   "test_data": "test_Data"
}
```
