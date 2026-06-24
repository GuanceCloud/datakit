---
title     : 'AWS Lambda 扩展'
summary   : '通过 AWS Lambda 扩展采集数据'
tags:
  - 'AWS'
__int_icon      : 'icon/awslambda'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

[:octicons-tag-24: Version-1.34.0](../datakit/changelog.md#cl-1.34.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

AWS Lambda 采集器是通过 AWS Lambda Extension 的方式采集 AWS Lambda 的指标与日志。

## 安装 {#installation}

> **DataKit 1.x 安装说明**：本节扩展包下载链接默认指向 DataKit `2.x` 的 `/datakit-v2/` 路径。如需使用 DataKit `1.x` 对应的扩展包，请将 URL 中的 `/datakit-v2/` 替换为 `/datakit/`。URL 中的静态资源域名会按当前文档品牌渲染。

### 添加 DataKit 层 {#layer}

- [通过 Zip 创建层](https://docs.aws.amazon.com/zh_cn/lambda/latest/dg/creating-deleting-layers.html#layers-create){:target="_blank"}

    - zip 下载地址：
        - [Linux amd64](https://static.<<<custom_key.brand_main_domain>>>/datakit-v2/datakit_aws_extension-linux-amd64.zip)
        - [Linux arm64](https://static.<<<custom_key.brand_main_domain>>>/datakit-v2/datakit_aws_extension-linux-arm64.zip)

    - 打开 Lambda 控制台的 [Layers page](https://console.amazonaws.cn/lambda/home#/layers){:target="_blank"}（层页面）。
    - 选择 **Create layer**（创建层）。
    - 在 **Layer configuration**（层配置）下，在 **Name**（名称）中，输入层的名称。
    - 请选择 **Upload a .zip file**（上传 .zip 文件）。然后，选择 **Upload**（上载）以选择本地 .zip 文件。
    - 选择 **Create**（创建）。

- [通过 ARN 添加层](https://docs.aws.amazon.com/zh_cn/lambda/latest/dg/adding-layers.html){:target="_blank"}

    - 打开 Lambda 控制台的[函数页面](https://console.amazonaws.cn/lambda/home#/functions){:target="_blank"}。
    - 选择要配置的函数。
    - 在**层**下，选择**添加层**。
    - 在**选择层**下，选择 **ARN** 层源。
    - 请在文本框中输入 ARN 并选择**验证**。然后，选择**添加**。

### 配置所需的环境变量 {#config-env}

- ENV_DATAWAY=`https://openway.<<<custom_key.brand_main_domain>>>?token=<your-token>`

## 指标 {#metric}

<!-- markdownlint-disable MD024 -->

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

<!-- markdownlint-disable MD024 -->

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

### 采集器支持 {#input}

- OpenTelemetry
- statsd
- ddtrace # 目前只支持 golang。由于 ddtrace 在 lambda 环境下会有特殊操作，需要添加 `tracer.WithLambdaMode(false)`。
