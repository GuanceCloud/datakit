
# Offline Deployment
---

In some cases, the target machine does not have a public access exit, so you can install DataKit offline as follows.

## Agent Installation {#install-via-proxy}

If there is a machine in the intranet that can access the extranet, a proxy can be deployed at the node to proxy the access traffic of the intranet machine through the machine.

At present, DataKit has a inner proxy collector; The same goal can also be achieved through Nginx forward proxy function. The basic network structure is as follows:

<figure markdown>
  ![](https://static.guance.com/images/datakit/dk-nginx-proxy.png){ width="700"}
</figure>

### Preconditions {#requrements}

- Install a DataKit on a machine with a public network exit [in the normal installation mode](datakit-install.md), and turn on the proxy collector on the DataKit, assuming that the [proxy](proxy.md) collector is located in Datakit IP 1.2. 3.4, with the following configuration:

```toml
[[inputs.proxy]]
  ## default bind ip address
  bind = "0.0.0.0" 
  ## default bind port
  port = 9530
```

- Or Nginx ready to configure the forward proxy

=== "Linux/Mac"

    - Use the datakit proxy
    
    Add the environment variable `HTTPS_PROXY="1.2.3.4:9530"`, and the installation command is as follows:
    
    ```shell
    export HTTPS_PROXY=http://1.2.3.4:9530; DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    ```
    
    - Using the Nginx proxy
    
    Add the environment variable `DK_PROXY_TYPE="nginx"; DK_NGINX_IP="1.2.3.4";`, and the installation command is as follows:
    
    ```shell
    export DK_PROXY_TYPE="nginx"; DK_NGINX_IP="1.2.3.4"; DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    ```

=== "Windows"

    - Using the datakit proxy
    
    Add the environment variable `$env:HTTPS_PROXY="1.2.3.4:9530"`, and the installation command is as follows:
    
    ```powershell
    $env:HTTPS_PROXY="1.2.3.4:9530"; $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
    ```
    
    - Using the Nginx proxy
    
    Add the environment variable `$env:DK_PROXY_TYPE="nginx"; $env:DK_NGINX_IP="1.2.3.4";`, and the installation command is as follows:
    
    ```powershell
    $env:DK_PROXY_TYPE="nginx"; $env:DK_NGINX_IP="1.2.3.4"; $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -ProxyUsage Override -ProxyList $env:HTTPS_PROXY -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
    ```
    
    > Note: Other setup parameter settings are the same as [normal setup](datakit-install.md).

---


## Full Offline Installation {#offline}

When there is no external network in the environment, the installation package can only be downloaded from the public network to the internal network by moving the hard disk (U disk).

There are two strategies to choose from for full offline installation:

- Simple mode: Directly copy the files in the U disk to each host and install DataKit. However, Simple Mode currently does **not support** [setting through environment variables](datakit-install.md#extra-envs) during installation.
- Advanced mode: Deploy an Nginx on the intranet and build a file server with Nginx instead of static.guance.com.

### Simple Mode {#offline-simple}

The address of the following files can be downloaded through wget and other download tools, or directly enter the corresponding URL to download in the browser.

???+ Attention

    When downloading from Safari browser, the suffix name may be different (for example, downloading the `. tar.gz ` file to `. tar `), which will cause the installation to fail. It is recommended to download with Chrome browser. 

- Download the packet [data.tar.gz](https://static.guance.com/datakit/data.tar.gz) first, which is the same for every platform.

- Then download two more installers:

=== "Windows 32 bit"

    - [Installer](https://static.guance.com/datakit/installer-windows-386.exe){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-windows-386-1.5.1.tar.gz){:target="_blank"}

=== "Windows 64 bit"

    - [Installer](https://static.guance.com/datakit/installer-windows-amd64.exe){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-windows-amd64-1.5.1.tar.gz){:target="_blank"}

=== "Linux X86 32 bit"

    - [Installer](https://static.guance.com/datakit/installer-linux-386){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-linux-386-1.5.1.tar.gz){:target="_blank"}

=== "Linux X86 64 bit"

    - [Installer](https://static.guance.com/datakit/installer-linux-amd64){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-linux-amd64-1.5.1.tar.gz){:target="_blank"}

=== "Linux Arm 32 bit"

    - [Installer](https://static.guance.com/datakit/installer-linux-arm){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-linux-arm-1.5.1.tar.gz){:target="_blank"}

=== "Linux Arm 64 bit"

    - [Installer](https://static.guance.com/datakit/installer-linux-arm64){:target="_blank"}
    - [DataKit](https://static.guance.com/datakit/datakit-linux-arm64-1.5.1.tar.gz){:target="_blank"}

After downloading, you should have three files (`<OS-ARCH>` here refers to the platform-specific installation package):

- `datakit-<OS-ARCH>.tar.gz`
- `installer-<OS-ARCH>` or `installer-<OS-ARCH>.exe`
- `data.tar.gz`

Copy these files to the corresponding machine (via USB flash drive or scp and other commands).

### Installation {#install}

=== "Windows"

    You need to run the Powershell with administrator privileges to execute:
    
    ```powershell
    .\installer-windows-amd64.exe --offline --dataway "https://openway.guance.com?token=<YOUR-TOKEN>" --srcs .\datakit-windows-amd64-1.5.1.tar.gz,.\data.tar.gz
    ```

=== "Linux"

    To run with root privileges:
    
    ```shell
    chmod +x installer-linux-amd64
    ./installer-linux-amd64 --offline --dataway "https://openway.guance.com?token=<YOUR-TOKEN>" --srcs datakit-linux-amd64-1.5.1.tar.gz,data.tar.gz
    ```

### Advanced Mode {#offline-advanced}

DataKit is currently installed on the public web, and all binary data and installation scripts are downloaded from the static.guance.com site. For machines that cannot access the site, you can replace the static.guance.com site by deploying a file server on the intranet.

The network traffic topology of advanced mode is as follows:

<figure markdown>
  ![](https://static.guance.com/images/datakit/nginx-file-server.png){ width="700"}
</figure>


Prepare a machine that can be accessed on the intranet, install Nginx on the machine, and download (or copy) the files required for DataKit installation to the Nginx server, so that other machines can download the installation files from the Nginx file server to complete the installation.

- Setting up the Nginx file server {#nginx-config}

Add configuration in nginx.conf

```
server {
    listen 8080;
    server_name _;
    ## Map to the following directory
    location / {
        root /;
        autoindex on;
        autoindex_exact_size off;
        autoindex_localtime on;
        charset utf-8,gbk;
    }
}
```

Restart Nginx：

```shell
nginx -t        # test configuration
nginx -s reload # reload configuration
```

- Download the files to the */datakit* directory where the Nginx server is located, taking wget downloading the Linux AMD64 platform installation package as an example:

```shell
#!/bin/bash

mkdir -p /datakit
wget -P /datakit https://static.guance.com/datakit/install.sh
wget -P /datakit https://static.guance.com/datakit/version
wget -P /datakit https://static.guance.com/datakit/data.tar.gz
wget -P /datakit https://static.guance.com/datakit/installer-linux-amd64-1.5.1
wget -P /datakit https://static.guance.com/datakit/datakit-linux-amd64-1.5.1.tar.gz

# Download other toolkits: sources is the installation package used to turn on the RUM sourcemap function. If this function is not turned on, you can choose not to download it.
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

mkdir -p /datakit/sourcemap/jdk \
  /datakit/sourcemap/R8       \
  /datakit/sourcemap/proguard \
  /datakit/sourcemap/ndk      \
  /datakit/sourcemap/libs     \
  /datakit/sourcemap/atosl

for((i=0;i<${#sources[@]};i++)); do
  wget https://static.guance.com${sources[$i]} -O ${sources[$i]}
done
```

- Prepare for installation

On the intranet machine, point it to the Nginx file server by setting `DK_INSTALLER_BASE_URL`:

=== "Linux/Mac"
    
    ```shell
    HTTPS_PROXY=http://1.2.3.4:9530 \
    DK_INSTALLER_BASE_URL="http://<nginxServer>:8080/datakit" \
    DK_DATAWAY="https://dataway?token=<TOKEN>" \
    bash -c "$(curl -L ${DK_INSTALLER_BASE_URL}/install.sh)"
    ```

=== "Windows"

    ```powershel
    $env:HTTPS_PROXY="1.2.3.4:9530";
    $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>";
    $env:DK_INSTALLER_BASE_URL="http://<nginxServer>:8080/datakit";
    Set-ExecutionPolicy Bypass -scope Process -Force;
    Import-Module bitstransfer;
    start-bitstransfer -source ${DK_INSTALLER_BASE_URL}/install.ps1 -destination .install.ps1;
    powershell .install.ps1;
    ```

So far, the offline installation is complete. Note that HTTPS_PROXY is additionally set here.

---

- Update DataKit

If there is a new version of DataKit, you can download it as above and execute the following command to upgrade:

=== "Linux/Mac"

    ```shell
    DK_INSTALLER_BASE_URL="http://<nginxServer>:8080/datakit" \
    DK_UPGRADE=1 \
    	bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    ```

=== "Windows"

    ```powershell
    $env:DK_INSTALLER_BASE_URL="http://<nginxServer>:8080/datakit";
    $env:DK_UPGRADE="1";
    Set-ExecutionPolicy Bypass -scope Process -Force;
    Import-Module bitstransfer;
    start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1;
    powershell .install.ps1;
    ```
