{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

## 下载 skywalking 客户端（目前支持版本：8.3.0）

- skywalking-java agent [DataFlux 下载](https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/plugins/skywalking-agent-8.3.0.tar.gz)（推荐）
- skywlking-java-apm [官方下载](https://archive.apache.org/dist/skywalking/8.3.0/apache-skywalking-apm-8.3.0.tar.gz)

## 在 Java 应用中添加 skywalking 支持

```shell
java -javaagent:/path/to/skywalking/agent -jar /path/to/your/service.jar
```

## 配置 skywalking

打开文件 /path/to/skywalking/agent/config/agent.config 进行配置

```conf
# The service name in UI
agent.service_name=${SW_AGENT_NAME:your-service-name}
# Backend service addresses.
collector.backend_service=${SW_AGENT_COLLECTOR_BACKEND_SERVICES:<datakit-ip:skywalking-port>}
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```
