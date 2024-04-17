# DataKit 采集的数据结构定义

## Point

Point 是 DataKit 中最常用的一种数据表示形式，目前 Point 有两种表现形式：

1. 行协议
1. Protobuf

由于行协议表达能力有限，后面会逐渐迁移到 Protobuf，同时中心仍然长期支持行协议写入（不同的 API 版本）。除此之外，行协议和 Protobuf 之间并无实质性的差异，效率上而言（非极端情况）：

- 编码体积上，Protobuf 和航协议相当（随机出来的数据点）
- 编码效率上，Protobuf 在效率（X10）和内存占用（X8）上，都较行协议更好
- 解码效率上，Protobuf 也更好（X5）

以上测试，参见 `TestEncodePayloadSize/BenchmarkEncode/BenchmarkDecode`。 可参考如下结果：

``` shell
### BAD
$ CGO_CFLAGS=-Wno-undef-prefix go test -run XXX -test.benchmem -test.v -bench BenchmarkDecode
goos: darwin
goarch: arm64
pkg: github.com/GuanceCloud/cliutils/point
BenchmarkDecode
BenchmarkDecode/bench-decode-lp
BenchmarkDecode/bench-decode-lp-10                   100          12091620 ns/op         4670056 B/op      90282 allocs/op
BenchmarkDecode/bench-decode-pb
BenchmarkDecode/bench-decode-pb-10                   550           2172161 ns/op         3052850 B/op      70024 allocs/op
BenchmarkDecode/bench-decode-json
BenchmarkDecode/bench-decode-json-10                  96          12690984 ns/op         6269919 B/op     137321 allocs/op
PASS
ok      github.com/GuanceCloud/cliutils/point   4.995s

$ CGO_CFLAGS=-Wno-undef-prefix go test -run XXX -test.benchmem -test.v -bench BenchmarkEncode
goos: darwin
goarch: arm64
pkg: github.com/GuanceCloud/cliutils/point
BenchmarkEncode
BenchmarkEncode/bench-encode-json
BenchmarkEncode/bench-encode-json-10                  98          10294642 ns/op         7724543 B/op      60288 allocs/op
BenchmarkEncode/bench-encode-lp
BenchmarkEncode/bench-encode-lp-10                   241           4900589 ns/op         8675253 B/op      41043 allocs/op
BenchmarkEncode/bench-encode-pb
BenchmarkEncode/bench-encode-pb-10                  2727            452257 ns/op         1115027 B/op         16 allocs/op
BenchmarkEncode/v2-encode-pb
BenchmarkEncode/v2-encode-pb-10                     2754            438385 ns/op              15 B/op          0 allocs/op
BenchmarkEncode/v2-encode-lp
BenchmarkEncode/v2-encode-lp-10                      268           4461083 ns/op         5258539 B/op      39066 allocs/op
PASS
ok      github.com/GuanceCloud/cliutils/point   7.483s
```

## Point 的约束 {#restrictions}

Point 构建函数：

```golang
pt := NewPointV2(name []byte, kvs KVs, opts... Option)
```

> `NewPoint()` 已 Deprecated.

由于 Point 最终需要写入后端存储，故而受到后端存储的各种限制，但原则上采集端仍然做尽量少的限制，以保证数据采集的多样性和完整性。目前 Point 限制如下（本次放开/新加的约束，粗体表示）：

1. 概念上，仍旧区分 tag 和 field。但在底层数据结构上，它们的类型一致，都是 key-value，只是在特定的 key-value 上，会标注其是否为 tag（`IsTag`）。可以简单理解为 tag/field 都是 key。field 概念不再重要，在这些 key-value 对中，tag 只是更特化的一种 key-value，非 tag 的 key-value 都可以视为 field。

1. kvs 本质上是一个数组结构，以增加顺序排列

1. 同一个 point 中，tag 和 field 之间不允许出现同名的 key。这一限制主要用来确保 DQL 查询时字段的唯一性。如果 tag/field 中出现同名 key，后出现的 key 将不再生效。

1. tag/field key 以及 value 目前不做协议上的限制，这一限制可以在应用层（比如 datakit/kodo）通过 `Option` 配置来控制

1. tag/field key/value 的内容均有几个方面的限制：

    - tag key 和 value 中均不允许出现换行字符，如果出现，SDK 会将其替换成空格。如字符串 `"abc\ndef"` 会被替换成 `"abc def"`。该限制无法通过 Option 控制。
    - Point **可以没有任何 key**（即可以没有 tag 和 field）。对于这样的 point，在上传过程中，自动忽略，但不会影响其它正常 point 的处理。
    - tag key 和 value 均不允许以反斜杠 `\` 结尾，如果出现，SDK 会将其移除掉。如字符串 `"abc\"` 会被替换成 `"abc"`。该限制无法通过 Option 控制。
    - 对日志类数据（非时序数据）而言，tag/field 的 key 中不允许出现 `.` 字符，如果出现，SDK 会将其替换成 `_`，如 key `"abc.def"` 会被替换成 `"abc_def"`。该限制可以通过 Option 控制。(**注：这一限制主要因后端使用 ES 作为存储所致，如果不用 ES 可以移除这一限制**)
    - tag/field 个数、 各个 tag/field key/value 长度（此处特指字符串长度）均可以通过 Option 设置。
    - 某些特定的数据类型（T/L/R/O/...），可以禁用某些 key（这里的 key 是带类型的，比如对象上不允许出现 key 为 `class` 的字段，一旦出现这些 key，最终的 Point 数据中，这些 key 将被自动移除掉。
		- 所有在构建点的过程中，如果出现自动调整（比如移除某些字段），均会在最终的 point 结构中出现对应的 warning 信息（通过 `point.Pretty()` 可以查看）

1. 构建 point 过程中出现的 warning/debug 信息，如果用 PB 结构上报，会一并上传到中心（kodo），中心在写入存储的时候，可以不用理会这些信息，主要是调试用

1. tag value 目前只支持 string 值

1. field value 目前支持:

    - int8/int16/int32/int64 表示有符号整数
    - uint8/uint16/uint32/uint64 表示无符号整数
    - float64：浮点数，暂无限制
    - bool：暂无限制
    - []byte：字节流，可以存放二进制数据
    - string：暂无限制
    - []any：数组类型，数组元素必须是基础类型（int/uint/float/string/bool），且数组内类型一致。
