
# 通过 DQL 查询数据
---

DataKit 支持以交互式方式执行 DQL 查询，在交互模式下，DataKit 自带语句补全功能：

> 通过 `datakit help dql` 可获取更多命令行参数帮助。

```shell
datakit dql      # 或者 datakit -Q
```

<figure markdown>
  ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/dk-dql-gif.gif){ width="800" }
  <figcaption> DQL 交互执行示例 </figcaption>
</figure>

Tips：

- 输入 `echo_explain` 即可看到后端查询语句
- 为避免显示太多 `nil` 查询结果，可通过 `disable_nil/enable_nil` 来开关
- 支持查询语句模糊搜，如 `echo_explain` 只需要输入 `echo` 或 `exp` 即可弹出提示，通过制表符（Tab）即可选择下拉提示
- DataKit 会自动保存前面多次运行的 DQL 查询历史（最大 5000 条），可通过上下方向键来选择

> 注：Windows 下，请在 Powershell 中执行 `datakit dql`

## 单次执行 DQL 查询 {#dql-once}

关于 DQL 查询，DataKit 支持运行单条 DQL 语句的功能：

```shell
# 单次执行一条查询语句
datakit dql --run 'cpu limit 1'

# 将执行结果写入 CSV 文件
datakit dql --run 'O::HOST:(os, message)' --csv="path/to/your.csv"

# 强制覆盖已有 CSV 文件
datakit dql --run 'O::HOST:(os, message)' --csv /path/to/xxx.csv --force

# 将结果写入 CSV 的同时，在终端也显示查询结果
datakit dql --run 'O::HOST:(os, message)' --csv="path/to/your.csv" --vvv
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

## DQL 查询结果 JSON 化 {#json-result}

以 JSON 形式输出结果，但 JSON 模式下，不会输出一些统计信息，如返回行数、时间消耗等（以保证 JSON 可直接解析）

```shell
datakit dql --run 'O::HOST:(os, message)' --json

# 如果字段值是 JSON 字符串，则自动做 JSON 美化（注意：JSON 模式下（即 --json），`--auto-json` 选项无效）
datakit dql --run 'O::HOST:(os, message)' --auto-json
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

## 查询特定工作空间的数据 {#query-on-wksp}

通过指定不同的 Token 来查询其它工作空间的数据：

```shell
datakit dql --run 'O::HOST:(os, message)' --token <your-token>
```
