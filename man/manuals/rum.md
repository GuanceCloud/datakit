{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# RUM 简介

RUM（Real User Monitor）采集器用于收集网页端或移动端上报的用户访问监测数据。

## 前置条件

- 将 DataKit 部署成公网可访问。

> 注意：可通过如下配置，禁用公网访问 DataKit 404 页面：

```toml
# datakit.conf
disable_404page = true
```

## 配置

RUM 采集默认已经支持，无需开启额外的采集器。

## 指标集

RUM 采集器默认会采集如下几个指标集：

- `error`
- `view`
- `resource`
- `long_task`
- `action`
