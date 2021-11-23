{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Pipeline 调试

Pipeline 编写较为麻烦，为此，DataKit 中内置了简单的调试工具，用以辅助大家来编写 Pipeline 脚本。

## 调试 grok 和 pipeline

指定 pipeline 脚本名称（`--pl`，pipeline 脚本必须放在 `<DataKit 安装目录>/pipeline` 目录下），输入一段文本（`--txt`）即可判断提取是否成功

```shell
datakit --pl your_pipeline.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms'
Extracted data(cost: 421.705µs): # 表示切割成功
{
	"code"   : "io/io.go: 458",       # 对应代码位置
	"level"  : "DEBUG",               # 对应日志等级
	"module" : "io",                  # 对应代码模块
	"msg"    : "post cost 6.87021ms", # 纯日志内容
	"time"   : 1610358231887000000    # 日志时间(Unix 纳秒时间戳)
}

# 提取失败示例
datakit --pl other_pipeline.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.g o:458  post cost 6.87021ms'
No data extracted from pipeline
```

由于 grok pattern 数量繁多，人工匹配较为麻烦。DataKit 提供了交互式的命令行工具 `grokq`（grok query）：

```Shell
datakit --grokq
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

> 注：Windows 下，请在 Powershell 中执行调试。

### Pipeline 字段命名注意事项

由于[行协议约束](apis#f54b954f)，在切割出来的字段中（在行协议中，它们都是 Field），不宜有任何 tag 字段，这些 Tag 包含如下几类：

- 各个具体采集器中，用户自行配置增加的 Tag，如 `[inputs.nginx.tags]` 下可增加各种 Tag
- DataKit 全局 Tag，如 `host`。当然，这个全局 Tag 用户也能自行配置
- 日志采集器默认会增加 `source/service` 这两个 Tag，在 Pipeline 中也不宜出现这两个字段切割

一旦 Pipeline 切割出来的字段中带有上述任何一个 Tag key（大小写敏感），都会导致如下数据报错，故建议在 Pipeline 切割中，绕开这些字段命名。

```shell
# 该错误在 DataKit monitor 中能看到
same key xxx in tag and field
```
