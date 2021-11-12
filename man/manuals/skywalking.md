{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

## 下载 SkyWalking

注意：APM v8.8.3 目前存在不兼容问题无法使用。开发测试版本有 APM v8.5.0 v8.6.0 v8.7.0

- skywlking-java-apm [官方下载](https://skywalking.apache.org/downloads/)

## 在 Java 应用中添加 skywalking 支持

```shell
java -javaagent:/path/to/skywalking/agent -jar /path/to/your/service.jar
```

## 配置 SkyWalking

打开文件 /path_to_skywalking_agent/config/agent.config 进行配置

```conf
# The service name in UI
agent.service_name=${SW_AGENT_NAME:your-service-name}
# Backend service addresses.
collector.backend_service=${SW_AGENT_COLLECTOR_BACKEND_SERVICES:<datakit-ip:skywalking-agent-port>}
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```
