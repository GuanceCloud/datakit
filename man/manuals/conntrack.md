# Conntrack
---

## 视图预览
Conntrack 性能指标展示，包括成功搜索条目数，插入的包数，连接数量等
![](imgs/input-conntrack-1.png) 

## 版本支持
操作系统支持：Linux 
## 前置条件

- 服务器 <[安装 Datakit](/datakit/datakit-install/)>
## 安装配置
说明：示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)
### 部署实施
#### 指标采集 (默认)

1. System 采集器会同时采集 system 和 conntrack 指标
2. Conntrack 数据采集默认开启，对应配置文件 /usr/local/datakit/conf.d/host/system.conf

参数说明

- interval：数据采集频率
```
[[inputs.system]]
  interval = '10s'
```

3. Conntrack 指标采集验证  /usr/local/datakit/datakit -M |egrep "最近采集|system"
![](imgs/input-conntrack-2.png) 

4. 指标预览
![](imgs/input-conntrack-3.png) 

#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 system 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag/)>
```
# 示例
[inputs.system.tags]
   app = "oa"
```
重启 Datakit
```
systemctl restart datakit
```
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Conntrack>
## 异常检测
<监控 - 模板新建 - 主机检测库>
## 指标详解
| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| `entries` | 当前连接数量 | int | count |
| `entries_limit` | 连接跟踪表的大小 | int | count |
| `stat_drop` | 跟踪失败被丢弃的包数目 | int | count |
| `stat_early_drop` | 由于跟踪表满而导致部分已跟踪包条目被丢弃的数目 | int | count |
| `stat_found` | 成功的搜索条目数目 | int | count |
| `stat_ignore` | 已经被跟踪的报数目 | int | count |
| `stat_insert` | 插入的包数目 | int | count |
| `stat_insert_failed` | 插入失败的包数目 | int | count |
| `stat_invalid` | 不能被跟踪的包数目 | int | count |
| `stat_search_restart` | 由于hash表大小修改而导致跟踪表查询重启的数目 | int | count |

## 常见问题排查
<[无数据上报排查](/datakit/why-no-data/)>

## 进一步阅读
<[主机可观测最佳实践](/best-practices/integrations/host/)>