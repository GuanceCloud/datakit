# DataKit 采集的数据结构定义

## Point

Point 是 DataKit 中最常用的一种数据表示形式，目前 Point 有两种表现形式：

1. 行协议
1. Protobuf

由于行协议表达能力有限，后面会逐渐迁移到 Protobuf，同时中心仍然长期支持行协议写入（不同的 API 版本）

## Point 的约束 {#restrictions}

Point 构建函数：

```golang
pt, err := NewPoint(name string, tags map[string]string, field map[string]interface{}, opt *Option)
```

由于 Point 最终需要写入后端存储，故而受到后端存储的各种限制，但原则上采集端仍然做尽量少的限制，以保证数据采集的多样性和完整性。目前 Point 限制如下（本次放开/新加的约束，粗体表示）：

1. tag 和 field 之间不允许出现同名的 key。这一限制主要用来确保 DQL 查询时字段的唯一性。如果 tag/field 中出现同名 key，最终的 Point 中会移除 field 中对应的 key。

1. tag/field key 以及 value 目前不做协议上的限制，这一限制可以在应用层（比如 datakit/kodo）通过 `Option` 配置来控制

1. tag/field key/value 的内容均有几个方面的限制：

    - tag key 和 value 中均不允许出现换行字符，如果出现，SDK 会将其替换成空格。如字符串 `"abc\ndef"` 会被替换成 `"abc def"`。该限制无法通过 Option 控制。
    - 每个 Point 至少要有一个 field。目前该行为无法通过 Option 控制。
    - tag key 和 value 均不允许以反斜杠 `\` 结尾，如果出现，SDK 会将其移除掉。如字符串 `"abc\"` 会被替换成 `"abc"`。该限制无法通过 Option 控制。
    - 对日志类数据（非时序数据）而言，tag/field 的 key 中不允许出现 `.` 字符，如果出现，SDK 会将其替换成 `_`，如字符串 `"abc.def"` 会被替换成 `"abc_def"`。该限制可以通过 Option 控制。
    - tag/field 个数、 各个 tag/field key/value 长度（此处特指字符串长度）均可以通过 Option 设置。
    - 可以禁用某些 tag/field key，一旦出现这些 key，最终的 Point 数据中，这些 key 将被移除掉。

1. field value 目前支持:

    - int：最大支持 `1<<63 - 1`，对于大于该数值的 field，建议将其进行单位转换，避免其越界。比如字节（Byte）可以转换成 MB。

      > 为什么此处限制了超过 `1<<63 - 1` 的整数？因为目前我们使用的 influxdb 版本（1.7）可能不支持 uint64。

    - float/double：浮点数，暂无限制
    - bool：暂无限制
    - string：暂无限制
    - []byte：如果用行协议表示，其 field 会显示成 base64 编码后的值；同样，如果是 Protobuf 的 json 形式，其值也是 base64 之后的值，此处可以用二进制（在不超出 value 长度限制的前提下）
		- nil：空值字段（即什么都没有，对应各种语言里面的 null/nil/NULL 等），对于值为 nil 的 field，Point 直接忽略该字段

### Option 说明

由于支持的数据类型较多，需要对各种数据字段做不同的约束尺度，可通过 Option 来调节这些差异。

目前可调节的维度有：

1. 数据内容维度：

    - 字段名（tag/field）限制：有些字段名在不同的数据类型中，是保留字段，采集器在构建数据时，不应（在 Tag 和 Field 中）使用这些字段，比如对象中的 `class` 字段，日志中的 `source` 字段等。
    - 最大字段个数限制：某些情况下，需要对字段个数做限制（tag 个数和 field 个数）
    - 单个字段 key 长度限制：某些情况下，key 命名过长会违反某些存储的命名约束
    - 单个字段 value 长度限制：某些情况下，value 过长（一般是字符串）会违反某些存储的命名约束

1. 编码维度：

    - 开启 protobuf（当前默认不开启）：如果开启 protobuf，Point 数据将以 protobuf 形式被上传到中心（中心需支持对应的 API）
		- 禁用 precheck（默认不开启）：如果不希望数据做任何规整，可以关闭数据检查（但最终会检查并对其做必要的数据调整）

## Protobuf Point

Protobuf 格式的 Point 结构更为宽松，相比行协议，它的形式相对更为友好，现将其跟行协议的形式做一下对比。

## 纯文本表示

纯文本表示有利于 debug 的时候对数据进行人工校验，也便于编写基本的测试数据。

对于一个确定的 Point 对象，以如下的点结构（为便于理解，以 JSON 形式表示）为例：

```json
{
	"measurement": "abc",
	"tags": {
		"tag1": "v1",
		"tag2": "v2",
	},
	"fields": {
		"f1": 1,
		"f2": 1.2,
		"f3": "hello",
		"f4": true,
	},
	"time": 1668391102000000000
}
```

行协议表示为：

```
abc,tag1=v1,tag2=v2 f1=1i,f2=1.2,f3="hello",f4=true 1668391102000000000
```

Protobuf 文本形式为：

```
{"name":"abc","tags":[{"key":"tag1","val":"v1"},{"key":"tag2","val":"v2"}],"fields":[{"key":"f1","i":"1"},{"key":"f2","f":1.2},{"key":"f3","s":"hello"},{"key":"f4","b":true}],"time":"1668391102000000000"}
```

这里针对 Protobuf 的文本形式做一下说明：

- 如果 Point 是以 Protobuf 形式构建的，那么通过 Go 代码 `fmt.Sprintf("%s", point.String())` 可得到该 JSON 结构。
- Protobuf JSON 基本结构为：

```
{
	"name": "指标集名字",
	"tags": [ {"key": "tag-key-name", "val": "tag-value"}, {... 下一对 tag-key-val} ],
	"fields": [
	    {"key": "key-name-1", "i": "int-value"}, # 有符号整数，注意，其值形式是用 JSON 字符串来表示的
			{"key": "key-name-3", "u": "uint-value"},# 无符号整数，注意，其值形式是用 JSON 字符串来表示的
			{"key": "key-name-2", "f": float-value}, # 浮点数，直接以 JSON 浮点表示
			{"key": "key-name-4", "b": boolean},     # bool 值，直接以 JSON bool 值表示
			{"key": "key-name-5", "s": "string-value"}, # 字符串值，直接以 JSON 字符串值表示
			{"key": "key-name-6", "d": "[]byte-base64-encoded"}, # 字节数据，以 base64 之后的 JSON 字符串表示
			{ ... more fields }
	]
}
```

此处，支持六种 field 数据类型，分别是有符号整数（`i`）、无符号整数（`u`）、浮点（`f`）、布尔（`b`）、字符串（`s`）以及二进制字节序列（`d`），在 JSON 中，**必须**通过第二个 key 的命名来识别 field 的类型，而不能直接使用其 value 来判断取值类型。

> 有符号/无符号整数为何用字符串表示？
> 在 Protobuf 的 JSON 规范中，[64-bit 整数均用 string 表示](https://developers.google.com/protocol-buffers/docs/proto3#json)，可能是某些编程语言无法处理巨大的整数导致的，但对 32-bit 整数仍然用 JSON int 表示。在现有的 Point 结构中，所有整数均采用 64-bit 整数，没有采用 32-bit 整数，故在 protobuf-JSON 中，所有整数均以字符串的形式来表示。

Protobuf 和行协议两种形态的差异如下：

- 行协议也能表示 uint64 这种类型，但在行协议中有所约束，对于超过 int64-max 的 uint64 数值，会被丢弃；而小于 int64-max 的 uint64 数值被强转为 int64。
- 行协议无法表示二进制数据，如果要表示二进制数据，需要业务层做特殊标记，单纯协议层无法识别二进制数据。意即，行协议仍然会将二进制数据进行 base64 编码并以字符串形式来表示，解析的时候，并不会将其自动还原成原始二进制，需要业务层主动识别并还原。而 Protobuf 协议本来就支持二进制数据，应用层能直接识别，无需特殊标记处理。
- 虽然表面上看，Protobuf 格式的 JSON 较为冗长，但传输的过程中，Protobuf 并非以 JSON 形式发送，测试标明，对 1000 个点（大部分是数值）进行字节编码，Protobuf 相比行协议，降低约 10% 左右的 payload。

## 编解码

对一组 Point 而言，当前 SDK 提供简单的编码功能，主要用于将批量数据点变成字节流，便于网络传输。

其主要有几个特征：

- 对 N 个数据点，支持从个数的维度对最终的字节流分包
    - 老的分包是以行协议当前字节流长度来分包，这个相对不好控制。但以点数来分包，包大小因点长度不一致，会参差不齐。但后者比前者更利于编码实现（特别是针对 Protobuf Point 而言）。
- 支持传入一个回调函数，直接在编码过程中以类似迭代器的方式进行数据处理

对行协议 Point 而言，其字节流内容为行协议。对 Protobuf Point 而言，其字节流是 protobuf 字节流。尤其需要注意的是，**每次编码，所有的 Point 必须是同样的类型，不支持行协议 和 protobuf 两类 Point 混合编码**。

### 解码

解码主要用于服务端用来处理外部输入的字节流数据，解码的时候需要提供对应的 option，用来指示输入的字节流是行协议还是 protobuf：

```golang
import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/point"
)

dec := point.Decoder{
	Opt: func() *Option{
		x := point.DefaultOption()
		x.Protobuf = true // 用来解码 protbuf，默认解码行协议
		return x
	},
}

pts, err := dec.Decode(data)

// 带回调函数的解码
dec := point.Decoder{
	Opt: point.DefaultOption(),
	Fn: func(pts []*point.Point) error {
		// do something on @pts
	}
}

pts, err := dec.Decode(data)
```

提供外挂一个回调函数，即可在解码完成后，调用对应的函数处理所有的 Point 数据。如果 Fn 返回 error，对应 Decode() 也将返回 error。如果数据本省能被解码，那么总会返回对应的 Point 数组，不管 Fn 是否调用成功。

如果 `data` 格式有问题，对行协议而言，一般的解析错误为：

```
lineproto parse error: missing tag value
```

> 当前 SDK 对行协议报错做了简化，报错中不再会报告原始数据，影响阅读。但在 Decoder 中增加了一个 DetailedError 字段，用来保留原始行协议解析的错误信息。

而 Protobuf 解析错误为：

```
proto: cannot parse invalid wire-format data
```

## DataKit 端的 Point 运行机制

DataKit 对采集到的数据原则上不做任何约束，即只要符合基本 Point 四元组结构，都支持上传到中心，所谓四元组指：

- measurement：指标集名称，不允许出现非 ascii 字符。如果出现，会对其 base64 编码
- tags：tag 可以没有
- fields：所有 Point 必须有至少一个 field
- time：time 实际上不是必须的，中心可以用接受时间作为 Point 时间（虽然不准）

即使采集到的数据中存在同名的 tag/field key，DataKit 仍坚持将其上传，但上传的过程中会自动修复这些瑕疵（按照[上述约束](#restrictions) 来修复），故上传到中心的数据，实际上跟采集到的数据会有一些差异，为什么这么做呢：

- 采集的过程中，这些数据存在，肯定有其相对合理的必要性
- 数据从采集到最终上传到中心，中间有一系列的处理操作（采样/过滤/文本处理/预聚合），这些操作可能依赖的那些字段，不应该过早被「自动修复」，这会给这些操作过程带来一些困扰
- 这些有瑕疵的字段，从数据本身来说，其存在只要不违反数据结构约束（比如 map 中不能存在同名的 key），就应该允许其存在。这些字段可能会违反中心的一些存储约束（比如字段值过长，某些存储不支持），但中心能选择合适的机制来修复这些数据，最大限度的保存住采集到的数据
