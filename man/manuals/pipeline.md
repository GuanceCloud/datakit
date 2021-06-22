{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Pipeline 使用文档

以下是文本处理器定义。随着不同语法的逐步支持，该文档会做不同程度的调整和增删。

基本规则：

- 函数名大小写不敏感
- 以 `#` 为行注释字符。不支持行内注释
- 标识符：只能出现 `[_a-zA-Z0-9]` 这些字符，且首字符不能是数字。如 `_abc, _abc123, _123ab`
- 字符串值可用双引号和单引号： `"this is a string"` 和 `'this is a string'` 是等价的
- 数据类型：支持浮点（`123.4`, `5.67E3`）、整形（`123`, `-1`）、字符串（`'张三'`, `"hello world"`）、Boolean（`true`, `false`）四种类型
- 多个函数之间，可以用空白字符（空格、换行、Tab 等）分割

## 快速开始

### 如何在 DataKit 中配置 pipeline：

- 第一步：编写如下 pipeline 文件，假定名为 `nginx.p`。将其存放在 `<datakit安装目录>/pipeline` 目录下。

```python
# 假定输入是一个 Nginx 日志（以下字段都是 yy 的...）
# 注意，脚本是可以加注释的

grok(_, "some-grok-patterns")  # 对输入的文本，进行 grok 提取
rename('client_ip', ip)        # 将 ip 字段改名成 client_ip
rename("网络协议", protocol)   # 将 protocol 字段改名成 `网络协议`

# 将时间戳(如 1610967131)换成 RFC3339 日期格式：2006-01-02T15:04:05Z07:00
datetime(access_time, "s", "RFC3339")

url_decode(request_url)      # 将 HTTP 请求路由翻译成明文

# 当 status_code 介于 200 ~ 300 之间，新建一个 http_status = "HTTP_OK" 的字段
group_between(status_code, [200, 300], "HTTP_OK", "http_status")

# 丢弃原内容
drop_origin_data()
```

- 第二步：配置对应的采集器来使用上面的 pipeline

以 tailf 采集器为例，配置字段 `pipeline_path` 即可，注意，这里配置的是 pipeline 的脚本名称，而不是路径。所有这里引用的 pipeline 脚本，必须存放在 `<DataKit 安装目录/pipeline>` 目录下：

```python
[[inputs.tailf]]
	logfiles = ["/path/to/nginx/log"]

	# required
	source = "nginx"
	from_beginning = false

	# 此处配置成 datakit 安装目录的相对路径，故所有脚本必须放在 /path/to/datakit/pipeline 目录下
	# 如果 pipeline 未配置，则在 pipeline 目录下寻找跟 source 同名的脚本（如 nginx -> nginx.p），
	# 作为其默认 pipeline 配置
	pipeline_ = "nginx.p"

	... # 其它配置
```

重启采集器，即可切割对应的日志。

### 在 datakit 中调试 pipeline

如果在编写 pipeline 的过程中，可能编写 pipeline 或者 grok 时，需要调试，DataKit 提供了对应的调试工具。 进入 DataKit 安装目录，执行：

```shell
./datakit --cmd --pl <pipeline-script-name.p> --txt <txt-to-be-pipelined>
```

参数说明：

- `cmd`: DataKit 命令模式
- `pl`: pipeline 文件名（即 DataKit 安装目录下 pipeline 里面的脚本），你只能调试这个目录下的 pipeline 脚本
- `txt`: 欲提取的原始文本，可以是 json 或纯文本（这里不是文件路径，直接传文本过来）

示例：

这里以 datakit 自身的日志切割为例。DataKit 自身的日志形式如下：

```
2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms
```

编写对应 pipeline：

```python
# pipeline for datakit log
# Mon Jan 11 10:42:41 CST 2021
# auth: tanb

grok(_, '%{_dklog_date:log_time}%{SPACE}%{_dklog_level:level}%{SPACE}%{_dklog_mod:module}%{SPACE}%{_dklog_source_file:code}%{SPACE}%{_dklog_msg:msg}')
rename("time", log_time) # 将 log_time 重名命名为 time
default_time(time)       # 将 time 字段作为输出数据的时间戳
drop_origin_data()       # 丢弃原始日志文本(不建议这么做)
```

这里引用了几个用户自定义的 pattern，如 `_dklog_date`、`_dklog_level`。我们将这些规则存放 `<datakit安装目录>/pipeline/pattern` 下（**注意，用户自定义 pattern 如果需要全局生效，必须放置在 `<DataKit安装目录/pipeline/pattern/>` 目录下**）:

```Shell
$ cat pipeline/pattern/datakit
# 注意：自定义的这些 pattern，命名最好加上特定的前缀，以免跟内置的命名冲突（内置 pattern 名称不允许覆盖）
# 自定义 pattern 格式为：
#    <pattern-name><空格><具体 pattern 组合>
_dklog_date %{YEAR}-%{MONTHNUM}-%{MONTHDAY}T%{HOUR}:%{MINUTE}:%{SECOND}%{INT}
_dklog_level (DEBUG|INFO|WARN|ERROR|FATAL)
_dklog_mod %{WORD}
_dklog_source_file (/?[\w_%!$@:.,-]?/?)(\S+)?
_dklog_msg %{GREEDYDATA}
```

现在 pipeline 以及其引用的 pattern 都有了，就能通过 DataKit 内置的 pipeline 调试工具，对这一行日志进行切割：

```Shell
# 提取成功示例
$ ./datakit --cmd --pl dklog_pl.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms'
Extracted data(cost: 421.705µs):
{
	"code": "io/io.go:458",
	"level": "DEBUG",
	"module": "io",
	"msg": "post cost 6.87021ms",
	"time": 1610358231887000000
}

# 提取失败示例
$ ./datakit --cmd --pl dklog_pl.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.g o:458  post cost 6.87021ms'
No data extracted from pipeline
```

### 利用 `grokq` 辅助 pipeline 编写

由于 grok pattern 数量繁多，人工匹配较为麻烦。DataKit 提供了交互式的命令行工具：

```Shell
$ ./datakit --cmd --grokq
grokq > Mon Jan 25 19:41:17 CST 2021   # 此处输入你希望匹配的文本
        2 %{DATESTAMP_OTHER: ?}        # 工具会给出对应对的建议，越靠前匹配月精确（权重也越大）。前面的数字表明权重。
        0 %{GREEDYDATA: ?}

grokq > 2021-01-25T18:37:22.016+0800
        4 %{TIMESTAMP_ISO8601: ?}      # 此处的 ? 表示你需要用一个字段来命名匹配到的文本
        0 %{NOTSPACE: ?}
        0 %{PROG: ?}
        0 %{SYSLOGPROG: ?}
        0 %{GREEDYDATA: ?}             # 像 GREEDYDATA 这种范围很广的 pattern，权重都较低
                                       # 权重越高，匹配的精确度越大

grokq > Q                              # Q 或 exit 退出
Bye!
```

## 脚本函数

函数参数说明：

- 函数参数中，匿名参数（`_`）指原始的输入文本数据
- json 路径，直接表示成 `x.y.z` 这种形式，无需其它修饰。例如 `{"a":{"first":2.3, "second":2, "third":"abc", "forth":true}, "age":47}`，json 路径为 `a.thrid` 表示待操作数据为 `abc`
- 所有函数参数的相对顺序，都是固定的，引擎会对其做具体检查
- 以下提到的所有 `key` 参数，都指已经过初次提取（通过 `grok()` 或 `json()`）之后，生成的 `key`
- 待处理json的路径，支持标识符的写法，不能使用字符串，如果是生成新key，需要使用字符串

### `add_pattern(name=required, pattern=required)`

函数说明: 创建自定义 grok 模式。

参数:

- `name`：模式命名
- `pattern`: 自定义模式内容

示例:

```python
# 待处理数据
data = "21:13:14"

# pipline脚本
script = `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{HOUR:hour}:%{MINUTE:minute}(?::%{SECOND:second})([^0-9]?)")
grok(_, "%{time}")`
`

# 处理结果
{
	"hour":"12",
	"minute":"13",
	"second":"14",
	"message":"21:13:14"
}
```

### `grok(input=required, pattern=required)`

函数说明: 通过 `pattern` 提取文本串 `input` 中的内容。

参数:

- `input`：待提取文本，可以是原始文本（`_`）或经过初次提取之后的某个 `key`
- `pattern`: grok 表达式

```python
grok(_, pattern)    # 直接使用输入的文本作为原始数据
grok(key, pattern)  # 对之前已经提取出来的某个 key，做再次 grok
```

示例:

```python
# 待处理数据
data = "21:13:14"

# pipline脚本
script = `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{HOUR:hour}:%{MINUTE:minute}(?::%{SECOND:second})([^0-9]?)")
grok(_, "%{time}")
`

# 处理结果
{
	"hour":"12",
	"minute":"13",
	"second":"14",
	"message":"21:13:14"
}
```

### `json(input=required, jsonPath=required, newkey=optional)`

函数说明: 提取 json 中的指定字段，并可将其命名成新的字段。

参数:

- `input`: 待提取 json，可以是原始文本（`_`）或经过初次提取之后的某个 `key`
- `jsonPath`: json 路径信息
- `newkey`：提取后数据写入新 key

```python
# 直接提取原始输入 json 中的x.y字段，并可将其命名成新字段abc
json(_, x.y, abc)

# 已提取出的某个 `key`，对其再提取一次 `x.y`，提取后字段名为 `x.y`
json(key, x.y) 
```

示例一:

```python
# 待处理数据
data = `{"info": {"age": 17, "name": "zhangsan", "height": 180}}`

# 处理脚本
script = `
json(_, info, "zhangsan")
json(zhangsan, name)
json(zhangsan, age, "年龄")
`

# 处理结果
{
	"message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}",
	"zhangsan": {
		"age": 17,
		"height": 180,
		"name": "zhangsan"
	}
}
```

示例二:

```python
# 待处理数据
data = `{
	"name": {"first": "Tom", "last": "Anderson"},
	"age":37,
	"children": ["Sara","Alex","Jack"],
	"fav.movie": "Deer Hunter",
	"friends": [
		{"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
		{"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
		{"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
]
	}`

# 处理脚本
script = `
json(_, name) json(name, first)
`
```

示例三:

```python
# 待处理数据
data = `[
	    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
	    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
	    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
	]`

# 处理脚本, json数组处理
script = `
json(_, [0].nets[-1])
`
```

### `json_all()`

函数说明：提取 json 中的所有字段，所有层次均被拉平。

参数：

- `input`：被提取的 json

示例：

```python
# 待处理数据
data = `
{
	"name": {"first": "Tom", "last": "Anderson"},
	"age":37,
	"身高": 180,
	"children": ["Sara","Alex","Jack"],
	"fav.movie": "Deer Hunter",
	"friends": [
		{"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
		{"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
		{"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
	]
}
`

# 处理脚本
script = `json_all()`

会提取出如下对象
{
	"age": 37,
	"身高": 180,
	"children[0]": "Sara",
	"children[1]": "Alex",
	"children[2]": "Jack",
	"fav.movie": "Deer Hunter",
	"friends[0].age": 44,
	"friends[0].first": "Dale",
	"friends[0].last": "Murphy",
	"friends[0].nets[0]": "ig",
	"friends[0].nets[1]": "fb",
	"friends[0].nets[2]": "tw",
	"friends[1].age": 68,
	"friends[1].first": "Roger",
	"friends[1].last": "Craig",
	"friends[1].nets[0]": "fb",
	"friends[1].nets[1]": "tw",
	"friends[2].age": 47,
	"friends[2].first": "Jane",
	"friends[2].last": "Murphy",
	"friends[2].nets[0]": "ig",
	"friends[2].nets[1]": "tw",
	"name.first": "Tom",
	"name.last": "Anderson"
}

引用某个字段

rename('年龄', age) # 将 age 重命名为 '年龄'

# 将 `friends[2].nets[1]` 重命名为 'f2nets'
# 注意：因为 friends[2].nets[1] 包含特殊的 json 路径字符，故需要用 `` 包围一下。
rename('f2nets', `friends[2].nets[1]`) 
rename('height', `身高`) # 身高因为是 Unicode 字符，需要 `` 包围一下 

# 经过上面 rename 之后，对象变成如下样子

{
	"年龄": 37,
	"height": 180,
	"children[0]": "Sara",
	"children[1]": "Alex",
	"children[2]": "Jack",
	"fav.movie": "Deer Hunter",
	"friends[0].age": 44,
	"friends[0].first": "Dale",
	"friends[0].last": "Murphy",
	"friends[0].nets[0]": "ig",
	"friends[0].nets[1]": "fb",
	"friends[0].nets[2]": "tw",
	"friends[1].age": 68,
	"friends[1].first": "Roger",
	"friends[1].last": "Craig",
	"friends[1].nets[0]": "fb",
	"friends[1].nets[1]": "tw",
	"friends[2].age": 47,
	"friends[2].first": "Jane",
	"friends[2].last": "Murphy",
	"friends[2].nets[0]": "ig",
	"f2nets": "tw",
	"name.first": "Tom",
	"name.last": "Anderson"
}

```

### `rename(new-key=required, old-key=required)`

函数说明: 将已提取的字段重新命名

参数:

- `new-key`: 新字段名
- `old-key`: 已提取的字段名

```python
# 把已提取的 abc 字段重新命名为 abc1
rename('abc1', abc)
```

示例：

```python
# 待处理数据
data = `{"info": {"age": 17, "name": "zhangsan", "height": 180}}`

# 处理脚本
script = `json(_, info.name, "姓名")`

# 处理结果
{
  "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}",
  "zhangsan": {
    "age": 17,
    "height": 180,
    "姓名": "zhangsan"
  }
}
```

### `url_decode(key=required)`

函数说明: 将已提取 `key` 中的 URL 解析成明文

参数:
- `key`: 已经提取的某个 `key`

示例：

```python
# 待处理数据
data = `{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}`

# 处理脚本
script = `json(_, url) url_decode(url)`

# 处理结果
{
  "message": "{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}",
  "url": "http://www.baidu.com/s?wd=测试"
}
```

### `geoip(key=required)`

函数说明: 在 IP 上追加更对 geo 信息。 `geoip()` 会额外产生多个字段，如：

- `isp`: 运营商
- `city`: 城市
- `province`: 省份
- `country`: 国家

参数:

- `key`: 已经提取出来的 IP 字段，支持 IPv4/6

示例：

```python
# 待处理数据
data = `{"ip":"116.228.89.206"}`

# 处理脚本
script = `json(_, ip) geoip(ip)`

# 处理结果
{
  "message": "{"ip":"116.228.89.206"}",
  "isp": "xxxx",
  "city": "xxxx",
  "province": "xxxx",
  "country": "xxxx"
}
```

### `datetime(key=required, precision=required, fmt=required)`

函数说明: 将时间戳转成指定日期格式

函数参数

- `key`: 已经提取的时间戳 (必选参数)
- `precision`：输入的时间戳精度(s, ms)
- `fmt`：日期格式，时间格式, 支持以下模版

```python
ANSIC       = "Mon Jan _2 15:04:05 2006"
UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
RFC822      = "02 Jan 06 15:04 MST"
RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
RFC3339     = "2006-01-02T15:04:05Z07:00"
RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
Kitchen     = "3:04PM"
```

示例:

```python
# 待处理数据
data = `{"a":{"timestamp": "1610960605000", "second":2},"age":47}`

# 处理脚本
script = `json(_, a.timestamp) datetime(a.timestamp, 'ms', 'RFC3339')`
```

### `expr(expr=required, key=required)`

函数说明: 计算表达式 expr 的值，并将计算结果写入 `key`

支持的表达式操作符：

- `+`
- `-`
- `*`
- `/`
- `%`
- `=`
- `!=`
- `<=`
- `<`
- `>=`
- `>`
- `^`
- `&&`
- `||`

函数参数

- `expr`: 表达式，如 `abc+2*3-def/5+x.y.z`
- `key`：结果写入指定字段

示例:

```python
expr(key1 * 2 + key3, result)
```

示例:

```python
# 待处理数据
data = `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`

# 处理脚本
script = `json(_, a.second)
expr(a.second*10+(2+3)*5, bb)
`

# 处理结果
{
   "bb": "45"
}
```

### `cast(key=required, type=required)`

函数说明: 将 key 值转换拆成指定类型

函数参数

- `key`: 已提取的某字段
- `type`：转换的目标类型，支持 `"str", "float", "int", "bool"` 这几种，目标类型需要用英文状态双引号括起来

示例:

```python
# 待处理数据
data = `{"first": 1,"second":2,"thrid":"aBC","forth":true}`

# 处理脚本
script = `json(_, first) cast(first, "str")`

# 处理结果
{
  "first":"1"
}
```

### `group_between(key=required, between=required, new-value=required, new-key=optional)`

函数说明： 如果 `key` 值在指定范围 `between` 内（注意：只能是单个区间，如 `[0,100]`），则可创建一个新字段，并赋予新值。若不提供新字段，则覆盖原字段值

示例一:

```python
# 待处理数据
data = `{"http_status": 200, "code": "success"}`

script = `json(_, http_status)
# 如果字段 http_status 值在指定范围内，则将其值改为 "OK"
group_between(http_status, [200, 300], "OK")
# 如果字段 http_status 值在指定范围内，则新建 status 字段，其值为 "OK"
`

# 处理结果
{
	"http_status": "OK"
}
```

示例二:

```python
# 待处理数据
data = `{"http_status": 200, "code": "success"}`

script = `json(_, http_status)
# 如果字段 http_status 值在指定范围内，则将其值改为 "OK"
# 如果字段 http_status 值在指定范围内，则新建 status 字段，其值为 "OK"
group_between(http_status, [200, 300], "OK", status)
`

# 处理结果
{
	"http_status": 200,
	"status": "OK"
}
```

### `group_in(key=required, in=required, new-value=required, new-key=optional)`

如果 `key` 值在列表 `in` 中，则可创建一个新字段，并赋予新值。若不提供新字段，则覆盖原字段值

示例:

```python
# 如果字段 log_level 值在列表中，则将其值改为 "OK"
group_in(log_level, ["info", "debug"], "OK")

# 如果字段 http_status 值在指定列表中，则新建 status 字段，其值为 "not-ok"
group_in(log_level, ["error", "panic"], "not-ok", status)
```

### `uppercase(key=required)`

函数说明: 将已提取 key 中内容转换成大写

函数参数

- `key`: 指定已提取的待转换字段名

将 key 内容转成大写

示例:

```python
# 待处理数据
data = `{"first": "hello","second":2,"thrid":"aBC","forth":true}`

# 处理脚本
script = `json(_, first) uppercase(first, "1")`

# 处理结果
{
   "first": "HELLO"
}
```

### `lowercase(key=required)`

函数说明: 将已提取 key 中内容转换成小写

函数参数

- `key`: 指定已提取的待转换字段名

示例:

```python
# 待处理数据
data = `{"first": "HeLLo","second":2,"thrid":"aBC","forth":true}`

# 处理脚本
script = `json(_, first) lowercase(first)`

# 处理结果
{
	"first": "hello"
}
```

### `nullif(key=required, value=required)`

函数说明: 若已提取 `key` 指定的字段内容等于 `value` 值，则删除此字段

函数参数

- `key`: 指定字段
- `value`: 目标值

示例:

```python
# 待处理数据
data = `{"first": 1,"second":2,"thrid":"aBC","forth":true}`

# 处理脚本
script = `json(_, first) json(_, second) nullif(first, "1")`

# 处理结果
{
	"second":2
}
```

### `strfmt(key=required, fmt=required, key1=optional, key2, ...)`

函数说明: 对已提取 `key1,key2...` 指定的字段内容根据 `fmt` 进行格式化，并把格式化后的内容写入 `key` 字段中

函数参数

- `key`: 指定格式化后数据写入字段名
- `fmt`: 格式化字符串模板
- `key1，key2`:已提取待格式化字段名

示例:

```python
# 待处理数据
data = `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`

# 处理脚本
script = `json(_, a.second)
json(_, a.thrid)
cast(a.second, "int")
json(_, a.forth)
strfmt(bb, "%v %s %v", a.second, a.thrid, a.forth)
`
```

### `drop_origin_data()`

函数说明: 丢弃初始化文本，否则初始文本放在 message 字段中

示例:

```python
# 待处理数据
data = `{"age": 17, "name": "zhangsan", "height": 180}`

# 处理脚本
script = `drop_origin_data()`

# 结果集中删除message内容
```

### `add_key(key-name=required, key-value=required)`

函数说明: 增加一个字段

函数参数
- `key-name`: 新增的 key 名称
- `key-value`：key 值（只能是 string/number/bool/nil 这几种类型）

示例:

```python
# 待处理数据
data = `{"age": 17, "name": "zhangsan", "height": 180}`

# 处理脚本
script = `add_key(city, "shanghai")`

# 处理结果
{
    "age": 17,
    "height": 180,
    "name": "zhangsan",
    "city": "shanghai"
}
```

### `default_time(key=required, timezone=optional)`

函数说明: 以提取的某个字段作为最终数据的时间戳

函数参数
- `key`: 指定的 key
- `timezone`: 指定的时区，默认为本机当前时区

待处理数据支持以下格式化时间

| 日期格式                                           | 日期格式                                                | 日期格式                                       | 日期格式                          |
| -----                                              | ----                                                    | ----                                           | ----                              |
| `2014-04-26 17:24:37.3186369`                      | `May 8, 2009 5:57:51 PM`                                | `2012-08-03 18:31:59.257000000`                | `oct 7, 1970`                     |
| `2014-04-26 17:24:37.123`                          | `oct 7, '70`                                            | `2013-04-01 22:43`                             | `oct. 7, 1970`                    |
| `2013-04-01 22:43:22`                              | `oct. 7, 70`                                            | `2014-12-16 06:20:00 UTC`                      | `Mon Jan  2 15:04:05 2006`        |
| `2014-12-16 06:20:00 GMT`                          | `Mon Jan  2 15:04:05 MST 2006`                          | `2014-04-26 05:24:37 PM`                       | `Mon Jan 02 15:04:05 -0700 2006`  |
| `2014-04-26 13:13:43 +0800`                        | `Monday, 02-Jan-06 15:04:05 MST`                        | `2014-04-26 13:13:43 +0800 +08`                | `Mon, 02 Jan 2006 15:04:05 MST`   |
| `2014-04-26 13:13:44 +09:00`                       | `Tue, 11 Jul 2017 16:28:13 +0200 (CEST)`                | `2012-08-03 18:31:59.257000000 +0000 UTC`      | `Mon, 02 Jan 2006 15:04:05 -0700` |
| `2015-09-30 18:48:56.35272715 +0000 UTC`           | `Thu, 4 Jan 2018 17:53:36 +0000`                        | `2015-02-18 00:12:00 +0000 GMT`                | `Mon 30 Sep 2018 09:09:09 PM UTC` |
| `2015-02-18 00:12:00 +0000 UTC`                    | `Mon Aug 10 15:44:11 UTC+0100 2015`                     | `2015-02-08 03:02:00 +0300 MSK m=+0.000000001` | `Thu, 4 Jan 2018 17:53:36 +0000`  |
| `2015-02-08 03:02:00.001 +0300 MSK m=+0.000000001` | `Fri Jul 03 2015 18:04:07 GMT+0100 (GMT Daylight Time)` | `2017-07-19 03:21:51+00:00`                    | `September 17, 2012 10:09am`      |
| `2014-04-26`                                       | `September 17, 2012 at 10:09am PST-08`                  | `2014-04`                                      | `September 17, 2012, 10:10:09`    |
| `2014`                                             | `2014:3:31`                                             | `2014-05-11 08:20:13,787`                      | `2014:03:31`                      |
| `3.31.2014`                                        | `2014:4:8 22:05`                                        | `03.31.2014`                                   | `2014:04:08 22:05`                |
| `08.21.71`                                         | `2014:04:2 03:00:51`                                    | `2014.03`                                      | `2014:4:02 03:00:51`              |
| `2014.03.30`                                       | `2012:03:19 10:11:59`                                   | `20140601`                                     | `2012:03:19 10:11:59.3186369`     |
| `20140722105203`                                   | `2014年04月08日`                                        | `1332151919`                                   | `2006-01-02T15:04:05+0000`        |
| `1384216367189`                                    | `2009-08-12T22:15:09-07:00`                             | `1384216367111222`                             | `2009-08-12T22:15:09`             |
| `1384216367111222333`                              | `2009-08-12T22:15:09Z`                                  |

JSON 提取示例:

```python
# 原始 json
{
	"time":"06/Jan/2017:16:16:37 +0000",
	"second":2,
	"thrid":"abc",
	"forth":true
}

# pipeline 脚本
json(_, time)      # 提取 time 字段
default_time(time) # 将提取到的 time 字段转换成时间戳

# 处理结果
{
  "time": 1483719397000000000,
}
```

文本提取示例:

```python
# 原始日志文本
2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms

# pipeline 脚本
grok(_, '%{TIMESTAMP_ISO8601:log_time}')   # 提取日志时间，并将字段命名为 log_time
default_time(log_time)                     # 将提取到的 log_time 字段转换成时间戳

# 处理结果
{
  "log_time": 1610358231887000000,
}

# 对于 tailf 采集的数据，最好将时间字段命名为 time，否则 tailf 采集器会以当前时间填充
rename("time", log_time)

# 处理结果
{
  "time": 1610358231887000000,
}
```

### `drop_key(key=required)`

函数说明: 删除已提取字段

函数参数

- `key`: 待删除字段名

示例:

```python
data = `{"age": 17, "name": "zhangsan", "height": 180}`

# 处理脚本
script = `
json(_, age,)
json(_, name)
json(_, height)
drop_key(height)
`

# 处理结果
{
	"age": 17,
	"name": "zhangsan"
}
```

### `user_agent(key=required)`

函数说明: 对指定字段上获取客户端信息

函数参数

- `key`: 待提取字段

`user_agent()` 会生产多个字段，如：

- `os`: 操作系统
- `browser`: 浏览器

示例:

```python
data = `{"userAgent":"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36", "second":2,"thrid":"abc","forth":true}`

script = `
json(_, userAgent) user_agent(userAgent)
`
```

### `parse_duration(key=required)`

函数说明: 如果 `key` 的值是一个 golang 的 duration 字符串（如 `123ms`），则自动将 `key` 解析成纳秒为单位的整数

目前 golang 中的 duration 单位如下：

- `ns` 纳秒
- `us/µs` 微秒
- `ms` 毫秒
- `s` 秒
- `m` 分钟
- `h` 小时

函数参数

- `key`: 待解析的字段

示例:

```python
# 假定 abc = "3.5s"
parse_duration(abc) # 结果 abc = 3500000000

# 支持负数: abc = "-3.5s"
parse_duration(abc) # 结果 abc = -3500000000

# 支持浮点: abc = "-2.3s"
parse_duration(abc) # 结果 abc = -2300000000

```

### `parse_date(new-key=required, yy=require, MM=require, dd=require, hh=require, mm=require, ss=require, ms=require, zone=require)`

函数说明: 将传入的日期字段各部分的值转化为时间戳

函数参数

- `key`: 新插入的字段
- `yy` : 年份数字字符串，支持四位或两位数字字符串，为空字符串，则处理时取当前年份
- `MM`:  月份字符串, 支持数字，英文，英文缩写
- `dd`: 日字符串
- `hh`: 小时字符串
- `mm`: 分钟字符串
- `ss`: 秒字符串
- `ms`: 毫秒字符串
- `zone`: 时区字符串，“+8”或"Asia/Shanghai"形式

示例:

```python
parse_date(aa, "2021", "May", "12", "10", "10", "34", "", "Asia/Shanghai") # 结果 aa=1620785434000000000

parse_date(aa, "2021", "12", "12", "10", "10", "34", "", "Asia/Shanghai") # 结果 aa=1639275034000000000

parse_date(aa, "2021", "12", "12", "10", "10", "34", "100", "Asia/Shanghai") # 结果 aa=1639275034000000100

parse_date(aa, "20", "February", "12", "10", "10", "34", "", "+8") 结果 aa=1581473434000000000
```

### `cover(key=required, range=require)`

函数说明: 对指定字段上获取的字符串数据按索引范围进行数据脱敏处理

函数参数

- `key`: 待提取字段
- `range`: 索引范围[start, end]，如：[3,5] 这种形式，strat和end是一个闭合区间, 下标从1开始，索引不区分半角和全角

注意：中文字符会使用全角替换


示例:

```python
# demo1
data = `{"str": "13789123014"}`

script = `
json(_, str) cover(str, [8, 13])
`

# demo2
data = `{"str": "13789123014"}`

script = `
json(_, str) cover(str, [2, 4])
`

# demo3
data = `{"str": "13789123014"}`

script = `
json(_, str) cover(str, [1, 1])
`

# demo4
data = `{"str": "小阿卡"}`

script = `
json(_, str) cover(str, [2, 2])
`
```

### `replace(key=required, regex=required, replaceStr=required)`

函数说明: 对指定字段上获取的字符串数据按正则进行替换

函数参数

- `key`: 待提取字段
- `regex`: 正则表达式
- `replaceStr`: 替换的字符串


示例:

```python
# 电话号码
data = `{"str": "13789123014"}`

script = `
json(_, str) replace(str, "(1[0-9]{2})[0-9]{4}([0-9]{4})", "$1****$2")
`

# 英文名
data = `{"str": "zhang san"}`

script = `
json(_, str) replace(str, "([a-z]*) \\w*", "$1 ***")
`

=======
# 身份证号
data = `{"str": "362201200005302565"}`

script = `
json(_, str) replace(str, "([1-9]{4})[0-9]{10}([0-9]{4})", "$1**********$2")

# 中文名
data = `{"str": "小阿卡"}`

script = `
json(_, str) replace(str, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])', "$1＊$2")
`
```

### grok 模式分类

DataKit 中 grok 模式可以分为两类：全局模式与局部模式，`pattern` 目录下的模式文件都是全局模式，所有 pipeline 脚本都可使用，而在 pipeline 脚本中通过 `add_pattern()` 函数新增的模式属于局部模式，只针对当前 pipeline 脚本有效。

当 DataKit 内置模式不能满足所有用户需求，用户可以自行在 pipeline 目录中增加模式文件来扩充。若自定义模式是全局级别，则需在 `pattern` 目录中新建一个文件并把模式添加进去，不要在已有内置模式文件中添加或修改，因为datakit启动过程会把内置模式文件覆盖掉。

### 添加局部模式

grok 本质是预定义一些正则表达式来进行文本匹配提取，并且给预定义的正则表达式进行命名，方便使用与嵌套引用扩展出无数个新模式。比如 DataKit 有 3 个如下内置模式：

```python
_second (?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)    #匹配秒数，_second为模式名
_minute (?:[0-5][0-9])                            #匹配分钟数，_minute为模式名
_hour (?:2[0123]|[01]?[0-9])                      #匹配年份，_hour为模式名
```

基于上面三个内置模式，可以扩展出自己内置模式且命名为 `time`:

```python
# 把 time 加到 pattern 目录下文件中，此模式为全局模式，任何地方都能引用 time
time ([^0-9]?)%{hour:hour}:%{minute:minute}(?::%{second:second})([^0-9]?)

# 也可以通过 add_pattern() 添加到 pipeline 文件中，则此模式变为局部模式，只有当前 pipeline 脚本能使用 time
add_pattern(time, "([^0-9]?)%{HOUR:hour}:%{MINUTE:minute}(?::%{SECOND:second})([^0-9]?)")

# 通过 grok 提取原始输入中的时间字段。假定输入为 12:30:59，则提取到 {"hour": 12, "minute": 30, "second": 59}
grok(_, %{time})
```

注意：

- 相同模式名以脚本级优先（即局部模式覆盖全局模式）
- pipeline 脚本中，`add_pattern()` 需在 `grok()` 函数前面调用，否则会导致第一条数据提取失败。

