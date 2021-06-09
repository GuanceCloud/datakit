# dql 定义

以下是 DataFlux 查询语言（dql）定义。随着不同语法的逐步支持，该文档会做不同程度的调整和增删。

全局约束如下：

- 非关键字（如指标名、标签名等）大小写敏感，**关键字及函数名大小写不敏感**

- 以 `#` 为行注释字符，不支持行内注释

- 支持的操作符：

	- `+`  - 加法
	- `-`  - 减法
	- `*`  - 乘法
	- `/`  - 除法
	- `%`  - 取模
	- `=` - 等于
	- `!=` - 不等于
	- `<=` - 大于等于
	- `<` - 小于
	- `>=` - 大于等于
	- `>` -  大于
	- `^` - 指数运算
	- `&&` - 逻辑与
	- `||` - 逻辑或

- 支持的关键字：

```
AND AS ASC AUTO
BY DESC FALSE FILTER
LIMIT LINK WITH LINEAR
NIL OFFSET OR PREVIOUS
SLIMIT SOFFSET TRUE WITH
```

- 标识符：标识符有几种形式，便于兼容各种变量命名形式

	- 正常变量名中只能出现 `[_a-zA-Z0-9]` 这些字符，且首字符不能是数字。如 `_abc, _abc123, _123ab`
	- 其它形式的变量名处理方式：
		- `this+is-a*xx/yy^zz?variable`，`by` 需写成 `` `this+is-a*xx/yy^zz?variable` ``，`` `by` ``，前者变量中带运算符，后者的 `by` 是 DQL 关键字
		- `这是一个中文或其它非英文字符变量` 需写成 `` `这是一个中文或其它非英文字符变量` ``
		- 变量中就带了一个反引号，`` this`is-a-vairalbe `` 需写成 `` `identifier("this`is-a-vairalbe")` `` 来修饰

- 字符串值可用双引号和单引号： `"this is a string"` 和 `'this is a string'` 是等价的

- 数据类型：支持浮点（`123.4`, `5.67E3`）、整形（`123`, `-1`）、字符串（`'张三'`, `"hello world"`）、Boolean（`true`, `false`）四种类型

- 特殊函数
	- `re('regex')` - 表示正则表达式，如 `re("*abc")`。对于比较复杂的正则，可用 `` `complex-regex` `` 这种形式来避免对 `'` 和 `"` 的转义。

	- `tz()` - 时区，有两种形式支持
	
		- `tz(+-12)` 以 24 个时区的偏移来指定，如 `tz(+8),tz(8), tz('Asia/Shanghai')` 是一样的，夏令时不能通过这种形式来指定。
		- `tz('Asia/Shanghai')` 以国际标准形式来指定时区。对于夏令时，只能通过这种形式来指定。

	- `identifier()` 用于修饰变量名中带 `` ` `` 字符的变量

	- `int()` 和 `float()` 对返回的数据做类型转换，仅适用于时序数据。

## 查询

查询遵循如下的语法范式，注意，各个部分之间的相对顺序不能调换，如 `time-expr` 不能出现在 `filter-clause` 之前。

```
namespace::
	data-source
	target-clause
	filter-clause
	time-expr
	by-clause
	limit-clause
	offset-clause
	slimit-clause
	soffset-clause
```

从语法角度而言， `data-source` 是必须的（类似于 SQL 中的 `FROM` 子句），其它部分都是可选的。但实际查询过程中，对查询的实际执行会施加一定约束（比如 `time_expr` 不允许时间跨度太大）

举例：

```python
# 获取指标集 cpu 最近 5 分钟所有字段的数据
M::cpu [5m]}

# 查找匹配正则表达式 *db 的所有指标最近 5 分钟的数据
M::re('*db') [5m]

# 获取指标集 cpu 10 分钟前到 5 分钟前的所有字段数据
M::cpu [10m:5m]

# 获取指标集 cpu 10 分钟前到 5 分钟前的所有字段数据，以 1分钟的间隔来聚合
M::cpu [10m:5m:1m]

# 查询时序数据指标集 cpu 最近 5分钟的两个字段 time_active, time_guest_nice，
# 以 host 和 cpu 两个 tag 来过滤，同时以 host 和 cpu 来分组显示结果。
M:: cpu:(time_active, time_guest_nice)
		{ host = "host-name", cpu = "cpu0" } [5m] BY host,cpu

# 以身高倒排，获取前十
O::human:(height, age) { age > 100, sex = "直男" } ORDER BY height LIMIT 10

M::cpu,mem:(time_active, time_guest_nice, host) { host = "host-name", cpu = "cpu0" } [5m] BY host,cpu
```

注意，`::` 和 `:` 两边都是可以添加空白字符的，如下语句是等价的：

```python
M::cpu:(time_active, time_guest_nice)
	{ host = "host-name", cpu = "cpu0" } [5m]

M   ::cpu : (time_active, time_guest_nice)
	{ host = "host-name", cpu = "cpu0" } [5m]

M   :: cpu :   (time_active, time_guest_nice)
	{ host = "host-name", cpu = "cpu0" } [5m]
```

## 语句

### namespace

语义层面，目前支持以下几种种数据源：

- M/metric - 时序指标数据
- O/object - 对象数据
- L/logging - 日志数据
- E/event - 事件数据
- T/tracing - 追踪数据
- R/rum - RUM 数据
- F/func - Func 函数计算

在语法层面，暂不对数据源做约束。数据源语法如下：

```python
data-source ::
	# 具体查询细节...
```

在具体的查询中，如果不指定数据源，则默认为 `metric` 或 `M`，即时序指标是 DataFlux 的默认数据源。

### target-clause

查询的结果列表：

```python
M::cpu:(time_active, system_usage) {host="biz_prod"} [5m]

# 这里支持同一个指标集上不同指标（类型要基本匹配）之间进行计算
M::cpu:(time_active+1, time_active/time_guest_nice) [5m]
```

### filter-clause

过滤子句用来对结果数据做过滤，类似 SQL 中的 `where` 条件：

```python
# 查询人口对象中（__class=human）中百岁直男的身高
O::human:(height) { age > 100, sex = "直男" }

# 带正则的过滤
O::human:(height) { age > 100, sex != re("男") }

# 带计算表达式的过滤
O::human:(height) { (age + 1)/2 > 31, sex != re("男") }

# 带或运算表达式的过滤
O::human:(height) { age > 31 || sex != re("男"), weight > 70}

# 带聚合的的结果列
M::cpu:(avg(time_active) AS time_active_avg, time_guest_nice) [1d::1h]

# 带聚合填充的的结果列
M::cpu:(fill(avg(time_active) AS time_active_avg, 0.1), time_guest_nice) [1d::1h]

# 带有 in 列表的查询,其中 in 中选项关系为逻辑 or, in 列表中只能是数值或者字符串
O::human:(height) { age in [30, 40, 50], weight > 70}
```

关于填充：

- 数值填充：形如 `cpu:(fill(f1, 123), fill(f2, "foo bar"), fill(f3, 123.456))`
- 线性填充：如 `cpu:(fill(f1, LINEAR))`
- 前值填充：如 `cpu:(fill(f1, PREVIOUS))`

> 注意：多个过滤条件之间。默认是 `AND` 的关系，但如果要表达 `OR` 的关系，就用 `||` 操作符即可。如下两个语句的意思是相等的：

```python
O::human:(height) { age > 31, sex != re("男") }
O::human:(height) { age > 31 && sex != re("男") }
```

来个复杂的过滤表达式：

```python
M::some_metric {(a>123.45 && b!=re("abc")) || (z!="abc"), c=re("xyz")} [1d::30m]
```

### time-expr

DataFlux 数据特点均有时间属性，故将时间的表达用单独的子句来表示：

- `[5m]` - 最近 5 分钟
- `[10m:5m]` - 最近 10 分钟到最近 5 分钟
- `[10m:5m:1m]` - 最近 10 分钟到最近 5 分钟，且结果按照 1 分钟的间隔聚合
- `[2019-01-01 12:13:14:5m:1w]` -  2019/1/1 12:13:14 到最近 5 分钟，且结果按照 1 周的间隔聚合。注意，指定日期时，只能精确到秒级别。且只有两种日期格式：
	- `2006-01-02 15:04:05`：这里的时间指 UTC 时区的时间，不支持指定时区。
	- `2006-01-02`

时间单位支持如下几种：

- `ns` - 纳秒
- `us` - 微秒
- `ms` - 毫秒
- `s` - 秒
- `m` - 分钟
- `h` - 小时
- `d` - 天
- `w` - 周
- `y` - 年，指定为 365d，不区分闰年。

### by-clause 语句

`BY` 子句用来对结果进行分类聚合。类似 MySQL 中的 `GROUP BY`

### filter-clause 语句

`FILTER ... WITH ...` 用来对不同数据集合做过滤计算：

```python
# 获取所有对象的 CPU 使用率
M::cpu:(host, usage) FILTER O::ecs:(hostname) WITH {host = hostname}
```

### link-with 语句

`LINK ... WITH ...` 用来对不同数据集做合并输出：

```python
O::ecs:(host, region) LINK M::cpu:(usage, hostname) [:5m] WITH {host = hostname}

# 多 LINK 写法
O::ecs:(host, region)  
    LINK M::cpu:(usage, hostname) [:5m]
    LINK M::mem:(percent, hostname) [:5m]
    WITH {host = hostname}
```

### SHOW 语句

`SHOW_xxx` 用来浏览数据：

- `SHOW_MEASUREMENT()` - 查看指标集列表，支持`filter-clause`、`limit`和`offset`语句
- `SHOW_TAG_KEY()` - 查看指标集 tag 列表，支持`filter-clause`、`limit`和`offset`语句
- `SHOW_TAG_VALUE()` - 查看指标集 tag-value 列表，支持`filter-clause`、`limit`和`offset`语句
- `SHOW_FIELD_KEY()` - 查看指标集 field-key 列表
- `SHOW_OBJECT_CLASS()` - 查看对象分类列表
- `SHOW_EVENT_SOURCE()` - 查看事件来源列表
- `SHOW_LOGGING_SOURCE()` - 查看日志来源列表
- `SHOW_TRACING_SERVICE()` - 查看 tracing 来源列表
- `SHOW_RUM_TYPE()` - 查看 RUM 数据类型列表


### 结果集函数结算

DQL 支持对查询结果进行二次计算：

```python
func::dataflux__dql:(EXPR_EVAL(expr='data1.f1+data1.f2', data=dql('M::cpu:(f1, f2)')))

# 通过 func 做跨数据集的计算
F::dataflux__dql:(SOME_FUNC(
	data1=dql('M::cpu:(f1, f2)'),
	data2=dql('O::ecs:(f1, f2)'), some_args))

# 通过 func 做跨数据集的复杂表达式计算
F::dataflux__dql:(EXPR_EVAL(
	expr='data1.f1/(data2.f2+3)',  # 表达式
	data1=dql('M::cpu:(f1, f2)'),
	data2=dql('O::ecs:(f1, f2)'),))
```

### 嵌套查询以及语句块

以 `()` 来表示子查询和外层查询的分隔，如两层嵌套

```python
metric::(
		# 子查询
		metric::cpu,mem:(f1, f2) {host="abcd"} [1m:2m:5s] BY f1 DESC 
	):(f1)              # 外层查询目标列
	{ host=re("abc*") } # 外城查询过滤条件
	[1m:2m:1s]          # 外层查询时间限制
```

三层嵌套

```python
metric::(     # 第二层查询
		metric::( # 第三层查询
				metric::a:(f1,f2,f3) {host="foo"} [10m::1m]
			):(f1,f2)
	):(f1)
```

原则上不对嵌套层次做限制。但**不允许某层嵌套中出现多个平级的子查询**，如：

```python
object::(     # 第二层查询
		object::( # 第三层查询
				object::a:(f1,f2,f3) {host="foo"} [10m::1m]
			):(f1,f2),

		object::( # 并列第三层查询：不支持
				object::b:(f1,f2,f3) {host="foo"} [10m::1m]
			):(f1,f2)
	):(f1)
```

## 函数说明

参见 [dql 实现的函数](./Funcs.md)

参见 [dql 实现的外层函数](./OuterFuncs.md)



