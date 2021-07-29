{{.CSS}}

- 发布日期：{{.ReleaseDate}}

# DataFlux 无数据排查

大家在部署完数据采集之后（通过 DataKit 或 Function 采集），有时候在 DataFlux 的页面上看不到对应的数据更新，每次排查起来都心力憔悴，为了缓解这一状况，可按照如下的一些步骤，来逐步围歼「为啥没有数据」这一问题。

## 检查 DataWay 连接是否正常

```shell
curl http[s]://your-dataway-addr:port
```

对 SAAS 而言，一般这样：

```shell
curl https://openway.dataflux.cn
```

如果得到如下结果，则表示网络是有问题的：

```
curl: (6) Could not resolve host: openway.dataflux.cn
```

## 检查机器时间

在 Linux/Mac 上，输入 `date` 即可查看当前系统时间：

```shell
date
Wed Jul 21 16:22:32 CST 2021
```

有些情况下，这里可能显示成这样：

```
Wed Jul 21 08:22:32 UTC 2021
```

这是因为，前者是中国东八区时间，后者是格林威治时间，两者相差 8 小时，但实际上，这两个时间的时间戳是一样的。

如果当前系统的时间跟你的手机时间相差甚远，特别是，它如果超前了，那么 DataFlux 上是看不到这些「将来」的数据的。

另外，如果时间滞后，你会看到一些老数据，不要以为发生了灵异事件，事实上，极有可能是 DataKit 所在机器的时间还停留在过去。

## 查看 Monitor 页面

如果是 Windows 系统，打开浏览器，输入

```
# 视实际绑定的网卡以及端口而定
http://localhost:9529/monitor
```

如果是 Linux/Mac，直接在终端输入（无需修改 `http_listen` 配置）：

> DataKit [1.1.7-rc7](changelog#494d6cd5) 才支持这一功能

```shell
datakit -M

# 或者

datakit --monitor --vvv
```

页面上会显示每个采集器的运行情况，如果某个采集器有误，在「当前错误（时间）」这一列，能看到具体的报错信息以及报错时间。

## 通过 DQL 查看是否有数据产生

在 Windows/Linux/Mac 上，这一功能均支持，其中 Windows 需在 Powershell 中执行

> DataKit [1.1.7-rc7](changelog#494d6cd5) 才支持这一功能

```shell
datakit --dql
> 这里即可输入 DQL 查询语句...
```

对于无数据排查，建议对照着采集器文档，看对应的指标集叫什么名字，以 MySQL 采集器为例，目前文档中有如下几个指标集：

- `mysql`
- `mysql_schema`
- `mysql_innodb`
- `mysql_table_schema`
- `mysql_user_status`

如果 MySQL 这个采集器没数据，可检查 `mysql` 这个指标集是否有数据：

``` python
#
# 查看 mysql 采集器上，指定主机上（这里是 tan-air.local）的最近一条 mysql 的指标
#
M::mysql {host='tan-air.local'} order by time desc limit 1
```

查看某个主机对象是不是上报了，这里的 `tan-air.local` 就是预期的主机名：

```python
O::HOST {host='tan-air.local'}
```

查看已有的 APM（tracing）数据分类：

```python
show_tracing_service()
```

以此类推，如果数据确实上报了，那么通过 DQL 总能找到，至于前端不显示，可能是其它过滤条件给挡掉了。通过 DQL，不管是 DataKit 采集的数据，还是其它手段（如 Function）采集的数据，都可以零距离查看原式数据，特别便于 Debug。

## 查看 DataKit 程序日志是否有异常

Windows 平台：

```powershell
# 通过 Powershell 给出最近 10 个 ERROR, WARN 级别的日志
Select-String -Path 'C:\Program Files\datakit\log' -Pattern "ERROR", "WARN"  | Select-Object Line -Last 10
```

Linux/Mac 平台：

```shell
# 给出最近 10 个 ERROR, WARN 级别的日志
cat /var/log/datakit/log | grep "WARN\|ERROR" | tail -n 10
```

如果日志中发现诸如 `Beyond...` 这样的描述，一般情况下，是因为数据量超过了免费额度。

对于远程给 DataKit 打数据的采集，可查看 gin.log 来查看是否有远程数据发送过来：

```shell
tail -f /var/log/datakit/gin.log
```
