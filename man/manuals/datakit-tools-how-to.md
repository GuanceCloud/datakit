{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit 各种小工具介绍

DataKit 内置很多不同的小工具，便于大家日常使用。可通过如下命令来查看 DataKit 的命令行帮助：

```shell
datakit -h
```

>注意：因不同平台的差异，具体帮助内容会有差别。

## 查看 DataKit 运行情况

> 当前的 monitor 查看方式已经废弃（仍然可用，不久将废弃），新的 monitor 功能[参见这里](datakit-monitor)

在终端即可查看 DataKit 运行情况，其效果跟浏览器端 monitor 页面相似：

```shell
datakit --monitor     # 或者 datakit -M

# 同时可查看采集器开启情况：
datakit -M --vvv
```

> 注：Windows 下暂不支持在终端查看 monitor 数据，只能在浏览器端查看。

## 检查采集器配置是否正确

编辑完采集器的配置文件后，可能某些配置有误（如配置文件格式错误），通过如下命令可检查是否正确：

```shell
sudo datakit --check-config
------------------------
checked 13 conf, all passing, cost 22.27455ms
```

## 查看帮助文档

为便于大家在服务端查看 DataKit 帮助文档，DataKit 提供如下交互式文档查看入口（Windows 不支持）：

```shell
datakit --man
man > nginx
(显示 Nginx 采集文档)
man > mysql
(显示 MySQL 采集文档)
man > Q               # 输入 Q 或 exit 退出
```

## 查看工作空间信息

为便于大家在服务端查看工作空间信息，DataKit 提供如下命令查看：

```shell
datakit --workspace-info
{
  "token": {
    "ws_uuid": "wksp_2dc431d6693711eb8ff97aeee04b54af",
    "bill_state": "normal",
    "ver_type": "pay",
    "token": "tkn_2dc438b6693711eb8ff97aeee04b54af",
    "db_uuid": "ifdb_c0fss9qc8kg4gj9bjjag",
    "status": 0,
    "creator": "",
    "expire_at": -1,
    "create_at": 0,
    "update_at": 0,
    "delete_at": 0
  },
  "data_usage": {
    "data_metric": 96966,
    "data_logging": 3253,
    "data_tracing": 2868,
    "data_rum": 0,
    "is_over_usage": false
  }
}
```

## 查看 DataKit 相关事件

DataKit 运行过程中，一些关键事件会以日志的形式进行上报，比如 DataKit 的启动、采集器的运行错误等。在命令行终端，可以通过 dql 进行查询。

```shell
sudo datakit --dql

dql > L::datakit limit 10;

-----------------[ r1.datakit.s1 ]-----------------
    __docid 'L_c6vvetpaahl15ivd7vng'
   category 'input'
create_time 1639970679664
    date_ns 835000
       host 'demo'
    message 'elasticsearch Get "http://myweb:9200/_nodes/_local/name": dial tcp 150.158.54.252:9200: connect: connection refused'
     source 'datakit'
     status 'warning'
       time 2021-12-20 11:24:34 +0800 CST
-----------------[ r2.datakit.s1 ]-----------------
    __docid 'L_c6vvetpaahl15ivd7vn0'
   category 'input'
create_time 1639970679664
    date_ns 67000
       host 'demo'
    message 'postgresql pq: password authentication failed for user "postgres"'
     source 'datakit'
     status 'warning'
       time 2021-12-20 11:24:32 +0800 CST
-----------------[ r3.datakit.s1 ]-----------------
    __docid 'L_c6tish1aahlf03dqas00'
   category 'default'
create_time 1639657028706
    date_ns 246000
       host 'zhengs-MacBook-Pro.local'
    message 'datakit start ok, ready for collecting metrics.'
     source 'datakit'
     status 'info'
       time 2021-12-20 11:16:58 +0800 CST       
          
          ...       
```

**部分字段说明**
 - category: 类别，默认为`default`, 还可取值为`input`， 表明是与采集器 (`input`) 相关
 - status: 事件等级，可取值为 `info`, `warning`, `error`

## DataKit 更新 IP 数据库文件

可直接使用如下命令更新数据库文件（仅 Mac/Linux 支持）

```shell
sudo datakit --update-ip-db
```

若 DataKit 在运行中，更新成功后会自动更新 IP-DB 文件。

## DataKit 安装第三方软件

### Telegraf 集成

> 注意：建议在使用 Telegraf 之前，先确 DataKit 是否能满足期望的数据采集。如果 DataKit 已经支持，不建议用 Telegraf 来采集，这可能会导致数据冲突，从而造成使用上的困扰。

安装 Telegraf 集成

```shell
sudo datakit --install telegraf
```

启动 Telegraf

```shell
cd /etc/telegraf
sudo cp telegraf.conf.sample telegraf.conf
sudo telegraf --config telegraf.conf
```

关于 Telegraf 的使用事项，参见[这里](telegraf)。

### Security Checker 集成

安装 Security Checker

```shell
sudo datakit --install scheck
sudo datakit --install sec-checker  # 该命名即将废弃
```

安装成功后会自动运行，Security Checker 具体使用，参见[这里](https://www.yuque.com/dataflux/sec_checker/install) 

## 上传 DataKit 运行日志

排查 DataKit 问题时，通常需要检查 DataKit 运行日志，为了简化日志搜集过程，DataKit 支持一键上传日志文件：

```shell
sudo datakit --upload-log
log info: path/to/tkn_xxxxx/your-hostname/datakit-log-2021-11-08-1636340937.zip # 将这个路径信息发送给我们工程师即可
```

运行命令后，会将日志目录下的所有日志文件进行打包压缩，然后上传至指定的存储。我们的工程师会根据上传日志的主机名以及 Token 传找到对应文件，进而排查 DataKit 问题。

## 查看云属性数据

如果安装 DataKit 所在的机器是一台云服务器（目前支持 `aliyun/tencent/aws/hwcloud/azure` 这几种），可通过如下命令查看部分云属性数据，如（标记为 `-` 表示该字段无效）：

```shell
datakit --show-cloud-info aws

           cloud_provider: aws
              description: -
     instance_charge_type: -
              instance_id: i-09b37dc1xxxxxxxxx
            instance_name: -
    instance_network_type: -
          instance_status: -
            instance_type: t2.nano
               private_ip: 172.31.22.123
                   region: cn-northwest-1
        security_group_id: launch-wizard-1
                  zone_id: cnnw1-az2
```
