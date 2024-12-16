# Lambda Extension 本地测试

为便于本地测试 Lambda extension 工作是否基本正常，我们可以在本地模拟 lambda 运行环境。

## 准备 aws lambda 本地服务

extension 通过 Lambda 提供的一组 api/env 来实现初始化，这些 api 包括：

- GET `/2020-01-01/extension/event/next`
- PUT `/2022-07-01/telemetry`
- POST `/2020-01-01/extension/register`

在 *scripts/v1-write.go* 中 mock 了这几个接口实现，可直接启动这个服务：

```shell
# 此处 -decode 用来解码上传的数据点，以确保上传的数据符合预期
$ go run scripts/v1-write.go -decode
```

启动完之后，本地的 localhost:54321 便充当了两个角色：

- Mock Lambda 环境的必要 API
- Mock Dataway 接收 Datakit 采集的数据

## 运行 lambda extension

编译完 Datakit 后，即可手动启动 Lambda extension:

```shell
os=$(__os)
arch=$(__arch)

AWS_LAMBDA_RUNTIME_API="localhost:54321" \
AWS_LAMBDA_FUNCTION_NAME=dk \
AWS_LAMBDA_FUNCTION_VERSION=0.1.9999 \
AWS_REGION="dk-debug" \
AWS_LAMBDA_FUNCTION_MEMORY_SIZE="10485760" \
EnvLambdaInitializationType="dk-lambda-ext-testing" \
ENV_DATAWAY=http://localhost:54321?token=tkn_xxxxxxxxxxxxxxxxxxxxxxx \
ENV_LOG_LEVEL="debug" \
ENV_DISABLE_LOG_COLOR=yes \
AWS_SAM_LOCAL=true \
DK_DEBUG_WORKDIR=~/datakit ./dist/datakit_aws_lambda-$os-$arch/extensions/datakit $@
```

此处，几个 `AWS_XXX` 环境变量设置是必须的，另外，`ENV_DATAWAY` 也必须设置，作为 Lambda extension，DK 是以容器方式（相当于 `-docker`）运行的，所以不会读取 *datakit.conf* 中的内容，其它几个 `ENV_XXX` 设置也是基于这个原因。

## 模拟 Lambda 日志数据

Lambda extension 启动完之后，它在 9529 上会注册一个 `/awslambda` API，我们可以直接往它推送数据：

```shell
curl http://localhost:9529/awslambda -XPOST --data-binary @lambda.log.json
```

其中示例数据如下：

```json
[
  {
    "time": "2022-10-12T00:03:50.000Z",
    "type": "function",
    "record": "[INFO] Hello world, I am a function!"
  },
  {
    "time": "2022-10-12T00:03:50.000Z",
    "type": "function",
    "record": {
      "timestamp": "2022-10-12T00:03:50.000Z",
      "level": "INFO",
      "requestId": "79b4f56e-95b1-4643-9700-2807f4e68189",
      "message": "Hello world, I am a function!"
    }
  },
  {
    "time": "2022-10-12T00:03:50.000Z",
    "type": "platform.runtimeDone",
    "record": {
      "timestamp": "2022-10-12T00:03:50.000Z",
      "level": "INFO",
      "requestId": "79b4f56e-95b1-4643-9700-2807f4e68189",
      "message": "Hello world, I am a function!"
    }
  }
]
```

此处 `type:platform.runtimeDone` 是必要的一个 event 类型，它会促使 Lambda extension 立即将前面两条日志发送给 Dataway。
