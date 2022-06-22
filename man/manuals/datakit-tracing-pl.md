{{.CSS}}
# Apply Pipeline Onto Datakit Tracing

- DataKit 版本：{{.Version}}
- 操作系统支持：全平台

## Prerequisit

**Span Data Structure**

Datakit 从第三方 Tracing Agent 收集到链路数据后会将所有数据统一到 Datakit Span 这一数据结构中去。所以请先仔细阅读有关[Datakit Tracing Structure](datakit-tracing-struct)的文档。

**Pipeline Data Source**

目前 Datakit Tracing Pipeline 的数据源是第三方 Tracing Agent 的原始数据，即 DatakitSpan.Content 字段，即 Line Protocol 中的 msg 字段。

**Pipeline How To**

运行在 Datakit 中的 Pipeline 脚本功能只是 Pipeline 脚本功能的一个子集。当前 Datakit 版本实现的功能仅限于数据 value 的修改和添加新的数据项。重复添加已有数据 key 可能造成系统功能错误，不正确的添加数据 key 可能造成[Datakit Tracing Backend](datakit-tracing#datakit-tracing-backend)的业务逻辑错误。

## Config Pipeline In Datakit

**Step 1:**

在 Datakit 安装目录下可以的 conf.d 目录的各个 Tracer Agent，例如：ddtrace ，下找到并修改相应的配置文件来启动 Pipeline 功能，注意仔细阅读配置文件中的文档注释。

```toml
  ## Piplines use to manipulate message and meta data. If this item configured right then
  ## the current input procedure will run the scripts wrote in pipline config file against the data
  ## present in span message.
  ## The string on the left side of the equal sign must be identical to the service name that
  ## you try to handle.
  [inputs.ddtrace.pipelines]
    service1 = "service1.p"
```

解注释相应配置条目。其中 service1 为当前希望操作的链路的服务名且必须与服务名保持一致，service1.p 为 Pipeline 脚本文件名（不是文件路径）将该文件存放在 Datakit 安装目录下的 pipeline 目录下即可。

**Step 2:**

编写 Pipeline 脚本文件，Pipeline 脚本使用详情参考[Pipeline](pipeline)和文档。

**Step 3:**

重启 Datakit 实例后去观测云平台查看运行结果。

## Pipeline Example

假如当前正在使用 Datakit DDTrace Agent 检测一个 x_service，x_service 服务下有一个 /v1/x/get_service_name 资源（resource），此 resource 下调用了 operation service_name。

Pipeline 文件如下：

**x_service.p**

```pipeline
add_key(new_customer_x, "Lucy")
set_tag(operation, "v1_service_name")
```

将此文件复制到 Datakit 安装目录下的 pipeline 子目录下。
重启 Datakit 实例后到观测云平台查看修改后的数据。
