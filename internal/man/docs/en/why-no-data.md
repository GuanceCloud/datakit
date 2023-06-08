{{.CSS}}
# How to Troubleshoot No Data Problems
---

After deploying data collection (collected through DataKit or Function), sometimes you can't see the corresponding data update on the page of Guance Cloud, and you are tired every time you check it. In order to alleviate this situation, you can adopt the following steps to gradually encircle the problem of "why there is no data".

## Checking input config {#check-input-conf}

[:octicons-tag-24: Version-1.9.0](changelog.md#cl-1.9.0)

We can run Datakit command to test if input configure able to collect data. For input `disk`, we can test like this:

``` shell
$ datakit debug --input-conf /usr/local/datakit/conf.d/host/disk.conf
loading /Users/tanbiao/datakit/conf.d/host/disk.conf with 1 inputs...
running input "disk"(0th)...
disk,device=/dev/disk3s1s1,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631702195i,inodes_total_mb=1631i,inodes_used=349475i,inodes_used_mb=0i,inodes_used_percent=0.02141781760611041,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064064000
disk,device=/dev/disk3s6,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631352732i,inodes_total_mb=1631i,inodes_used=12i,inodes_used_mb=0i,inodes_used_percent=0.0000007355858585707753,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064243000
disk,device=/dev/disk3s2,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631353840i,inodes_total_mb=1631i,inodes_used=1120i,inodes_used_mb=0i,inodes_used_percent=0.00006865463350366712,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064254000
disk,device=/dev/disk3s4,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631352837i,inodes_total_mb=1631i,inodes_used=117i,inodes_used_mb=0i,inodes_used_percent=0.000007171961659450622,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064260000
disk,device=/dev/disk1s2,fstype=apfs free=503996416i,inodes_free=4921840i,inodes_free_mb=4i,inodes_total=4921841i,inodes_total_mb=4i,inodes_used=1i,inodes_used_mb=0i,inodes_used_percent=0.00002031760067015574,total=524288000i,used=20291584i,used_percent=3.8703125 1685509141064266000
disk,device=/dev/disk1s1,fstype=apfs free=503996416i,inodes_free=4921840i,inodes_free_mb=4i,inodes_total=4921873i,inodes_total_mb=4i,inodes_used=33i,inodes_used_mb=0i,inodes_used_percent=0.000670476462923769,total=524288000i,used=20291584i,used_percent=3.8703125 1685509141064271000
disk,device=/dev/disk1s3,fstype=apfs free=503996416i,inodes_free=4921840i,inodes_free_mb=4i,inodes_total=4921892i,inodes_total_mb=4i,inodes_used=52i,inodes_used_mb=0i,inodes_used_percent=0.0010565042873756677,total=524288000i,used=20291584i,used_percent=3.8703125 1685509141064276000
disk,device=/dev/disk3s5,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1634318356i,inodes_total_mb=1634i,inodes_used=2965636i,inodes_used_mb=2i,inodes_used_percent=0.18146011694186712,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064280000
disk,device=/dev/disk2s1,fstype=apfs free=3697000448i,inodes_free=36103520i,inodes_free_mb=36i,inodes_total=36103578i,inodes_total_mb=36i,inodes_used=58i,inodes_used_mb=0i,inodes_used_percent=0.00016064889745830732,total=5368664064i,used=1671663616i,used_percent=31.137422570532436 1685509141064285000
disk,device=/dev/disk3s1,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631702197i,inodes_total_mb=1631i,inodes_used=349477i,inodes_used_mb=0i,inodes_used_percent=0.0214179401512444,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064289000
# 10 points("M") from disk, cost 1.544792ms | Ctrl+c to exit.
```

Here Datakit will start input `disk` and print the collected points to terminal. At the buttom, there will show some result about these point:

- Collected points and its category(Here we got 10 points of category `Metric`)
- Input name, here is `disk`
- Collect cost, here is *1.544ms*

We can interrupt the test by *Ctrl + c*. We can also change the `interval`(if exist) config to get test result more fequently.

<!-- markdownlint-disable MD046 -->
???+ attention

    Some inputs that accept HTTP requests(such as DDTrace/RUM), we need to setup a HTTP server(`--http-listen=[IP:Port]`), then use some tools (such as `curl`) to post testing data to Datakit's HTTP server. For detailed info, see help with `datakit help debug`.
<!-- markdownlint-enable -->

## Check whether the DataWay connection is normal  {#check-connection}

```shell
curl http[s]://your-dataway-addr:port
```

For SAAS, this is generally the case:

```shell
curl https://openway.guance.com
```

If the following results are obtained, the network is problematic:

```
curl: (6) Could not resolve host: openway.guance.com
```

If you find an error log such as the following, it indicates that there is something wrong with the connection with DataWay, which may be restricted by the firewall:

```shell
request url https://openway.guance.com/v1/write/xxx/token=tkn_xxx failed:  ... context deadline exceeded...
```

## Check Machine Time {#check-local-time}

On Linux/Mac, enter `date` to view the current system time:

```shell
date
Wed Jul 21 16:22:32 CST 2021
```

In some cases, this may appear as follows:

```
Wed Jul 21 08:22:32 UTC 2021
```

This is because the former is China's East Eight District Time, while the latter is Greenwich Mean Time, with a difference of 8 hours, but in fact, the timestamps of these two times are the same.

If the time of the current system is far from that of your mobile phone, especially if it is ahead of time, there is no "future" data on Guance Cloud.

In addition, if the time lag, you will see some old data. Don't think that paranormal happened. In fact, it is very likely that the time of DataKit's machine is still in the past.

## See if the data is blacklisted or discarded by Pipeline {#filter-pl}

If [a blacklist](datakit-filter)(such as a log blacklist) is configured, newly collected data may be filtered out by the blacklist.

Similarly, if data is [discarded](pipeline#fb024a10) in Pipeline, it may also cause the center to see the data.

## View monitor page {#monitor}

See [here](datakit-monitor.md)

## Check whether there is data generated through dql {#dql}

This function is supported on Windows/Linux/Mac, where Windows needs to be executed in Powershell.

> This feature is supported only in DataKit [1.1.7-rc7](changelog#cl-1.1.7-rc7).

```shell
datakit dql
> Here you can enter the DQL query statement...
```

For non-data investigation, it is recommended to compare the collector document to see the name of the corresponding indicator set. Take MySQL collector as an example. At present, there are the following indicator sets in the document:

- `mysql`
- `mysql_schema`
- `mysql_innodb`
- `mysql_table_schema`
- `mysql_user_status`

If mysql does not have data, check whether `mysql` has data:

``` python
#
# Look at the metrics of the most recent mysql on the mysql collector, specifying the host (in this case tan-air.local)
#
M::mysql {host='tan-air.local'} order by time desc limit 1
```

To see if a host object has been reported, where `tan-air.local` is the expected host name:

```python
O::HOST {host='tan-air.local'}
```

View the existing APM (tracing) data classification:

```python
show_tracing_service()
```

By analogy, if the data is reported, it can always be found through DQL. As for the front end, it may be blocked by other filtering conditions. Through DQL, no matter the data collected by DataKit or other means (such as Function), the original data can be viewed at zero distance, which is especially convenient for Debug.

## Check the DataKit program log for exceptions {#check-log}

The last 10 ERROR and WARN logs are given through Shell/Powershell

```shell
# Shell
cat /var/log/datakit/log | grep "WARN\|ERROR" | tail -n 10

# Powershell
Select-String -Path 'C:\Program Files\datakit\log' -Pattern "ERROR", "WARN"  | Select-Object Line -Last 10
```

- If a description such as `Beyond...` is found in the log, it is generally because the amount of data exceeds the free amount.
- If some words such as `ERROR/WARN` appear, it indicates that DataKit has encountered some problems in general.

### View the running log of a single collector {#check-input-log}

If no exception is found, you can directly view the operation log of a single collector:

```shell
# shell
tail -f /var/log/datakit/log | grep "<采集器名称>" | grep "WARN\|ERROR"

# powershell
Get-Content -Path "C:\Program Files\datakit\log" -Wait | Select-String "<采集器名称>" | Select-String "ERROR", "WARN"
```

也You can also remove the filter such as `ERROR/WARN` and directly view the corresponding collector log. If you don't have enough logs, open the debug log in `datakit.conf` to see more logs:

```
# DataKit >= 1.1.8-rc0
[logging]
	...
	level = "debug" # 将默认的 info 改为 debug
	...

# DataKit < 1.1.8-rc0
log_level = "debug"
```

### View gin.log {#check-gin-log}
 
For remote data collection to DataKit, you can check gin.log to see if there is remote data sent: 

```shell
tail -f /var/log/datakit/gin.log
```
