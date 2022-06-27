{{.CSS}}
# TDengine
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

TDEngine 是一款高性能、分布式、支持 SQL 的时序数据库 (Database)。在开通采集器之前请先阅读一些：[TDEngine-基本概念](https://docs.taosdata.com/concept/)

tdengine 采集器需要的连接 `taos_adapter` 才可以正常工作，taosAdapter 从 TDengine v2.4.0.0 版本开始成为 TDengine 服务端软件 的一部分   **注意版本**

本文主要是指标集的详细介绍，tdengine 集群安装不在本篇之内。

## 开启采集配置文件

```shell
cd /usr/local/datakit

cp conf.d/tdengine/tdengine.conf.sample conf.d/tdengine/tdengine.conf

vim conf.d/tdengine/tdengine.conf

### 配置文件具体配置

[[inputs.tdengine]]
  ## adapter config (Required)
  adapter_endpoint = "http://taosadapter.test.com"
  user = "<username>"
  password = "<password>"

  ## add tag (optional)
  [inputs.cpu.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
```

正确配置user、password 和 adapter_endpoint 并重启 datakit 之后,在观测云中指标集中可看到指标数据。

也可以使用 [仪表板模版](https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/dashboard/TDEngine-dashboard.json)。并在观测云-场景-仪表板-导入仪表板 导入仪表板即可。也可以在导入的仪表板中调整和修改。

## 指标集汇总

- [checkHealth](#checkHealth) 健康检查
- [td_cluster](#td_cluster) 集群状态
- [td_node dnode](#td_node) 状态
- [td_request](#td_request) tdengine request 指标集
- [td_database](#td_database) 数据库指标集
- [td_node_usage](#td_node_usage) node 资源使用情况指标集
- [td_adapter](#td_adapter) taosAdapter 监控指标集

### checkHealth

采集器在启动时会先进行一次健康检查，确保 taosAdapter 是可连接状态并使用用户名密码完成登录。

| 指标  | 说明                | queryType | sql             | 单位  | fields | tags |
|:----|:------------------|:---------:|:----------------|:----|:------:|:----:|
| -   | 检查数据库连接并使用用户名密码登陆 |    sql    | show databases; | -   |   -    |  -   |

sql 示例

```shell
taos> show databases;
              name              |      created_time       |   ntables   |   vgroups   | replica | quorum |  days  |           keep           |  cache(MB)  |   blocks    |   minrows   |   maxrows   | wallevel |    fsync    | comp | cachelast | precision | update |   status   |
====================================================================================================================================================================================================================================================================================
 biz_j2sxjtz3nfu9dumspjdthg_... | 2022-06-15 13:51:39.768 |           0 |           0 |       1 |      1 |     10 | 180                      |          16 |           6 |         100 |        4096 |        1 |        3000 |    2 |         0 | ns        |      0 | ready      |
 biz_cagq904o5bi4hbvp89rg_3d    | 2022-06-13 14:13:48.287 |           2 |           1 |       1 |      1 |     10 | 3650                     |          16 |           6 |         100 |        4096 |        1 |        3000 |    2 |         0 | ns        |      2 | ready      |
 biz_j2sxjtz3nfu9dumspjdthg_30d | 2022-06-15 13:51:29.999 |           0 |           0 |       1 |      1 |     10 | 30                       |          16 |           6 |         100 |        4096 |        1 |        3000 |    2 |         0 | ns        |      0 | ready      |
 biz_j2sxjtz3nfu9dumspjdthg_... | 2022-06-15 13:51:45.399 |           0 |           0 |       1 |      1 |     10 | 360                      |          16 |           6 |         100 |        4096 |        1 |        3000 |    2 |         0 | ns        |      0 | ready      |
 biz_j2sxjtz3nfu9dumspjdthg_14d | 2022-06-15 13:51:35.590 |           0 |           0 |       1 |      1 |     10 | 14                       |          16 |           6 |         100 |        4096 |        1 |        3000 |    2 |         0 | ns        |      0 | ready      |
Query OK, 5 row(s) in set (0.000586s)
```

### td_cluster

tdengine 集群状态

| 指标                               | 说明               | queryType | sql                                                                                                                          | 单位   |              fields              |       tags       |               补充               |
|:---------------------------------|:-----------------|:---------:|:-----------------------------------------------------------------------------------------------------------------------------|:-----|:--------------------------------:|:----------------:|:------------------------------:|
| first_ep                         | First EP Name    |    sql    | select last(first_ep),last(version),last(master_uptime) from log.cluster_info;                                               | s(秒) |          master_uptime           | first_ep,version | 指标只能是数值类型,将 first_ep 放到 tags 中 |
| expire_time                      | 企业版到期时间          |    sql    | select last(expire_time) from log.grants_info;                                                                               | s(秒) |           expire_time            |        -         |          非企业版则不会有这项指标          |
| timeseries_used,timeseries_total | 企业版已使用的测点数,总数    |    sql    | select max(timeseries_used) as used ,max(timeseries_total) as total from log.grants_info where ts >= now-10m and ts <= now ; | 个数   | timeseries_used,timeseries_total |        -         |          非企业版则不会有这项指标          |
| database_count                   | 数据库库总数           |    sql    | show databases ;                                                                                                             | 个数   |          database_count          |        -         |                                |
| table_count                      | 所有数据库的表数量之和      |    sql    | show databases;                                                                                                              | 个数   |           table_count            |        -         |          每个库中所包含表总数之和          |
| connections_total                | 当前连接个数           |    sql    | select * from log.cluster_info where ts >= now-10m and ts <= now;                                                            | 个数   |        connections_total         | first_ep,version |               -                |
| dnodes_total,dnodes_alive        | 每种资源的总数和存活数      |    sql    | select last(dnodes_total),last(dnodes_alive) from log.cluster_info where ts >= now-5m and ts <= now;                         | 个数   |    dnodes_total,dnodes_alive     |        -         |               -                |

该指标集需要查询的sql：

```shell
## 集群状态表
taos> select * from log.cluster_info limit 1;
             ts             |            first_ep            |   version    |    master_uptime     | monitor_interval | dnodes_total | dnodes_alive | mnodes_total | mnodes_alive | vgroups_total | vgroups_alive | vnodes_total | vnodes_alive | connections_total |
=====================================================================================================================================================================================================================================================================
 2022-05-18 07:58:51.488393 | tdengine-test-01:6030          | 2.4.0.18     |         683040.00000 |               30 |            2 |            2 |            1 |            1 |             3 |             3 |            3 |            3 |                 4 |
Query OK, 1 row(s) in set (0.002745s) 
```

### td_node

td_node 指标集主要是收集节点(node)的版本、end_point、status、create_time 等信息。这些信息都是字符串形式，并且在表中只有修改没有插入。

| 指标        | 说明             | queryType | sql          | 单位  |  fields   |                        tags                        | 补充  |
|:----------|:---------------|:---------:|:-------------|:----|:---------:|:--------------------------------------------------:|:---:|
| id,vnodes | dnode状态        |    sql    | show dnodes; | -   | id,vnodes | first_ep,version,status,create_time,offline_reason |     |
| id        | master node 状态 |    sql    | show mnodes; | -   |    id     |        end_point,role,role_time,create_time        |  -  |

该指标集需要查询的sql：

```shell 
## 节点状态
taos> show dnodes;
   id   |           end_point            | vnodes | cores  |   status   | role  |       create_time       |      offline reason      |
======================================================================================================================================
      1 | tdengine-test-01:6030          |      8 |      4 | ready      | any   | 2022-05-10 10:14:51.306 |                          |
      2 | tdengine-test-02:6030          |     13 |      4 | ready      | any   | 2022-05-10 10:16:24.309 |                          |
Query OK, 2 row(s) in set (0.000483s)

### master 状态 
taos> show mnodes;
   id   |           end_point            |     role     |        role_time        |       create_time       |
=============================================================================================================
      1 | tdengine-test-01:6030          | master       | 2022-06-15 02:54:07.178 | 2022-05-10 10:14:51.306 |
Query OK, 1 row(s) in set (0.000544s)
```

### td_request

数据的插入，查询，http请求等指标

| 指标                                    | 说明     | queryType | sql                                                                                                             | 单位  |                fields                 |   tags   | 补充  |
|:--------------------------------------|:-------|:---------:|:----------------------------------------------------------------------------------------------------------------|:----|:-------------------------------------:|:--------:|:---:|
| req_insert_rate,req_insert_batch_rate | 数据插入频率 |    sql    | select ts,req_insert_rate,req_insert_batch_rate,dnode_ep from log.dnodes_info where ts >= now-1m and ts <= now; | -   | req_insert_rate,req_insert_batch_rate | dnode_ep |  -  |
| req_select,req_select_rate            | 查询指标   |    sql    | select ts,req_select,req_select_rate,dnode_ep from log.dnodes_info where ts >= now-1m and ts <= now;            | -   |      req_select,req_select_rate       | dnode_ep |  -  |
| req_http,req_http_rate                | http请求 |    sql    | select ts,req_http,req_http_rate,dnode_ep from log.dnodes_info where ts >= now-1m and ts <= now;                | -   |        req_http,req_http_rate         | dnode_ep |  -  |

该指标集需要查询的sql：
```shell 
## insert
taos> select ts,req_insert_rate,req_insert_batch_rate,dnode_ep from log.dnodes_info where ts >= now-1m and ts <= now;
             ts             |   req_insert_rate    | req_insert_batch_rate |            dnode_ep            |
=============================================================================================================
 2022-06-17 08:21:30.000000 |              0.00000 |               0.00000 | tdengine-test-01:6030          |
 2022-06-17 08:21:30.000000 |              4.10000 |               1.96667 | tdengine-test-02:6030          |
Query OK, 2 row(s) in set (0.002442s)

## select
taos> select ts,req_select,req_select_rate,dnode_ep from log.dnodes_info where ts >= now-1m and ts <= now;
             ts             |      req_select       |   req_select_rate    |            dnode_ep            |
=============================================================================================================
 2022-06-17 08:22:00.000000 |                     0 |              0.00000 | tdengine-test-01:6030          |
 2022-06-17 08:22:00.000000 |                    11 |              0.36667 | tdengine-test-02:6030          |
 2022-06-17 08:22:30.000000 |                     1 |              0.03333 | tdengine-test-02:6030          |
Query OK, 3 row(s) in set (0.001362s)

## http
taos> select ts,req_http,req_http_rate,dnode_ep from log.dnodes_info where ts >= now-1m and ts <= now ;
             ts             |       req_http        |    req_http_rate     |            dnode_ep            |
=============================================================================================================
 2022-06-17 08:22:30.000000 |                     0 |              0.00000 | tdengine-test-01:6030          |
 2022-06-17 08:23:00.000000 |                     0 |              0.00000 | tdengine-test-01:6030          |
 2022-06-17 08:22:30.000000 |                     0 |              0.00000 | tdengine-test-02:6030          |
 2022-06-17 08:23:00.000000 |                     0 |              0.00000 | tdengine-test-02:6030          |
Query OK, 4 row(s) in set (0.001155s)
```

### td_database

数据库指标

| 指标         | 说明          | queryType | sql                                                                                                                            | 单位  |   fields   |          tags           |         补充          |
|:-----------|:------------|:---------:|:-------------------------------------------------------------------------------------------------------------------------------|:----|:----------:|:-----------------------:|:-------------------:|
| tables_num | VGroups 变化图 |    sql    | select last(ts),last(database_name),last(tables_num),last(status) from log.vgroups_info where ts > now-30s group by vgroup_id; | -   | tables_num | database_name,vgroup_id | 查询所有库的状态、ntable_num |


该指标集需要查询的sql：

```shell 
taos> select last(ts),last(database_name),last(tables_num),last(status) from log.vgroups_info where ts > now-30s group by vgroup_id;
             ts             |         database_name          | tables_num  |             status             |  vgroup_id  |
===========================================================================================================================
 2022-06-17 08:26:25.834388 | log                            |         750 | ready                          |           2 |
 2022-06-17 08:26:25.835130 | biz_ca0vgmko5bia4h133ge0       |          26 | ready                          |           4 |
 2022-06-17 08:26:25.831894 | biz_cagq904o5bi4hbvp89rg_3d    |           2 | ready                          |          16 |
 2022-06-17 08:26:25.836825 | biz_uwse9r634axkrve9h5awmf_7d  |           7 | ready                          |          33 |
 2022-06-17 08:26:25.835506 | biz_9exko7jznmzagtdjjyzfjf_30d |           0 | ready                          |          54 |
 2022-06-17 08:26:25.836315 | biz_9exko7jznmzagtdjjyzfjf_3d  |           0 | ready                          |          55 |
 2022-06-17 08:26:25.835908 | biz_9exko7jznmzagtdjjyzfjf_7d  |           1 | ready                          |          56 |
 2022-06-17 08:26:25.838038 | biz_9exko7jznmzagtdjjyzfjf_14d |           1 | ready                          |          57 |
 2022-06-17 08:26:25.837647 | biz_awiand9yzv7ubjh7ny9pje_7d  |           8 | ready                          |          59 |
 2022-06-17 08:26:25.838823 | biz_awiand9yzv7ubjh7ny9pje_14d |           1 | ready                          |          73 |
 2022-06-17 08:26:25.832763 | test                           |           1 | ready                          |          85 |
 2022-06-17 08:26:25.839316 | biz_9kuhanpzzcwoxvdpctxyrn_... |           0 | ready                          |          86 |
 2022-06-17 08:26:25.833571 | biz_9kuhanpzzcwoxvdpctxyrn_14d |           1 | ready                          |          87 |
 2022-06-17 08:26:25.838429 | biz_9kuhanpzzcwoxvdpctxyrn_3d  |           0 | ready                          |          88 |
 2022-06-17 08:26:25.834002 | biz_9kuhanpzzcwoxvdpctxyrn_... |           1 | ready                          |          89 |
 2022-06-17 08:26:25.833189 | biz_9kuhanpzzcwoxvdpctxyrn_7d  |           8 | ready                          |          90 |
 2022-06-17 08:26:25.837220 | biz_j2sxjtz3nfu9dumspjdthg_7d  |           3 | ready                          |          92 |
 2022-06-17 08:26:25.839836 | biz_j2sxjtz3nfu9dumspjdthg_14d |           3 | ready                          |          93 |
 2022-06-17 08:26:25.832332 | biz_j2sxjtz3nfu9dumspjdthg_30d |           1 | ready                          |          94 |
 2022-06-17 08:26:25.834763 | biz_sae3abxbbqe96bw6nfurhn_30d |           1 | ready                          |         142 |
 2022-06-17 08:26:25.831440 | biz_sae3abxbbqe96bw6nfurhn_7d  |          15 | ready                          |         143 |
Query OK, 21 row(s) in set (0.004188s)
```

### td_node_usage

资源使用情况

| 指标                                                   | 说明               | queryType | sql                                                                                                                                                                           | 单位  |              fields               |   tags   |     补充     |
|:-----------------------------------------------------|:-----------------|:----------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----|:---------------------------------:|:--------:|:----------:|
| uptime                                               | dnode 启动时间       | sql       | select last(ts),last(uptime)  from log.dnodes_info where errors=0 group by dnode_ep;                                                                                          | s   |              uptime               | dnode_ep |     -      |
| cpu_cores,vnodes_num,cpu_engine,mem_engine,mem_total | dnode 的 基本信息     | sql       | select last(ts),last(cpu_cores),last(vnodes_num),last(cpu_engine),last(mem_engine),last(mem_total) from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep;   | -   |             cpu_cores             | dnode_ep |   cpu核心数   |
| disk_used,disk_total,dick_percent                    | 磁盘使用率            | sql       | select last(ts),last(disk_used),last(disk_total), last(disk_used) / last(disk_total) as dick_percent from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep; | GB  | disk_used,disk_total,dick_percent | dnode_ep | 磁盘使用和总磁盘大小 |
| cpu_engine,cpu_system                                | CPU使用率           | sql       | select last(ts),avg(cpu_engine), avg(cpu_system) from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep;                                                     | GB  |       cpu_engine,cpu_system       | dnode_ep | 磁盘使用和总磁盘大小 |
| mem_engine,mem_system                                | 内存使用情况           | sql       | select last(ts),avg(mem_engine) as mem_engine, avg(mem_system) as mem_system from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep;                         | GB  |       mem_engine,mem_system       | dnode_ep |     内存     |
| io_read_taosd,io_write_taosd                         | io使用情况-磁盘读写      | sql       | select last(ts),avg(io_read_disk) as io_read_taosd, avg(io_write_disk) as io_write_taosd from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep;             | GB  |   io_read_taosd,io_write_taosd    | dnode_ep |     -      |
| net_in,net_out                                       | 网络 IO，总合网络 IO 速率 | sql       | select last(ts),avg(net_in) as net_in,avg(net_out) as net_out from log.dnodes_info where ts >= now-1m and ts <= now group by dnode_ep;                                        | GB  |          net_in,net_out           | dnode_ep |     -      |

相关的sql 示例：

```shell 
taos> select *  from log.dnodes_info where ts > now-1m limit 2;
             ts             |        uptime        |      cpu_engine      |      cpu_system      |  cpu_cores  |      mem_engine      |      mem_system      |      mem_total       |     disk_engine      |      disk_used       |      disk_total      |        net_in        |       net_out        |       io_read        |       io_write       |     io_read_disk     |    io_write_disk     |       req_http        |    req_http_rate     |      req_select       |   req_select_rate    |      req_insert       |  req_insert_success   |   req_insert_rate    |   req_insert_batch    | req_insert_batch_success | req_insert_batch_rate |        errors         | vnodes_num  |   masters   | has_mnode |  dnode_id   |            dnode_ep            |
===================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================================
 2022-06-15 14:29:30.000000 |        3096890.00000 |              0.05013 |              0.40942 |           4 |             32.74219 |           2416.22656 |           7688.00000 |              0.00000 |              4.97984 |            196.73599 |              0.01864 |              0.03961 |              0.00637 |              0.00755 |              0.00000 |              0.00104 |                     0 |              0.00000 |                     0 |              0.00000 |                    39 |                    39 |              1.30000 |                    37 |                       37 |               1.23333 |                     0 |           4 |           4 |      true |           1 | tdengine-test-01:6030          |
 2022-06-15 14:29:30.000000 |        3096808.00000 |              0.04187 |              0.10049 |           4 |             29.75781 |           1533.48828 |           7688.00000 |              0.00000 |              4.34319 |            196.73599 |              0.03511 |              0.02350 |              0.00255 |              0.00129 |              0.00000 |              0.00000 |                     0 |              0.00000 |                     1 |              0.03333 |                     0 |                     0 |              0.00000 |                     0 |                        0 |               0.00000 |                     0 |           9 |           9 |     false |           2 | tdengine-test-02:6030          |
Query OK, 2 row(s) in set (0.001577s)
```

### td_adapter

taosAdapter 监控指标集

| 指标                      | 说明           | queryType | sql                                                                                                                                                        | 单位  |         fields          |              tags              | 补充  |
|:------------------------|:-------------|:---------:|:-----------------------------------------------------------------------------------------------------------------------------------------------------------|:----|:-----------------------:|:------------------------------:|:---:|
| total_req_count         | 总请求数         |    sql    | select ts,count as total_req_count,endpoint,status_code,client_ip from log.taosadapter_restful_http_total where ts >= now-5m and ts <= now;                | s   |     total_req_count     | endpoint,status_code,client_ip |  -  |
| req_fail                | 请求失败数 code   |    sql    | select ts,count as req_fail,endpoint,status_code,client_ip from log.taosadapter_restful_http_fail where ts >= now-5m and ts <= now;                        | s   |        req_fail         | endpoint,status_code,client_ip |  -  |
| request_in_flight       | 正在处理的请求数     |    sql    | select ts,count as request_in_flight,endpoint,status_code,client_ip  from log.taosadapter_restful_http_request_in_flight where ts >= now-1m and ts <= now; | s   |    request_in_flight    | endpoint,status_code,client_ip |  -  |
| cpu_percent,mem_percent | % CPU和内存使用情况 |    sql    | select * from log.taosadapter_system where ts >= now-5m and ts <= now;                                                                                     | s   | cpu_percent,mem_percent |            endpoint            |  -  |

其中涉及到4张表：

```shell 
## log.taosadapter_restful_http_total
taos> select ts,count as total_req_count,endpoint,status_code,client_ip from log.taosadapter_restful_http_total where ts >= now-1m and ts <= now;
             ts             |    total_req_count    |            endpoint            | status_code |           client_ip            |
=====================================================================================================================================
 2022-06-17 09:12:54.776636 |                     6 | tdengine-test-01:6041          |         200 | 172.16.5.29                    |
 2022-06-17 09:13:24.775822 |                     1 | tdengine-test-01:6041          |         200 | 172.16.5.29                    |
 2022-06-17 09:13:24.775822 |                     1 | tdengine-test-01:6041          |         200 | 172.16.5.29                    |
 2022-06-17 09:13:24.775822 |                     1 | tdengine-test-01:6041          |         200 | 172.16.5.29                    |
 2022-06-17 09:13:24.775822 |                     2 | tdengine-test-01:6041          |         200 | 172.16.5.29                    |
 2022-06-17 09:12:54.776636 |                     1 | tdengine-test-01:6041          |         204 | 172.16.5.29                    |


## log.taosadapter_restful_http_fail
taos> select ts,count as req_fail,endpoint,status_code,client_ip from log.taosadapter_restful_http_fail where ts >= now-1m and ts <= now;
             ts             |       req_fail        |            endpoint            | status_code |           client_ip            |
=====================================================================================================================================
 2022-06-17 09:13:54.776692 |                     1 | tdengine-test-01:6041          |         200 | 172.16.5.29                    |
 2022-06-17 09:13:54.727716 |                     1 | tdengine-test-02:6041          |         200 | 172.16.5.29                    |
 2022-06-17 09:13:54.727716 |                     1 | tdengine-test-02:6041          |         200 | 172.16.5.29                    |
Query OK, 3 row(s) in set (0.002229s)

### log.taosadapter_restful_http_request_in_flight
taos> select *  from log.taosadapter_restful_http_request_in_flight where ts >= now-10h and ts <= now;
             ts             |         count         |            endpoint            |
======================================================================================
 2022-06-17 05:36:54.776171 |                     1 | tdengine-test-01:6041          |
Query OK, 1 row(s) in set (0.001103s)

## log.taosadapter_system
taos> select * from log.taosadapter_system where ts >= now-1m and ts <= now;
             ts             |        cpu_percent        |        mem_percent        |            endpoint            |
======================================================================================================================
 2022-06-17 09:17:54.776728 |               0.250068000 |               1.158662000 | tdengine-test-01:6041          |
 2022-06-17 09:18:24.775850 |               0.083382000 |               1.165318000 | tdengine-test-01:6041          |
 2022-06-17 09:17:54.728596 |               0.166702000 |               1.185488000 | tdengine-test-02:6041          |
 2022-06-17 09:18:24.727699 |               0.000000000 |               1.185488000 | tdengine-test-02:6041          |
Query OK, 4 row(s) in set (0.001381s)
```

>- 有些表中没有 `ts` 字段，会使用 `time.now()` 代替
