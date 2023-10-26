### `xml()` {#fn-xml}

Function prototype: `fn xml(input: str, xpath_expr: str, key_name)`

Function description: Extract fields from XML through xpath expressions.

Function parameters:

- input: XML to extract
- xpath_expr: xpath expression
- key_name: The extracted data is written to a new key

Example one:

```python
# data to be processed
        <entry>
         <fieldx>valuex</fieldx>
         <fieldy>...</fieldy>
         <fieldz>...</fieldz>
         <field array>
             <fielda>element_a_1</fielda>
             <fielda>element_a_2</fielda>
         </fieldarray>
     </entry>

# process script
xml(_, '/entry/fieldarray//fielda[1]/text()', field_a_1)

# process result
{
   "field_a_1": "element_a_1", # extracted element_a_1
   "message": "\t\t\u003centry\u003e\n \u003cfieldx\u003evaluex\u003c/fieldx\u003e\n \u003cfieldy\u003e...\u003c/fieldy\u003e\n \u003cfieldz\u003e...\ u003c/fieldz\u003e\n \u003cfieldarray\u003e\n \u003cfielda\u003eelement_a_1\u003c/fielda\u003e\n \u003cfielda\u003eelement_a_2\u003c/fielda\u003e\n \u003c/fieldarray\n\c\u003 u003e",
   "status": "unknown",
   "time": 1655522989104916000
}
```

Example two:

```python
# data to be processed
<OrderEvent actionCode = "5">
  <OrderNumber>ORD12345</OrderNumber>
  <VendorNumber>V11111</VendorNumber>
</OrderEvent>

# process script
xml(_, '/OrderEvent/@actionCode', action_code)
xml(_, '/OrderEvent/OrderNumber/text()', OrderNumber)

# process result
{
   "OrderNumber": "ORD12345",
   "action_code": "5",
   "message": "\u003cOrderEvent actionCode = \"5\"\u003e\n \u003cOrderNumber\u003eORD12345\u003c/OrderNumber\u003e\n \u003cVendorNumber\u003eV11111\u003c/VendorNumber\n\u003e\u003e"
   "status": "unknown",
   "time": 1655523193632471000
}
```
