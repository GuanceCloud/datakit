{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

Prom 采集器可以获取各种 Prometheus Exporters 暴露出来的指标数据，只要配置相应的 Exporter 地址，就可以将指标数据接入。

## 前置条件

只能接入 Prometheus 形式的指标数据。

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 协议转换说明

由于 Prometheus 的数据格式跟 Influxdb 的行协议格式存在一定的差别。 对 Prometheus 而言，以下为一个 K8s 集群中一段分暴露出来的数据：

```
node_filesystem_avail_bytes{device="/dev/disk1s1",fstype="apfs",mountpoint="/"} 1.21585664e+08
node_filesystem_avail_bytes{device="/dev/disk1s4",fstype="apfs",mountpoint="/private/var/vm"} 1.2623872e+08
node_filesystem_avail_bytes{device="/dev/disk3s1",fstype="apfs",mountpoint="/Volumes/PostgreSQL 13.2-2"} 3.7269504e+07
node_filesystem_avail_bytes{device="/dev/disk5s1",fstype="apfs",mountpoint="/Volumes/Git 2.15.0 Mavericks Intel Universal"} 1.2808192e+07
node_filesystem_avail_bytes{device="map -hosts",fstype="autofs",mountpoint="/net"} 0
node_filesystem_avail_bytes{device="map auto_home",fstype="autofs",mountpoint="/home"} 0
# HELP node_filesystem_device_error Whether an error occurred while getting statistics for the given device.
# TYPE node_filesystem_device_error gauge
node_filesystem_device_error{device="/dev/disk1s1",fstype="apfs",mountpoint="/"} 0
node_filesystem_device_error{device="/dev/disk1s4",fstype="apfs",mountpoint="/private/var/vm"} 0
node_filesystem_device_error{device="/dev/disk3s1",fstype="apfs",mountpoint="/Volumes/PostgreSQL 13.2-2"} 0
node_filesystem_device_error{device="/dev/disk5s1",fstype="apfs",mountpoint="/Volumes/Git 2.15.0 Mavericks Intel Universal"} 0
node_filesystem_device_error{device="map -hosts",fstype="autofs",mountpoint="/net"} 0
node_filesystem_device_error{device="map auto_home",fstype="autofs",mountpoint="/home"} 0
# HELP node_filesystem_files Filesystem total file nodes.
# TYPE node_filesystem_files gauge
node_filesystem_files{device="/dev/disk1s1",fstype="apfs",mountpoint="/"} 9.223372036854776e+18
node_filesystem_files{device="/dev/disk1s4",fstype="apfs",mountpoint="/private/var/vm"} 9.223372036854776e+18
node_filesystem_files{device="/dev/disk3s1",fstype="apfs",mountpoint="/Volumes/PostgreSQL 13.2-2"} 9.223372036854776e+18
node_filesystem_files{device="/dev/disk5s1",fstype="apfs",mountpoint="/Volumes/Git 2.15.0 Mavericks Intel Universal"} 9.223372036854776e+18
node_filesystem_files{device="map -hosts",fstype="autofs",mountpoint="/net"} 0
node_filesystem_files{device="map auto_home",fstype="autof
```

对 Influxdb 而言，上面数据的一种组织方式为

```
node_filesystem,tag-list available_bytes=1.21585664e+08,device_error=0,files=9.223372036854776e+18 time
```

其组织依据是：

- 在 Prometheus 暴露出来的指标中，如果名称前缀都是 `node_filesystem`，那么就将其规约到行协议指标集 `node_filesystem` 上
- 将切割掉前缀的原 Prometheus 指标，都放到指标集 `node_filesystem` 的指标中
- 默认情况下，Prometheus 中的所有 tags（即 `{}` 中的部分）在 Influxdb 的行协议中，都保留下来

要达到这样的切割目的，可以这样配置 `prom.conf`

```
  [[inputs.prom.measurements]]
    prefix = "node_filesystem_"
    name = "node_filesystem"
```

## 命令行调试指标集 

由于 Prometheus 暴露出来的指标非常多，大家不一定需要所有的指标，故 DataKit 提供一个简单的调试 `prom.conf` 的工具，如果不断调整 `prom.conf` 的配置，已达到如下几个目的：

- 只采集符合一定名称规则的 Prometheus 指标
- 只采集部分计量数据（`metric_types`），如 `gauge` 类指标和 `counter` 类指标

Datakit 支持命令行直接调试 prom 采集器的配置文件，从 conf.d/{{.Catalog}} 拷贝出一份 prom.conf 模板，填写对应 Exporter 地址，即可通过 DataKit 调试这个 `prom.conf`：

执行如下命令，即可调试 `prom.conf`

```
datakit --prom-conf prom.conf
```

参数说明：

- `prom-conf`: 指定配置文件，默认在当前目录下寻找 `prom.conf` 文件，如果未找到，会去 `<datakit-install-dir>/conf.d/{{.Catalog}}` 目录下查找相应文件。

输出示例：

```
================= Line Protocol Points ==================

 prom_node,device=disk0 disk_written_sectors_total=146531.087890625 1623379432917573000
 prom_node,device=disk2 disk_written_sectors_total=0 1623379432917573000
 prom_node,device=disk4 disk_written_sectors_total=0 1623379432917573000
 prom_node memory_total_bytes=8589934592 1623379432917573000
 prom_node,device=XHC20 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=awdl0 network_transmit_bytes_total=1527808 1623379432917573000
 prom_node,device=bridge0 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=en0 network_transmit_bytes_total=2847181824 1623379432917573000
 prom_node,device=en1 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=en2 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=gif0 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=lo0 network_transmit_bytes_total=6818923520 1623379432917573000
 prom_node,device=p2p0 network_transmit_bytes_total=0 1623379432917573000
 ....
================= Summary ==================

Total time series: 58
Total line protocol points: 261
Total measurements: 3 (prom_node, prom_go, prom_promhttp)
```

输出说明：

- Line Protocol Points： 产生的行协议点
- Summary： 汇总结果
  - Total time series: 时间线数量
  - Total line protocol points: 行协议点数
  - Total measurements: 指标集个数及其名称。
