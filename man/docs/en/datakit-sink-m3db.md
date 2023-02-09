# M3DB
---

M3DB is a distributed time series database open source by Uber, which is mainly used to store Metric type data and has been used in Uber for many years.

Datakit supports writing the collected metric data into M3db, which can be configured to the specified database in the form of configuration file and environment variable.

For more information and documentation on M3DB, please refer to:

- [m3db-github-source](https://github.com/m3db/m3){:target="_blank"}
- [m3db-official document](https://m3db.io/docs){:target="_blank"}

## Install the Stand-alone M3db {#install-storage}

``` shell 
# Download and start
wget https://s3-gz01.didistatic.com/n9e-pub/tarball/m3dbnode-single-v0.0.1.tar.gz
tar zxvf m3dbnode-single-v0.0.1.tar.gz
cd m3dbnode-single 
./scripts/install.sh #install.sh is a self-written script. It is recommended to check the steps for yourself
systemctl enable m3dbnode

# Initialization
curl -X POST http://localhost:7201/api/v1/database/create -d '{
  "type": "local",
  "namespaceName": "default",
  "retentionTime": "48h"
}'


# View status
systemctl status m3dbnode

# Or
ss -tlnp|grep m3dbnode
```

## Start Sink-m3db on Datakit {#enable}

### Specify M3DB Through Configuration File {#dk-config}

1. Modify the datakit configuration file

``` shell 
vim /usr/local/datakit/conf/datakit.conf
```

2. Modify the sink configuration. Note that if sink correlation has never been configured, you can add a configuration item

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "http"
    host = "localhost:7201"
    path = "/api/v1/prom/remote/write"
    categories = ["M"] # M3DB 目前只支持时序时序（metric）
    target = "m3db"
```

3. Restart datakit

``` shell
datakit --restart
```

## Specify M3db Settings During Installation {#install}

```shell
DK_SINK_M="m3db://localhost:7201?scheme=http" \
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

Datakit installed through environment variables automatically generates the corresponding configuration in the configuration file.

## M3db Visualization {#view}

It is recommended that you use [prometheus](https://prometheus.io/download/){:target="_blank"} and [grafana](https://grafana.com/){:target="_blank"} to query and display data.
