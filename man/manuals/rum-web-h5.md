# Web页面 (HTML5)
---

## 视图预览
![image.png](imgs/input-web-h5-01.png)<br />![image.png](imgs/input-web-h5-02.png)

## 安装部署

### 前置条件

- 至少拥有一台内网服务器，且已<[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>。
- **开放9529端口**：（RUM数据传输端口）**生产环境建议采用域名或slb进行数据接收，然后转发至datakit所在服务器的9529端口。**
- **测试环境**：开发或测试时可将数据发送至datakit所在服务器的内网9529端口
- **生产环境**：因涉及外网RUM数据收集，需开放datakit所在服务器的外网9529端口（可利用slb对外转发数据至datakit所在服务器的9529端口，或者用域名收集数据并转发至datakit所在服务器的9529端口，同时建议用https加密协议进行传输）


### 配置实施
**登录DF平台——选择用户访问监测——新建应用——输入应用名称——创建——选择web类型**

备注：单个project中理论上只有一个html文档，需要在该html文档中添加可观测js，如果存在多个项目均需要接入，则需要在多个项目的project中添加js，建议不同的项目在DF可观测平台上创建不同的应用，方便后期的管理以及问题的排查。<br />![image.png](imgs/input-web-h5-03.png)

| 接入方式 | 简介 |
| --- | --- |
| NPM | 通过把 SDK 代码一起打包到你的前端项目中，此方式可以确保对前端页面的性能不会有任何影响，不过可能会错过 SDK 初始化之前的的请求、错误的收集。 |
| CDN 异步加载 | 通过 CDN 加速缓存，以异步脚本引入的方式，引入 SDK 脚本，此方式可以确保 SDK 脚本的下载不会影响页面的加载性能，不过可能会错过 SDK 初始化之前的的请求、错误的收集。 |
| CDN 同步加载 | 通过 CDN 加速缓存，以同步脚本引入的方式，引入 SDK 脚本，此方式可以确保能够收集到所有的错误，资源，请求，性能指标。不过可能会影响页面的加载性能。 |


| 参数 | 类型 | 是否必须 | 默认值 | 描述 |
| --- | --- | --- | --- | --- |
| `applicationId` | String | 是 |  | 从 dataflux 创建的应用 ID |
| `datakitOrigin` | String | 是 |  | datakit 数据上报 Origin 注释: <br />`协议（包括：//），域名（或IP地址）[和端口号]`<br /> 例如：<br />https://www.datakit.com, <br />http://100.20.34.3:8088 |
| `env` | String | 是 |  | web 应用当前环境， 如 prod：线上环境；gray：灰度环境；pre：预发布环境 common：日常环境；local：本地环境； |
| `version` | String | 是 |  | web 应用的版本号 |
| `resourceSampleRate` | Number | 否 | `100` | 资源指标数据采样率百分比: <br />`100` 表示全收集，<br />`0` 表示不收集 |
| `sampleRate` | Number | 否 | `100` | 指标数据采样率百分比: <br />`100` 表示全收集，<br />`0` 表示不收集 |
| `trackSessionAcrossSubdomains` | Boolean | 否 | `false` | 同一个域名下面的子域名共享缓存 |
| `allowedDDTracingOrigins` | Array | 否 | `[]` | 允许注入<br />`ddtrace`<br /> 采集器所需header头部的所有请求列表。可以是请求的origin，也可以是是正则，origin: <br />`协议（包括：//），域名（或IP地址）[和端口号]`<br /> 例如：<br />`["https://api.example.com", /https:\\/\\/.*\\.my-api-domain\\.com/]` |
| `trackInteractions`<br /> | Boolean | 否 | `false` | 是否开启用户行为采集，开启后可采集用户在web页面中的多种操作事件。 |

**接入示例（同步载入）：**<br />![image.png](imgs/input-web-h5-04.png)


### [Web应用分析](https://www.yuque.com/dataflux/doc/htei6b)


### 高级功能
<[自定义用户标识](https://www.yuque.com/dataflux/doc/az7ofn)> 此方法需保证在rum-js初始化之后可以读到。<br /><[自定义设置会话](https://www.yuque.com/dataflux/doc/dx8cqi)><br /><[自定义添加额外的数据TAG](https://www.yuque.com/dataflux/doc/gxavg8)>

## 场景视图
DF平台已默认内置，无需手动创建

**如有需要，可参照以下步骤进行创建**<br />场景 - 新建空白场景 - 场景模板-Web应用用户访问监测<br />相关文档 <[DataFlux 场景管理](https://www.yuque.com/dataflux/doc/trq02t)> 

## 异常检测
异常检测库 - 新建检测库 - 主机检测库<br />相关文档 <[DataFlux 内置检测库](https://www.yuque.com/dataflux/doc/br0rm2)> 

| 序号 | 规则名称 | 触发条件 | 级别 | 检测频率 |
| --- | --- | --- | --- | --- |
| 1 | RUM页面耗时异常 | 页面加载平均耗时 > 7s | 警告 | 5m |
| 2 | RUM页面耗时异常 | 页面加载平均耗时 > 3s | 紧急 | 5m |

| 序号 | 规则名称 | 触发条件 | 级别 | 检测频率 |
| --- | --- | --- | --- | --- |
| 1 | RUM页面JS错误异常次数过多 | js错误次数 > 50 | 警告 | 5m |
| 2 | RUM页面JS错误异常次数过多 | js错误次数 > 100 | 紧急 | 5m |


## 数据类型详情
<[WEB应用-数据类型详情](https://www.yuque.com/dataflux/doc/hlge69)>

## 最佳实践
<[web应用监控（RUM）最佳实践](https://www.yuque.com/dataflux/bp/web)><br /><[JAVA应用-RUM-APM-LOG 联动分析](https://www.yuque.com/dataflux/bp/java-rum-apm-log)><br /><[Kubernetes应用的RUM-APM-LOG联动分析](https://www.yuque.com/dataflux/bp/k8s-rum-apm-log)>

## 故障排查

### 1、[产生 Script error 消息的原因](https://www.yuque.com/dataflux/doc/eqs7v2#f4de165f)

### 2、[资源数据(ssl, tcp, dns, trans,ttfb)收集不完整问题](https://www.yuque.com/dataflux/doc/eqs7v2#421cacac)

### 3、[针对跨域资源的问题](https://www.yuque.com/dataflux/doc/eqs7v2#5a624065)

