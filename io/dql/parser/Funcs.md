

## 一、名词

- `M` - 指时序数据中的指标集
- `L` - 日志数据，以字段 `source` 作为逻辑意义上的分类
- `O` - 对象数据，以字段 `class` 作为逻辑意义上的分类
- `E` - 事件数据，以字段 `source` 作为逻辑意义上的分类
- `T` - 追踪数据，以字段 `service` 作为逻辑意义上的分类
- `R` - RUM 数据，以字段 `source` 作为逻辑意义上的分类
- `S` - security数据，以字段`category`作为逻辑意义上的分类

## 二、show 函数列表

### 2.1 object数据

show_object_source() 

说明: 

展示 object 数据的指标集合，该函数不需要参数

示例: 

```golang
// (1)请求
show_object_source()

// (2)返回
{
	"content": [
		{
			"series": [
				{
					"name": "measurements",
					"columns": [
						"name"
					],
					"values": [
						[
							"Servers"
						]
					]
				}
			],
			"cost": "",
			"raw_query": ""
		}
	]
}

```


show_object_class() 

说明: 

展示 object 数据的指标集合，该函数不需要参数,

注意: 

将遗弃，使用 show_object_source() 代替

show_object_field(`source_value`) 

说明: 

展示 source_value 指标下的所有 fileds 列表

示例: 

```golang
// (1)请求
show_object_field(Servers)

// (2)返回
{
	"content": [
		{
			"series": [
				{
					"name": "fields",
					"columns": [
						"fieldKey",
						"fieldType"
					],
					"values": [
						[
							"__class",
							"keyword"
						]
					]
				}
			],
			"cost": "",
			"raw_query": ""
		}
	]
}

```

 

### 2.2 logging数据

show_logging_source()

说明: 

展示日志数据的指标集合，该函数不需要参数

示例: 

show_logging_source(), 返回结构同 show_object_source()

show_logging_field(`source_value`) 

说明: 

展示 source_value 指标下的所有 fileds 列表

示例: 

show_logging_field(nginx), 返回结构同 show_object_field(Servers)

### 2.3 keyevent

show_event_source()

说明: 

展示keyevent数据的指标集合，该函数不需要参数

示例: 

show_event_source(), 返回结构同 show_object_source()

show_event_field(`source_value`) 

说明: 

展示 source_value 指标下的所有 fields 列表

示例: 

show_event_field(datafluxTrigger), 返回结构同 show_object_field(Servers)

### 2.4 tracing数据

show_tracing_source()

说明: 

展示 tracing 数据的指标集合，该函数不需要参数

示例: 

show_tracing_source(), 返回结构同 show_object_source()

show_tracing_service()

说明: 

展示 tracing 数据的指标集合，该函数不需要参数

注意: 

将遗弃，使用 show_tracing_source() 代替

show_tracing_field(`source_value`) 

说明: 

展示 source_value 指标下的所有 fields 列表

示例: 

show_tracing_field(mysql), 返回结构同 show_object_field(Servers)

### 2.5 rum数据

show_rum_source()

说明: 

展示rum数据的指标集合，该函数不需要参数

示例: 

show_rum_source(), 返回结构同 show_object_source()

show_rum_type()

说明: 

展示rum数据的指标集合，该函数不需要参数

注意: 

将遗弃，使用 show_rum_source() 代替

show_rum_field(`source_value`) 

说明: 

展示 source_value 指标下的所有 fields 列表

示例: 

show_rum_field(js_error), 返回结构同 show_object_field(Servers)

### 2.6 时序数据

show_measurement()  

说明: 

展示时序数据的指标集合

示例: 

show_measurement(), 返回结构同 show_object_source()

show_tag_key()      

说明: 

查看指标集 tag 列表, 可以指定具体的指标

示例: 

```golang

// (1) 请求
show_tag_key(from=['cpu'])

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "cpu",
          "columns": [
            "tagKey"
          ],
          "values": [
            [
              "cpu"
            ],
            [
              "host"
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

show_tag_value()    

说明: 

返回数据库中指定 tag key 的 tag value 列表

示例: 

```golang

// (1) 请求
show_tag_value(from=['cpu'], keyin=['host'])

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "cpu",
          "columns": [
            "key",
            "value"
          ],
          "values": [
            [
              "host",
              "jydubuntu"
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}
```

show_field_key()    

说明: 

查看指标集 field-key 列表

示例: 

show_field_key(from=['cpu']), 返回结构同 show_object_field(Servers)

## 三、函数列表

### avg

说明: 

返回字段的平均值

注意: 

avg函数应用的字段需要是数值类型，

如果该字段 field1 值为 '10' 字符串类型，可以使用 avg(int(field1))实现

如果该字段 field1 值为 '10.0' 字符串类型，可以使用 avg(float(field1))实现

场景: 

(1) `avg(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例: 

```golang
// (1) 请求
L::nginx:(avg(connect_total)) {__errorCode='200'}

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginx",
          "columns": [
            "time",
            "avg_connect_total"
          ],
          "values": [
            [
              null,
              50.16857454347234
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### bottom

说明: 

返回最小的 n 个 field 值

场景: 

(1) `bottom(field1, n)`

参数有且只有两个，

第一个参数表示字段名称，类型为字符串

第二个参数表示返回数量，类型为数值

注意: 

field1不能是time字段，即bottom(time, 10)不支持，可以使用query实现

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例: 

```golang

// (1) 请求
L::nginx:(bottom(host, 2)) {__errorCode='200'}

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginx",
          "columns": [
            "time",
            "host"
          ],
          "values": [
            [
              1609154974839,
              "csoslinux"
            ],
            [
              1609154959048,
              "csoslinux"
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### count

说明: 

返回非空字段值的汇总值

场景:

(1) `count(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

(2) `count(func1())`

参数可以是一个内置函数，例如: `count(distinct(field1))`, 适用范围是 `M`

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例: 

```golang

// (1) 请求
L::nginx:(count(host)) {__errorCode='200'}

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginx",
          "columns": [
            "time",
            "count_host"
          ],
          "values": [
            [
              null,
              36712
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### count_distinct

说明: 

统计字段不同值的数量

场景:

(1) `count_distinct(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求
L::nginx:(count_distinct(host)) {__errorCode='200'}

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginx",
          "columns": [
            "time",
            "count_distinct(host)"
          ],
          "values": [
            [
              null,
              3
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```


### derivative

说明: 

返回字段的相邻两个点的变化率

场景:

(1) `derivative(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

适用范围: 

`M`

注意: `field1`对应值为数值类型

示例: 

```golang

// (1) 请求
M::cpu:(derivative(usage_idle)) limit 2

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "cpu",
          "columns": [
            "time",
            "derivative"
          ],
          "values": [
            [
              1608612970000,
              -0.06040241121018255
            ],
            [
              1608612980000,
              0.020079912763694096
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### difference

说明:

差值

适用范围: 

`M`

示例:

```golang 

// (1) 请求
M::cpu:(difference(usage_idle)) limit 2

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "cpu",
          "columns": [
            "time",
            "difference"
          ],
          "values": [
            [
              1608612970000,
              -0.6040241121018255
            ],
            [
              1608612980000,
              0.20079912763694097
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### distinct

说明: 

返回`field value`的不同值列表

场景:

(1) `distinct(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求
R::js_error:(distinct(error_message))

// (2) 返回
{
	"content": [
		{
			"series": [
				{
					"name": "js_error",
					"columns": [
						"time",
						"distinct_error_message"
					],
					"values": [
						[
							null,
							"sdfs is not defined"
						],
						[
							null,
							"xxxxxxx console error:"
						]
					]
				}
			],
			"cost": "",
			"raw_query": ""
		}
	]
}
```


### exists

说明:

文档中，指定字段必须存在

场景:

(1) `field1=exists()`

不需要参数

适用范围: 

`L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求, 注意: exists函数位于filter-clause语句中
rum::js_error:(sdk_name, error_message){sdk_name=exists()} limit 1

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "js_error",
          "columns": [
            "time",
            "sdk_name",
            "error_message"
          ],
          "values": [
            [
              1609227006093,
              "小程序 SDK",
              "sdfs is not defined"
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}
```

### first

说明:

返回时间戳最早的值

场景:

(1) `first(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

注意: 

field1不能是time字段，即first(time)无意义

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求
L::nginx:(first(host)) {__errorCode='200'}

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginx",
          "columns": [
            "time",
            "host"
          ],
          "values": [
            [
              1609837113498,
              "wangjiaoshou"
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### float

说明: 

cast 函数，将 string 类型数据转为 float 数值

场景: 

(1) `avg(float(fieldName))`

参数有且只有一个，参数值为字段名称, 该函数只能应用于`sum`, `max`, `min`, `avg`中，作为嵌套内层函数使用，即`float(fieldName)`目前不支持

适用范围: 

`L`, `O`, `E`, `T`, `R`

### histogram

说明:

直方图，范围聚合

场景:

(1) `histogram(fieldName, startValue, endValue, interval, minDoc)`

必填参数，

a. `fieldName` x轴对应的字段名称

b. `startValue` x轴最小值边界

c. `endValue` x轴最大值边界

d. `interval` 间隔范围

缺省参数，

e. `minDoc` 表示低于`minDoc`的值不返回


适用范围: 

`L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求
E::`monitor`:(histogram(date_range, 300, 6060, 100, 1))

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "monitor",
          "columns": [
            "time", // 字段名称为time，但是实际表示y轴的数值
            "histogram(date_range, 300, 6060, 100, 1)"
          ],
          "values": [
            [
              300,
              11183
            ],
            [
              600,
              93
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": "",
      "total_hits": 10000,
      "group_by": null
    }
  ]
}
```

### int

说明:

cast 函数，将 string 类型数据转为 int 数值

场景:

(1) `avg(int(fieldName))`

参数有且只有一个，参数值为字段名称, 该函数只能应用于`sum`, `max`, `min`, `avg`中，作为嵌套内层函数使用，即`int(fieldName)`目前不支持

适用范围: 

`L`, `O`, `E`, `T`, `R`

### last

说明:

返回时间戳最近的值

场景:

(1) `last(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

注意: field1不能是time字段，即last(time)无意义

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例: 

L::nginx:(last(host)) {__errorCode='200'}, 返回结构同 first 函数

### log

说明:

求对数

适用范围: 

`M`

示例:

```golang

// (1) 请求
M::cpu:(log(usage_idle, 10)) limit 2

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "cpu",
          "columns": [
            "time",
            "log"
          ],
          "values": [
            [
              1608612960000,
              1.9982417203437028
            ],
            [
              1608612970000,
              1.995599815632755
            ]
          ]
        }
      ],
      "cost": " ",
      "raw_query": ""
    }
  ]
}

```

### match

说明:

全文搜索（模糊搜索）

场景:

(1) `field1=match(field1_value)`

参数有且只有一个, 表示查询的字段值

适用范围: 

`L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求, 注意: match函数位于filter-clause语句中
rum::js_error:(sdk_name, error_message){error_message=match('not defined')} limit 1

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "js_error",
          "columns": [
            "time",
            "sdk_name",
            "error_message"
          ],
          "values": [
            [
              1609227006093,
              "小程序 SDK",
              "sdfs is not defined"
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}
```

### max

说明:

返回最大的字段值

场景:

(1) `max(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求
L::nginx:(max(connect_total)) {__errorCode='200'}

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginx",
          "columns": [
            "time",
            "max_connect_total"
          ],
          "values": [
            [
              null,
              99
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### min

说明:

返回最小的字段值

场景:

(1) `min(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例:

L::nginx:(min(connect_total)) {__errorCode='200'}, 返回结构同 max 函数

### moving_average

说明:

平均移动

适用范围: 

`M`

示例:

```golang

// (1) 请求
M::cpu:(moving_average(usage_idle, 2)) limit 2

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "cpu",
          "columns": [
            "time",
            "moving_average"
          ],
          "values": [
            [
              1608612970000,
              99.29394753991822
            ],
            [
              1608612980000,
              99.09233504768578
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### non_negative_derivative

说明:

数据的非负变化率

适用范围: 

`M`

示例:

```golang

// (1) 请求
M::cpu:(non_negative_derivative(usage_idle)) limit 2

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "cpu",
          "columns": [
            "time",
            "non_negative_derivative"
          ],
          "values": [
            [
              1608612980000,
              0.020079912763694096
            ],
            [
              1608613000000,
              0.010417976581746303
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### percentile

说明:

返回较大百分之 n 的字段值

适用范围: 

`M`

示例:

```golang

// (1) 请求
M::cpu:(percentile(usage_idle, 5)) limit 2

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "cpu",
          "columns": [
            "time",
            "percentile"
          ],
          "values": [
            [
              1609133610000,
              97.75280898882501
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```


### re

说明:

查询时候，正则过滤

场景:

(1) `field1=re(field1_value)`

参数有且只有一个, 表示查询的字段值

适用范围: `M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求, 注意: re函数位于filter-clause语句中
rum::js_error:(sdk_name, error_message){error_message=re('.*not defined.*')} limit 1

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "js_error",
          "columns": [
            "time",
            "sdk_name",
            "error_message"
          ],
          "values": [
            [
              1609227006093,
              "小程序 SDK",
              "sdfs is not defined"
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}
```

### queryString

说明:

字符串查询,

dql将使用特殊语法解析器，解析输入的字符串，查询文档

场景:

(1) 普通的全文查询

`field1=queryString(field1_value)`

参数有且只有一个, 表示查询的字段值, 类似于上面的函数 match

(2) 查询条件逻辑组合

`status=queryString("info OR warnning")`

逻辑操作符为(需要使用大写字符串): 

a. `OR`

b. `AND`, 默认值

默认的逻辑操作符为 `OR`, 查询字符串中 空格, 逗号都表示逻辑与关系

(3) 通配查询

`message=queryString("error*")`

`message=queryString("error?")`

通配字符 `*`表示匹配 0 或多个任意字符，`?` 表示匹配1个任意字符

适用范围: 

`L`, `O`, `E`, `T`, `R`

详见: 

[elasticsearch queryString](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html)

示例:

```golang

// (1) 请求
L::datakit:(host,message) {message=queryString('/[telegraf|GIN]/ OR /[rum|GIN]/')} limit 1

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "datakit",
          "columns": [
            "time",
            "host",
            "message"
          ],
          "values": [
            [
              1616412168015,
              "aaffb5b0ce0b",
              "[GIN] 2021/03/22 - 11:22:48 | 500 |         2m52s |   210.26.121.20 | POST     \"/v1/write/rum?precision=ms\""
            ]
          ]
        }
      ],
      "cost": "26ms",
      "raw_query": "",
      "total_hits": 12644,
      "group_by": null
    }
  ]
}
```

### sum

说明:

返回字段值的和

场景:

(1) `sum(field1)`

参数有且只有一个，参数类型是字符串，参数值为字段名称

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang

// (1) 请求
L::nginx:(sum(connect_total)) {__errorCode='200'}

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginx",
          "columns": [
            "time",
            "sum_connect_total"
          ],
          "values": [
            [
              null,
              1844867
            ]
          ]
        }
      ],
      "cost": "",
      "raw_query": ""
    }
  ]
}

```

### top

说明:

返回最大的 n 个 field 值

场景:

(1) `top(field1, n)`

参数有且只有两个，

第一个参数表示字段名称，类型为字符串

第二个参数表示返回数量，类型为数值

注意: 

field1不能是time字段，即top(time, 10)不支持，可以使用query实现

适用范围: 

`M`, `L`, `O`, `E`, `T`, `R`

示例:

L::nginx:(top(host, 2)) {__errorCode='200'}, 返回结构同 bottom 函数


### wildcard

说明:

通配查询


场景:

通配查询

`field1=wildcard(field1_value)`

通配字符 `*`表示匹配 0 或多个任意字符，`?` 表示匹配1个任意字符

适用范围: 

`L`, `O`, `E`, `T`, `R`

详见: 

[elasticsearch wildcard query](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-query-string-query.html)

示例:

```golang

// (1) 请求
L::datakit:(host,message) {message=wildcard('*write*')} limit 1

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "datakit",
          "columns": [
            "time",
            "host",
            "message"
          ],
          "values": [
            [
              1616412168015,
              "aaffb5b0ce0b",
              "[GIN] 2021/03/22 - 11:22:48 | 500 |         2m52s |   210.26.121.20 | POST     \"/v1/write/rum?precision=ms\""
            ]
          ]
        }
      ],
      "cost": "26ms",
      "raw_query": "",
      "total_hits": 12644,
      "group_by": null
    }
  ]
}
```
