---
title     : 'AWS Lambda Extension'
summary   : 'Extend data collection through AWS Lambda'
tags:
  - 'AWS'
__int_icon      : 'icon/awslambda'
dashboard :
  - desc  : 'None'
    path  : '-'
monitor   :
  - desc  : 'None'
    path  : '-'
---


{{.AvailableArchs}}

---

[:octicons-tag-24: Version-1.34.0](../datakit/changelog.md#cl-1.34.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

The AWS Lambda collector collects AWS Lambda metrics and logs through the Lambda extension.

## Installation {#installation}

### Adding a Datakit Layer {#layer}

- [Create a Layer via Zip](https://docs.aws.amazon.com/lambda/latest/dg/creating-deleting-layers.html#layers-create){:target="_blank"}

    - Zip download links:
        - [Linux amd64](https://static.guance.com/datakit/datakit_aws_extension-linux-amd64.zip)
        - [Linux arm64](https://static.guance.com/datakit/datakit_aws_extension-linux-arm64.zip)

    - Open the Lambda console [Layers page](https://console.aws.amazon.com/lambda/home#/layers){:target="_blank"}.
    - Select **Create layer**.
    - Under **Layer configuration**, enter the layer name in **Name**.
    - Choose **Upload a .zip file**. Then, select **Upload** to choose the local .zip file.
    - Select **Create**.

- [Add a Layer via ARN](https://docs.aws.amazon.com/lambda/latest/dg/adding-layers.html){:target="_blank"}

    - Open the Lambda console [Functions page](https://console.aws.amazon.com/lambda/home#/functions){:target="_blank"}.
    - Select the function you want to configure.
    - Under **Layers**, select **Add Layer**.
    - Under **Select a layer**, choose **ARN** as the layer source.
    - Enter the ARN in the text box, select **Verify**, and then choose **Add**.

### Configure the Required Environment Variables {#env}

- ENV_DATAWAY=`https://openway.guance.com?token=<your-token>`

## Metrics {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}

### Collector Support {#input}

- OpenTelemetry
- statsd
- ddtrace # Currently, only Go is supported. Due to special operations required by ddtrace in the lambda environment, you need to add `tracer.WithLambdaMode(false)`.
