{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit DQL 查询

DataKit 支持以交互式方式执行 DQL 查询，在交互模式下，DataKit 自带语句补全功能：

```shell
datakit --dql      # 或者 datakit -Q
dql > cpu limit 1
-----------------[ 1.cpu ]-----------------
             cpu 'cpu-total'
            host 'tan-air.local'
            time 2021-06-23 10:06:03 +0800 CST
       usage_irq 0
      usage_idle 56.928839
      usage_nice 0
      usage_user 19.825218
     usage_guest 0
     usage_steal 0
     usage_total 43.071161
    usage_iowait 0
    usage_system 23.245943
   usage_softirq 0
usage_guest_nice 0
---------
1 rows, cost 13.55119ms
```

Tips：

- 输入 `echo_explain` 即可看到后端查询语句
- 为避免显示太多 `nil` 查询结果，可通过 `disable_nil/enable_nil` 来开关
- 支持查询语句模糊搜，如 `echo_explain` 只需要输入 `echo` 或 `exp` 即可弹出提示，**通过 `Tab` 即可选择下拉提示**
- DataKit 会自动保存前面多次运行的 DQL 查询历史（最大 5000 条），可通过上下方向键来选择

> 注：Windows 下，请在 Powershell 中执行 `datakit --dql` 或 `datakit -Q`

#### 单次执行 DQL 查询

关于 DQL 查询，DataKit 支持运行单条 DQL 语句的功能：

```shell
# 单次执行一条查询语句
datakit --run-dql 'cpu limit 1'

# 将执行结果写入 CSV 文件
datakit --run-dql 'O::HOST:(os, message)' --csv="path/to/your.csv"

# 强制覆盖已有 CSV 文件
datakit --run-dql 'O::HOST:(os, message)' --csv /path/to/xxx.csv --force

# 将结果写入 CSV 的同时，在终端也显示查询结果
datakit --run-dql 'O::HOST:(os, message)' --csv="path/to/your.csv" --vvv
```

导出的 CSV 文件样式示例：

```shell
name,active,available,available_percent,free,host,time
mem,2016870400,2079637504,24.210166931152344,80498688,achen.local,1635242524385
mem,2007961600,2032476160,23.661136627197266,30900224,achen.local,1635242534385
mem,2014437376,2077097984,24.18060302734375,73502720,achen.local,1635242544382
```

注意：

- 第一列是查询的指标集名称
- 之后各列是该采集器对应的各项数据
- 当字段为空时，对应列也为空

#### DQL 查询结果 JSON 化

以 JSON 形式输出结果，但 JSON 模式下，不会输出一些统计信息，如返回行数、时间消耗等（以保证 JSON 可直接解析）

```shell
datakit --run-dql 'O::HOST:(os, message)' --json
datakit -Q --json

# 如果字段值是 JSON 字符串，则自动做 JSON 美化（注意：JSON 模式下（即 --json），`--auto-json` 选项无效）
datakit --run-dql 'O::HOST:(os, message)' --auto-json
-----------------[ r1.HOST.s1 ]-----------------
message ----- json -----  # JSON 开始处有明显标志，此处 message 为字段名
{
  "host": {
    "meta": {
      "host_name": "www",
  ....                    # 此处省略长文本
  "config": {
    "ip": "10.100.64.120",
    "enable_dca": false,
    "http_listen": "localhost:9529",
    "api_token": "tkn_f2b9920f05d84d6bb5b14d9d39db1dd3"
  }
}
----- end of json -----   # JSON 结束处有明显标志
     os 'darwin'
   time 2021-09-13 16:56:22 +0800 CST
---------
8 rows, 1 series, cost 4ms
```

#### 查询特定工作空间的数据

通过指定不同的 Token 来查询其它工作空间的数据：

```shell
datakit --run-dql 'O::HOST:(os, message)' --token <your-token>
datakit -Q --token <your-token>
```

