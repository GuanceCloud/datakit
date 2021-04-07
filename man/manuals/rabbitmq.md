- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# 简介

`rabbitmq` 采集器 是通过 插件`rabbitmq-management` 采集数据监控 `rabbitmq` ,它能够：
- 展示 `rabbitmq` 基础数据，比如 连接数 ，队列数，消息总数等
- 跟踪`queue`信息,比如队列大小，消费者计数等
- 跟踪`node`信息，比如使用的 `socket` `mem` 等
- 跟踪 `exchange`信息 ，比如 `message_publish_count` 等`


## 前置条件
- 安装 `rabbitmq` 以 `Ubuntu` 为例
    ```
    sudo apt-get update
    sudo apt-get install rabbitmq-server
    sudo service rabbitmq-server start
    ```
      
- 开启 `REST API plug-ins` 

    ```
    sudo rabbitmq-plugins enable rabbitmq-management
    ```
      
- 创建 user，比如：

    ```
    rabbitmqctl add_user dataflux <SECRET>
    rabbitmqctl set_permissions  -p / dataflux "^aliveness-test$" "^amq\.default$" ".*"
    rabbitmqctl set_user_tags dataflux monitoring
    ```


## 配置

进入 DataKit 安装目录下的 `conf.d/{{.InputName}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
