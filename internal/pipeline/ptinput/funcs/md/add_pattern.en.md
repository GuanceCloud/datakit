### `add_pattern()` {#fn-add-pattern}

Function prototype: `fn add_pattern(name: str, pattern: str)`

Function description: Create custom grok patterns. The grok pattern has scope restrictions, such as a new scope will be generated in the if else statement, and the pattern is only valid within this scope. This function cannot overwrite existing grok patterns in the same scope or in the previous scope

Function parameters:

- `name`: pattern naming
- `pattern`: custom pattern content

Example:

```python
# input data: "11,abc,end1", "22,abc,end1", "33,abc,end3"

# script
add_pattern("aa", "\\d{2}")
grok(_, "%{aa:aa}")
if false {

} else {
    add_pattern("bb", "[a-z]{3}")
    if aa == "11" {
        add_pattern("cc", "end1")
        grok(_, "%{aa:aa},%{bb:bb},%{cc:cc}")
    } elif aa == "22" {
        # Using pattern cc here will cause compilation failure: no pattern found for %{cc}
        grok(_, "%{aa:aa},%{bb:bb},%{INT:cc}")
    } elif aa == "33" {
        add_pattern("bb", "[\\d]{5}") # Overwriting bb here fails
        add_pattern("cc", "end3")
        grok(_, "%{aa:aa},%{bb:bb},%{cc:cc}")
    }
}

# result
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
