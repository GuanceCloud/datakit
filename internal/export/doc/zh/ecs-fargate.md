
# 集成到 AWS ECS Fargate
---

## 前置条件 {#req}

- 任务元数据终端节点版本 4

## 简述 {#intro}

Datakit 可以集成到 AWS ECS Fargate 环境中，只需要简单的配置就能开启采集。

从 ECS Fargate 能采集到的数据有：

- 容器指标、对象数据，包括容器的基础信息、CPU 和内存使用率等
- 接收当前 Task 的其他容器发送的 APM 数据
- 开启 logstreaming 接收容器日志数据

ECS Fargate 的任务元数据端点（Task metadata Endpoint）只能在任务定义（Task definitions）内部使用，所以需要在每个任务定义中都部署一个 Datakit 容器。

启用的唯一配置是给 Datakit 添加一个环境变量 `ENV_ECS_FARGATE` 为 `"on"`，Datakit 会自动切换到此采集模式。

## 部署和配置 {#config}

通常情况下，只需要将 Datakit 以容器方式集成到任务定义中，且在任务定义指定需要任务角色。可以分为 3 步，如下：

1. 创建或修改 [IAM 策略](https://docs.aws.amazon.com/zh_cn/IAM/latest/UserGuide/introduction.html){:target="_blank"}，Datakit 至少需要以下 3 种权限：

- ecs:ListClusters 列出可用的集群。
- ecs:ListContainerInstances 列出集群的实例。
- ecs:DescribeContainerInstances 描述实例以添加有关正在运行的资源和任务的指标。

2. 在任务定义中，添加 Datakit 容器，示例配置项如下：

- 名称：`datakit`
- 镜像：`pubrepo.guance.com/datakit/datakit:<指定版本>`
- 主要容器：`"否"`
- 端口映射，容器端口：`9529（按需配置，默认是 9529）`
- 资源分配限制：CPU `2`vCPU，内存限制 `4`GB

3. 使用环境变量配置 Datakit，必要的环境变量如下：

- "ENV_ECS_FARGATE": "on"
- "ENV_DATAWAY": "https://openway.guance.com?token=<your-token>"
- "ENV_HTTP_LISTEN": "0.0.0.0:9529"
- "ENV_DEFAULT_ENABLED_INPUTS": "dk,container,ddtrace"

这是一份运行的 Datakit 和 trace 的任务定义示例：

```json
{
    "family": "datakit-dev",
    "containerDefinitions": [
        {
            "name": "trace",
            "image": "pubrepo.guance.com/datakit-dev/ddtrace-golang-demo:v1",
            "cpu": 0,
            "portMappings": [
                {
                    "name": "ddtrace-80-tcp",
                    "containerPort": 80,
                    "hostPort": 80,
                    "protocol": "tcp",
                    "appProtocol": "http"
                }
            ],
            "essential": true,
            "environment": [
                {
                    "name": "DD_TRACE_AGENT_PORT",
                    "value": "9529"
                }
            ],
            "mountPoints": [],
            "volumesFrom": [],
            "readonlyRootFilesystem": false
        },
        {
            "name": "dk",
            "image": "pubrepo.guance.com/datakit/datakit:1.21.0",
            "cpu": 2048,
            "memory": 4096,
            "portMappings": [
                {
                    "name": "dk-9529-tcp",
                    "containerPort": 9529,
                    "hostPort": 9529,
                    "protocol": "tcp"
                }
            ],
            "essential": false,
            "environment": [
                {
                    "name": "ENV_ECS_FARGATE",
                    "value": "on"
                },
                {
                    "name": "ENV_DATAWAY",
                    "value": "https://openway.guance.com?token=<your-token>"
                },
                {
                    "name": "ENV_HTTP_LISTEN",
                    "value": "0.0.0.0:9529"
                },
                {
                    "name": "ENV_DEFAULT_ENABLED_INPUTS",
                    "value": "dk,container,ddtrace"
                }
            ],
            "mountPoints": [],
            "volumesFrom": []
        }
    ],
    "taskRoleArn": "arn:aws-cn:iam::123123123:role/datakit-dev",
    "executionRoleArn": "arn:aws-cn:iam::123123123:role/datakit-dev",
    "networkMode": "awsvpc",
    "requiresCompatibilities": [
        "FARGATE"
    ],
    "cpu": "4096",
    "memory": "8192",
    "runtimePlatform": {
        "cpuArchitecture": "ARM64",
        "operatingSystemFamily": "LINUX"
    }
}
```
