# 将指标型数据存入 m3db 数据库中

## m3db 数据库

M3DB 是 Uber 开源的一款分布式时序数据库，主要用来存储 Metric 类型数据，已在 Uber 内部使用多年。

Datakit 支持将采集到的指标性数据写入 M3db 中，可以通过配置文件和环境变量两种形式配置到指定的数据库中。

M3DB更多介绍和文档 请参考：

- [m3db-github-源码](https://github.com/m3db/m3)
- [m3db-官方文档](https://m3db.io/docs)

### 快速上手 安装单机版 m3db

``` shell 
# 下载并启动
wget https://s3-gz01.didistatic.com/n9e-pub/tarball/m3dbnode-single-v0.0.1.tar.gz
tar zxvf m3dbnode-single-v0.0.1.tar.gz
cd m3dbnode-single 
./scripts/install.sh #install.sh为自行编写的脚本，建议自己查看一下步骤
systemctl enable m3dbnode

# 初始化
curl -X POST http://localhost:7201/api/v1/database/create -d '{
  "type": "local",
  "namespaceName": "default",
  "retentionTime": "48h"
}'


# 查看状态
systemctl status m3dbnode

# 或者
ss -tlnp|grep m3dbnode
```

## 在 datakit 上开启 sink-m3db

### 通过配置文件指定 M3DB

1. 修改配置 datakit 配置文件

``` shell 
vim /usr/local/datakit/conf/datakit.conf
```

2. 修改 sink 配置，注意 如果从没有配置过 sink 相关，新增一个配置项即可

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "http"
    host = "localhost:7201"
    path = "/api/v1/prom/remote/write"
    categories = ["M"] # M3DB 目前只支持时序时序（metric）
    target = "m3db"
```

3. 重启 datakit

``` shell
datakit --restart
```

### 安装阶段指定 M3DB 设置

```shell
DK_SINK_M="m3db://localhost:7201?scheme=http" \
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

通过环境变量安装的 Datakit，会在自动在配置文件中生成相应的配置。

## M3DB 可视化

这里推荐您使用 [prometheus](https://prometheus.io/download/) 和 [grafana](https://grafana.com/) 去查询和展示数据。
