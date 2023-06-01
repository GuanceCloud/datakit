### `append()` {#fn-append}

函数原型：`fn append(arr, elem) arr`

函数说明：往数组 arr 末尾添加元素 elem。

参数：

- `arr`: 要添加元素的数组。
- `elem`: 添加的元素。

示例：

```python
# 例 1
abc = ["1", "2"]
abc = append(abc, 5.1)
# abc = ["1", "2", 5.1]

# 例 2
a = [1, 2]
b = [3, 4]
c = append(a, b)
# c = [1, 2, [3, 4]]
```
