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

### 安全限制

由于 RUM DataKit 一般部署在公网环境，但是只会使用其中特定的 [DataKit API](apis) 接口，其它接口是不能开放的。通过如下方式可加强 API 访问控制，在 *datakit.conf* 中，修改如下 *public_apis* 字段配置：

```toml
[http_api]
  rum_origin_ip_header = "X-Forwarded-For"
  listen = "0.0.0.0:9529"
  disable_404page = true
  rum_app_id_white_list = []

  public_apis = [  # 如果该列表为空，则所有 API 不做访问控制
    "/v1/write/rum",
    "/some/other/apis/..."

    # 除此之外的其他 API，只能 localhost 访问，比如 datakit -M 就需要访问 /stats 接口
    # 另外，DCA 不受这个影响，因为它是独立的 HTTP server
  ]
```

其它接口依然可用，但只能通过 DataKit 本机访问，比如[查询 DQL](datakit-dql-how-to) 或者查看 [DataKit 运行状态](datakit-tools-how-to#44462aae)。

## 配置

RUM 采集默认已经支持，无需开启额外的采集器。

## 指标集

RUM 采集器默认会采集如下几个指标集：

- `error`
- `view`
- `resource`
- `long_task`
- `action`

## Sourcemap 转换

通常生产环境的 js 文件会经过各种转换和压缩，与开发时的源代码差异较大，不便于排错(`debug`)。如果需要定位错误至源码中，就得借助于`sourcemap`文件。

DataKit 支持这种源代码文件信息的映射，方法是将对应 js 的`sourcemap`文件进行 zip 压缩打包，命名格式为 `<app_id>-<env>-<version>.zip`，上传至`<DataKit安装目录>/data/rum/`，这样就可以对上报的`error`指标集数据自动进行转换，并追加 `error_stack_source` 字段至该指标集中。

**打包说明** 

将`sourcemap`文件进行 zip 压缩打包，必须要保证该压缩包解压后的文件路径与`error_stack`中 URL 的路径一致。

假设如下 `error_stack`：

```
ReferenceError
  at a.hideDetail @ http://localhost:8080/static/js/app.7fb548e3d065d1f48f74.js:1:1037
  at a.showDetail @ http://localhost:8080/static/js/app.7fb548e3d065d1f48f74.js:1:986
  at <anonymous> @ http://localhost:8080/static/js/app.7fb548e3d065d1f48f74.js:1:1174
```

需要转换的路径是`/static/js/app.7fb548e3d065d1f48f74.js`，与其对应的`sourcemap`路径为`/static/js/app.7fb548e3d065d1f48f74.js.map`，那么对应压缩包解压后的目录结构如下：

```
static/
└── js
    └── app.7fb548e3d065d1f48f74.js.map

```

转换后的`error_stack_source`：

```

ReferenceError
  at a.hideDetail @ webpack:///src/components/header/header.vue:94:0
  at a.showDetail @ webpack:///src/components/header/header.vue:91:0
  at <anonymous> @ webpack:///src/components/header/header.vue:101:0
```

**文件上传和删除**

打包完成后，除了手动上传至 DataKit 相关目录，还可通过 http 接口上传和删除该文件，前提是开启 DCA 服务。

上传：

```
curl -X POST '<dca_address>/v1/rum/sourcemap?app_id=<app_id>&env=<env>&version=<version>' -F "file=@<sourcemap_path>" -H "Content-Type: multipart/form-data"
```

删除：

```
curl -X DELETE '<dca_address>/v1/rum/sourcemap?app_id=<app_id>&env=<env>&version=<version>'
```

变量说明：

- `<dca_address>`: DCA 服务的地址，如 `http://localhost:9531`
- `<app_id>`: 对应 RUM 的 `applicationId`
- `<env>`: 对应 RUM 的 `env`
- `<version>`: 对应 RUM 的 `version`
- `<sourcemap_path>`: 待上传的`sourcemap` 压缩包文件路径

**注意：**

- 该转换过程，只针对 `error` 指标集。
- 当前只支持 js 的 `sourcemap` 转换。
- `sourcemap` 文件名称需要与原文件保持一致，如果未找到对应的`sourcemap`文件，将不进行转换。
- 通过接口上传的`sourcemap`压缩包，不需要重启 DataKit即可生效，但如果是手动上传，需要重启 DataKit，方可生效。


