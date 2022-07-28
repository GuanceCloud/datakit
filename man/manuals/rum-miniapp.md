# 小程序 (MiniApp)
---

操作系统支持：Mac OS / Windows

# 视图预览
收集各个小程序应用的指标数据，监控小程序性能指标，错误log，以及资源请求情况数据，以可视化的方式分析各个小程序应用端的性能。

### 概览
![image.png](imgs/input-miniapp-01.png)

### 性能分析
![image.png](imgs/input-miniapp-02.png)

### 查看器
![image.png](imgs/input-miniapp-03.png)


## 

# 安装部署

## 前置条件

- 安装 DataKit（[DataKit 安装文档](https://www.yuque.com/dataflux/datakit/datakit-how-to)）

## 应用接入
总共分两步

### 第1步：创建
登录 DataFlux 控制台，进入「应用监测」页面，点击右上角「新建应用」，在新窗口输入「应用名称」，点击「创建」，即创建完成。以下有操作截屏可参考<br />![image.png](imgs/input-miniapp-04.png)<br />![image.png](imgs/input-miniapp-05.png)

### ![image.png](imgs/input-miniapp-06.png)第2步：接入
两种方式，自由选择

#### 方式1:**npm引入方式**
> 
#### **提示：npm可参考微信 <**[点击进入](https://developers.weixin.qq.com/miniprogram/dev/devtools/npm.html)**>**

```html
const { datafluxRum } = require('@cloudcare/rum-miniapp')
// 初始化 Rum
datafluxRum.init({

	// 必填，Datakit域名地址(这里为示例) 需要在微信小程序管理后台加上域名白名单
  datakitOrigin: 'http:x.x.x.x:9529',
 
  // 必填，当你创建应用后，可以看到自己的applicationId
  applicationId: 'xxxxx', 

	// 选填，小程序的环境
  env: 'testing', 

	// 选填，小程序版本
  version: '1.0.0', 

	// 选填，是否开启trackInteractions
  trackInteractions: true
})
```

#### 方式2:**CDN 本地方式引入（推荐）**
> JS下载地址：[https://static.dataflux.cn/js-sdk/dataflux-rum-miniapp.js](https://static.dataflux.cn/js-sdk/dataflux-rum-miniapp.js)

```html
const { datafluxRum } = require('./js拖到项目中的路径')
// 初始化 Rum
datafluxRum.init({

	// 必填，Datakit域名地址(这里为示例) 需要在微信小程序管理后台加上域名白名单
  datakitOrigin: 'http:x.x.x.x:9529',
 
  // 必填，当你创建应用后，可以看到自己的applicationId
  applicationId: 'xxxxx', 

	// 选填，小程序的环境
  env: 'testing', 

	// 选填，小程序版本
  version: '1.0.0', 

	// 选填，是否开启trackInteractions
  trackInteractions: true
})
```

## Demo展示
> **CDN 下载文件本地方式引入demo演示**

1:下载<br />![image.png](imgs/input-miniapp-07.png)<br />**2:集成**<br />![image.png](imgs/input-miniapp-08.png)<br />3:集成结束

# 场景视图
场景 - 新建场景 -小程序监控场景 <br />![image.png](imgs/input-miniapp-09.png)

# 异常检测
暂无

# 指标详解
<[指标详细说明](https://www.yuque.com/dataflux/doc/dmhe97)>

# 最佳实践
<[最佳实践](https://www.yuque.com/dataflux/doc/dmhe97)>
