
# Host Installation
---

This article describes the basic installation of DataKit.

## Register/log in to Guance Cloud {#login-guance}

The browser visits the [Guance Cloud registration](https://auth.guance.com/redirectpage/register){:target="_blank"} portal, fills in the corresponding information, and then [logs in to Guance Cloud](https://console.guance.com/pageloading/login){:target="_blank"}.

## Get the Installation Command {#get-install}

Log in to the workspace, click "Integration" on the left and select "Datakit" at the top, and you can see the installation commands of various platforms.

> Note that the following Linux/Mac/Windows installer can automatically identify the hardware platform (arm/x86, 32bit/64bit) without making a hardware platform selection.

=== "Linux"

    The order is roughly as follows:
    
    ```shell
    DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    ```
    
    After the installation is completed, you will see a prompt that the installation is successful at the terminal.

=== "Mac"

    The Mac installation command is basically the same as Linux:
    
    ```shell
    DK_DATAWAY=https://openway.guance.com?token=<TOKEN> bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    ```
    
    After the installation is completed, you will see a prompt that the installation is successful at the terminal.

=== "Windows"

    Installation on Windows requires a Powershell command line installation and must run Powershell as an administrator. Press the Windows key, enter powershell to see the pop-up powershell icon, and right-click and select "Run as an administrator".
    
    ```powershell
    $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
    ```

???+ attention "Mac Installation Problems"

    If it appears during the installation/upgrade process when installing on Mac:
    
    ```shell
    "launchctl" failed with stderr: /Library/LaunchDaemons/cn.dataflux.datakit.plist: Service is disabled
    # or
    "launchctl" failed with stderr: /Library/LaunchDaemons/com.guance.datakit.plist: Service is disabled
    ```
    
    execute
    
    ```shell
    sudo launchctl enable system/datakit
    ```
    
    Then execute the following command
    
    ```shell
    sudo launchctl load -w /Library/LaunchDaemons/cn.dataflux.datakit.plist
    # or。/
    sudo launchctl load -w /Library/LaunchDaemons/com.guance.datakit.plist
    ```

## Additional Supported Installation Variable {#extra-envs}

If you need to define some DataKit configuration during the installation phase, you can add environment variables to the installation command, just append them before `DK_DATAWAY` For example, append the `DK_NAMESPACE` setting:

```shell
# Linux/Mac
DK_NAMESPACE="<namespace>" DK_DATAWAY="https://openway.guance.com?token=<TOKEN>" bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"

# Windows
$env:DK_NAMESPACE="<namespace>"; $env:DK_DATAWAY="https://openway.guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

The setting format of the two environment variables is:

```shell
# Windows: Multiple environment variables are divided by semicolons
$env:NAME1="value1"; $env:Name2="value2"

# Linux/Mac: Multiple environment variables are divided by spaces
NAME1="value1" NAME2="value2"
```

The environment variables supported by the installation script are as follows (supported by the whole platform).

???+ attention

    These environment variable settings are not supported for [full offline installation](datakit-offline-install.md#offline). However, these environment variables can be set by [proxy](datakit-offline-install.md#with-datakit) and [setting local installation address](datakit-offline-install.md#with-nginx).

### Most Commonly Used Environment Variables {#common-envs}

- `DK_DATAWAY`: Specify the DataWay address, and the DataKit installation command has been brought by default
- `DK_GLOBAL_TAGS`: Deprecated, DK_GLOBAL_HOST_TAGS instead
- `DK_GLOBAL_HOST_TAGS`: Support the installation phase to fill in the global host tag, format example: `host=__datakit_hostname,host_ip=__datakit_ip` (multiple tags are separated by English commas)
- `DK_GLOBAL_ELECTION_TAGS`: Support filling in the global election tag during the installation phase，format example: `project=my-porject,cluster=my-cluster` (support filling in the global election tag during the installation phase)
- `DK_DEF_INPUTS`: List of collector names opened by default, format example: `cpu,mem,disk`
  - Since [many collectors will be turned on by default](datakit-input-conf.md#default-enabled-inputs), this environment variable is used to adjust the default collector list. For example, if you can choose to turn on only three collectors: cpu, mem and disk, pass in `DK_DEF_INPUTS="cpu,mem,disk"`.
- `DK_CLOUD_PROVIDER`: Support filling in cloud vendors during installation (`aliyun/aws/tencent/hwcloud/azure`)

### On DataKit's Own Log  {#env-logging}

- `DK_LOG_LEVEL`: Optional info/debug
- `DK_LOG`: If changed to stdout, the log will not be written to the file, but will be output by the terminal.
- `DK_GIN_LOG`: If changed to stdout, the log will not be written to the file, but will be output by the terminal.

### On DataKit pprof  {#env-pprof}

- `DK_ENABLE_PPROF`: whether to turn on `pprof`
- `DK_PPROF_LISTEN`: `pprof` service listening address

### On DataKit Election  {#env-election}

- `DK_ENABLE_ELECTION`: Open the election, not by default. If you need to open it, give any non-empty string value to the environment variable. (eg `True`/`False`)
- `DK_NAMESPACE`: Supports namespaces specified during installation (for election)

### On HTTP/API  Environment {#env-http-api}
- `DK_HTTP_LISTEN`: Support the installation-stage specified DataKit HTTP service binding network card (default `localhost`)
- `DK_HTTP_PORT`: Support specifying the port of the DataKit HTTP service binding during installation (default `9529`)
- `DK_RUM_ORIGIN_IP_HEADER`: RUM-specific
- `DK_DISABLE_404PAGE`: Disable the DataKit 404 page (commonly used when deploying DataKit RUM on the public network. Such as `True`/`False`)
- `DK_INSTALL_IPDB`: Specify the IP library at installation time (currently only `iploc` and `geolite2` is supported)

### On DCA  {#env-dca}
- `DK_DCA_ENABLE`: Support DCA service to be turned on during installation (not turned on by default)
- `DK_DCA_LISTEN`: Support custom configuration of DCA service listening addresses and ports during installation (default `0.0.0.0:9531`）
- `DK_DCA_WHITE_LIST`: Support setup of DCA service access whitelist, multiple whitelists split (e.g. `192.168.0.1/24,10.10.0.1/24`)

### On External Collector  {#env-external-inputs}
- `DK_INSTALL_EXTERNALS`: Used to install external collectors not packaged with DataKit

### On Confd Configuration  {#env-connfd}

| Environment Variable Name                 | Type   | Applicable Scenario            | Description     | Sample Value |
| ----                     | ----   | ----               | ----     | ---- |
| DK_CONFD_BACKEND        | string |  All              | Backend Source Type  | `etcdv3`, `zookeeper`, `redis` or `consul` |
| DK_CONFD_BASIC_AUTH     | string | `etcdv3`, `consul` | Optional      | |
| DK_CONFD_CLIENT_CA_KEYS | string | `etcdv3`, `consul` | Optional      | |
| DK_CONFD_CLIENT_CERT    | string | `etcdv3`, `consul` | Optional      | |
| DK_CONFD_CLIENT_KEY     | string | `etcdv3`, `consul` or `redis` | Optional      | |
| DK_CONFD_BACKEND_NODES  | string |  All              | Backend Source Address | `[IP地址:2379,IP address 2:2379]` |
| DK_CONFD_PASSWORD       | string | `etcdv3`, `consul` | Optional      |  |
| DK_CONFD_SCHEME         | string | `etcdv3`, `consul` | Optional      |  |
| DK_CONFD_SEPARATOR      | string | `redis`            | Optional default 0 |  |
| DK_CONFD_USERNAME       | string | `etcdv3`, `consul` | Optional      |  |

### On Git Configuration {#env-gitrepo}

- `DK_GIT_URL`: The remote git repo address for managing configuration files. (e.g. `http://username:password@github.com/username/repository.git`）
- `DK_GIT_KEY_PATH`: The full path of the local PrivateKey. (e.g.  `/Users/username/.ssh/id_rsa`）
- `DK_GIT_KEY_PW`: The password to use the local PrivateKey. (e.g.  `passwd`）
- `DK_GIT_BRANCH`: Specify the branch to pull. <stong>If it is empty, it is the default</strong>, and the default is the remotely specified main branch, which is usually `master`.
- `DK_GIT_INTERVAL`: The interval of the timed pull. (e.g. `1m`)

### On Sinker Configuration {#env-sink}

- `DK_SINKER`: Used to setup Dataway sinker, it's a JSON string, please refer to [here](datakit-daemonset-deploy.md#env-sinker) for more info.

See [M3DB example](datakit-sink-m3db.md)

### On Cgroup Configuration {#env-cgroup}

The following installation options are supported only on Linux platforms:

- `DK_CGROUP_DISABLED`: Turn off Cgroup function on Linux system (on by default)
- `DK_LIMIT_CPUMAX`: Maximum CPU power supported on Linux system, default 30.0
- `DK_LIMIT_CPUMIN`: Minimum CPU power supported on Linux system, default 5.0
- `DK_LIMIT_MEMMAX`: Limit memory (including swap) on Linux, default 4096 (4GB)

### Other Installation Options {#env-others}

- `DK_INSTALL_ONLY`: Install only, not run
- `DK_HOSTNAME`: Support custom configuration hostname during installation
- `DK_UPGRADE`: Upgrade to the latest version (Note: Once this option is turned on, all other options are invalid)
- `DK_INSTALLER_BASE_URL`: You can choose the installation script for different environments, default to `https://static.guance.com/datakit`
- `DK_PROXY_TYPE`: Proxy type. The options are: "datakit" or "nginx", both lowercase
- `DK_NGINX_IP`: Proxy server IP address (only need to fill in IP but not port). With highest priority, this is mutually exclusive with the above "HTTP_PROXY" and "HTTPS_PROXY" and will override both.
- `DK_INSTALL_LOG`: Set the setup log path, default to *install.log* in the current directory, if set to `stdout`, output to the command line terminal.
- `HTTPS_PROXY`: Installed through the Datakit agent

## FAQ {#faq}

### :material-chat-question: How to Deal with the Unfriendly Host Name {#bad-hostname}

Because DataKit uses Hostname as the basis for data concatenation, in some cases, some host names are not very friendly, such as  `iZbp141ahn....`, but for some reasons, these host names cannot be modified, which brings some troubles to use. In DataKit, this unfriendly host name can be overwritten in the main configuration.

In `datakit.conf`, modify the following configuration and the DataKit will read `ENV_HOSTNAME` to overwrite the current real hostname:

```toml
[environments]
	ENV_HOSTNAME = "your-fake-hostname-for-datakit"
```

> Note: If a host has collected data for a period of time, after changing the host name, the historical data will no longer be associated with the new host name. Changing the host name is equivalent to adding a brand-new host.

## :material-chat-question: More Readings {#more-reading}

- [Getting started with DataKit](datakit-service-how-to.md)
