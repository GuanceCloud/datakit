{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

Datakit 内嵌的 SkyWalking Agent 用于接收，运算，分析 Skywalking Tracing 协议数据。

## SkyWalking 文档

> APM v8.8.3 目前存在不兼容问题无法使用。目前已支持 v8.5.0 v8.6.0 v8.7.0

- [Quickstart](https://skywalking.apache.org/docs/skywalking-showcase/latest/readme/)
- [Docs](https://skywalking.apache.org/docs/)
- [Clients Download](https://skywalking.apache.org/downloads/)
- [Souce Code](https://github.com/apache/skywalking)

## 配置 SkyWalking Client

打开文件 /path_to_skywalking_agent/config/agent.config 进行配置

```conf
# The service name in UI
agent.service_name=${SW_AGENT_NAME:your-service-name}
# Backend service addresses.
collector.backend_service=${SW_AGENT_COLLECTOR_BACKEND_SERVICES:<datakit-ip:skywalking-agent-port>}
```

## 配置 SkyWaking Agent

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

## 启动 Java Client

```command
java -javaagent:/path/to/skywalking/agent -jar /path/to/your/service.jar
```
