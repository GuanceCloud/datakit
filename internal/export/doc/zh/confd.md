# 通过配置中心分发配置
---

## 配置中心介绍 {#intro}

配置中心的思路就是把项目中各种配置、各种参数、各种开关，全部都放到一个集中的地方进行统一管理，并提供一套标准的接口。当各个服务需要获取配置的时候，就来配置中心的接口拉取。当配置中心中的各种参数有更新的时候，也能通知到各个服务实时的过来同步最新的信息，使之动态更新。

采用“配置集中管理”，可以很好的解决传统的“配置文件过于分散”的问题。所有的配置都集中在配置中心这一个地方管理，不需要每一个项目都自带一个，这样极大的减轻了开发成本。

采用“配置与应用分离”，可以很好的解决传统的“配置文件无法区分环境”的问题，配置并不跟着环境走，当不同环境有不同需求的时候，就到配置中心获取即可，极大的减轻了运维部署成本。

具备“实时更新”的功能，就是用来解决传统的“静态化配置”的问题。线上系统需要调整参数的时候，只需要在配置中心动态修改即可。

DataKit 支持 `etcd-v3` `consul` `redis` `zookeeper` `aws secrets manager` `nacos` `file` 等多种配置中心，并且可以同时采用多种配置中心协同工作。
当配置中心数据发生变化，DataKit 可以自动更改配置，增加或删除采集器，相关采集器进行必要的重启。

## 引入配置中心 {#Configuration-Center-Import}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    DataKit 通过修改 `/datakit/conf.d/datakit.conf` 引入配置中心的资源。例如：

    ```
    # 原有的其他配置信息
    [[confds]]
      enable = true
      backend = "zookeeper"
      nodes = ["IP:2181","IP2:2181"...]
    [[confds]]
      enable = true
      backend = "etcdv3"
      nodes = ["IP:2379","IP2:2379"...]
      # client_cert = "可选"
      # client_key = "可选"
      # client_ca_keys = "可选"
      # basic_auth = "可选"
      # username = "可选"
      # password = "可选"
    [[confds]]
      enable = true
      backend = "redis"
      nodes = ["IP:6379","IP2:6379"...]
      # client_key = "可选"
      # separator = "可选|默认是 0"
    [[confds]]
      enable = true
      backend = "consul"
      nodes = ["IP:8500","IP2:8500"...]
      # scheme = "可选"
      # client_cert = "可选"
      # client_key = "可选"
      # client_ca_keys = "可选"
      # basic_auth = "可选"
      # username = "可选"
      # password = "可选"
    [[confds]]
      enable = true
      backend = "aws"
      region = "cn-north-1"
      # Access key ID    : must use the key file /root/.aws/config or ENV
      # Secret access key: must use the key file /root/.aws/config or ENV
      circle_interval = 60
    [[confds]]
      enable = true
      backend = "nacos"
      nodes = ["http://IP:8848","https://IP2:8848"...]
      # username = "可选"
      # password = "可选"
      circle_interval = 60 
      confd_namespace =    "confd namespace ID"
      pipeline_namespace = "pipeline namespace ID"
    # 原有的其他配置信息
    ```
    ???+ note

        如果配置多个的数据中心后端，只有排列第一的配置内容会生效。

=== "Kubernetes"

    由于 Kubernetes 环境的特殊性，采用环境变量传递的安装/配置方式最为简单。
    
    在 Kubernetes 里面安装的时候需要设置如下的环境变量，把配置信息带进去。具体参见[Kubernetes 文档](datakit-daemonset-deploy.md#env-confd)

=== "程序安装时引入"

    如果需要在安装阶段定义一些 DataKit 配置，可在安装命令中增加环境变量，在 `DK_DATAWAY` 前面追加即可。如：

    ```shell
    # Linux/Mac
{{ InstallCmd 4
(.WithPlatform "unix")
(.WithEnvs "DK_CONFD_BACKEND" "etcd3")
(.WithEnvs "DK_CONFD_BACKEND_NODES" "[127.0.0.1:2379]")
}}

    # Windows
{{ InstallCmd 4
(.WithPlatform "windows")
(.WithEnvs "DK_CONFD_BACKEND" "etcd3")
(.WithEnvs "DK_CONFD_BACKEND_NODES" "[127.0.0.1:2379]")
}}
    ```

    两种环境变量的设置格式为：

    ```shell
    # Windows: 多个环境变量之间以分号分割
    $env:NAME1="value1"; $env:Name2="value2"

    # Linux/Mac: 多个环境变量之间以空格分割
    NAME1="value1" NAME2="value2"
    ```

    具体参见[主机安装文档](datakit-install.md#env-confd)

<!-- markdownlint-enable -->

## 默认开启的采集器 {#default-enabled-inputs}
DataKit 安装完成后，会默认开启一批主机相关的采集器，无需手动配置，如 `cpu`、`disk`、`diskio`、`mem` 等。具体参见[采集器配置](datakit-input-conf.md#default-enabled-inputs)

配置中心可以修改这些采集器的配置，但是无法删除、停止这些采集器。若要删除默认采集器，可以在 DataKit *conf.d* 目录下，打开 *datakit.conf* 文件，在 `default_enabled_inputs` 中删除该采集器。

## 采集器单例运行控制 {#input-singleton}

有些采集器，只允许单例运行，例如全部的默认开启采集器、`netstat` 等。有些是可以多例运行的，例如 `nginx`、`nvidia_smi` 等。

对于只能单例运行的采集器配置，只有第一个被加载的采集配置会生效（按照文件名排序）。

## 数据格式 {#data-format}

DataKit 配置信息以 Key-Value 形式存储在配置中心。Key 的前缀必须是 `/datakit/`，例如 `/datakit/XXX` , `XXX` 不重复就可，推荐使用对应的采集器的名名字，例如 `/datakit/netstat` 。

Value 的内容就是 *conf.d* 子目录下各种配置文件的完整内容。例如：

```toml
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

file 模式，file 文件内容就是原有的 *.conf* 文件内容。

## 配置中心如何更新配置(Golang 为例) {#update-config}

### zookeeper {#update-zookeeper}

```golang
import (
    "github.com/samuel/go-zookeeper/zk"
)

func zookeeperDo(index int) {
    hosts := []string{ip + ":2181"}
    conn, _, err := zk.Connect(hosts, time.Second*5)
    if err != nil {
        fmt.Println("conn, _, err := zk.Connect error: ", err)
    }
    defer conn.Close()
    // 创建一级目录节点
    add(conn, "/datakit", "")
    // 创建二级目录节点
    add(conn, "/datakit/confd", "")
    add(conn, "/datakit/pipeline", "")
    // 创建三级目录节点
    add(conn, "/datakit/pipeline/metrics", "")
    add(conn, "/datakit/pipeline/metric", "")
    add(conn, "/datakit/pipeline/network", "")
    add(conn, "/datakit/pipeline/keyevent", "")
    add(conn, "/datakit/pipeline/object", "")
    add(conn, "/datakit/pipeline/custom_object", "")
    add(conn, "/datakit/pipeline/logging", "")
    add(conn, "/datakit/pipeline/tracing", "")
    add(conn, "/datakit/pipeline/rum", "")
    add(conn, "/datakit/pipeline/security", "")
    add(conn, "/datakit/pipeline/profiling", "")
    // 创建一个节点
    key := "/datakit/confd/netstat.conf"
    value := `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
    add(conn, key, value)
}

func add(conn *zk.Conn, path, value string) {
    if path == "" {
        return
    }

    var data = []byte(value)
    var flags int32 = 0
    acls := zk.WorldACL(zk.PermAll)
    s, err := conn.Create(path, data, flags, acls)
    if err != nil {
        fmt.Println("创建 error: ", err)
        modify(conn, path, value)
        return
    }
    fmt.Println("创建成功", s)
}

func modify(conn *zk.Conn, path, value string) {
    if path == "" {
        return
    }
    var data = []byte(value)
    _, sate, _ := conn.Get(path)
    s, err := conn.Set(path, data, sate.Version)
    if err != nil {
        fmt.Println("修改 error: ", err)
        return
    }
    fmt.Println("修改成功", s)
}

```

### etcd-v3 {#update-etcdv3}

``` golang
import (
    etcdv3 "go.etcd.io/etcd/client/v3"
)

func etcdv3Do(index int) {
    cli, err := etcdv3.New(etcdv3.Config{
        Endpoints:   []string{ip + ":2379"},
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        fmt.Println(" error: ", err)
    }
    defer cli.Close()
    key := "/datakit/confd/host/netstat.conf"
    // key := "/datakit/pipeline/metric/netstat.p"
    value := `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

    // put
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    _, err = cli.Put(ctx, key, data)
    cancel()
    if err != nil {
        fmt.Println(" error: ", err)
    }
}
```

### Redis {#update-redis}

``` golang
import (
    "github.com/go-redis/redis/v8"
)

func redisDo(index int) {
    // 初始化 context
    ctx := context.Background()

    // 初始化 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr:     ip + ":6379",
        Password: "654123", // no password set
        DB:       0,        // use default DB
    })

    // 操作 Redis
    key := "/datakit/confd/host/netstat.conf"
    value := `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

    err := rdb.Set(ctx, key, value, 0).Err()
    if err != nil {
        panic(err)
    }

    // 发布订阅
    n, err := rdb.Publish(ctx, "__keyspace@0__:/datakit*", "set").Result()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Printf("%d clients received the message\n", n)
}
```

### Consul {#update-consul}

``` golang
import (
    "github.com/hashicorp/consul/api"
)

func consulDo(index int) {
    // 创建终端
    client, err := api.NewClient(&api.Config{
        Address: "http://" + ip + ":8500",
    })
    if err != nil {
        fmt.Println(" error: ", err)
    }

    // 获得 KV 句柄
    kv := client.KV()
  
    // 注意 datakit 前面 没有 /
    key := "/datakit/confd/host/netstat.conf"
    value := `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
    // 写入数据
    p := &api.KVPair{Key: key, Value: []byte(data), Flags: 32}
    _, err = kv.Put(p, nil)
    if err != nil {
        fmt.Println(" error: ", err)
    }

    p1 := &api.KVPair{Key: key1, Value: []byte(data), Flags: 32}
    _, err = kv.Put(p1, nil)
    if err != nil {
        fmt.Println(" error: ", err)
    }
}
```

### AWS Secrets Manager  {#update-aws}

```golang
import (
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
    "github.com/aws/smithy-go"
)

func consulDo(index int) {
    // 创建终端
    region := "cn-north-1"
    config, err := config.LoadDefaultConfig(context.TODO(),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(《AccessKeyID》, 《SecretAccessKey》, "")),
        config.WithRegion(region),
    )
    // will use secret file like ~/.aws/config
    // config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
    if err != nil {
        fmt.Printf("ERROR config.LoadDefaultConfig : %v\n", err)
    }

    // 获得 KV 句柄
    conn := secretsmanager.NewFromConfig(config)
  
    key := "/datakit/confd/host/netstat.conf"
    // key := "/datakit/pipeline/metric/netstat.p"
    value := `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

    // 写入数据
    input := &secretsmanager.CreateSecretInput{
        // Description:  aws.String(""),
        Name:         aws.String(key),
        SecretString: aws.String(value),
    }

    result, err := conn.CreateSecret(context.TODO(), input)
    if err != nil {
        fmt.Println(" error: ", err)
    }
}
```

### Nacos {#update-nacos}

1. 通过网址登入 Nacos 管理页面
1. 创建 `/datakit/confd` 和 `/datakit/pipeline` 两个空间
1. 分组名按照 `datakit/conf.d` 和 `datakit/pipeline` 子目录的样式创建
1. `dataID` 按照 `.conf` 文件和 `.p` 文件的规则创建。不可省略后缀
1. 通过管理页面增/删/改 `dataID` 即可

## 配置中心更新 Pipeline {#update-config-pipeline}

参考 [配置中心如何更新配置](confd.md#update-config)

键名 `datakit/confd` 字样改为 `datakit/pipeline` ，再加上「类型/文件名」即可。例如 *datakit/pipeline/logging/nginx.p* 键值就是 Pipeline 的文本。

更新 Pipeline 支持 etcdV3/Consul/Redis/Zookeeper/AWS Secrets Manager/Nacos。

## 后端数据源软件版本说明 {#backend-version}

在开发、测试过程中，后端数据源软件使用了如下版本。

- Redis     : v6.0.16
- etcd      : v3.3.0
- Consul    : v1.13.2
- Zookeeper : v3.7.0
- Nacos     : v2.1.2
