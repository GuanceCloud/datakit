### `drop_origin_data()` {#fn-drop-origin-data}

函数原型：`fn drop_origin_data()`

函数说明：丢弃初始化文本，否则初始文本放在 message 字段中

示例：

```python
# 待处理数据：{"age": 17, "name": "zhangsan", "height": 180}

# 结果集中删除 message 内容
drop_origin_data()
```

