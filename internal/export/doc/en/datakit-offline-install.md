
# Offline Deployment
---

In some cases, the target machine does not have a public access exit, so you can install DataKit offline as follows.

## Agent Installation {#install-via-proxy}

If there is a machine in the intranet that can access the external, a proxy can be deployed at the node to proxy the access traffic of the intranet machine through the machine.

At present, DataKit has a inner proxy collector; The same goal can also be achieved through Nginx forward proxy function. The basic network structure is as follows:

```mermaid
flowchart LR
dk1(Datakit)
dk2(Datakit)
dk3(Datakit)
proxy(Nginx or Datakit.Proxy)
cdn(<<<custom_key.brand_name>>> CDN)
studio(openway.<<<custom_key.brand_name>>>.com)
%%%

dk1--> proxy
dk2--> proxy
dk3--> proxy

proxy --> cdn
proxy --> studio
```

### Preconditions {#requrements}

- Install a DataKit on a machine with a public network exit [in the normal installation mode](datakit-install.md), and turn on the proxy collector on the DataKit, assuming that the [proxy](../integrations/proxy.md) collector is located in DataKit IP 1.2. 3.4, with the following configuration:

```toml
[[inputs.proxy]]
  ## default bind ip address
  bind = "0.0.0.0" 
  ## default bind port
  port = 9530
```

- Or Nginx ready to configure the forward proxy
<!-- markdownlint-disable MD046 -->
=== "Linux/Mac"

    - Use the DataKit proxy
    
    Add the environment variable `HTTPS_PROXY="1.2.3.4:9530"`, and the installation command is as follows:
    
    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithEnvs "HTTPS_PROXY" "http://1.2.3.4:9530") (.WithProxy true) }}
    ```
    
    - Using the Nginx proxy
    
    Add the environment variable `DK_PROXY_TYPE="nginx"; DK_NGINX_IP="1.2.3.4";`, and the installation command is as follows:
    
    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithEnvs "HTTPS_PROXY" "http://1.2.3.4:9530") (.WithProxy true) (.WithEnvs "DK_NGINX_IP" "1.2.3.4") }}
    ```

=== "Windows"

    - Using the DataKit proxy
    
    Add the environment variable `$env:HTTPS_PROXY="1.2.3.4:9530"`, and the installation command is as follows:
    
    ```powershell
{{ InstallCmd 4
(.WithPlatform "windows")
(.WithBitstransferOpts "-ProxyUsage Override -ProxyList $env:HTTPS_PROXY")
(.WithEnvs "HTTPS_PROXY" "1.2.3.4:9530")
}}
    ```
    
    - Using the Nginx proxy
    
    Add the environment variable `$env:DK_PROXY_TYPE="nginx"; $env:DK_NGINX_IP="1.2.3.4";`, and the installation command is as follows:
    
    ```powershell
{{ InstallCmd 4
(.WithPlatform "windows")
(.WithBitstransferOpts "-ProxyUsage Override -ProxyList $env:DK_NGINX_IP")
(.WithEnvs "DK_PROXY_TYPE" "nginx")
(.WithEnvs "DK_NGINX_IP" "1.2.3.4")
}}
    ```
    
    > Note: Other setup parameter settings are the same as [normal setup](datakit-install.md).

---
<!-- markdownlint-enable -->

## Full Offline Installation {#offline}

When there is no external network in the environment, the installation package can only be downloaded from the public network to the internal network by moving the hard disk (U disk).

There are two strategies to choose from for full offline installation:

- Simple mode: Directly copy the files in the U disk to each host and install DataKit. However, Simple Mode currently does **not support** [setting through environment variables](datakit-install.md#extra-envs) during installation.
- Advanced mode: Deploy a Nginx on the intranet and build a file server with Nginx instead of static.<<<custom_key.brand_main_domain>>>.

### Simple Mode {#offline-simple}

The address of the following files can be downloaded through wget and other download tools, or directly enter the corresponding URL to download in the browser.
<!-- markdownlint-disable MD046 -->
???+ note

    When downloading from Safari browser, the suffix name may be different (for example, downloading the `. tar.gz ` file to `. tar `), which will cause the installation to fail. It is recommended to download with Chrome browser. 
<!-- markdownlint-enable -->
- Download the packet [data.tar.gz](https://static.<<<custom_key.brand_main_domain>>>/datakit/data.tar.gz) first, which is the same for every platform.

- Then download more installers as below:
<!-- markdownlint-disable MD046 -->

=== "Linux"

    - **X86 32 bit**
        - [`Installer`](https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-linux-386){:target="_blank"}
        - [`DataKit`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-linux-386-{{ .Version }}.tar.gz){:target="_blank"}
        - [`DataKit-Lite`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit_lite-linux-386-{{ .Version }}.tar.gz){:target="_blank"}
        - [`Upgrader`](https://static.<<<custom_key.brand_main_domain>>>/datakit/dk_upgrader-linux-386-{{ .Version }}.tar.gz){:target="_blank"}

    - **X86 64 bit**
        - [`Installer`](https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-linux-amd64){:target="_blank"}
        - [`DataKit`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-linux-amd64-{{ .Version }}.tar.gz){:target="_blank"}
        - [`DataKit-Lite`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit_lite-linux-amd64-{{ .Version }}.tar.gz){:target="_blank"}
        - [`Upgrader`](https://static.<<<custom_key.brand_main_domain>>>/datakit/dk_upgrader-linux-amd64-{{ .Version }}.tar.gz){:target="_blank"}
        - [`APM Auto Instrumentation`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-apm-inject-linux-amd64-{{ .Version }}.tar.gz){:target="_blank"}

    - **Arm 32 bit**
        - [`Installer`](https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-linux-arm){:target="_blank"}
        - [`DataKit`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-linux-arm-{{ .Version }}.tar.gz){:target="_blank"}
        - [`DataKit-Lite`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit_lite-linux-arm-{{ .Version }}.tar.gz){:target="_blank"}
        - [`Upgrader`](https://static.<<<custom_key.brand_main_domain>>>/datakit/dk_upgrader-linux-arm-{{ .Version }}.tar.gz){:target="_blank"}

    - **Arm 64 bit**
        - [`Installer`](https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-linux-arm64){:target="_blank"}
        - [`DataKit`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-linux-arm64-{{ .Version }}.tar.gz){:target="_blank"}
        - [`DataKit-Lite`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit_lite-linux-arm64-{{ .Version }}.tar.gz){:target="_blank"}
        - [`Upgrader`](https://static.<<<custom_key.brand_main_domain>>>/datakit/dk_upgrader-linux-arm64-{{ .Version }}.tar.gz){:target="_blank"}
        - [`APM Auto Instrumentation`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-apm-inject-linux-arm64-{{ .Version }}.tar.gz){:target="_blank"}

=== "Windows"

    For Windows, only X86 is available.

    - **32 bit**
        - [`Installer`](https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-windows-386.exe){:target="_blank"}
        - [`DataKit`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-windows-386-{{ .Version }}.tar.gz){:target="_blank"}
        - [`DataKit-Lite`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit_lite-windows-386-{{ .Version }}.tar.gz){:target="_blank"}
        - [`Upgrader`](https://static.<<<custom_key.brand_main_domain>>>/datakit/dk_upgrader-windows-386-{{ .Version }}.tar.gz){:target="_blank"}

    - **64 bit**
        - [`Installer`](https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-windows-amd64.exe){:target="_blank"}
        - [`DataKit`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-windows-amd64-{{ .Version }}.tar.gz){:target="_blank"}
        - [`DataKit-Lite`](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit_lite-windows-amd64-{{ .Version }}.tar.gz){:target="_blank"}
        - [`Upgrader`](https://static.<<<custom_key.brand_main_domain>>>/datakit/dk_upgrader-windows-amd64-{{ .Version }}.tar.gz){:target="_blank"}

<!-- markdownlint-enable -->

After downloading, you should have a few files as below (*[OS-ARCH]* here refers to the platform-specific installation package):

- *datakit-[OS-ARCH].tar.gz*
- *dk_upgrader-[OS-ARCH].tar.gz*
- *installer-[OS-ARCH]` or `installer-[OS-ARCH].exe*
- *data.tar.gz*
- *datakit-apm-inject-[OS-ARCH].tar.gz*

Copy these files to the corresponding machine (via USB flash drive or `scp` and other commands).

<!-- markdownlint-disable MD046 -->
???+ note

    It is crucial to download each of these files completely. They may not be reusable between different versions. For example, the installer program behaves differently across various DataKit versions because it may adjust the default configurations of DataKit, which can have varying degrees of additions and deletions. It is best to use the installer program corresponding to version 1.2.3 of DataKit for the installation or upgrade of DataKit 1.2.3.
<!-- markdownlint-enable -->

#### Installation {#simple-install}

> If you are performing an offline install of the lite version of DataKit, you need to specify the installation package with a `_lite` suffix, such as *datakit_lite-linux-amd64-{{.Version}}.tar.gz*.


<!-- markdownlint-disable MD046 -->
=== "Linux"

    To run with root privileges:
    
    ```shell
    chmod +x installer-linux-amd64
    ./installer-linux-amd64 --offline --dataway "https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>" --srcs datakit-linux-amd64-{{ .Version }}.tar.gz,dk_upgrader-linux-amd64-{{ .Version }}.tar.gz,data.tar.gz
    ```

=== "Windows"

    You need to run the Powershell with administrator privileges to execute:
    
    ```powershell
    .\installer-windows-amd64.exe --offline --dataway "https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>" --srcs datakit-windows-amd64-{{.Version}}.tar.gz,dk_upgrader-windows-amd64-{{.Version}}.tar.gz,data.tar.gz
    ```
<!-- markdownlint-enable -->
#### Upgrade {#simple-upgrade}

> If you are performing an offline upgrade of the lite version of DataKit, you need to specify the installation package with a `_lite` suffix, such as `datakit_lite-linux-amd64-{{.Version}}.tar.gz`.

<!-- markdownlint-disable MD046 -->
=== "Linux"

    To run with root privileges:

    ```shell
    chmod +x installer-linux-amd64
    ./installer-linux-amd64 --offline --upgrade --srcs datakit-linux-amd64-{{ .Version }}.tar.gz,data.tar.gz
    ```

=== "Windows"

    You need to run the Powershell with administrator privileges to execute:

    ```powershell
    .\installer-windows-amd64.exe --offline --upgrade --srcs datakit-windows-amd64-{{.Version}}.tar.gz,data.tar.gz
    ```
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ tip "How to Specify More Configuration Parameters for Offline Installation"

    During online installation, we can specify some default parameters through [environment variables `DK_XXX=YYY`](datakit-install.md#extra-envs). These default parameters actually take effect through the *install.sh* script (on Windows, it's *install.ps1*). However, these environment variables are ineffective for the installation program *installer-xxx*. We can only use the command-line arguments of *installer-xxx* to add these options. By using the following command, we can find out the parameters supported by the installation program:

    ```shell
    ./installer-linux-amd64 --help
    ```

    For example, the Dataway address we specified above is set through the `--dataway` option. Additionally, these extra command-line parameter settings are only effective in installation mode and do not take effect in (offline) upgrade mode.

<!-- markdownlint-enable -->

### Advanced Mode {#offline-advanced}

DataKit is currently installed on the public web, and all binary data and installation scripts are downloaded from the static.<<<custom_key.brand_main_domain>>> site. For machines that cannot access the site, you can replace the static.<<<custom_key.brand_main_domain>>> site by deploying a file server on the intranet.

The network traffic topology of advanced mode is as follows:

<figure markdown>
  ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/nginx-file-server.png){ width="700"}
</figure>

Prepare a machine that can be accessed on the intranet, install Nginx on the machine, and download (or copy) the files required for DataKit installation to the Nginx server, so that other machines can download the installation files from the Nginx file server to complete the installation.

- Setting up the Nginx file server {#nginx-config}

Add configuration in nginx.conf

```txt
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
mkdir -p /datakit/apm_lib
wget -P /datakit https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh
wget -P /datakit https://static.<<<custom_key.brand_main_domain>>>/datakit/version
wget -P /datakit https://static.<<<custom_key.brand_main_domain>>>/datakit/data.tar.gz
wget -P /datakit https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-linux-amd64-{{ .Version }}
wget -P /datakit https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-linux-amd64-{{ .Version }}.tar.gz
wget -P /datakit https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit_lite-linux-amd64-{{ .Version }}.tar.gz
wget -P /datakit https://static.<<<custom_key.brand_main_domain>>>/datakit/dk_upgrader-linux-amd64-{{ .Version }}.tar.gz
wget -P /datakit https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-apm-inject-linux-amd64-{{ .Version }}.tar.gz
wget -P /datakit/apm_lib https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar

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
  "/datakit/sourcemap/atosl/atosl-darwin-x64"
  "/datakit/sourcemap/atosl/atosl-darwin-arm64"
  "/datakit/sourcemap/atosl/atosl-linux-x64"
  "/datakit/sourcemap/atosl/atosl-linux-arm64"
)

mkdir -p /datakit/sourcemap/jdk \
  /datakit/sourcemap/R8       \
  /datakit/sourcemap/proguard \
  /datakit/sourcemap/ndk      \
  /datakit/sourcemap/atosl

for((i=0;i<${#sources[@]};i++)); do
  wget https://static.<<<custom_key.brand_main_domain>>>${sources[$i]} -O ${sources[$i]}
done
```

<!-- markdownlint-disable MD046 -->
???+ Attention

    You must append suffix **.exe** to the download link of `Installer` on Windows, for example: [*https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-windows-386-{{.Version}}.exe*](https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-windows-386-{{.Version}}.exe) for Windows 32bit and
    [*https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-windows-amd64-{{.Version}}.exe*](https://static.<<<custom_key.brand_main_domain>>>/datakit/installer-windows-amd64-{{.Version}}.exe) for Windows 64bit.
<!-- markdownlint-enable -->

#### Install {#advance-install}

On the intranet machine, point it to the Nginx file server by setting `DK_INSTALLER_BASE_URL`:

<!-- markdownlint-disable MD046 MD034 -->
=== "Linux/Mac"

    ```shell
{{ InstallCmd 4
(.WithPlatform "unix")
(.WithSourceURL "${DK_INSTALLER_BASE_URL}")
(.WithEnvs "DK_INSTALLER_BASE_URL" "http://[Nginx-Server]:8080/datakit")
(.WithEnvs "HTTPS_PROXY" "http://1.2.3.4:9530")
(.WithProxy true)
}}
    ```

=== "Windows"

    ```powershell
{{ InstallCmd 4
(.WithPlatform "windows")
(.WithSourceURL "${DK_INSTALLER_BASE_URL}")
(.WithEnvs "HTTPS_PROXY" "1.2.3.4:9530")
(.WithEnvs "DK_INSTALLER_BASE_URL" "http://[Nginx-Server]:8080/datakit")
}}
    ```
<!-- markdownlint-enable -->

So far, the offline installation is complete. Note that HTTPS_PROXY is additionally set here.

---

#### Upgrade {#advance-upgrade}

If there is a new version of DataKit, you can download it as above and execute the following command to upgrade:
<!-- markdownlint-disable MD046 -->
=== "Linux/Mac"

    ```shell
    DK_INSTALLER_BASE_URL="http://<nginxServer>:8080/datakit" \
{{ InstallCmd 4 (.WithPlatform "unix") (.WithUpgrade true) (.WithSourceURL "${DK_INSTALLER_BASE_URL}") }}
    ```

=== "Windows"

    ```powershell
    $env:DK_INSTALLER_BASE_URL="http://<nginxServer>:8080/datakit";
{{ InstallCmd 4 (.WithPlatform "windows") (.WithUpgrade true) (.WithSourceURL "${DK_INSTALLER_BASE_URL}") }}
    ```
<!-- markdownlint-enable -->
## Kubernetes Offline Deployment {#k8s-offline}

### Bash Script Assisted Installation {#Auxiliary-installation}

Here is a simple script to help you complete the tasks of password free login, file distribution and image decompression.
<!-- markdownlint-disable MD046 -->
???- note "`datakit_tools.sh` (Stand-alone open)"
    ```shell
    #!/bin/bash
    # Please modify the host IP to be password-free
    host_ip=(
      10.200.14.112
      10.200.14.113
      10.200.14.114
    )
    # Please change the login password
    psd='123.com'

    menu() {
      echo -e "\e[33m------Please select the required operation------\e[0m"
      echo -e "\e[33m1. Set SSH remote keyless login\e[0m"
      echo -e "\e[33m2. Scp remote transfer file\e[0m"
      echo -e "\e[33m3. Remote decompression image\e[0m"
      read -p "Please enter an option:" num
    }

    # Send ssh-copy-id to the host
    SSH-COPY(){
    yum install -y expect
    ssh-keygen -t rsa -P '' -f ~/.ssh/id_rsa
    for i in ${host_ip[@]}
      do
        /usr/bin/expect<<EOF
        spawn ssh-copy-id ${i}
        expect {
              "(yes/no)" {send "yes\r";exp_continue}
              "password" {send "${psd}\r"}
    }
    expect eof
    EOF
    done
    }

    SCP(){
    read  -p "Please enter the file name to transfer (multiple files can be passed in): " file_name
    for i in ${host_ip[@]}
      do
      for j in ${file_name[@]}
        do
          echo -e "\e[33m------${i}---${j}------\e[0m"
          scp ${j} root@${i}:/root/
      done
    done
    }

    SSH(){
    read -p "Please enter the file name to extract: " file_name
    # Remotely unzip image packets
    for i in ${host_ip[@]}
      do
        echo -e "\e[33m------${i}------\e[0m"
        # ssh root@${i} "docker load -i ${file_name}"
        ssh root@${i} " ctr -n=k8s.io image import ${file_name} "
    done
    }
    #menu
    CASE(){
    case ${num} in
    1)
    SSH-COPY
            echo -e "\e[33m------------------------------------------------------\e[0m"
    ;;
    2)
    SCP
            echo -e "\e[33m------------------------------------------------------\e[0m"
    ;;
    3)
    SSH
            echo -e "\e[33m------------------------------------------------------\e[0m"
      ;;
    *)
            
      echo -e "\e[31mPlease enter the number in the option{1|2|3}\e[0m"
    esac
    }

    menu
    CASE
    read -p "Do you want to continue with the list operation? [y/n]：" a
    while [ "${a}" == "y" ]
      do
        menu 
        CASE
        read -p "Do you want to continue with the list operation? [y/n]：" a
        continue
    done
    ```

```shell
# You need to modify the host IP and login password in the script, and then complete the operation according to the guidance.
chmod +x datakit_tools.sh
./datakit_tools.sh
```
<!-- markdownlint-enable -->
### Agent Installation {#k8s-install-via-proxy}

**If there is a machine in the intranet that can connect to the internet, you can deploy a nginx server on this node to use as the image acquisition.**

- Download `datakit.yaml` and DataKit image files

```shell
wget https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit.yaml -P /home/guance/
```

- Download the DataKit image and make it into a package

```shell
# Pull the image of the amd64 architecture and make it into an image package
docker pull --platform amd64 pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:{{.Version}}
docker save -o datakit-amd64-{{.Version}}.tar pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:{{.Version}}
mv datakit-amd64-{{.Version}}.tar /home/guance

# Pull the image of the arm64 architecture and make it into an image package
docker pull --platform arm64 pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:{{.Version}}
docker save -o datakit-arm64-{{.Version}}.tar pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:{{.Version}}
mv datakit-arm64-{{.Version}}.tar /home/guance

# Check whether the image architecture is correct
docker image inspect pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:{{.Version}} |grep Architecture

```

- Modify Nginx configuration agent
<!-- markdownlint-disable MD046 -->
???- note "/etc/nginx/nginx.conf"
    ```shell
    #user  nobody;
    worker_processes  1;

    #error_log  logs/error.log;
    #error_log  logs/error.log  notice;
    #error_log  logs/error.log  info;

    #pid        logs/nginx.pid;


    events {
        worker_connections  1024;
    }


    http {
        include       mime.types;
        default_type  application/octet-stream;

        #log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
        #                  '$status $body_bytes_sent "$http_referer" '
        #                  '"$http_user_agent" "$http_x_forwarded_for"';

        #access_log  logs/access.log  main;

        sendfile        on;
        #tcp_nopush     on;

        #keepalive_timeout  0;
        keepalive_timeout  65;

        #gzip  on;

        server {
            listen       8080;
            server_name  localhost;
            root /home/guance;
            autoindex on;

            #charset koi8-r;

            #access_log  logs/host.access.log  main;


            #error_page  404              /404.html;

            # redirect server error pages to the static page /50x.html
            #

            # proxy the PHP scripts to Apache listening on 127.0.0.1:80
            #
            #location ~ \.php$ {
            #    proxy_pass   http://127.0.0.1;
            #}

            # pass the PHP scripts to FastCGI server listening on 127.0.0.1:9000
            #
            #location ~ \.php$ {
            #    root           html;
            #    fastcgi_pass   127.0.0.1:9000;
            #    fastcgi_index  index.php;
            #    fastcgi_param  SCRIPT_FILENAME  /scripts$fastcgi_script_name;
            #    include        fastcgi_params;
            #}

            # deny access to .htaccess files, if Apache's document root
            # concurs with nginx's one
            #
            #location ~ /\.ht {
            #    deny  all;
            #}
        }


        # another virtual host using mix of IP-, name-, and port-based configuration
        #
        #server {
        #    listen       8000;
        #    listen       somename:8080;
        #    server_name  somename  alias  another.alias;

        #    location / {
        #        root   html;
        #        index  index.html index.htm;
        #    }
        #}


        # HTTPS server
        #
        #server {
        #    listen       443 ssl;
        #    server_name  localhost;

        #    ssl_certificate      cert.pem;
        #    ssl_certificate_key  cert.key;

        #    ssl_session_cache    shared:SSL:1m;
        #    ssl_session_timeout  5m;

        #    ssl_ciphers  HIGH:!aNULL:!MD5;
        #    ssl_prefer_server_ciphers  on;

        #    location / {
        #        root   html;
        #        index  index.html index.htm;
        #    }
        #}

    }
    ```
<!-- markdownlint-enable -->
- Other intranet machines execute commands.

```shell
wget http://<nginx-server-ip>:8080/datakit.yaml 
wget http://<nginx-server-ip>:8080/datakit-amd64-{{.Version}}.tar 
```

- Unzip image command

```shell
# docker 
docker load -i /k8sdata/datakit/datakit-amd64-{{.Version}}.tar

# containerd
ctr -n=k8s.io image import /k8sdata/datakit/datakit-amd64-{{.Version}}.tar

```

- Start DataKit container

```shell
kubectl apply -f datakit.yaml
```

### Full Offline Installation {#k8s-offilne-all}

When there is no external network in the environment, the installation package needs be downloaded from the public network to the internal network through mobile hard disk (U disk).

- Unzip image command

```shell
# docker 
docker load -i datakit-amd64-{{.Version}}.tar

# containerd
ctr -n=k8s.io image import datakit-amd64-{{.Version}}.tar

```

- The cluster controller executes the start command

```shell
kubectl apply -f datakit.yaml
```

- Update command

```shell
# You need to decompress the image first
kubectl patch -n datakit daemonsets.apps datakit -p '{"spec": {"template": {"spec": {"containers": [{"image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:<version>","name": "datakit"}]}}}}'
```
