{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
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

> 注意，切割过程中，需避免[可能出现的跟 tag key 重名的问题](datakit-pl-how-to#5cf855c0)

- 第二步：配置对应的采集器来使用上面的 pipeline

以 logging 采集器为例，配置字段 `pipeline_path` 即可，注意，这里配置的是 pipeline 的脚本名称，而不是路径。所有这里引用的 pipeline 脚本，必须存放在 `<DataKit 安装目录/pipeline>` 目录下：

```python
[[inputs.logging]]
    logfiles = ["/path/to/nginx/log"]

    # required
    source = "nginx"

    # 所有脚本必须放在 /path/to/datakit/pipeline 目录下
    # 如果开启了 gitrepos 功能，则优先以 gitrepos 中的同名文件为准
    # 如果 pipeline 未配置，则在 pipeline 目录下寻找跟 source 同名
    # 的脚本（如 nginx -> nginx.p），作为其默认 pipeline 配置
    pipeline = "nginx.p"

    ... # 其它配置
```

重启采集器，即可切割对应的日志。关于 Pipeline 编写、调试以及注意事项，参见[这里](datakit-pl-how-to)。

## Grok 模式分类

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

## 脚本执行流

pipeline 支持 `if/elif/else` 语法，`if` 后面的语句仅支持条件表达式，即 `<`、`<=`、`==`、`>`、`>=` 和 `!=`， 且支持小括号优先级和多个条件表达式的 `AND` 和 `OR` 连接。
表达式两边可以是已存在的 key 或固定值（数值、布尔值、字符串和 nil ），例如：

```python
# 数值比较
add_key(score, 95)

if score == 100  {
	add_key(level, "S")
} elif score >= 90 && score < 100 {
	add_key(level, "A")
} elif score >= 60 {
	add_key(level, "C")
} else {
	add_key(level, "D")
}

# 字符串比较
add_key(name, "张三")

if name == "法外狂徒" {
	# 这是不可能的，不要污蔑我
}
```

和大多数编程/脚本语言相同，根据 `if/elif` 的条件是否成立，来决定其执行顺序。

注意：如果是进行数值比较，需要先用 `cast()` 进行类型转换，比如：

```
# status_code 是 grok 切出来的 string 类型
cast(status_code, "int")

if status == 200 {
	add_key(level, "OK")
} elif status >= 400 && status < 500 {
	add_key(level, "ERROR")
} elif stauts > 500 {
	add_key(level, "FATAL")
}

```

## 脚本函数

函数参数说明：

- 函数参数中，匿名参数（`_`）指原始的输入文本数据
- json 路径，直接表示成 `x.y.z` 这种形式，无需其它修饰。例如 `{"a":{"first":2.3, "second":2, "third":"abc", "forth":true}, "age":47}`，json 路径为 `a.thrid` 表示待操作数据为 `abc`
- 所有函数参数的相对顺序，都是固定的，引擎会对其做具体检查
- 以下提到的所有 `key` 参数，都指已经过初次提取（通过 `grok()` 或 `json()`）之后，生成的 `key`
- 待处理json的路径，支持标识符的写法，不能使用字符串，如果是生成新key，需要使用字符串

{{.PipelineFuncs}}
