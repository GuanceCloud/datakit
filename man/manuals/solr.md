{{.CSS}}
# Solr
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

solr 采集器，用于采集 solr cache 和 request times 等的统计信息。

## 前置条件

DataKit 使用 Solr Metrics API 采集指标数据，支持 Solr 7.0 及以上版本。可用于 Solr 6.6，但指标数据不完整。

## 安装部署

#### 指标采集 (必选)

1. 开启 Datakit Solr 插件，复制 sample 文件

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

2. 修改 `solr.conf` 配置文件
```bash
vi solr.conf
```
参数说明

- interval：采集指标频率
- servers：solr server地址
- username：用户名
- password：密码

```yaml
[[inputs.solr]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  ## specify a list of one or more Solr servers
  servers = ["http://localhost:8983"]

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

```

3. 重启 Datakit (如果需要开启日志，请配置日志采集再重启)
```bash
systemctl restart datakit
```

4. Solr 指标采集验证 `/usr/local/datakit/datakit -M |egrep "最近采集|solr"`

![image.png](../imgs/solr-1.png)

5. 指标预览

![image.png](../imgs/solr-2.png)


### 日志采集

如需采集 Solr 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 Solr 日志文件的绝对路径。比如：

```toml
[inputs.solr.log]
    # 填入绝对路径
    files = ["/path/to/demo.log"] 
```


参数说明

- files：日志文件路径 (通常填写访问日志和错误日志)
- pipeline：日志切割文件(内置)，实际文件路径 /usr/local/datakit/pipeline/solr.p

切割日志示例：

```
2013-10-01 12:33:08.319 INFO (org.apache.solr.core.SolrCore) [collection1] webapp.reporter
```

切割后字段：

| 字段名   | 字段值                        |
| -------- | ----------------------------- |
| Reporter | webapp.reporter               |
| status   | INFO                          |
| thread   | org.apache.solr.core.SolrCore |
| time     | 1380630788319000000           |


日志预览

![image.png](../imgs/solr-3.png)

## 场景视图
场景 - 新建场景 - Solr 监控场景
## 异常检测
异常检测库 - 新建检测库 - Solr 检测库

## 故障排查
- [无数据上报排查](why-no-data.md)

## 进一步阅读
[DataFlux pipeline 文本数据处理](/datakit/pipeline.md)