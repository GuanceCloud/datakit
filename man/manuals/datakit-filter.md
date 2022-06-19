{{.CSS}}
# 行协议过滤器
---

- DataKit 版本：{{.Version}}
- 操作系统支持：全平台

本文档主要描述 DataKit Filter 基本使用以及注意事项。

## 简介

DataKit Filter 用于对采集到的行协议数据进行筛选，用于过滤掉一些不想要的数据，它的功能跟 Pipeline 有一点类似，但有所区别：

| 数据处理组件 | 支持本地配置 | 支持中心下发 | 支持数据丢弃 | 支持数据改写 | 使用方法                                                        |
| ----         | ----         | ----         | ----         | ----         | ----                                                            |
| Pipeline     | Y            | Y            | Y            | Y            | 通过在采集器中配置 Pipeline 或者在观测云 Studio 编写 Pipeline   |
| Filter       | Y            | Y            | Y            | N            | 在观测云 Studio 编写 Pipeline 或者在 datakit.conf 中配置 filter |

从表中可以看出，相比 Pipeline，如果只是简单的过滤掉部分数据，那么 Filter 是一种更便捷的数据筛选工具。

## Filter 具体使用方法

Filter 的主要功能就是数据筛选，其筛选依据是通过一定的筛选条件，对采集到的数据进行判定，符合筛选条件的数据，将被丢弃。

过滤器的基本语法模式为：

```
{ conditions [AND/OR conditons] }
```

其中 `conditions` 又可以是其它各种条件的组合。以下是一些过滤器示例：

```python
# 这条一般针对日志数据，用于判定所有日志类型，将其中符合条件的 key1/key2 过滤掉
# 注意，这里的 key1 和 key2 均为行协议字段中的 tag 或 field
{ source = re('.*')  AND ( key1 = "abc" OR key2 = "def") }

# 这条一般针对 Tracing 数据，用于名为 app1 的 service，将其中符合条件的 key1/key2 过滤掉
{ service = "app-1"  AND ( key1 = "abc" OR key2 = "def") }
```

### 过滤器操作的数据范围

由于 DataKit 采集到的（绝大部分）数据均以行协议的方式上报，故所有过滤器均工作于行协议之上。过滤器支持在如下数据上做数据筛选：

- 指标集名称：对于不同类型的数据，指标集的业务归属有所不同，分别如下：
  - 对时序数据（M）而言，在过滤器运行的时候，会在其 tag 列表中注入一个 `measurement` 的 tag，故可以这样来写基于指标集的过滤器：`{  measurement = re('abc.*') AND ( tag1='def' and field2 = 3.14)}`
  - 对对象数据（O）而言，在过滤器运行的时候，会在其 tag 列表中注入一个 `class` 的 tag，故可以这样来写基于对象的过滤器：`{  class = re('abc.*') AND ( tag1='def' and field2 = 3.14)}`
  - 对日志数据（L）而言，在过滤器运行的时候，会在其 tag 列表中注入一个 `source` 的 tag，故可以这样来写基于对象的过滤器：`{  trace = re('abc.*') AND ( tag1='def' and field2 = 3.14)}`

> 如果原来 tag 中就存在一个名为 `measurement/class/source` 的 tag，那么==在过滤器运行过程中，原来的 measurement/class/source 这些 tag 值将不存在==

- Tag（标签）：对所有的数据类型，均可以在其 Tag 上执行过滤。
- Field（指标）：对所有的数据类型，均可以在其 Field 上执行过滤。

### DataKit 中手动配置 filter

在 `datakt.conf` 中，可手动配置黑名单过滤，示例如下：

```toml
[io]
  [io.Filters]
    logging = [ # 针对日志数据的过滤
    	"{ source = 'datakit' or f1 IN [ 1, 2, 3] }"
    ]
    metric = [ # 针对指标的过滤
    	"{ measurement IN ['datakit', 'disk'] }",
    	"{ measurement CONTAIN ['host.*', 'swap'] }",
    ]
    object = [ # 针对对象过滤
    	"{ class CONTAIN ['host_.*'] }",
    ]
    tracing = [ # 针对 tracing 过滤
    	"{ service = re("abc.*") AND some_tag CONTAIN ['def_.*'] }",
    ]
```

一旦 *datakit.conf* 中配置了过滤器，那么则以该过滤器为准，==观测云 Studio 配置的过滤器将不再生效==。

这里的配置需遵循如下规则：

- 具体的一组过滤器，==必须指定它所过滤的数据类型==，目前只支持 logging/metric/tracing/object 这四种
- 同一个数据类型，不要配置多个入口（即配置了多组 logging 过滤器），否则 *datakit.conf* 会解析报错，导致 DataKit 无法启动
- 单个数据类型下，能配置多个过滤器（如上例中的 metric）
- 对于语法错误的过滤器，DataKit 默认忽略，它将不生效，但不影响 DataKit 其它功能

## 过滤器基本语法规则

### 基本语法规则

过滤器基本语法规则，跟 Pipeline 基本一致，参见[这里](pipeline#basic-syntax)。

### 操作符

支持基本的数值比较操作：

- 判断相等
  - `=`
  - `!=`

- 判断数值大小
  - `>`
  - `>=`
  - `<`
  - `<=`

- 括号表达式：用于任意关系之间的逻辑组合，如

```
{ service = re('.*') AND ( abc IN [1,2,'foo', 2.3] OR def CONTAIN ['foo.*', 'bar.*']) }
```

除此之外，还支持如下列表操作：

| 操作符                  | 支持数值类型   | 说明                                                   | 示例                                |
| ----                    | ----           | ----                                                   | ----                                |
| `IN`, `NOTIN`           | 数值列表列表   | 指定的字段是否在列表中，列表中支持多类型混杂           | `{ abc IN [1,2, "foo", 3.5]}`       |
| `CONTAIN`, `NOTCONTAIN` | 正则表达式列表 | 指定的字段是否匹配列表中的正则，该列表只支持字符串类型 | `{ abc CONTAIN ["foo.*", "bar.*"]}` |

> 列表中==只能出现普通的数据类型==，如字符串、整数、浮点，其它表达式均不支持。 

`IN/NOTIN/CONTAIN/NOTCONTAIN` 这些关键字==大小写不敏感==，即 `in` 和 `IN` 以及 `In` 效果是一样的。除此之外，其它操作数的大小写都是敏感的，比如如下两个过滤器表达的意思不同：

```
{ abc IN [1,2, "foo", 3.5]} # 字段 abc（tag 或 field）是否在列表中
{ ABC IN [1,2, "foo", 3.5]} # 字段 ABC（tag 或 field）是否在列表中
```

在行协议中，所有字段都是大小写敏感的。

