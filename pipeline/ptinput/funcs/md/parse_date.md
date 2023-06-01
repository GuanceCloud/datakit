### `parse_date()` {#fn-parse-date}

函数原型：`fn parse_date(key: str, yy: str, MM: str, dd: str, hh: str, mm: str, ss: str, ms: str, zone: str)`

函数说明：将传入的日期字段各部分的值转化为时间戳

函数参数

- `key`: 新插入的字段
- `yy` : 年份数字字符串，支持四位或两位数字字符串，为空字符串，则处理时取当前年份
- `MM`:  月份字符串，支持数字，英文，英文缩写
- `dd`: 日字符串
- `hh`: 小时字符串
- `mm`: 分钟字符串
- `ss`: 秒字符串
- `ms`: 毫秒字符串
- `us`: 微秒字符串
- `ns`: 纳秒字符串
- `zone`: 时区字符串，“+8”或\"Asia/Shanghai\"形式

示例：

```python
parse_date(aa, "2021", "May", "12", "10", "10", "34", "", "Asia/Shanghai") # 结果 aa=1620785434000000000

parse_date(aa, "2021", "12", "12", "10", "10", "34", "", "Asia/Shanghai") # 结果 aa=1639275034000000000

parse_date(aa, "2021", "12", "12", "10", "10", "34", "100", "Asia/Shanghai") # 结果 aa=1639275034000000100

parse_date(aa, "20", "February", "12", "10", "10", "34", "", "+8") 结果 aa=1581473434000000000
```
