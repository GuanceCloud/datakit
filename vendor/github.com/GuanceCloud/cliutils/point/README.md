# DataKit 采集的数据结构定义

## Point

Point 是 DataKit 中最常用的一种数据表示形式，目前 Point 有两种表现形式：

1. 行协议
1. Protobuf

由于行协议表达能力有限，后面会逐渐迁移到 Protobuf，同时中心仍然长期支持行协议写入（不同的 API 版本）。除此之外，行协议和 Protobuf 之间并无实质性的差异，效率上而言（非极端情况）：

- 编码体积上，Protobuf 稍占上风（10% ~ 20%）
- 编码效率上，Protobuf 在效率（~2.5 倍）和内存占用（~8 倍）上，都较行协议更好
- 解码效率上，Protobuf 更慢一点，行协议解码更快（~1.3 倍），但其内存开销更大（~1.44 倍）。而 Protobuf 解码较慢，内存开销较小，但内存分配次数更多（可能导致碎片）。

以上测试，参见 `TestSize/BenchmarkEncode/BenchmarkDecode`。

## Point 的约束 {#restrictions}

Point 构建函数：

```golang
pt := NewPointV2(name []byte, kvs KVs, opts... Option)
```

> `NewPoint()` 已 Deprecated.

由于 Point 最终需要写入后端存储，故而受到后端存储的各种限制，但原则上采集端仍然做尽量少的限制，以保证数据采集的多样性和完整性。目前 Point 限制如下（本次放开/新加的约束，粗体表示）：

1. 概念上，仍旧区分 tag 和 field。但在底层数据结构上，它们的类型一致，都是 key-value，只是在特定的 key-value 上，会标注其是否为 tag（`IsTag`）。可以简单理解为 tag/field 都是 key。field 概念不再重要，在这些 key-value 对中，tag 只是更特化的一种 key-value，非 tag 的 key-value 都可以视为 field。

1. kvs 本质上是一个数组结构，在增加/删除的过程中，它**一直是有序存放**（以 key 排序）

1. 同一个 point 中，tag 和 field 之间不允许出现同名的 key。这一限制主要用来确保 DQL 查询时字段的唯一性。如果 tag/field 中出现同名 key，最终的 Point 中会移除 field 中对应的 key。

1. tag/field key 以及 value 目前不做协议上的限制，这一限制可以在应用层（比如 datakit/kodo）通过 `Option` 配置来控制

1. tag/field key/value 的内容均有几个方面的限制：

    - tag key 和 value 中均不允许出现换行字符，如果出现，SDK 会将其替换成空格。如字符串 `"abc\ndef"` 会被替换成 `"abc def"`。该限制无法通过 Option 控制。
    - Point **可以没有任何 key**（即可以没有 tag 和 field）。对于这样的 point，在上传过程中，自动忽略，但不会影响其它正常 point 的处理。
    - tag key 和 value 均不允许以反斜杠 `\` 结尾，如果出现，SDK 会将其移除掉。如字符串 `"abc\"` 会被替换成 `"abc"`。该限制无法通过 Option 控制。
    - 对日志类数据（非时序数据）而言，tag/field 的 key 中不允许出现 `.` 字符，如果出现，SDK 会将其替换成 `_`，如字符串 `"abc.def"` 会被替换成 `"abc_def"`。该限制可以通过 Option 控制。(**注：这一限制主要因后端使用 ES 作为存储所致，如果不用 ES 可以移除这一限制**)
    - tag/field 个数、 各个 tag/field key/value 长度（此处特指字符串长度）均可以通过 Option 设置。
    - 某些特定的数据类型（T/L/R/O/...），可以禁用某些 key（这里的 key 是带类型的，比如对象上不允许出现 key 为 `class` value 为 string 类型的 key-value，但如果某个 `class` 字段为 int，原则上还是允许存在的），一旦出现这些 key，最终的 Point 数据中，这些 key 将被自动移除掉。
		- 所有在构建点的过程中，如果出现自动调整（比如移除某些字段），均会在最终的 point 结构中出现对应的 warning 信息（通过 `point.Pretty()` 可以查看）

1. 构建 point 过程中出现的 warning/debug 信息，如果用 PB 结构上报，会一并上传到中心（kodo），中心在写入存储的时候，可以不用理会这些信息，主要是调试用

1. tag value 目前只支持 []byte/string 两种值。对 string，底层会转成 []byte 存放

1. field value 目前支持:

    - int8/int16/int32/uint8/uint16/uint32/ 均被转换成 int64 来处理
    - int64：最大支持 math.MaxInt64，对于大于该数值的 field，建议将其进行单位转换，避免其越界。比如字节（Byte）可以转换成 MB
      > 为什么此处限制了超过 `1<<63 - 1` 的整数？因为目前我们使用的 influxdb 版本（1.7）可能不支持 uint64
    - uint64：行协议编码时，如果该 uint64 数值小于 math.MaxInt64，会将其类型转成 int64（目前我们采用的行协议不支持 uint 解析）。否则丢弃该字段。PB 协议编码时，无此限制
    - float64：浮点数，暂无限制
    - bool：暂无限制
    - []byte/string：暂无限制
		- nil：空值字段（即什么都没有，对应各种语言里面的 null/nil/NULL 等），对于值为 nil 的 kv，编码过程中会将其丢弃

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

## Protobuf VS Line-Protocol

这两种不同的编码形态只影响数据的编码和解码，在日常的 point 使用过程中，基本上不用关注它们。

## 编解码

对一组 Point 而言，当前 SDK 提供简单的编码功能，主要用于将批量数据点变成字节流，便于网络传输。

其主要有几个特征：

- 对 N 个数据点，支持从个数的维度对最终的字节流分包
    - 老的分包是以行协议当前字节流长度来分包，这个相对不好控制。但以点数来分包，包大小因点长度不一致，会参差不齐。但后者比前者更利于编码实现（特别是针对 Protobuf Point 而言）。
- 支持传入一个回调函数，直接在编码过程中以类似迭代器的方式进行数据处理

示例：

```go
import (
	"github.com/GuanceCloud/cliutils/point"
)

enc := GetEncoder(WithEncEncoding(Protobuf),
	WithEncBatchSize(100)) // 100 point
defer PutEncoder(enc)

bufs, err := enc.Encode(points)
if err != nil {
	// error handling
}

for _, buf := range bufs {
	// send them to somewhere
}
```

### 解码

解码主要用于服务端用来处理外部输入的字节流数据，解码的时候需要提供对应的 option，用来指示输入的字节流是行协议还是 protobuf：

```golang
import (
	"github.com/GuanceCloud/cliutils/point"
)

dec := GetDecoder(WithDecEncoding(Protobuf)) // set encoding to protobuf
defer PutDecoder(dec)

points, err := dec.Decode(buffer)
if err != nil {
	// error handling
}

for _, p := range points {
	// do something with p
}
```

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

DataKit 对采集到的数据原则上不做任何约束，即只要符合基本 Point 结构，都支持上传到中心，所谓 Point 结构指：

- measurement：指标集名称，不允许出现非 ascii 字符。如果出现，会对其 base64 编码
- kvs：该 Point 对应的一组 key-value 值（数组），某些 key-value 可以被标注为 tag
- time：该 Point 对应的时间。time 实际上不是必须的，中心可以用接受时间作为 Point 时间（虽然不准）
