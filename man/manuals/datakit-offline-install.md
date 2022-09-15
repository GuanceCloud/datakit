{{.CSS}}
# 离线部署
---

某些时候，目标机器没有公网访问出口，按照如下方式可离线安装 DataKit。

## 代理安装 {#install-via-proxy}

### 1. 通过 Datakit 代理安装 {#with-datakit}

#### 前置条件 {#requrements}

- 通过[正常安装方式](datakit-install.md)，在有公网出口的机器上安装一个 DataKit
- 开通该 DataKit 上的 [proxy](proxy.md) 采集器，假定 proxy 采集器所在 Datakit IP 为 1.2.3.4，有如下配置：

```toml
[[inputs.proxy]]
  ## default bind ip address
  bind = "0.0.0.0" 
  ## default bind port
  port = 9530
```

=== "Linux/Mac"

    - 使用 datakit 代理
    
    增加环境变量 `HTTPS_PROXY="1.2.3.4:9530"`，安装命令如下：
    
    ```shell
    export HTTPS_PROXY=http://1.2.3.4:9530; DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    ```
    
    - 使用 nginx 代理
    
    增加环境变量 `DK_PROXY_TYPE="nginx"; DK_NGINX_IP="1.2.3.4";`，安装命令如下：
    
    ```shell
    export DK_PROXY_TYPE="nginx"; DK_NGINX_IP="1.2.3.4"; DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    ```

=== "Windows"

    - 使用 datakit 代理
    
    增加环境变量 `$env:HTTPS_PROXY="1.2.3.4:9530"`，安装命令如下：
    
    ```powershell
    $env:HTTPS_PROXY="1.2.3.4:9530"; $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
    ```

    - 使用 nginx 代理
    
    增加环境变量 `$env:DK_PROXY_TYPE="nginx"; $env:DK_NGINX_IP="1.2.3.4";`，安装命令如下：
    
    ```powershell
    $env:DK_PROXY_TYPE="nginx"; $env:DK_NGINX_IP="1.2.3.4"; $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
    ```
    
    > 注意：其它安装参数设置，跟[正常安装](datakit-install.md) 无异。

### 2. 通过 Nginx 代理安装 {#with-nginx}

先准备一台可以访问外网的机器，所有内网机节点都可以通内网形式访问到这台上。

在公网的机器上安装 nginx， 将 datakit 安装所需的文件下载到nginx服务器上，这样节点机从 nginx 服务器上下载安装文件既可完成安装。

在安装完成后可以通过 nginx 的正向代理功能将所有的节点机器采集上的数据转发到观测云上（需要 ssl 证书）。

也可以通过 Datakit 的代理功能将数据发送出来。

接下来 就是操作步骤。

> 注意：没有 nginx， 需要先自行安装 nginx。


#### 在nginx机器中配置并下载全量 datakit 文件 {#nginx-config}

在 nginx.conf 中添加配置，用来让节点机下载 dk 安装文件：
```txt
server {
    listen 8080;
    server_name _;
    ## 映射到跟目录下
    location / {
        root /;
        autoindex on;
        autoindex_exact_size off;
        autoindex_localtime on;
        charset utf-8,gbk;
    }
}
```

加载新配置及测试

```shell
nginx -t        # 测试配置
nginx -s reload # reload配置
```


下载文件到 nginx 服务器所在的 /datakit 目录下：

这里准备了一个脚本，其中的 `sources` 是开启 sourcemap 功能使用的安装包，如果未开启此功能，可选择不下载。

```shell
#!/bin/bash

mkdir -p /datakit

# "download install.sh ...."
wget -P /datakit https://static.guance.com/datakit/install.sh

# "download vesion ...."
wget -P /datakit https://static.guance.com/datakit/version

# data
wget -P /datakit https://static.guance.com/datakit/data.tar.gz

# version
version=`cat /datakit/version |grep \"version\" |awk -F "\"" '{print$4}'` && echo "version is: '${version}'"

# "download installer"
wget -P /datakit https://static.guance.com/datakit/installer-linux-amd64-${version}

# "download datakit ...."
wget -P /datakit https://static.guance.com/datakit/datakit-linux-amd64-${version}.tar.gz


# download datakit tools
sources=(
  "/datakit/sourcemap/jdk/OpenJDK11U-jdk_x64_mac_hotspot_11.0.16_8.tar.gz"
  "/datakit/sourcemap/jdk/OpenJDK11U-jdk_aarch64_mac_hotspot_11.0.15_10.tar.gz"
  "/datakit/sourcemap/jdk/OpenJDK11U-jdk_x64_linux_hotspot_11.0.16_8.tar.gz"
  "/datakit/sourcemap/jdk/OpenJDK11U-jdk_aarch64_linux_hotspot_11.0.16_8.tar.gz"
  "/datakit/sourcemap/R8/commandlinetools-mac-8512546_simplified.tar.gz"
  "/datakit/sourcemap/R8/commandlinetools-linux-8512546_simplified.tar.gz"
  "/datakit/sourcemap/proguard/proguard-7.2.2.tar.gz"
  "/datakit/sourcemap/ndk/android-ndk-r22b-x64-mac-simplified.tar.gz"
  "/datakit/sourcemap/ndk/android-ndk-r25-x64-linux-simplified.tar.gz"
  "/datakit/sourcemap/libs/libdwarf-code-20200114.tar.gz"
  "/datakit/sourcemap/libs/binutils-2.24.tar.gz"
  "/datakit/sourcemap/atosl/atosl-20220804-x64-linux.tar.gz"
)

mkdir -p /datakit/sourcemap/jdk
mkdir -p /datakit/sourcemap/R8
mkdir -p /datakit/sourcemap/proguard
mkdir -p /datakit/sourcemap/ndk
mkdir -p /datakit/sourcemap/libs
mkdir -p /datakit/sourcemap/atosl

for((i=0;i<${#sources[@]};i++));
  do
    ## use datakit --symbol-tools 
  wget https://static.guance.com${sources[$i]} -O ${sources[$i]}
done

```


#### 在节点机上 通过 DK_INSTALLER_BASE_URL 下载并安装 {#node-install}

注意修改命令行中的 `nginxServer` 和 `DK_DATAWAY`

=== "Linux/Mac"
    
    ```shell
    DK_INSTALLER_BASE_URL="http://<nginxServer>:8080/datakit" \
    DK_DATAWAY="https://user/PAAS/dataway?token=<TOKEN>" \
    bash -c "$(curl -L ${DK_INSTALLER_BASE_URL}/install.sh)"
    ```

=== "Windows"

    ```powershel
    $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; $env:DK_INSTALLER_BASE_URL="http://<nginxServer>:8080/datakit"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source ${DK_INSTALLER_BASE_URL}/install.ps1 -destination .install.ps1; powershell .install.ps1;
    ```

到此为止，离线安装完成。

---

#### 通过 nginx 升级节点中 Datakit 版本 {#node-upgrade}
下载最新的 Datakit 版本。替换版本即可。

=== "Linux/Mac"
    
    - 下载最新版本 datakit：

    ```shell
    # version 为最新版本的版本号
    wget https://static.guance.com/datakit/datakit-linux-amd64-${version}.tar.gz
    
    # 解压
    tar -zxvf datakit-linux-amd64-${version}.tar.gz
    
    # 通过 ssh 命令行形式下载并重启
    # 注意：没有做免密登录，需要手动输入密码。
    ssh root@<node_ip> "wget http://<nginxServer>:8080/datakit/datakit -O /usr/local/datakit/datakit && systemctl restart datakit"
    
    ```

=== "Windows"

    手动下载最新版的 Datakit 并覆盖之前版本。重启即可。

#### symbol-tools {#symbol-tools}

其它安装文件，如 ipdb/sourcemap 等那一堆文件，都应使用该地址来下载，也即支持在调用 datakit tool 命令时，指定 ENV：

```shell
DK_INSTALLER_BASE_URL=http://<nginxServer>:8080 datakit install --symbol-tools
```


----


## 全离线安装 {#offline}

当环境完全没有外网的情况下，只能通过移动硬盘（U 盘）等方式。

### 下载安装包 {#download}

以下文件的地址，可通过 wget 等下载工具，也可以直接在浏览器中输入对应的 URL 下载。

???+ Attention

    Safari 浏览器下载时，后缀名可能不同（如将 `.tar.gz` 文件下载成 `.tar`），会导致安装失败。建议用 Chrome 浏览器下载。

- 先下载数据包 [data.tar.gz](https://static.guance.com/datakit/data.tar.gz)，每个平台都一样。

- 然后再下载俩个安装程序：

=== "Windows 32 位"

    - [Installer](https://static.guance.com/datakit/installer-windows-386.exe){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-windows-386-{{.Version}}.tar.gz){:target="_blank"}

=== "Windows 64 位"

    - [Installer](https://static.guance.com/datakit/installer-windows-amd64.exe){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-windows-amd64-{{.Version}}.tar.gz){:target="_blank"}

=== "Linux X86 32 位"

    - [Installer](https://static.guance.com/datakit/installer-linux-386){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-linux-386-{{.Version}}.tar.gz){:target="_blank"}

=== "Linux X86 64 位"

    - [Installer](https://static.guance.com/datakit/installer-linux-amd64){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-linux-amd64-{{.Version}}.tar.gz){:target="_blank"}

=== "Linux Arm 32 位"

    - [Installer](https://static.guance.com/datakit/installer-linux-arm){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-linux-arm-{{.Version}}.tar.gz){:target="_blank"}

=== "Linux Arm 64 位"

    - [Installer](https://static.guance.com/datakit/installer-linux-arm64){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-linux-arm64-{{.Version}}.tar.gz){:target="_blank"}

下载完后，应该有三个文件（此处 `<OS-ARCH>` 指特定平台的安装包）：

- `datakit-<OS-ARCH>.tar.gz`
- `installer-<OS-ARCH>` 或 `installer-<OS-ARCH>.exe`
- `data.tar.gz`

将这些文件拷贝到对应机器上（通过 U 盘或 scp 等命令）。

### 安装 {#install}

=== "Windows"

    需以 administrator 权限运行 Powershell 执行：

    ```powershell
    .\installer-windows-amd64.exe --offline --dataway "https://openway.guance.com?token=<YOUR-TOKEN>" --srcs .\datakit-windows-amd64-{{.Version}}.tar.gz,.\data.tar.gz
    ```

=== "Linux"

    需以 root 权限运行：

    ```shell
    chmod +x installer-linux-amd64
    ./installer-linux-amd64 --offline --dataway "https://openway.guance.com?token=<YOUR-TOKEN>" --srcs datakit-linux-amd64-{{.Version}}.tar.gz,data.tar.gz
    ```
