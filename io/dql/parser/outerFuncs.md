## 一、dql 外层函数说明

基本语法规则

outerFunc1(参数列表).outerFunc2(参数列表)

外层函数可以链接调用

## 二、外层函数列表

### abs

说明:

计算处理集每个元素的绝对值

场景:

(1) `abs(dql查询字符串)` 或者 `abs(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `abs()`

当作为链接调用的非第 1 个函数时，

不需要参数

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
abs(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614566985356,
              8004050000
            ],
            [
              1614566982596,
              79325000
            ],
            [
              1614566922891,
              90110000
            ]
          ]
        }
      ],
      "cost": "43.333168ms",
      "group_by": null
    }
  ]
}

```

### avg

说明:

计算处理集的平均值

场景:

(1) `avg(dql查询字符串)` 或者 `avg(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `avg()`

当作为链接调用的非第 1 个函数时，

不需要参数

注意:

返回值 time 列的值为 0

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
avg(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              0,
              2674188333.3333335
            ]
          ]
        }
      ],
      "cost": "43.380748ms",
      "group_by": null
    }
  ]
}
```


### count

说明:

对返回结果，统计数量

场景:

(1) `count(dql查询字符串)` 或者 `count(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `count()`

当作为链接调用的非第 1 个函数时，

不需要参数

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
count(`L::nginxlog:(status) {client_ip='127.0.0.1'}`)

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginxlog",
          "columns": [
            "time",
            "status"
          ],
          "values": [
            [
              0,
              20
            ]
          ]
        }
      ],
      "cost": "21.159579ms",
      "group_by": null
    }
  ]
}

```

### count_distinct

说明:

对处理集, 去重统计数量

场景:

(1) `count_distinct(dql查询字符串)` 或者 `count_distinct(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `count_distinct()`

当作为链接调用的非第 1 个函数时，

不需要参数

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
count_distinct(`L::nginxlog:(status) {client_ip='127.0.0.1'}`)

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "nginxlog",
          "columns": [
            "time",
            "status"
          ],
          "values": [
            [
              0,
              2
            ]
          ]
        }
      ],
      "cost": "21.159579ms",
      "group_by": null
    }
  ]
}

```

### cumsum

说明:

对处理集累计求和

场景:

(1) `cumsum(dql查询字符串)` 或者 `cumsum(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `cumsum()`

当作为链接调用的非第 1 个函数时，

不需要参数

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
cumsum(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614250017606,
              985025000
            ],
            [
              1614250017602,
              1104300000
            ],
            [
              1614250017599,
              2253690000
            ]
          ]
        }
      ],
      "cost": "25.468929ms",
      "group_by": null
    }
  ]
}


```

### derivative

说明:

计算处理集相邻元素的导数

场景:

(1) `derivative(dql查询字符串)` 或者 `derivative(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `derivative()`

当作为链接调用的非第 1 个函数时，

不需要参数

注意:

求导的时间单位为 秒（s)

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
derivative(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614250560828,
              2159233.6119981343
            ],
            [
              1614250560818,
              -11357500000
            ]
          ]
        }
      ],
      "cost": "24.817991ms",
      "group_by": null
    }
  ]
}

```

### difference

说明:

计算处理集相邻元素的差值

场景:

(1) `difference(dql查询字符串)` 或者 `difference(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `difference()`

当作为链接调用的非第 1 个函数时，

不需要参数

注意:

处理集至少大于一行，否则返回空值

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
difference(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614250788967,
              88595000
            ],
            [
              1614250788854,
              -89940000
            ]
          ]
        }
      ],
      "cost": "24.738317ms",
      "group_by": null
    }
  ]
}

```

### first

说明:

计算处理集的最早有意义的值

场景:

(1) `first(dql查询字符串)` 或者 `first(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `first()`

当作为链接调用的非第 1 个函数时，

不需要参数

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
first(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614567497285,
              8003885000
            ]
          ]
        }
      ],
      "cost": "34.99329ms",
      "group_by": null
    }
  ]
}
```

### last

说明:

计算处理集的最近有意义的值

场景:

(1) `last(dql查询字符串)` 或者 `last(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `last()`

当作为链接调用的非第 1 个函数时，

不需要参数

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
last(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614567720225,
              50490000
            ]
          ]
        }
      ],
      "cost": "35.016794ms",
      "group_by": null
    }
  ]
}
```

### log10

说明:

计算处理集每个元素的 log10 值

场景:

(1) `log10(dql查询字符串)` 或者 `log10(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `log10()`

当作为链接调用的非第 1 个函数时，

不需要参数

注意:

处理集至少大于一行，否则返回空值

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
log10(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614251956207,
              7.317750028842234
            ],
            [
              1614251955227,
              8.191939809656507
            ],
            [
              1614251925530,
              8.133810257633591
            ]
          ]
        }
      ],
      "cost": "717.257675ms",
      "group_by": null
    }
  ]
}

```

### log2

说明:

计算处理集每个元素的 log2 值

场景:

(1) `log2(dql查询字符串)` 或者 `log2(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `log2()`

当作为链接调用的非第 1 个函数时，

不需要参数

注意:

处理集至少大于一行，否则返回空值

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
log2(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614251925530,
              27.019932813316046
            ],
            [
              1614251865510,
              26.439838744891972
            ],
            [
              1614251805516,
              29.703602660685803
            ]
          ]
        }
      ],
      "cost": "1.01630157s",
      "group_by": null
    }
  ]
}

```

### max

说明:

计算处理集的最大值

场景:

(1) `max(dql查询字符串)` 或者 `max(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `max()`

当作为链接调用的非第 1 个函数时，

不需要参数

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
max(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614567387278,
              1006975000
            ]
          ]
        }
      ],
      "cost": "43.857171ms",
      "group_by": null
    }
  ]
}
```

### min

说明:

计算处理集的最小值

场景:

(1) `min(dql查询字符串)` 或者 `min(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `min()`

当作为链接调用的非第 1 个函数时，

不需要参数

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
min(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614567507202,
              86480000
            ]
          ]
        }
      ],
      "cost": "42.551151ms",
      "group_by": null
    }
  ]
}
```

### moving_average

说明:

计算处理集的移动平均值

场景:

(1) `moving_average(dql查询字符串, 3)` 或者 `moving_average(dql=dql查询字符串, size=3)`

当作为链接调用的第 1 个函数时，

参数有且只有两个，

第一个参数表示 dql 查询，类型为字符串

第二个参数表示窗口大小，类型为数值

(2) `moving_average(size=3)` , `moving_average(3)`

当作为链接调用的非第 1 个函数时，

参数有且只有 1 个，表示窗口大小，类型为数值

注意:

窗口的大小需要不小于处理集的行数，否则返回空值

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
moving_average(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`,size=2)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614251505520,
              106675000
            ],
            [
              1614251445542,
              102757500
            ]
          ]
        }
      ],
      "cost": "24.738867ms",
      "group_by": null
    }
  ]
}

```

### non_negative_derivative

说明:

计算处理集相邻元素的非负导数

场景:

(1) `non_negative_derivative(dql查询字符串)` 或者 `non_negative_derivative(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `non_negative_derivative()`

当作为链接调用的非第 1 个函数时，

不需要参数

注意:

求导的时间单位为 秒（s)

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
non_negative_derivative(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)


// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614250901131,
              234697986.57718122
            ]
          ]
        }
      ],
      "cost": "25.706837ms",
      "group_by": null
    }
  ]
}

```

### non_negative_difference

说明:

计算处理集相邻元素的非负差值

场景:

(1) `non_negative_difference(dql查询字符串)` 或者 `non_negative_difference(dql=dql查询字符串)`

当作为链接调用的第 1 个函数时，

参数有且只有 1 个，

第一个参数表示 dql 查询，类型为字符串

(2) `non_negative_difference()`

当作为链接调用的非第 1 个函数时，

不需要参数

注意:

处理集至少大于一行，否则返回空值

适用范围:

`M`, `L`, `O`, `E`, `T`, `R`

示例:

```golang
// (1) 请求
non_negative_difference(dql=`R::resource:(resource_load) {resource_load > 100} limit 3`)

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614250900989,
              87595000
            ]
          ]
        }
      ],
      "cost": "23.694907ms",
      "group_by": null
    }
  ]
}

```

### 链接调用

```golang
// (1) 请求
difference(dql=`R::resource:(resource_load) {resource_load > 100} [1614239472:1614239531] limit 3`).cumsum()

// (2) 返回
{
  "content": [
    {
      "series": [
        {
          "name": "resource",
          "columns": [
            "time",
            "resource_load"
          ],
          "values": [
            [
              1614239530215,
              -364330000
            ],
            [
              1614239530135,
              -889240000
            ]
          ]
        }
      ],
      "cost": "16.873202ms",
      "group_by": null
    }
  ]
}

```
