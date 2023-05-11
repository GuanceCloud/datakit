### `append()` {#fn-append}

Function prototype: `fn append(arr, elem) arr`

Function description: Add the element elem to the end of the array arr.

Function parameters:

- `arr`: array
- `elem`: element being added.

Example:

```python
# Example 1
abc = ["1", "2"]
abc = append(abc, 5.1)
# abc = ["1", "2", 5.1]

# Example 2
a = [1, 2]
b = [3, 4]
c = append(a, b)
# c = [1, 2, [3, 4]]
```
