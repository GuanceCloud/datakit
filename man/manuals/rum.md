{{.CSS}}
# 采集器配置
---

{{.AvailableArchs}}

---

RUM（Real User Monitor）采集器用于收集网页端或移动端上报的用户访问监测数据。

## 接入方式 {#supported-platforms}

<div class="grid cards" markdown>
- :material-web: [__JavaScritp__](../real-user-monitoring/web/app-access.md)
- :material-wechat: [__微信小程序__](../real-user-monitoring/miniapp/app-access.md)
- :material-android: [__Android__](../real-user-monitoring/android/app-access.md)
- :material-apple-ios: [__iOS__](../real-user-monitoring/ios/app-access.md)
- [__Flutter__](../real-user-monitoring/flutter/app-access.md)
- :material-react:[__ReactNative__](../real-user-monitoring/react-native/app-access.md)
</div>

## 前置条件 {#requirements}

- 将 DataKit 部署成公网可访问

建议将 RUM 以单独的方式部署在公网上，==不要跟已有的服务部署在一起==（如 Kubernetes 集群）。因为 RUM 这个接口上的流量可能很大，集群内部的流量会被它干扰到，而且一些可能的集群内部资源调度机制，可能影响 RUM 服务的运行。

- 在 DataKit 上[安装 IP 地理信息库](datakit-tools-how-to.md#install-ipdb)
- 自 [1.2.7](changelog.md#cl-1.2.7) 之后，由于调整了 IP 地理信息库的安装方式，默认安装不再自带 IP 信息库，需手动安装

## 配置 {#config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    或者直接在 *datakit.conf* 中默认采集器中开启即可：

    ``` toml
    default_enabled_inputs = [ "rum", "cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes" ]
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    在 datakit.yaml 中，环境变量 `ENV_DEFAULT_ENABLED_INPUTS` 增加 rum 采集器名称（如下 `value` 中第一个所示）：

    ```yaml
    - name: ENV_DEFAULT_ENABLED_INPUTS
      value: rum,cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container
    ```

## 安全限制 {#security-setting}

由于 RUM DataKit 一般部署在公网环境，但是只会使用其中特定的 [DataKit API](apis.md) 接口，其它接口是不能开放的。通过如下方式可加强 API 访问控制，在 *datakit.conf* 中，修改如下 *public_apis* 字段配置：

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

其它接口依然可用，但只能通过 DataKit 本机访问，比如[查询 DQL](datakit-dql-how-to.md) 或者查看 [DataKit 运行状态](datakit-tools-how-to.md#using-monitor)。

### 禁用 DataKit 404 页面 {#disable-404}

可通过如下配置，禁用公网访问 DataKit 404 页面：

```toml
# datakit.conf
disable_404page = true
```

## 指标集 {#measurements}

RUM 采集器默认会采集如下几个指标集：

- `error`
- `view`
- `resource`
- `long_task`
- `action`

## Sourcemap 转换 {#sourcemap}

通常生产环境的 js 文件或移动端App代码会经过混淆和压缩以减小应用的尺寸，发生错误时的调用堆栈与开发时的源代码差异较大，不便于排错(`troubleshoot`)。如果需要定位错误至源码中，就得借助于`sourcemap`文件。

DataKit 支持这种源代码文件信息的映射，方法是将对应符号表文件进行 zip 压缩打包，命名格式为 `<app_id>-<env>-<version>.zip`，上传至`<DataKit安装目录>/data/rum/<platform>`，这样就可以对上报的`error`指标集数据自动进行转换，并追加 `error_stack_source` 字段至该指标集中。

### 安装 sourcemap 工具集 {#install-tools}

首先需要安装相应的符号还原工具，datakit 提供了一键安装命令来简化工具的安装：

```shell
sudo datakit install --symbol-tools
```

如果安装过程中出现某个软件安装失败的情况，你可能需要根据错误提示手动安装对应的软件


### Zip 包打包说明 {#zip}

=== "Web"

    将js文件经 webpack 混淆和压缩后生成的 `.map` 文件进行 zip 压缩打包，再拷贝到 `<DataKit安装目录>/data/rum/web`目录下，必须要保证该压缩包解压后的文件路径与`error_stack`中 URL 的路径一致。 假设如下 `error_stack`：

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

=== "Android"
    
    Android 目前存在两种 `sourcemap` 文件，一种是 Java 字节码经 `R8`/`Proguard` 压缩混淆后产生的 mapping 文件，另一种为 C/C++ 原生代码编译时未清除符号表和调试信息的（unstripped） `.so` 文件，如果你的安卓应用同时包含这两种 `sourcemap` 文件， 打包时需要把这两种文件都打包进 zip 包中，之后再把 zip 包拷贝到 `<DataKit安装目录>/data/rum/android` 目录下，zip 包解压后的目录结构类似：
    
    ```
    <app_id>-<env>-<version>/
    ├── mapping.txt
    ├── armeabi-v7a/
    │   ├── libgameengine.so
    │   ├── libothercode.so
    │   └── libvideocodec.so
    ├── arm64-v8a/
    │   ├── libgameengine.so
    │   ├── libothercode.so
    │   └── libvideocodec.so
    ├── x86/
    │   ├── libgameengine.so
    │   ├── libothercode.so
    │   └── libvideocodec.so
    └── x86_64/
        ├── libgameengine.so
        ├── libothercode.so
        └── libvideocodec.so
    ```

    默认情况下，`mapping` 文件将位于： `<项目文件夹>/<Module>/build/outputs/mapping/<build-type>/`，`.so` 文件在用CMake编译项目时位于： `<项目文件夹>/<Module>/build/intermediates/cmake/debug/obj/`，用NDK编译时位于：`<项目文件夹>/<Module>/build/intermediates/ndk/debug/obj/`（debug编译） 或 `<项目文件夹>/<Module>/build/intermediates/ndk/release/obj/`（release编译）。

    转换的效果如下：

    === "Java/Kotlin"

        转换前 `error_stack` :

        ```
        java.lang.ArithmeticException: divide by zero
            at prof.wang.activity.TeamInvitationActivity.o0(Unknown Source:1)
            at prof.wang.activity.TeamInvitationActivity.k0(Unknown Source:0)
            at j9.f7.run(Unknown Source:0)
            at java.lang.Thread.run(Thread.java:1012)
        ```
        
        转换后 `error_stack_source` :
    
        ```
        java.lang.ArithmeticException: divide by zero
        at prof.wang.activity.TeamInvitationActivity.onClick$lambda-0(TeamInvitationActivity.java:1)
        at java.lang.Thread.run(Thread.java:1012)
        ```

    === "C/C++ 原生代码"

        转换前 `error_stack` :
    
        ```
        backtrace:
        #00 pc 00000000000057fc  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_4+12)
        #01 pc 00000000000058a4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_3+8)
        #02 pc 00000000000058b4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_2+12)
        #03 pc 00000000000058c4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_1+12)
        #04 pc 0000000000005938  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_crash+112)
        ...
        ```
        
        转换后 `error_stack_source` :
    
        ```
        backtrace:
        
        Abort message: 'abort message for ftNative internal testing'
        #00 0x00000000000057fc /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_4+12)
        xc_test_call_4
        /Users/Brandon/Documents/workplace/working/StudioPlace/xCrash/xcrash_lib/src/main/cpp/xcrash/xc_test.c:65:9
        #01 0x00000000000058a4 /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_3+8)
        xc_test_call_3
        /Users/Brandon/Documents/workplace/working/StudioPlace/xCrash/xcrash_lib/src/main/cpp/xcrash/xc_test.c:73:13
        #02 0x00000000000058b4 /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_2+12)
        xc_test_call_2
        /Users/Brandon/Documents/workplace/working/StudioPlace/xCrash/xcrash_lib/src/main/cpp/xcrash/xc_test.c:79:13
        #03 0x00000000000058c4 /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_1+12)
        xc_test_call_1
        /Users/Brandon/Documents/workplace/working/StudioPlace/xCrash/xcrash_lib/src/main/cpp/xcrash/xc_test.c:85:13
        #04 0x0000000000005938 /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_crash+112)
        xc_test_crash
        /Users/Brandon/Documents/workplace/working/StudioPlace/xCrash/xcrash_lib/src/main/cpp/xcrash/xc_test.c:126:9
        ...
        ```

=== "iOS"

    iOS平台上的 `sourcemap` 文件是以 `.dSYM` 为后缀的带有调试信息的符号表文件，一般情况下，项目编译完和 `.app` 文件在同一个目录下，如下所示：

    ```
    $ ls -l Build/Products/Debug-iphonesimulator/
    total 0
    drwxr-xr-x   6 zy  staff  192  8  9 15:27 Fishing.app
    drwxr-xr-x   3 zy  staff   96  8  9 14:02 Fishing.app.dSYM
    drwxr-xr-x  15 zy  staff  480  8  9 15:27 Fishing.doccarchive
    drwxr-xr-x   6 zy  staff  192  8  9 13:55 Fishing.swiftmodule
    ```

    需要注意，XCode Release编译默认会生成 `.dSYM` 文件，而Debug编译默认不会生成，需要对 XCode 做如下相应的设置：

    ```
    Build Settings -> Code Generation -> Generate Debug Symbols -> Yes
    Build Settings -> Build Option -> Debug Information Format -> DWARF with dSYM File
    ```

    进行 zip 打包时，把相应的 `.dSYM` 文件打包进 zip 包即可，如果你的项目涉及多个 `.dSYM` 文件，需要一起打包到 zip 包内，之后再把 zip 包拷贝到 `<DataKit安装目录>/data/rum/ios` 目录下，zip 包解压后的目录结构类似如下(`.dSYM` 文件本质上是一个目录，和macOS下的可执行程序 `.app` 文件类似)：


    ```
    <app_id>-<env>-<version>/
    ├── AFNetworking.framework.dSYM
    │   └── Contents
    │       ├── Info.plist
    │       └── Resources
    │           └── DWARF
    │               └── AFNetworking
    └── App.app.dSYM
        └── Contents
            ├── Info.plist
            └── Resources
                └── DWARF
                    └── App
    
    ```

### 文件上传和删除 {#upload-delete}

打包完成后，除了手动拷贝至 DataKit 相关目录，还可通过 http 接口上传和删除该文件，前提是 Datakit 开启了 DCA 服务。

上传：

```shell
curl -X POST '<dca_address>/v1/rum/sourcemap?app_id=<app_id>&env=<env>&version=<version>&platform=<platform>' -F "file=@<sourcemap_path>" -H "Content-Type: multipart/form-data"
```

删除：

```shell
curl -X DELETE '<dca_address>/v1/rum/sourcemap?app_id=<app_id>&env=<env>&version=<version>&platform=<platform>'
```

变量说明：

- `<dca_address>`: DCA 服务的地址，如 `http://localhost:9531`
- `<app_id>`: 对应 RUM 的 `applicationId`
- `<env>`: 对应 RUM 的 `env`
- `<version>`: 对应 RUM 的 `version`
- `<platform>` 应用平台，当前支持 `web`/ `android` / `ios`
- `<sourcemap_path>`: 待上传的`sourcemap` 压缩包文件路径

???+ attention

    - 该转换过程，只针对 `error` 指标集
    - 当前只支持 Javascript/Android/iOS 的 sourcemap 转换
    - 如果未找到对应的 sourcemap 文件，将不进行转换
    - 通过接口上传的 sourcemap 压缩包，不需要重启 DataKit 即可生效。但如果是手动上传，需要重启 DataKit，方可生效
