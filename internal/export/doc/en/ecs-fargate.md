Translation:

# Integrated with AWS ECS Fargate
---

## Prerequisites {#req}

- Task metadata endpoint version 4

## Overview {#intro}

Datakit can be integrated with AWS ECS Fargate environment with simple configuration to enable data collection.

The following data can be collected from ECS Fargate:

- Container metrics and object data, including container's basic information, CPU and memory usage, etc.
- APM data sent by other containers receiving the current task
- Enabled logstreaming to receive container log data

The Task Metadata Endpoint of ECS Fargate can only be used internally within the task definitions, so a Datakit container needs to be deployed in each task definition.

The only configuration required is to add an environment variable `ENV_ECS_FARGATE` with the value `"on"` to Datakit, and Datakit will automatically switch to this collection mode.

## Deployment and Configuration {#config}

In most cases, you just need to integrate Datakit into the task definition as a container and specify the required task roles. It can be divided into 3 steps as follows:

1. Create or modify an [IAM policy](https://docs.aws.amazon.com/zh_cn/IAM/latest/UserGuide/introduction.html){:target="_blank"}. Datakit requires at least the following 3 permissions:

- ecs:ListClusters to list available clusters.
- ecs:ListContainerInstances to list instances in a cluster.
- ecs:DescribeContainerInstances to describe instances to add metrics about running resources and tasks.

2. In the task definition, add a Datakit container with the following sample configuration:

- Name: `datakit`
- Image: `pubrepo.guance.com/datakit/datakit:<specified version>`
- Essential container: `No`
- Port mapping, container port: `9529 (configure as needed, defaults to 9529)`
- Resource allocation limit: CPU `2`vCPU, Memory limit `4`GB

3. Configure Datakit using environment variables. The necessary environment variables are as follows:

- "ENV_ECS_FARGATE": "on"
- "ENV_DATAWAY": "https://openway.guance.com?token=<your-token>"
- "ENV_HTTP_LISTEN": "0.0.0.0:9529"
- "ENV_DEFAULT_ENABLED_INPUTS": "dk,container,ddtrace"

This is an example of a running Datakit and trace task definition:

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
