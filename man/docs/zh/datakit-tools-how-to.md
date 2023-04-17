# 各种其它工具使用
---

DataKit 内置很多不同的小工具，便于大家日常使用。可通过如下命令来查看 DataKit 的命令行帮助：

```shell
datakit help
```

>注意：因不同平台的差异，具体帮助内容会有差别。

## 查看 DataKit 运行情况 {#using-monitor}

> 当前的 monitor 查看方式已经废弃（仍然可用，不久将废弃），新的 monitor 功能[参见这里](datakit-monitor.md)

在终端即可查看 DataKit 运行情况，其效果跟浏览器端 monitor 页面相似：

DataKit 新的 monitor 用法[参见这里](datakit-monitor.md)。

## 检查采集器配置是否正确 {#check-conf}

编辑完采集器的配置文件后，可能某些配置有误（如配置文件格式错误），通过如下命令可检查是否正确：

```shell
datakit tool --check-config
------------------------
checked 13 conf, all passing, cost 22.27455ms
```

### 采集一次 SNMP 配置 {#check-snmp}

编辑完 SNMP 采集器的配置文件后，可能某些配置有误（如配置文件格式错误），通过如下命令可以采集一次 SNMP 设备用于检查是否正确：

```shell
datakit tool --test-snmp /usr/local/datakit/conf.d/snmp/snmp.conf
# 以下会打印采集到的信息...
......
```

## 查看工作空间信息 {#workspace-info}

为便于大家在服务端查看工作空间信息，DataKit 提供如下命令查看：

```shell
datakit tool --workspace-info
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

## 查看 DataKit 相关事件 {#event}

DataKit 运行过程中，一些关键事件会以日志的形式进行上报，比如 DataKit 的启动、采集器的运行错误等。在命令行终端，可以通过 dql 进行查询。

```shell
datakit dql

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

## DataKit 更新 IP 数据库文件 {#install-ipdb}

=== "主机安装"

    - 可直接使用如下命令安装/更新 IP 地理信息库（此处可选择另一个 IP 地址库 `geolite2`，只需把 `iploc` 换成 `geolite2` 即可）：
    
    ```shell
    datakit install --ipdb iploc
    ```
    
    - 更新完 IP 地理信息库后，修改 datakit.conf 配置：
    
    ``` toml
    [pipeline]
      ipdb_type = "iploc"
    ```
    
    - 重启 DataKit 生效

    - 测试 IP 库是否生效

    ```shell
    $ datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```

    如果安装失败，其输出如下：
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province: 
       country: 
    ```

=== "Kubernetes(yaml)"

    - 修改 *datakit.yaml*，打开 4 处带 `---iploc-start ` 和 `---iploc-end` 中间注释。
    
    - 重新安装 DataKit：
    
    ```shell
    $ kubectl apply -f datakit.yaml
    
    # 确保 DataKit 容器启动
    $ kubectl get pod -n datakit
    ```

    - 进入容器，测试 IP 库是否生效

    ```shell
    $ datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```

    如果安装失败，其输出如下：
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province:
       country:
    ```
    
=== "Kubernetes(helm)"

    - helm 部署添加 `--set iploc.enable`
    
    ```shell
    $ helm install datakit datakit/datakit -n datakit \
    --set datakit.dataway_url="https://openway.guance.com?token=<YOUR-TOKEN>" \
    --set iploc.enable true \
    --create-namespace 
    ```
    
    关于 helm 的部署事项，参见[这里](datakit-daemonset-deploy.md/#__tabbed_1_2)。
    
    - 进入容器，测试 IP 库是否生效

    ```shell
    $ datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```

    如果安装失败，其输出如下：
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province:
       country:
    ```

## DataKit 安装第三方软件 {#extras}

### Telegraf 集成 {#telegraf}

> 注意：建议在使用 Telegraf 之前，先确 DataKit 是否能满足期望的数据采集。如果 DataKit 已经支持，不建议用 Telegraf 来采集，这可能会导致数据冲突，从而造成使用上的困扰。

安装 Telegraf 集成

```shell
datakit install --telegraf
```

启动 Telegraf

```shell
cd /etc/telegraf
cp telegraf.conf.sample telegraf.conf
telegraf --config telegraf.conf
```

关于 Telegraf 的使用事项，参见[这里](telegraf.md)。

### Security Checker 集成 {#scheck}

安装 Security Checker

```shell
datakit install --scheck
```

安装成功后会自动运行，Security Checker 具体使用，参见[这里](../scheck/scheck-install.md)

### DataKit eBPF 集成 {#ebpf}

安装 DataKit eBPF 采集器, 当前只支持 `linux/amd64 | linux/arm64` 平台，采集器使用说明见 [DataKit eBPF 采集器](ebpf.md)

```shell
datakit install --ebpf
```

如若提示 `open /usr/local/datakit/externals/datakit-ebpf: text file busy`，停止 DataKit 服务后再执行该命令

???+ warning

    该命令在 [:octicons-tag-24: Version-1.5.6](changelog.md#cl-1.5.6-brk) 已经被移除。新版本默认就内置了 eBPF 集成。


## 上传 DataKit 运行日志 {#upload-log}

> Deprecated: 请使用 [Bug-Report 功能](datakit-tools-how-to.md#bug-report)来代替。

排查 DataKit 问题时，通常需要检查 DataKit 运行日志，为了简化日志搜集过程，DataKit 支持一键上传日志文件：

```shell
datakit tool --upload-log
log info: path/to/tkn_xxxxx/your-hostname/datakit-log-2021-11-08-1636340937.zip # 将这个路径信息发送给我们工程师即可
```

运行命令后，会将日志目录下的所有日志文件进行打包压缩，然后上传至指定的存储。我们的工程师会根据上传日志的主机名以及 Token 传找到对应文件，进而排查 DataKit 问题。

## 收集 DataKit 运行信息 {#bug-report}

[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9) · [:octicons-beaker-24: Experimental](index.md#experimental)

在排查 DataKit 故障原因时，需要手动收集各种相关信息（如日志、配置文件和监控数据等），这通常比较繁琐。为了简化这个过程，DataKit 提供了一个命令，可以一次性获取所有相关信息并将其打包到一个文件中。使用方式如下：


```shell
datakit tool --bug-report
```

执行成功后，在当前目录下生成一个 zip 文件，命名格式为 `info-<时间戳毫秒数>.zip`。

解压后的文件列表参考如下：

```shell

├── config
│   ├── container
│   │   └── container.conf
│   ├── datakit.conf
│   ├── db
│   │   ├── kafka.conf
│   │   ├── mysql.conf
│   │   └── sqlserver.conf
│   ├── host
│   │   ├── cpu.conf
│   │   ├── disk.conf
│   │   └── system.conf
│   ├── network
│   │   └── dialtesting.conf
│   ├── profile
│   │   └── profile.conf
│   ├── pythond
│   │   └── pythond.conf
│   └── rum
│       └── rum.conf
├── env.txt
├── metrics 
│   ├── metric-1680513455403 
│   ├── metric-1680513460410
│   └── metric-1680513465416 
├── log
│   ├── gin.log
│   └── log
└── profile
    ├── allocs
    ├── heap
    └── profile

```

文件说明

| 文件名称  | 是否目录 | 说明                                                          |
| ---:      | ---:     | ---:                                                          |
| `config`  | 是       | 配置文件，包括主配置和已开启的采集器配置                      |
| `env.txt` | 否       | 运行环境的环境变量信息                                        |
| `log`     | 是       | 最新的日志文件，包括 log 和 gin log，暂不支持 `stdout`        |
| `profile` | 是       | pprof 开启时，会采集 profile 数据                             |
| `metrics` | 是       | `/metrics` 接口返回的数据，命名格式为 `metric-<时间戳毫秒数>` |

**敏感信息处理**

信息收集时，敏感信息（如token、密码等）会被自动过滤替换，具体规则如下：

- 环境变量

只获取以 `ENV_` 开头的环境变量，且对环境变量名称中包含 `password`, `token`, `key`, `key_pw`, `secret` 的环境变量进行脱敏处理，替换为 `******`

- 配置文件

配置文件内容进行正则替换处理：

如：

```
https://openway.guance.com?token=tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` => `https://openway.guance.com?token=******
pass = "1111111"` => `pass = "******"
postgres://postgres:123456@localhost/test` => `postgres://postgres:******@localhost/test
```

经过上述处理，能够去除绝大部分敏感信息。尽管如此，如果导出的文件还存在敏感信息，可以手动将敏感信息移除，请务必确认。

## 查看云属性数据 {#cloudinfo}

如果安装 DataKit 所在的机器是一台云服务器（目前支持 `aliyun/tencent/aws/hwcloud/azure` 这几种），可通过如下命令查看部分云属性数据，如（标记为 `-` 表示该字段无效）：

```shell
datakit tool --show-cloud-info aws

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

## 解析行协议数据 {#parse-lp}

[:octicons-tag-24: Version-1.5.6](changelog.md#cl-1.5.6)

通过如下命令可解析行协议数据：

```shell
datakit tool --parse-lp /path/to/file
Parse 201 points OK, with 2 measurements and 201 time series
```

可以以 JSON 形式输出：

```shell
datakit tool --parse-lp /path/to/file --json
{
  "measurements": {  # 指标集列表
    "testing": {
      "points": 7,
      "time_series": 6
    },
    "testing_module": {
      "points": 195,
      "time_series": 195
    }
  },
  "point": 202,        # 总点数
  "time_serial": 201   # 总时间线数
}
```

## DataKit 自动命令补全 {#completion}

> DataKit 1.2.12 才支持该补全，且只测试了 Ubuntu 和 CentOS 两个 Linux 发行版。其它 Windows 跟 Mac 均不支持。

在使用 DataKit 命令行的过程中，因为命令行参数很多，此处我们添加了命令提示和补全功能。

主流的 Linux 基本都有命令补全支持，以 Ubuntu 和 CentOS 为例，如果要使用命令补全功能，可额外安装如下软件包：

- Ubuntu：`apt install bash-completion`
- CentOS: `yum install bash-completion bash-completion-extras`

如果安装 DataKit 之前，这些软件已经安装好了，则 DataKit 安装时会自动带上命令补全功能。如果这些软件包是在 DataKit 安装之后才更新的，可执行如下操作来安装 DataKit 命令补全功能：

```shell
datakit tool --setup-completer-script
```

补全使用示例：

```shell
$ datakit <tab> # 输入 \tab 即可提示如下命令
dql       help      install   monitor   pipeline  run       service   tool

$ datakit dql <tab> # 输入 \tab 即可提示如下选项
--auto-json   --csv         -F,--force    --host        -J,--json     --log         -R,--run      -T,--token    -V,--verbose
```

以下提及的所有命令，均可使用这一方式来操作。

### 获取自动补全脚本 {#get-completion}

如果大家的 Linux 系统不是 Ubuntu 和 CentOS，可通过如下命令获取补全脚本，然后再按照对应平台的 shell 补全方式，一一添加即可。

```shell
# 导出补全脚本到本地 datakit-completer.sh 文件中
datakit tool --completer-script > datakit-completer.sh
```
