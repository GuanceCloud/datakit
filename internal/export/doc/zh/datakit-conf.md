
# DataKit 主配置
---

DataKit 主配置用来配置 DataKit 自己的运行行为。

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    其目录一般位于：
    
    - Linux/Mac: `/usr/local/datakit/conf.d/datakit.conf`
    - Windows: `C:\Program Files\datakit\conf.d\datakit.conf`

=== "Kubernetes"

    DaemonSet 安装时，虽然在对应目录下也存在这个文件，**但实际上 DataKit 并不加载这里的配置**。这些配是通过在 *datakit.yaml* 中[注入环境变量](datakit-daemonset-deploy.md#using-k8-env)来生成的。下面所有的配置，都能在 Kubernetes 部署文档中找到[对应的环境变量](datakit-daemonset-deploy.md#using-k8-env)配置。
<!-- markdownlint-enable -->

## Datakit 主配置示例 {#maincfg-example}

Datakit 主配置示例如下，我们可以根据该示例来开启各种功能（当前版本 {{ .Version }}）：

<!-- markdownlint-disable MD046 -->
??? info "*datakit.conf*"

    ```toml linenums="1"
    {{ CodeBlock .DatakitConfSample 4 }}
    ```
<!-- markdownlint-enable -->

## HTTP 服务的配置 {#config-http-server}

DataKit 会开启 HTTP 服务，用来接收外部数据，或者对外提供基础的数据服务。

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    ### 修改 HTTP 服务地址 {#update-http-server-host}
    
    默认的 HTTP 服务地址是 `localhost:9529`，如果 9529 端口被占用，或希望从外部访问 DataKit 的 HTTP 服务（比如希望接收 [RUM](../integrations/rum.md) 或 [Tracing](../integrations/datakit-tracing.md) 数据），可将其修改成：
    
    ```toml
    [http_api]
       listen = "0.0.0.0:<other-port>"
       # 或使用 IPV6 地址
       # listen = "[::]:<other-port>"
    ```

    注意，IPv6 支持需 [Datakit 升级到 1.5.7](changelog.md#cl-1.5.7-new)。

    #### 使用 Unix domain socket {#uds}

    Datakit 支持 UNIX domain sockets 访问。开启方式如下：`listen` 字段配置为<b>一个不存在文件的全路径</b>，这里以 `datakit.sock` 举例，可以为任意文件名。
    ```toml
    [http_api]
       listen = "/tmp/datakit.sock"
    ```
    配置完成后可以使用 `curl` 命令测试是否配置成功：`sudo curl --no-buffer -XGET --unix-socket /tmp/datakit.sock http:/localhost/v1/ping`。更多关于 `curl` 的测试命令的信息可以参阅[这里](https://superuser.com/a/925610){:target="_blank"}。
    
    ### HTTP 请求频率控制 {#set-http-api-limit}
    
    由于 DataKit 需要大量接收外部数据写入，为了避免给所在节点造成巨大开销，可修改如下 HTTP 配置（默认不开启）：
    
    ```toml
    [http_api]
      request_rate_limit = 1000.0 # 限制每个 HTTP API 每秒只接收 1000 次请求
    ```

    ### 其它设置 {#http-other-settings}

    ```toml
    [http_api]
        close_idle_connection = true # 关闭闲置连接
        timeout = "30s"              # 设置服务端 HTTP 超时
    ```

=== "Kubernetes"

    参见[这里](datakit-daemonset-deploy.md#env-http-api)
<!-- markdownlint-enable -->

## 全局标签（Tag）修改 {#set-global-tag}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

Datakit 允许给其采集的所有数据配置全局标签，全局标签分为两类：

- 主机类全局标签：采集的数据跟当前主机绑定，比如 CPU/内存等指标数据
- 选举类全局标签：采集的数据来自某个公共（远程）实体，比如 MySQL/Redis 等，这些采集一般都参与选举，故这些数据上不会带上当前主机相关的标签

```toml
[global_host_tags] # 这里面的我们称之为「全局主机标签」：GHT
  ip   = "__datakit_ip"
  host = "__datakit_hostname"

[election]
  [election.tags] # 这里面的我们称之为「全局选举标签」：GET
    project = "my-project"
    cluster = "my-cluster"
```

加全局标签时，有几个地方要注意：

1. 这些全局标签的值可以用 Datakit 目前已经支持的几个通配（双下划线（`__`）前缀和 `$` 都是可以的）：

    1. `__datakit_ip/$datakit_ip`：标签值会设置成 DataKit 获取到的第一个主网卡 IP
    1. `__datakit_hostname/$datakit_hostname`：标签值会设置成 DataKit 的主机名

1. 由于 [DataKit 数据传输协议限制](apis.md#lineproto-limitation)，不要在全局标签（Tag）中出现任何指标（Field）字段，否则会因为违反协议导致数据处理失败。具体参见具体采集器的字段列表。当然，也不要加太多标签，而且每个标签的 Key 以及 Value 长度都有限制。
1. 如果被采集上来的数据中，本来就带有同名的标签，那么 DataKit 不会再追加这里配置的全局标签
1. 即使 GET 中没有任何配置，DataKit 仍然会在所有数据上尝试添加一个 `host=__datakit_hostname` 的标签
1. 这俩类全局标签（GHT/GET）是可以有交集的，比如都可以在其中设置一个 `project = "my-project"` 的标签
1. 当没有开启选举的情况下，GET 沿用 GHT（它至少有一个 `host` 的标签）中的所有标签
1. 选举类采集器默认追加 GET，非选举类采集器默认追加 GHT。

<!-- markdownlint-disable MD046 -->
???+ tip "如何区分选举和非选举采集器？"

    在采集器文档中，在顶部有类似如下标识，它们表示当前采集器的平台适配情况以及采集特性：

    :fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:  · :fontawesome-solid-flag-checkered:

    若带有 :fontawesome-solid-flag-checkered: 则表示当前采集器是选举类采集器。
<!-- markdownlint-enable -->

### 全局 Tag 在远程采集时的设置 {#notice-global-tags}

因为 DataKit 会默认给采集到的所有数据追加标签 `host=<DataKit 所在主机名>`，但某些情况这个默认追加的 `host` 会带来困扰。

以 MySQL 为例，如果 MySQL 不在 DataKit 所在机器，但又希望这个 `host` 标签是被采集的 MySQL 的真实主机名（或云数据库的其它标识字段），而非 DataKit 所在的主机名。

对这种情况，我们有两种方式可以绕过 DataKit 上的全局 tag：

- 在具体采集器中，一般都有一个如下配置，我们可以在这里面新增 Tag，比如，如果不希望 DataKit 默认添加 `host=xxx` 这个 Tag，可以在这里覆盖这个 Tag，以 MySQL 为例：

```toml
[[inputs.mysql.tags]]
  host = "real-mysql-host-name" 
```

- 以 [HTTP API 方式往 DataKit 推送数据](apis.md#api-v1-write)时，可以通过 API 参数 `ignore_global_tags` 来屏蔽所有全局 Tag

<!-- markdownlint-disable MD046 -->
???+ tip

    自 [1.4.20](changelog.md#cl-1.4.20) 之后，DataKit 默认会以被采集服务连接地址中的的 IP/Host 作为 `host` 的标签值。
<!-- markdownlint-enable -->

## DataKit 自身运行日志配置 {#logging-config}

DataKit 自身日志有两个，一个是自身运行日志（*/var/log/datakit/log*），一个是 HTTP Access 日志（*/var/log/datakit/gin.log*）。

DataKit 默认日志等级为 `info`。编辑 `datakit.conf`，可修改日志等级以及分片大小：

```toml
[logging]
  level = "debug" # 将 info 改成 debug
  rotate = 32     # 每个日志分片为 32MB
```

- `level`：置为 `debug` 后，即可看到更多日志（目前只支持 `debug/info` 两个级别）。
- `rotate`：DataKit 默认会对日志进行分片，默认分片大小为 32MB，总共 6 个分片（1 个当前写入分片加上 5 个切割分片，分片个数尚不支持配置）。如果嫌弃 DataKit 日志占用太多磁盘空间（最多 32 x 6 = 192MB），可减少 `rotate` 大小（比如改成 4，单位为 MB）。HTTP 访问日志也按照同样的方式自动切割。

## 高级配置 {#advance-config}

下面涉及的内容涉及一些高级配置，如果对配置不是很有把握，建议咨询我们的技术专家。

### Point 缓存 {#point-pool}

[:octicons-tag-24: Version-1.28.0](changelog.md#cl-1.28.0) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

为了优化 Datakit 高负载情况下的内存占用，可以开启 Point Pool 来缓解：

```toml
# datakit.conf
[point_pool]
    enable = true
    reserved_capacity = 4096
```

同时，[Datakit 配置](datakit-conf.md#dataway-settings)中可以开启 `content_encoding = "v2"` 的传输编码（[:octicons-tag-24: Version-1.32.0](changelog.md#cl-1.32.0) 已默认启用 v2），相比 v1，它的内存和 CPU 开销都更低。

<!-- markdownlint-disable MD046 -->
???+ attention

    在低负载（Datakit 内存占用 100MB 左右）的情况下，开启 Point-Pool 会增加 Datakit 自身的内存占用，但也不至于太多。所谓的高负载，一般指占用内存在 2GB+ 的场景。同时开启后也能改善 Datakit 自身的 CPU 消耗。
<!-- markdownlint-enable -->

### IO 模块调参 {#io-tuning}

[:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    某些情况下，DataKit 的单机数据采集量非常大，如果网络带宽有限，可能导致部分数据的采集中断或丢弃。可以通过配置 io 模块的一些参数来缓解这一问题：

    ```toml
    [io]
      feed_chan_size  = 4096  # 数据处理队列（一个 job 一般都有多个 point）长度
      max_cache_count = 512   # 数据批量发送点数的阈值，缓存中超过该值即触发发送
      flush_interval  = "10s" # 数据发送的间隔阈值，每隔 10s 至少发送一次
      flush_workers   = 8     # 数据上传 worker 数（默认 CPU-core * 2 + 1）
    ```

    阻塞模式参见 [k8s 中的对应说明](datakit-daemonset-deploy.md#env-io)

=== "Kubernetes"

    参见[这里](datakit-daemonset-deploy.md#env-io)
<!-- markdownlint-enable -->

#### IO 磁盘缓存 {#io-disk-cache}

[:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8) · [:octicons-beaker-24: Experimental](index.md#experimental)

当 DataKit 发送数据失败后，为了不丢失关键数据，可以开启磁盘缓存。磁盘缓存的目的在于将发送失败的数据暂时存入磁盘，待条件允许时，再将数据发送出去。

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    ```toml
    [io]
      enable_cache      = true   # 开启磁盘缓存
      cache_all         = false  # 是否全类缓存（默认情况下，指标/对象/拨测数据不缓存）
      cache_max_size_gb = 5      # 指定每个分类磁盘大小为 5GB
    ```

=== "Kubernetes"

    参见[这里](datakit-daemonset-deploy.md#env-io)

---

???+ attention

    这里的 `cache_max_size_gb` 指每个分类（Category）的缓存大小，总共 10 个分类的话，如果每个指定 5GB，理论上会占用 50GB 左右的空间。
<!-- markdownlint-enable -->

### 资源限制  {#resource-limit}

由于 DataKit 上处理的数据量无法估计，如果不对 DataKit 消耗的资源做物理限制，将有可能消耗所在节点大量资源。这里我们可以借助 Linux 的 cgroup 和 Windows 的 job object 来限制，在 *datakit.conf* 中有如下配置：

```toml
[resource_limit]
  path = "/datakit" # Linux cgroup 限制目录，如 /sys/fs/cgroup/memory/datakit, /sys/fs/cgroup/cpu/datakit

  # 允许 CPU 最大使用率（百分制）
  cpu_max = 20.0

  # 默认允许 4GB 内存(memory + swap)占用
  # 如果置为 0 或负数，则不启用内存限制
  mem_max_mb = 4096 
```

如果 DataKit 超出内存限制后，会被操作系统强制杀掉，通过命令可以看到如下结果，此时需要[手动启动服务](datakit-service-how-to.md#when-service-failed)：

```shell
$ systemctl status datakit 
● datakit.service - Collects data and upload it to DataFlux.
     Loaded: loaded (/etc/systemd/system/datakit.service; enabled; vendor preset: enabled)
     Active: activating (auto-restart) (Result: signal) since Fri 2022-02-30 16:39:25 CST; 1min 40s ago
    Process: 3474282 ExecStart=/usr/local/datakit/datakit (code=killed, signal=KILL)
   Main PID: 3474282 (code=killed, signal=KILL)
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - 资源限制只在[宿主机安装](datakit-install.md)的时候会默认开启
    - 只支持 CPU 使用率和内存使用量（mem+swap）控制，且只支持 Linux 和 windows ([:octicons-tag-24: Version-1.15.0](changelog.md#cl-1.15.0)) 操作系统。
    - CPU 使用率控制目前不支持这些 windows 操作系统： Windows 7, Windows Server 2008 R2, Windows Server 2008, Windows Vista, Windows Server 2003 和 Windows XP。
    - 非 root 用户改资源限制配置时，必须重装 service。
    - CPU 核心数限制会影响 Datakit 部分子模块的 worker 数配置（一般是 CPU 核心数的整数倍）。比如数据上传 worker 就是 CPU 核心数 * 2。而单个上传 worker 会占用默认 10MB 的内存用于数据发送，故 CPU 核心数如果开放较多，会影响 Datakit 整体内存的占用

???+ tip

    Datakit 自 [1.5.8](changelog.md#cl-1.5.8) 开始支持 cgroup v2。如果不确定 cgroup 版本，可通过命令 `mount | grep cgroup` 来确认。
<!-- markdownlint-enable -->

#### Datakit 用量计量标准 {#dk-usage-count}

[:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)

为了规范 Datakit 用量统计，现对 Datakit 的逻辑计量方法进行如下说明：

- 如果没有开启以下这些采集器，则 Datakit 逻辑计量个数为 1
- 如果 Datakit 运行时长（中间断档不超过 30 分钟）超过 12 小时，则参与计量，否则不参与计量
- 对于以下开启的采集器，按照 Datakit [当前配置的 CPU 核心数](datakit-conf.md#resource-limit)进行计量，最小值为 1，最大值为物理 CPU 核数 [^1]，小数点按照四舍五入规则取整：
    - [RUM 采集器](../integrations/rum.md)
    - 通过 [TCP/UDP 收取日志数据的采集器](../integrations/logging.md##socket)
    - 通过 [kafkamq 采集器](../integrations/kafkamq.md)同步日志/指标/RUM 等数据的采集器
    - 通过 [prom_remote_write 采集器](../integrations/prom_remote_write.md)同步 Prometheus 指标的采集器
    - 通过 [beats_output](../integrations/beats_output.md) 同步日志数据的采集器

通过上述规则，可以更加合理地反映 Datakit 的实际使用情况，为用户提供更加透明、公平的计费方式。

### 选举配置 {#election}

参见[这里](election.md#config)

### DataWay 参数配置 {#dataway-settings}

Dataway 部分有如下几个配置可以配置，其它部分不建议改动：

- `timeout`：上传观测云的超时时间，默认 30s
- `max_retry_count`：设置 Dataway 发送的重试次数（默认 4 次）[:octicons-tag-24: Version-1.17.0](changelog.md#cl-1.17.0)
- `retry_delay`：设置重试间隔基础步长，默认 200ms。所谓基础步长，即第一次 200ms，第二次 400ms，第三次 800ms，以此类推（以 $2^n$ 递增）[:octicons-tag-24: Version-1.17.0](changelog.md#cl-1.17.0)
- `max_raw_body_size`：控制单个上传包的最大大小（压缩前），单位字节 [:octicons-tag-24: Version-1.17.1](changelog.md#cl-1.17.1)
- `content_encoding`：可选择 v1 或 v2 [:octicons-tag-24: Version-1.17.1](changelog.md#cl-1.17.1)
    - v1 即行协议（默认 v1）
    - v2 即 Protobuf 协议，相比 v1，它各方面的性能都更优越。运行稳定后，后续将默认采用 v2

Kubernetes 下部署相关配置参见[这里](datakit-daemonset-deploy.md#env-dataway)。

### Sinker 配置 {#dataway-sink}

参见[这里](../deployment/dataway-sink.md)

### 使用 Git 管理 DataKit 配置 {#using-gitrepo}

参见[这里](git-config-how-to.md)

### 设置打开的文件描述符的最大值 {#enable-max-fd}

Linux 环境下，可以在 Datakit 主配置文件中配置 `ulimit` 项，以设置 Datakit 的最大可打开文件数，如下：

```toml
ulimit = 64000
```

ulimit 默认配置为 64000。在 Kubernetes 中，通过[设置 `ENV_ULIMIT`](datakit-daemonset-deploy.md#env-others) 即可。

### :material-chat-question: 资源限制 CPU 使用率说明 {#cgroup-how}

CPU 使用率是百分比制（最大值 100.0），以一个 8 核心的 CPU 为例，如果限额 `cpu_max` 为 20.0（即 20%），则 DataKit 最大的 CPU 消耗，在 top 命令上将显示为 160% 左右。

### 采集器密码保护 {#secrets_management}

[:octicons-tag-24: Version-1.31.0](changelog.md#cl-1.31.0)


如果您希望避免在配置文件中以明文存储密码，则可以使用该功能。

DataKit 在启动加载采集器配置文件时遇到 `ENC[]` 时会在文件、env、或者 AES 加密得到密码后替换文本并重新加载到内存中，以得到正确的密码。

ENC 目前支持三种方式：

- 文件形式（推荐）：

    配置文件中密码格式： ENC[file:///path/to/enc4dk] ，在对应的文件中填写正确的密码即可。

- AES 加密方式。

    需要在主配置文件 `datakit.conf`  中配置秘钥： crypto_AES_key 或者 crypto_AES_Key_filePath, 秘钥长度是 16 位。
    密码处的填写格式为： `ENC[aes://5w1UiRjWuVk53k96WfqEaGUYJ/Oje7zr8xmBeGa3ugI=]`


接下来以 `mysql` 为例，说明两种方式如何配置使用：

1 文件形式

首先，将明文密码放到文件 `/usr/local/datakit/enc4mysql` 中，然后修改配置文件 mysql.conf:

```toml
# 部分配置
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "ENC[file:///usr/local/datakit/enc4mysql]"
  port = 3306
  # sock = "<SOCK>"
  # charset = "utf8"
```

DK 会从 `/usr/local/datakit/enc4mysql` 中读取密码并替换密码，替换后为 `pass = "Hello*******"`

2 AES 加密方式

首先在 `datakit.conf` 中配置秘钥：

```toml
# crypto key or key filePath.
[crypto]
  # 配置秘钥
  aes_key = "0123456789abcdef"
  # 或者，将秘钥放到文件中并在此配置文件位置。
  aes_Key_file = "/usr/local/datakit/mykey"
```

`mysql.conf` 配置文件：

```toml
pass = "ENC[aes://5w1UiRjWuVk53k96WfqEaGUYJ/Oje7zr8xmBeGa3ugI=]"
```

注意，通过 `AES` 加密得到的密文需要完整的填入。以下是代码示例：
<!-- markdownlint-disable MD046 -->
=== "Golang"

    ```go
    // AESEncrypt  加密。
    func AESEncrypt(key []byte, plaintext string) (string, error) {
        block, err := aes.NewCipher(key)
        if err != nil {
            return "", err
        }
    
        // PKCS7 padding
        padding := aes.BlockSize - len(plaintext)%aes.BlockSize
        padtext := bytes.Repeat([]byte{byte(padding)}, padding)
        plaintext += string(padtext)
        ciphertext := make([]byte, aes.BlockSize+len(plaintext))
        iv := ciphertext[:aes.BlockSize]
        if _, err := io.ReadFull(rand.Reader, iv); err != nil {
            return "", err
        }
        mode := cipher.NewCBCEncrypter(block, iv)
        mode.CryptBlocks(ciphertext[aes.BlockSize:], []byte(plaintext))
    
        return base64.StdEncoding.EncodeToString(ciphertext), nil
    }
    
    // AESDecrypt AES  解密。
    func AESDecrypt(key []byte, cryptoText string) (string, error) {
        ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
        if err != nil {
            return "", err
        }
    
        block, err := aes.NewCipher(key)
        if err != nil {
            return "", err
        }
    
        if len(ciphertext) < aes.BlockSize {
            return "", fmt.Errorf("ciphertext too short")
        }
    
        iv := ciphertext[:aes.BlockSize]
        ciphertext = ciphertext[aes.BlockSize:]
    
        mode := cipher.NewCBCDecrypter(block, iv)
        mode.CryptBlocks(ciphertext, ciphertext)
    
        // Remove PKCS7 padding
        padding := int(ciphertext[len(ciphertext)-1])
        if padding > aes.BlockSize {
            return "", fmt.Errorf("invalid padding")
        }
        ciphertext = ciphertext[:len(ciphertext)-padding]
    
        return string(ciphertext), nil
    }
    ```

=== "Java"

    ```java
    import javax.crypto.Cipher;
    import javax.crypto.spec.IvParameterSpec;
    import javax.crypto.spec.SecretKeySpec;
    import java.security.SecureRandom;
    import java.util.Base64;
    
    public class AESUtils {
        public static String AESEncrypt(byte[] key, String plaintext) throws Exception {
            javax.crypto.Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
            SecretKeySpec secretKeySpec = new SecretKeySpec(key, "AES");
    
            SecureRandom random = new SecureRandom();
            byte[] iv = new byte[16];
            random.nextBytes(iv);
            IvParameterSpec ivParameterSpec = new IvParameterSpec(iv);
            cipher.init(Cipher.ENCRYPT_MODE, secretKeySpec, ivParameterSpec);
            byte[] encrypted = cipher.doFinal(plaintext.getBytes());
            byte[] ivAndEncrypted = new byte[iv.length + encrypted.length];
            System.arraycopy(iv, 0, ivAndEncrypted, 0, iv.length);
            System.arraycopy(encrypted, 0, ivAndEncrypted, iv.length, encrypted.length);
    
            return Base64.getEncoder().encodeToString(ivAndEncrypted);
        }
    
        public static String AESDecrypt(byte[] key, String cryptoText) throws Exception {
            byte[] ciphertext = Base64.getDecoder().decode(cryptoText);
    
            SecretKeySpec secretKeySpec = new SecretKeySpec(key, "AES");
    
            if (ciphertext.length < 16) {
                throw new Exception("ciphertext too short");
            }
    
            byte[] iv = new byte[16];
            System.arraycopy(ciphertext, 0, iv, 0, 16);
            byte[] encrypted = new byte[ciphertext.length - 16];
            System.arraycopy(ciphertext, 16, encrypted, 0, ciphertext.length - 16);
    
            Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
            IvParameterSpec ivParameterSpec = new IvParameterSpec(iv);
            cipher.init(Cipher.DECRYPT_MODE, secretKeySpec, ivParameterSpec);
    
            byte[] decrypted = cipher.doFinal(encrypted);
    
            return new String(decrypted);
        }
    }
    public static void main(String[] args) {
        try {
            String key = "0123456789abcdef"; // 16, 24, or 32 bytes AES key
            String plaintext = "HelloAES9*&.";
            byte[] keyBytes = key.getBytes("UTF-8");

            String encrypted = AESEncrypt(keyBytes, plaintext);
            System.out.println("Encrypted text: " + encrypted);

            String decrypt = AESDecrypt(keyBytes, encrypted);
            System.out.println("解码后的是："+decrypt);
        } catch (Exception e) {
            System.out.println(e);
            e.printStackTrace();
        }
    }
    ```
<!-- markdownlint-enable -->

K8S 环境下可以通过环境变量方式添加私钥：`ENV_CRYPTO_AES_KEY` 和 `ENV_CRYPTO_AES_KEY_FILEPATH` 可以参考：[DaemonSet 安装-其他](datakit-daemonset-deploy.md#env-others)


## 延伸阅读 {#more-reading}

- [DataKit 宿主机安装](datakit-install.md)
- [DataKit DaemonSet 安装](datakit-daemonset-deploy.md)
- [DataKit 行协议过滤器](datakit-filter.md)

[^1]: 如果没有配置 CPU 限额，则 N 取物理机/Node 的 CPU 核心数
