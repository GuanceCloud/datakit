{{.CSS}}
# Memcached
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

Memcached 采集器可以从 Memcached 实例中采集实例运行状态指标，并将指标采集到观测云，帮助监控分析 Memcached 各种异常情况

## 前置条件

- Memcached 版本 >= 1.5.0

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 视图预览
Memcached 性能指标展示：包括连接数，命令数，网络流量，线程数，命中率信息等

![image](imgs/input-memcached-1.png)

## 版本支持

操作系统支持：Linux / Windows / Mac

## 前置条件

- 服务器 <[安装 Datakit](../datakit/datakit-install.md)>

## 安装配置

说明：示例 Memcached 版本为 Linux 环境 memcached 1.4.15，Windows 版本请修改对应的配置文件

### 部署实施

#### 指标采集 (必选)

1、开启 Datakit Memcache 插件，复制 sample 文件

```
cd /usr/local/datakit/conf.d/db
cp memcached.conf.sample memcached.conf
```

2、修改 memcached 配置文件

```
vi memcached.conf
```

参数说明

- servers：服务连接地址
- unix_sockets：socket 文件路径
- interval：数据采集频率

```
[[inputs.memcached]]
  servers = ["localhost:11211"]
  # unix_sockets = ["/var/run/memcached.sock"]
  interval = '10s'
```

3、Memcached 指标采集验证  /usr/local/datakit/datakit -M |egrep "最近采集|memcached"

![image](imgs/input-memcached-2.png)

指标预览

![image](imgs/input-memcached-3.png)

#### 插件标签 (非必选)

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 memcached 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](../best-practices/guance-skill/tag.md)>

```
# 示例
[inputs.memcached.tags]
   app = "oa"
```

重启 Datakit

```
systemctl restart datakit
```

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Memcached 监控视图>

## 异常检测

<监控 - 模板新建 - Memcached 检测库>

## 指标详解

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 

## 常见问题排查

<[[无数据上报排查](why-no-data.md)>

## 进一步阅读

<[Memcached 简介](https://www.jianshu.com/p/bf648b4e60ad)>
