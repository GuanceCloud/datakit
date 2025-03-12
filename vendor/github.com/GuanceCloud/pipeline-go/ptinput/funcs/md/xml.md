### `xml()` {#fn-xml}

函数原型：`fn xml(input: str, xpath_expr: str, key_name)`

函数说明：通过 xpath 表达式，从 XML 中提取字段。

参数：

- input: 待提取的 XML
- xpath_expr: xpath 表达式
- key_name: 提取后数据写入新 key

示例一：

```python
# 待处理数据
       <entry>
        <fieldx>valuex</fieldx>
        <fieldy>...</fieldy>
        <fieldz>...</fieldz>
        <fieldarray>
            <fielda>element_a_1</fielda>
            <fielda>element_a_2</fielda>
        </fieldarray>
    </entry>

# 处理脚本
xml(_, '/entry/fieldarray//fielda[1]/text()', field_a_1)

# 处理结果
{
  "field_a_1": "element_a_1",  # 提取了 element_a_1
  "message": "\t\t\u003centry\u003e\n        \u003cfieldx\u003evaluex\u003c/fieldx\u003e\n        \u003cfieldy\u003e...\u003c/fieldy\u003e\n        \u003cfieldz\u003e...\u003c/fieldz\u003e\n        \u003cfieldarray\u003e\n            \u003cfielda\u003eelement_a_1\u003c/fielda\u003e\n            \u003cfielda\u003eelement_a_2\u003c/fielda\u003e\n        \u003c/fieldarray\u003e\n    \u003c/entry\u003e",
  "status": "unknown",
  "time": 1655522989104916000
}
```

示例二：

```python
# 待处理数据
<OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
</OrderEvent>

# 处理脚本
xml(_, '/OrderEvent/@actionCode', action_code)
xml(_, '/OrderEvent/OrderNumber/text()', OrderNumber)

# 处理结果
{
  "OrderNumber": "ORD12345",
  "action_code": "5",
  "message": "\u003cOrderEvent actionCode = \"5\"\u003e\n \u003cOrderNumber\u003eORD12345\u003c/OrderNumber\u003e\n \u003cVendorNumber\u003eV11111\u003c/VendorNumber\u003e\n\u003c/OrderEvent\u003e",
  "status": "unknown",
  "time": 1655523193632471000
}
```
