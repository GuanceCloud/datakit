
# DataKit 更新
---

DataKit 支持手动更新和自动更新两种方式。

## 前置条件 {#req}

- 远程更新要求 DataKit 版本 >= 1.5.9
- 自动更新要求 DataKit 版本 >= 1.1.6-rc1
- 手动更新暂无版本要求

### 手动更新 {#manual}

直接执行如下命令查看当前 DataKit 版本。如果线上有最新版本，则会提示对应的更新命令，如：

> - 如果 [DataKit < 1.2.7](changelog.md#cl-1.2.7)，此处只能用 `datakit --version`
> - 如果 DataKit < 1.2.0，请[直接使用更新命令](changelog.md#cl-1.2.0-break-changes)

<!-- markdownlint-disable MD046 -->

=== "Linux/macOS"

    ``` shell
    $ datakit version
    
           Version: 1.2.8
            Commit: e9ccdfbae4
            Branch: testing
     Build At(UTC): 2022-03-11 11:07:06
    Golang Version: go version go1.18.3 linux/amd64
          Uploader: xxxxxxxxxxxxx/xxxxxxx/xxxxxxx
    ReleasedInputs: all
    ---------------------------------------------------
    
    Online version available: 1.2.9, commit 9f5ac898be (release at 2022-03-10 12:03:12)
    
    Upgrade:
{{ InstallCmd 4 (.WithPlatform "unix") (.WithUpgrade true) }}
    ```

=== "Windows"

    ``` powershell
    $ datakit.exe version
    
           Version: 1.2.8
            Commit: e9ccdfbae4
            Branch: testing
     Build At(UTC): 2022-03-11 11:07:36
    Golang Version: go version go1.18.3 linux/amd64
          Uploader: xxxxxxxxxxxxx/xxxxxxx/xxxxxxx
    ReleasedInputs: all
    ---------------------------------------------------
    
    Online version available: 1.2.9, commit 9f5ac898be (release at 2022-03-10 12:03:12)
    
    Upgrade:
{{ InstallCmd 4 (.WithPlatform "windows") (.WithUpgrade true) }}
    ```
<!-- markdownlint-enable -->

---

如果当前 DataKit 处于被代理模式，自动更新的提示命令中，会自动加上代理设置：

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
    HTTPS_PROXY=http://10.100.64.198:9530 DK_UPGRADE=1 ...
    ```

=== "Windows"

    ``` powershell
    $env:HTTPS_PROXY="http://10.100.64.198:9530"; $env:DK_UPGRADE="1" ...
    ```
<!-- markdownlint-enable -->

### 远程更新服务 {#auto}

[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9) · [:octicons-beaker-24: Experimental](index.md#experimental)

> 注意：伺服服务不支持 k8s 中安装的 DataKit。

在 DataKit 安装过程中，默认会安装一个远程更新的伺服服务，专用于升级 DataKit 版本。如果是较老的 DataKit 版本，则在 DataKit 升级命令中，可以额外指定参数来安装该伺服服务：

<!-- markdownlint-disable MD046 -->

=== "公网安装"

    ```shell hl_lines="2"
    DK_UPGRADE=1 \
      DK_UPGRADE_MANAGER=1 \
      bash -c "$(curl -L https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh)"
    ```

=== "离线更新"

    [:octicons-tag-24: Version-1.38.0](changelog.md#cl-1.38.0)

    如果已经[线下同步了 DataKit 的安装包](datakit-offline-install.md#offline-advanced)，假定线下安装包地址是 `http://my.static.com/datakit`，则此处的升级命令是

    ```shell hl_lines="3"
    DK_UPGRADE=1 \
      DK_UPGRADE_MANAGER=1 \
      DK_INSTALLER_BASE_URL="http://my.static.com/datakit" \
      bash -c "$(curl -L https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh)"
    ```

???+ note

    伺服服务默认会绑定在 `0.0.0.0:9542` 地址上，如果该地址被占用，可以额外指定：
    
    ```shell hl_lines="3"
    DK_UPGRADE=1 \
      DK_UPGRADE_MANAGER=1 \
      DK_UPGRADE_LISTEN=0.0.0.0:19542 \
      bash -c "$(curl -L https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh)"
    ```

---

由于伺服服务提供了 HTTP API，它有如下参数可选（[:octicons-tag-24: Version-1.38.0](changelog.md#cl-1.38.0)）：

- **`version`**：将 DataKit 升级/降级到指定的版本号（如果是离线安装，需确保指定的版本是否已经同步）
- **`force`**：如果当前 DataKit 尚未启动或行为异常，我们可以通过该参数强制升级它并且拉起服务

我们可以手动调用其接口来实现远程更新，或者通过 DCA 来实现远程更新。

=== "手动调用"

    ```shell
    # 更新到最新 DataKit 版本
    curl -XPOST "http://<datakit-ip>:9542/v1/datakit/upgrade"

    {"msg":"success"}

    # 更新到指定 DataKit 版本
    curl -XPOST "http://<datakit-ip>:9542/v1/datakit/upgrade?version=3.4.5"

    # 强制升级一个 DataKit 版本
    curl -XPOST "http://<datakit-ip>:9542/v1/datakit/upgrade?force=1"
    ```

=== "DCA"

    参见 [DCA 文档](../dca/index.md)。

---

???+ info

    - 升级过程根据网络带宽情况，可能耗时较长（基本等同于手动调用 DataKit 升级命令），请耐心等待 API 返回。如果中途中断，**其行为是未定义的**。
    - 升级过程中，如果指定版本不存在，请求会报错（`3.4.5` 这个版本不存在）：

    ```json
    {
      "error_code": "datakit.upgradeFailed",
      "message": "unable to download script file http://my.static.com/datakit/install-3.4.5.sh: resonse status: 404 Not Found"
    }
    ```

    - 如果当前 DataKit 未启动，则会报错：

    ```json
    {
      "error_code": "datakit.upgradeFailed",
      "message": "get datakit version failed: unable to query current DataKit version: Get \"http://localhost:9529/v1/ping\": dial tcp localhost:9529 connect: connection refused)"
    }
    ```
<!-- markdownlint-enable -->

### 离线更新 {#offline-upgrade}

参见[离线安装](datakit-offline-install.md)相关的章节。

### 更新到指定版本 {#downgrade}

如果需要升级或回退到指定版本，可以通过如下命令进行操作：

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithUpgrade true) (.WithVersion "-3.4.5") }}
    ```
=== "Windows"

    ```powershell
{{ InstallCmd 4 (.WithPlatform "windows") (.WithUpgrade true) (.WithVersion "-3.4.5") }}
    ```
<!-- markdownlint-enable -->

上述命令中的 `<版本号>`，可以从 [DataKit 的发布历史](changelog-{{.Year}}.md)页面找到。

若要回退 DataKit 版本，目前只支持退回到 [1.2.0](changelog.md#cl-1.2.0) 以后的版本，之前的 rc 版本不建议回退。

### 额外支持的环境变量 {#extra-envs}

目前在升级命令中也支持和安装命令一致的环境变量[安装命令支持的环境变量](datakit-install.md#extra-envs)，从 [1.62.1](changelog.md#cl-1.62.1) 版本开始支持。

## FAQ {#faq}

### 更新和安装的差异 {#upgrade-vs-install}

如果要升级较新版本的 DataKit，可以通过：

- 重新安装
- [执行升级命令](datakit-update.md#manual)

在已经安装好 DataKit 的主机上，建议通过升级命令来升级到较新的版本，而不是重新安装。如果重新安装，所有 [*datakit.conf* 里面的配置](datakit-conf.md#maincfg-example)都会被重置为默认设置，比如全局 tag 配置、端口设置等等。这可能不是我们所期望的。

不过，不管是重新安装，还是执行升级命令，所有采集相关的配置，都不会因此变更。

### 版本检测失败的处理 {#version-check-failed}

在 DataKit 安装/升级过程中，安装程序会对当前运行的 DataKit 版本进行检测，以确保当前运行的 DataKit 版本就是升级后的版本。

但是某些情况下，老版本的 DataKit 服务并未卸载成功，导致检测过程中中发现，当前运行的 DataKit 版本号还是老的版本号：

```shell
2022-09-22T21:20:35.967+0800    ERROR   installer  installer/main.go:374  checkIsNewVersion: current version: 1.4.13, expect 1.4.16
```

此时我们可以强制停止老版本的 DataKit，并重启 DataKit：

``` shell
datakit service -T # 停止服务
datakit service -S # 启动新的服务

# 如若不行，可以先卸载 DataKit 服务，并重装服务
datakit service -U # 卸载服务
datakit service -I # 重装服务

# 以上操作完成后，再确认下 DataKit 版本是否是最新版本

datakit version # 确保当前运行的 DataKit 已经是最新的版本

       Version: 1.4.16
        Commit: 1357544bd6
        Branch: master
 Build At(UTC): 2022-09-20 11:43:20
Golang Version: go version go1.18.3 linux/amd64
      Uploader: zy-infra-gitlab-prod-runner/root/xxx
ReleasedInputs: checked
```
