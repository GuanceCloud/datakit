### `match()` {#fn-match}

Function prototype: `fn match(pattern: str, s: str) bool`

Function description: Use the specified regular expression to match the string, return true if the match is successful, otherwise return false

Function parameters:

- `pattern`: regular expression
- `s`: string to match

Example:

```python
# script
test_1 = "pattern 1,a"
test_2 = "pattern -1,"

add_key(match_1, match('''\w+\s[,\w]+''', test_1)) 

add_key(match_2, match('''\w+\s[,\w]+''', test_2)) 

# result
{
    "match_1": true,
    "match_2": false
}
```
