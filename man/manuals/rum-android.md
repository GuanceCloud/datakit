# Android
---

操作系统支持：Mac OS/ Windows

# 视图预览

### 概览
Android应用的概览场景统计Android访问的启动次数、PV数、页面错误率、页面加载时间、会话分析、性能分析、错误分析等指标，从Android应用启动、会话分布、访问用户设备、受欢迎页面排行、页面访问量排行、资源错误排行等方面，可视化的展示用户访问Android应用的数据统计，快速定位用户访问Android应用的问题，提高用户访问性能。可通过环境、版本筛选查看已经接入的Android应用。<br />![10.android_overview.png](imgs/input-android-01.png)

### 性能分析

Android应用的性能分析，通过统计PV数、页面加载时间、最受关注页面会话数、页面长任务分析、资源分析等指标，可视化的实时查看整体的Android应用页面性能情况，更精准的定位需要优化的页面，可通过环境、版本等筛选查看已经接入的Android应用。<br />![10.android_performance.png](imgs/input-android-02.png)

### 资源分析

Android应用的资源分析，通过统计资源分类、XHR & Fetch 分析、资源耗时分析等指标，可视化的实时查看整体的Android应用资源情况；通过统计资源请求排行，更精准的定位需要优化的资源；可通过环境、版本等筛选查看已经接入的Android应用。<br />![10.android_resource.png](imgs/input-android-03.png)


### 错误分析

Android应用的JS错误分析，通过统计错误率、Crash、Crash版本、网络错误状态分布等指标，可视化的实时查看整体的Android应用错误情况；通过受影响的资源错误统计，可快速定位资源错误；可通过环境、版本等筛选查看已经接入的Android应用。<br />![10.android_error.png](imgs/input-android-04.png)

# 安装部署

## 前置条件

- 安装 DataKit（[DataKit 安装文档](https://www.yuque.com/dataflux/datakit/datakit-how-to)）


## 应用接入
总共分两步

### 第1步：创建一个Android应用
登录 DataFlux 控制台，进入「应用监测」页面，点击右上角「新建应用」，在新窗口输入「应用名称」，点击「创建」，即可开始配置。<br />![image.png](imgs/input-android-05.png)

### 第2步：安装

#### 2.1 Gradle 配置

在项目的根目录的 `build.gradle` 文件中添加 `DataFlux SDK` 的远程仓库地址
```groovy
buildscript {
    //...省略部分代码
    repositories {
        //...省略部分代码
        //添加 DataFlux SDK 的远程仓库地址
        maven {
            url 'https://mvnrepo.jiagouyun.com/repository/maven-releases'
        }
    }
    dependencies {
        //...省略部分代码
        //添加 DataFlux Plugin 的插件依赖
        classpath 'com.cloudcare.ft.mobile.sdk.tracker.plugin:ft-plugin:1.0.2-alpha02'
    }
}
allprojects {
    repositories {
        //...省略部分代码
        //添加 DataFlux SDK 的远程仓库地址
        maven {
            url 'https://mvnrepo.jiagouyun.com/repository/maven-releases'
        }
    }
}
```

在项目主模块( app 模块)的 `build.gradle` 文件中添加 `DataFlux SDK` 的依赖及 `DataFlux Plugin` 的使用 和 Java 8 的支持

```groovy
dependencies {
    //添加 DataFlux SDK 的依赖
    implementation 'com.cloudcare.ft.mobile.sdk.tracker.agent:ft-sdk:1.2.0-alpha06'
    //捕获 native 层崩溃信息的依赖，需要配合 ft-sdk 使用不能单独使用
    implementation 'com.cloudcare.ft.mobile.sdk.tracker.agent:ft-native:1.0.0-alpha04'
    //推荐使用这个版本，其他版本未做过充分兼容测试
    implementation 'com.google.code.gson:gson:2.8.5'

}
//应用插件
apply plugin: 'ft-plugin'
//配置插件使用参数
FTExt {
    //是否显示 Plugin 日志，默认为 false
    showLog = true
}
android{
	//...省略部分代码
	defaultConfig {
        //...省略部分代码
        ndk {
            //当使用 ft-native 捕获 native 层的崩溃信息时，应该根据应用适配的不同的平台
            //来选择支持的 abi 架构，目前 ft-native 中包含的 abi 架构有 'arm64-v8a',
            // 'armeabi-v7a', 'x86', 'x86_64'
            abiFilters 'armeabi-v7a'
        }
    }
    compileOptions {
        sourceCompatibility = 1.8
        targetCompatibility = 1.8
    }
}
```

> 最新的版本请看上方的 Agent 和 Plugin 的版本名



#### 2.2 基础参数配置

```kotlin
class DemoApplication : Application() {
    override fun onCreate() {
        val config = FTSDKConfig
            .builder(DATAKIT_URL)//Datakit 安装地址
            .setUseOAID(true)//是否使用OAID
            .setDebug(true)//是否开启Debug模式（开启后能查看调试数据）
            .setXDataKitUUID("ft-dataKit-uuid-001");

        FTSdk.install(config)
    }
}
```

#### 

#### 2.3 RUM 配置

```kotlin
FTSdk.initRUMWithConfig(
            FTRUMConfig()
                .setRumAppId(RUM_APP_ID)
                .setEnableTraceUserAction(true)
                .setSamplingRate(0.8f)
                .setExtraMonitorTypeWithError(MonitorType.ALL)
                .setEnableTrackAppUIBlock(true)
                .setEnableTrackAppCrash(true)
                .setEnableTrackAppANR(true)
        )
```

#### 

#### 2.4 Log 配置

```kotlin
   FTSdk.initLogWithConfig(
            FTLoggerConfig()
                .setEnableConsoleLog(true)
                .setServiceName("ft-sdk-demo")
                .setEnableLinkRumData(true)
                .setEnableCustomLog(true)
                .setSamplingRate(0.8f)
        )
```

#### 

#### 2.5 Trace 配置

```kotlin
   FTSdk.initTraceWithConfig(
            FTTraceConfig()
                .setServiceName("ft-sdk-demo")
                .setSamplingRate(0.8f)
                .setEnableLinkRUMData(true)
        )
```


#### 2.6 RUM 用户信息绑定与解绑

```kotlin
//可以在用户登录成功后调用此方法用来绑定用户信息
FTSdk.bindRumUserData("001")

//可以在用户退出登录后调用此方法来解绑用户信息
FTSdk.unbindRumUserData()
```


#### 2.7 关闭 SDK

```kotlin
//如果动态改变 SDK 配置，需要先关闭，以避免错误数据的产生
FTSdk.shutDown()
```


#### 2.8 自定义日志上报

```kotlin
//上传单个日志
FTLogger.getInstance().logBackground("test", Status.INFO)

//批量上传日志
FTLogger.getInstance().logBackground(mutableListOf(LogData("test",Status.INFO)))
```

## Demo演示
github地址：[https://github.com/DataFlux-cn/datakit-android](https://github.com/DataFlux-cn/datakit-android)

# 场景视图
场景 - 新建空白场景 - 系统视图 - Apache 监控视图<br />相关文档 <[DataFlux 场景管理](https://www.yuque.com/dataflux/doc/trq02t)> 

# 异常检测
暂无


# 指标详解
<[Android应用数据采集指标](https://www.yuque.com/dataflux/doc/nnlr2x)>


# 最佳实践
<最佳实践>

## 调用方法说明

## FTSdk
| **方法名** | **含义** | **必须** | **注意** |
| --- | --- | --- | --- |
| install | 安装初始化配置项 | 是 | 需要在 Application 中执行 |
| initRUMWithConfig | 设置 RUM 配置 | 否 | 需要在 Application 中执行 |
| initLogWithConfig | 设置 Log 配置 | 否 | 需要在 Application 中执行 |
| initTraceWithConfig | 设置 Trace 配置 | 否 | 需要在 Application 中执行 |
| bindRumUserData | 绑定用户信息 | 否 |  |
| unbindRumUserData | 解绑用户信息 | 否 |  |
| shutDown | 关闭SDK中正在执行的操作 | 否 |  |


## FTSDKConfig
| **方法名** | **含义** | **必须** | **注意** |
| --- | --- | --- | --- |
| setUseOAID | 是否使用 `OAID` 唯一识别 | 否 | 默认为 `false`，开启后替换 deviceUUID 进行使用，[了解 OAID](#as3yK) |
| setXDataKitUUID | 设置数据采集端的识别 ID | 否 | 默认为随机`uuid` |
| setDebug | 是否开启调试模式 | 否 | 默认为 `false`，开启后方可打印 SDK 运行日志 |
| setEnv | 设置采集环境 | 否 | 默认为 `EnvType.PROD` |
| setOnlySupportMainProcess | 是否只支持在主进程运行 | 否 | 默认为 `true` ，如果需要在其他进程中执行需要将该字段设置为 `false` |


## FTRUMConfig
| **方法名** | **含义** | **必须** | **注意** |
| --- | --- | --- | --- |
| setRumAppId | 设置`Rum AppId` | 是 | 对应设置 RUM `appid`，才会开启`RUM`的采集功能，[获取 appid 方法](#e2b33ee3) |
| setEnableTrackAppCrash | 是否上报 App 崩溃日志 | 否 | 默认为 `false`，开启后会在错误分析中显示错误堆栈数据。<br /> [关于崩溃日志中混淆内容转换的问题](#lR0q8) |
| setExtraMonitorTypeWithError | 设置辅助监控信息 | 否 | 添加附加监控数据到 `Rum` 崩溃数据中，<br />`MonitorType.BATTERY` 为电池余量，`MonitorType.Memory` 为内存用量， `MonitorType.CPU` 为 CPU 占有率 |
| setEnableTrackAppANR | 是否开启  ANR 检测 | 否 | 默认为 `false` |
| setEnableTrackAppUIBlock | 是否开启 UI 卡顿检测 | 否 | 默认为 `false` |
| setEnableTraceUserAction | 是否追踪用户操作 | 否 | 目前只支持用户启动和点击操作 |


## FTLoggerConfig
| **方法名** | **含义** | **必须** | **注意** |
| --- | --- | --- | --- |
| setServiceName | 设置服务名 | 否 | 默认为 `df_rum_android`  |
| setSampleRate | 设置采集率 | 否 | 采集率的值范围为>= 0、<= 1，默认值为 1 |
| setTraceConsoleLog | 是否上报控制台日志 | 否 | 日志等级对应关系<br />Log.v -> ok;<br />Log.i、Log.d -> info;<br />Log.e -> error;<br />Log.w -> warning |
| setEnableLinkRUMData | 是否与 RUM 数据关联 | 否 | 默认为 `false` |
| setLogCacheDiscardStrategy | 设置频繁日志丢弃规则 | 否 | 默认为 <br />`LogCacheDiscard.DISCARD`  ，<br />`DISCARD` 为丢弃追加数据，`DISCARD_OLDEST` 丢弃老数据 |
| setEnableCustomLog | 是否上传自定义日志 | 否 | 默认为 `false` |


## FTTraceConfig
| **方法名** | **含义** | **必须** | **注意** |
| --- | --- | --- | --- |
| setServiceName | 设置服务名 | 否 | 默认为 `df_rum_android`  |
| setSampleRate | 设置采集率 | 否 | 采集率的值范围为>= 0、<= 1，默认值为 1 |
| setTraceType | 设置链路追踪的类型 | 否 | 默认为 `DDTrace`，目前支持 `Zipkin` , `Jaeger`, `DDTrace`  |
| setEnableLinkRUMData | 是否与 RUM 数据关联 | 否 | 默认为 `false` |
| setTraceContentType | 设置过滤规则 | 否 | 过滤链路的规则 content-type 类型 |


## FTLogger
| **方法名** | **含义** | **必须** | **注意** |
| --- | --- | --- | --- |
| logBackground | 自定义日志 | 否 | 日志过快会触发丢弃机制，支持单条日志和批量日志 |


# <br />

## Proguard 混淆配置
```c
-dontwarn com.ft.sdk.**

-keep class com.ft.sdk.**{*;}

-keep class ftnative.*{*;}

-keep class com.bun.miitmdid.core.**{*;}

-keepnames class * extends android.view.View
```


## 权限配置说明
| **名称** | **使用原因** |
| --- | --- |
| `READ_PHONE_STATE` | 用于获取手机的设备信息，便于精准分析数据信息 |
| `WRITE_EXTERNAL_STORAGE` | 用户存储缓存数据 |

> 关于如何申请动态权限，具体详情参考[Android Developer](https://developer.android.google.cn/training/permissions/requesting?hl=en)


# 故障排查


### 关于 OAID


#### 介绍

在 `Android 10` 版本中，非系统应用将不能获取到系统的 `IMEI`、`MAC`等信息。面对该问题移动安全联盟联合国内的手机厂商推出了<br />补充设备标准体系方案，选择用 `OAID` 字段作为IMEI等系统信息的替代字段。`OAID` 字段是由中国信通院联合华为、小米、OPPO、<br />VIVO 等厂商共同推出的设备识别字段，具有一定的权威性。目前 DataFlux SDK 使用的是 `oaid_sdk_1.0.22.aar`

> 关于 OAID 可移步参考[移动安全联盟](http://www.msa-alliance.cn/col.jsp?id=120)



#### 使用

使用方式和资源下载可参考[移动安全联盟的集成文档](http://www.msa-alliance.cn/col.jsp?id=120)


#### 示例

下载好资源文件后，将 oaid_sdk_1.0.22.aar 拷贝到项目的 libs 目录下，并设置依赖<br />[获取最新版本](http://www.msa-alliance.cn/col.jsp?id=120)

![image.png](imgs/input-android-06.png)

将下载的资源中的 `supplierconfig.json` 文件拷贝到主项目的 `assets` 目录下，并修改里面对应的内容，特别是需要设置 `appid` 的部分。需要设置 `appid` 的部分需要去对应厂商的应用商店里注册自己的 `app`。



![image.png](imgs/input-android-07.png)<br />![image.png](imgs/input-android-08.png)


##### 设置依赖

```groovy
implementation files('libs/oaid_sdk_1.0.22.arr')
```


##### 混淆设置

```
 -keep class com.bun.miitmdid.core.**{*;}
```


##### 设置 gradle

编译选项，这块可以根据自己的对平台的选择进行合理的配置

```groovy
ndk {
    abiFilters 'armeabi-v7a','x86','arm64-v8a','x86_64','armeabi'
}
packagingOptions {
    doNotStrip "*/armeabi-v7a/*.so"
    doNotStrip "*/x86/*.so"
    doNotStrip "*/arm64-v8a/*.so"
    doNotStrip "*/x86_64/*.so"
    doNotStrip "armeabi.so"
}
```

> 以上步骤配置完成后，在配置 FT SDK 时调用 FTSDKConfig 的 setUseOAID(true) 方法即可



### HttpClient RUM 问题

目前 `HttpClient` 无法做到 `RUM` 网络请求的性能分析，目前只支持 `okhttp3` 做网络数据的追踪


### 日志混淆内容转换


#### 问题描述

当你的应用发生崩溃且你已经接入 `DataFlux SDK` 同时你也开启了崩溃日志上报的开关时，你可以到 `DataFlux` 后台你的工作空间下的日志模块找到相应<br />的崩溃日志。如果你的应用开启了混淆，此时你会发现崩溃日志的堆栈信息也被混淆，无法直接定位具体的崩溃位置，因此你需要按以下方式来解决该问题。


#### 解决方式

-  找到 mapping 文件。如果你开启了混淆，那么在打包的时候会在该目录下（`module-name` -> `build`-> `outputs` -> `mapping`）生成一个混淆文件映射表（`mapping.txt`）,该文件就是源代码与混淆后的类、方法和属性名称之间的映射。因此每次发包后应该根据应用的版本号保存好对应 `mapping.txt` 文件，以备根据后台日志 `tag` 中的版本名字段（`app_version_name`）来找到对应 `mapping.txt` 文件。 
-  下载崩溃日志文件。到 DataFlux 的后台中把崩溃日志下载到本地，这里假设下载到本地的文件名为 `crash_log.txt` 
-  运行 `retrace` 命令转换崩溃日志。`Android SDK` 中自带 `retrace` 工具（该工具在目录 `sdk-root/tools/proguard` 下，`windows` 版本是 `retrace.bat,Mac/Linux` 版本是 `retrace.sh`），通过该工具可以恢复崩溃日志的堆栈信息。命令示例  
```
retrace.bat -verbose mapping.txt crash_log.txt
```

-  上一步是通过 `retrace` 命令行来执行，当然也可以通过 `GUI` 工具。在 `<sdk-root>/tools/proguard/bin` 目录下有个 `proguardgui.bat` 或 `proguardgui.sh` GUI 工具。运行 `proguardgui.bat` 或者 `./proguardgui.sh` -> 从左侧的菜单中选择“`ReTrace`” -> 在上面的 `Mapping file` 中选择你的 `mapping` 文件，在下面输入框输入要还原的代码 ->点击右下方的“`ReTrace!`” 
