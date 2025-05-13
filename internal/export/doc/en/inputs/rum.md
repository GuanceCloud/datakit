---
title     : 'RUM'
summary   : 'Collect user behavior data'
tags:
  - 'RUM'
__int_icon      : 'icon/rum'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

RUM (Real User Monitor) collector is used to collect user access monitoring data reported by web page or mobile terminal.

## Configuration {#config}

### Access Mode {#supported-platforms}

<div class="grid cards" markdown>
- :material-web: [__JavaScript__](../real-user-monitoring/web/app-access.md)
- :material-wechat: [__WeChat Mini-Program__](../real-user-monitoring/miniapp/app-access.md)
- :material-android: [__Android__](../real-user-monitoring/android/app-access.md)
- :material-apple-ios: [__iOS__](../real-user-monitoring/ios/app-access.md)
- [__Flutter__](../real-user-monitoring/flutter/app-access.md)
- :material-react:[__ReactNative__](../real-user-monitoring/react-native/app-access.md)
</div>

### Preconditions {#requirements}

- Deploy DataKit to be publicly accessible

It is recommended that RUM be deployed separately on the public network, not with existing services (such as Kubernetes cluster). As the traffic on RUM interface may be very large, the traffic within the cluster will be disturbed by it, and some possible resource scheduling mechanisms within the cluster may affect the operation of RUM services.

- On the DataKit [install IP geo-Repository](../datakit/datakit-tools-how-to.md#install-ipdb)
- Since [1.2.7](../datakit/changelog.md#cl-1.2.7), due to the adjustment of the installation method of IP geographic information base, the default installation no longer comes with its own IP information base, but needs to be installed manually.

### Collector Configuration {#input-config}
<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    We can also enable RUM input in *datakit.conf*:

    ``` toml
    default_enabled_inputs = [ "rum", "cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes" ]
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    In *datakit.yaml*, the environment variable `ENV_DEFAULT_ENABLED_INPUTS` adds the rum collector name (as shown in the first in `value` below):

    ```yaml
    - name: ENV_DEFAULT_ENABLED_INPUTS
      value: rum,cpu,disk,diskio,mem,swap,system,hostobject,net,host_processes,container
    ```
<!-- markdownlint-enable -->
### Security Restrictions {#security-setting}

See [DataKit API access control](../datakit/datakit-conf.md#public-apis).

### Disable DataKit 404 Page {#disable-404}

You can disable public network access to DataKit 404 pages with the following configuration:

```toml
# datakit.conf
disable_404page = true
```

## RUM {#rum}

The RUM collector collects the following metric sets by default:

- `error`
- `view`
- `resource`
- `long_task`
- `action`

## Sourcemap Transformation {#sourcemap}

Usually, js files in production environment or App code on mobile side will be confused and compressed to reduce the size of application. The call stack when an error occurs is quite different from the source code at development time, which is inconvenient for debugging (`troubleshoot`). If you need to locate errors in the source code, you have to rely on the `sourcemap` file.

DataKit supports this mapping of source code file information by zipping the corresponding symbol table file, named *<app_id\>-<env\>-<version\>.zip* and uploading it to *<DataKit Installation Directory\>/data/rum/<platform\>* so that the reported `error` measurement data can be automatically converted and the `error_stack_source` field appended to the metric set.


<!-- markdownlint-disable MD046 -->
???+ warning "Limits on Sourcemap files"

    All Sourcemap files must with extension *.map*, and single *.map* file(after unzip) should not exceed 4GiB.
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD025 -->
### Install the sourcemap tools {#install-tools}
<!-- markdownlint-enable -->
First, you need to install the corresponding symbol restoration tool. DataKit provides a one-click installation command to simplify the installation of the tool:

```shell
sudo datakit install --symbol-tools
```

If a software installation fails during the installation process, you may need to manually install the corresponding software according to the error prompt.

### Zip Packaging Instructions {#zip}
<!-- markdownlint-disable MD046 -->
=== "Web"

    After the js file is obfuscated and compressed by webpack, the `.map` file is zip compressed and packaged, and then copied to the *<DataKit installation directory\>/data/rum/web* directory. It is necessary to ensure that the uncompressed file path of the compressed package is consistent with the URL path in `error_stack`. Assume the following `error_stack`：

    ```
    ReferenceError
      at a.hideDetail @ http://localhost:8080/static/js/app.7fb548e3d065d1f48f74.js:1:1037
      at a.showDetail @ http://localhost:8080/static/js/app.7fb548e3d065d1f48f74.js:1:986
      at <anonymous> @ http://localhost:8080/static/js/app.7fb548e3d065d1f48f74.js:1:1174
    ```

    The path to be converted is `/static/js/app.7fb548e3d065d1f48f74.js`, and its corresponding `sourcemap` path is `/static/js/app.7fb548e3d065d1f48f74.js.map`, so the directory structure of the corresponding compressed package after decompression is as follows:

    ```
    static/
    └── js
    └── app.7fb548e3d065d1f48f74.js.map

    ```

    After conversion `error_stack_source`：

    ```

    ReferenceError
      at a.hideDetail @ webpack:///src/components/header/header.vue:94:0
      at a.showDetail @ webpack:///src/components/header/header.vue:91:0
      at <anonymous> @ webpack:///src/components/header/header.vue:101:0
    ```

=== "Mini Program"

    Same as Web except you should copy the `.zip` archive into *<DataKit installation directory\>/data/rum/miniapp* directory.

=== "Android"

    Android currently has two types of `sourcemap` files. One is the mapping file produced by Java bytecode obfuscated by `R8`/`Proguard` compression. The other is an (unstripped) `.so` file that does not clear the symbol table and debugging information when compiling C/C + + native code. If your android application contains these two `sourcemap` files at the same time, you need to package these two files into a zip package when packaging, and then copy the zip package to the *<DataKit installation directory\>/data/rum/android* directory. The directory structure after zip package decompression is similar:

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

    By default, the `mapping` file will be in: *<project folder\>/<Module\>/build/outputs/mapping/<build-type\>/*, the `.so` file will be in: *<project folder\>/<Module\>/build/intermediates/cmake/debug/obj/* when compiling with ndk: *<project folder\>/<Module\>/build/intermediates/ndk/debug/obj/*（debug compilation) or *<project folder\>/<Module\>/build/intermediates/ndk/release/obj/*(release compile).

    The effect of the transformation is as follows:

    === "Java/Kotlin"

        Before conversion `error_stack` :

        ```
        java.lang.ArithmeticException: divide by zero
            at prof.wang.activity.TeamInvitationActivity.o0(Unknown Source:1)
            at prof.wang.activity.TeamInvitationActivity.k0(Unknown Source:0)
            at j9.f7.run(Unknown Source:0)
            at java.lang.Thread.run(Thread.java:1012)
        ```

        After conversion `error_stack_source` :

        ```
        java.lang.ArithmeticException: divide by zero
        at prof.wang.activity.TeamInvitationActivity.onClick$lambda-0(TeamInvitationActivity.java:1)
        at java.lang.Thread.run(Thread.java:1012)
        ```

    === "C/C++ Native Code"

        Before conversion `error_stack` :

        ```
        backtrace:
        #00 pc 00000000000057fc  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_4+12)
        #01 pc 00000000000058a4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_3+8)
        #02 pc 00000000000058b4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_2+12)
        #03 pc 00000000000058c4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_1+12)
        #04 pc 0000000000005938  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_crash+112)
        ...
        ```

        After conversion `error_stack_source` :

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

    The `sourcemap` file on the iOS platform is a symbol table file with debugging information suffixed `.dSYM`. Typically, the project is compiled in the same directory as the `.app` file, as follows:

    ```
    $ ls -l Build/Products/Debug-iphonesimulator/
    total 0
    drwxr-xr-x   6 zy  staff  192  8  9 15:27 Fishing.app
    drwxr-xr-x   3 zy  staff   96  8  9 14:02 Fishing.app.dSYM
    drwxr-xr-x  15 zy  staff  480  8  9 15:27 Fishing.doccarchive
    drwxr-xr-x   6 zy  staff  192  8  9 13:55 Fishing.swiftmodule
    ```

    Note that XCode Release builds the `.dSYM` file by default, while Debug compilation will not be generated by default, so you need to set XCode accordingly:

    ```
    Build Settings -> Code Generation -> Generate Debug Symbols -> Yes
    Build Settings -> Build Option -> Debug Information Format -> DWARF with dSYM File
    ```

    When packaging zip, you can package the corresponding `.dSYM` files into the zip package. If your project involves multiple `.dSYM` files, you need to package them together into the zip package, and then copy the zip package to the *<DataKit installation directory\>/data/rum/ios* directory. The directory structure after zip package decompression is similar to the following (the`.dSYM` file is essentially a directory, which is similar to the executable program `.app` file under macOS):


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

---

???+ info "For RUM Headless"

    For [RUM headless](../dataflux-func/headless.md), you can upload these package files on web pages, and following upload/delete operations are not required.
<!-- markdownlint-enable -->

### File Upload and Delete {#upload-delete}

After packaging, in addition to manually copying to DataKit related directories, the file can also be uploaded and deleted through http interface.

> From DataKit [:octicons-tag-24: Version-1.16.0](../datakit/changelog.md#cl-1.16.0), sourcemap related apis were moved from DCA service to DataKit service.

[Upload](../datakit/apis.md#api-sourcemap-upload):

```shell
curl -X PUT '<datakit_address>/v1/sourcemap?app_id=<app_id>&env=<env>&version=<version>&platform=<platform>&token=<token>' -F "file=@<sourcemap_path>" -H "Content-Type: multipart/form-data"
```

[Delete](../datakit/apis.md#api-sourcemap-delete):

```shell
curl -X DELETE '<datakit_address>/v1/sourcemap?app_id=<app_id>&env=<env>&version=<version>&platform=<platform>&token=<token>'
```

[Verify sourcemap](../datakit/apis.md#api-sourcemap-check):

```shell
curl -X GET '<datakit_address>/v1/sourcemap/check?app_id=<app_id>&env=<env>&version=<version>&platform=<platform>&error_stack=<error_stack>'
```

Variable description:

- `<datakit_address>`: DataKit host，such as `http://localhost:9529`
- `<token>`: The token is specified by dataway field in `datakit.conf`
- `<app_id>`: RUM's application ID
- `<env>`: RUM's tag `env`
- `<version>`: RUM's tag `version`
- `<platform>` RUM supported platform, currently support `web/miniapp/android/ios`
- `<sourcemap_path>`: Path of zipped file path
- `<error_stack>`: The error stack string

<!-- markdownlint-disable MD046 -->
???+ note

    - This conversion process is only for the `error` measurement.
    - Currently only Javascript/Android/iOS sourcemap conversion is supported.
    - If the corresponding sourcemap file is not found, no conversion will be performed.
    - Sourcemap compressed package uploaded through the interface, which does not need to restart DataKit to take effect. However, if it is uploaded manually, you need to restart the DataKit before it can take effect.
<!-- markdownlint-enable -->

## CDN resolve {#cdn-resolve}

For the `resource` indicator, DataKit attempts to analyze whether the resource uses CDN and the corresponding CDN manufacturer. When the `provider_type` field value in the indicator set is "CDN", it indicates that The resource uses CDN, and the `provider_name` field value is the specific CDN manufacturer name.

### Customize the CDN lookup dictionary {#customize-cdn-map}

DataKit has a built-in list of CDN manufacturers. If you find that the CDN you used cannot be recognized, you can modify the list in the configuration file, which is located at */usr/local/datakit/conf.d/rum/rum.conf*, which is determined according to DataKit installation location, where `cdn_map` configuration item is used to customize the CDN dictionary. The CDN list seems like the following JSON:

```json
[
  {
    "domain": "alicdn.com",
    "name": "Aliyun CDN",
    "website": "https://www.aliyun.com"
  },
  ...
]
```

We can easily copy and modify the [built-in CDN Dict](built-in_cdn_dict_config.md){:target="_blank"} config, then paste all the content to the configuration file, remember to restart the DataKit after modification.

## RUM Session Replay {#rum-session-replay}

As of version [:octicons-tag-24: Version-1.5.5](../datakit/changelog.md#cl-1.5.5), DataKit support to collect the data of RUM Session Replay. It needs you to add item `session_replay_endpoints` to RUM configuration as bellow and then restart DataKit.

```toml
[[inputs.rum]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v1/write/rum"]

  ## use to upload rum screenshot,html,etc...
  session_replay_endpoints = ["/v1/write/rum/replay"]

  ...
```

<!-- markdownlint-disable MD046 -->
???+ info

    RUM configuration file is located at */usr/local/datakit/conf.d/rum/rum.conf*(Linux/macOS) and *C:\\Program Files\\datakit\\conf.d\\rum*（Windows） by default, which depend on the operating system you use and the installation location of DataKit.
<!-- markdownlint-enable -->

### RUM Session Replay Filter {#rum-session-replay-filter}

Starting from the DataKit [:octicons-tag-24: Version-1.20.0](../datakit/changelog.md#cl-1.20.0) version, it is supported to use configuration to filter out unnecessary session replay data. New The configuration item name is `filter_rules`, and the format is similar to the following (please refer to `rum.conf.sample` RUM sample configuration file):

```toml
[inputs.rum.session_replay]
#   cache_path = "/usr/local/datakit/cache/session_replay"
#   cache_capacity_mb = 20480
#   clear_cache_on_start = false
#   upload_workers = 16
#   send_timeout = "75s"
#   send_retry_count = 3
   filter_rules = [
       "{ service = 'xxx' or version IN [ 'v1', 'v2'] }",
       "{ app_id = 'yyy' and env = 'production' }"
   ]
```

`filter_rules` is an array of rules. There is an "OR" logical relationship between each rule. That is to say, a certain session replay data will be discarded as long as it hits any one of the rules. It will be discarded only if all the rules fail to hit. reserve. The fields currently supported by filtering rules are as shown in the following table:

| Field name          | Type   | Description                                   | Example         |
| ------------------- | ------ | --------------------------------------------- | --------------- |
| `app_id`            | string | Application ID                                | appid_123456789 |
| `service`           | string | service name                                  | user_center     |
| `version`           | string | Service version                               | v1.0.0          |
| `env`               | string | Service deployment environment                | production      |
| `sdk_name`          | string | RUM SDK name                                  | df_web_rum_sdk  |
| `sdk_version`       | string | RUM SDK version                               | 3.1.5           |
| `source`            | string | data source                                   | browser         |
| `has_full_snapshot` | string | Whether it is full data                       | false           |
| `raw_segment_size`  | int    | Size of raw session replay data (unit: bytes) | 656             |
