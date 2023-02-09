# Distribute Configuration through Configuration Center
---

## Introduction to Configuration Center {#intro}

The idea of configuration center is to put all kinds of configurations, parameters and switches in a centralized place for unified management and provide a set of standard interfaces. 
When each service needs to obtain configuration, it will configure the interface pull of the center. When various parameters in the configuration center are updated, it can also inform each service to synchronize the latest information in real time, so that it can be dynamically updated.

Adopting "centralized configuration management" can solve the traditional problem of "too scattered configuration files". All configurations are centralized in the configuration center, which does not need to bring one for each project, thus greatly reducing the development cost.

Adopting "separation of configuration and application" can solve the problem that the traditional "configuration file can't distinguish the environment". Configuration does not follow the environment. When different environments have different requirements, it can be obtained from the configuration center, which greatly reduces the operation and maintenance deployment cost.

With the function of "real-time update", it is used to solve the problem of traditional "static configuration". When the online system needs to adjust the parameters, it only needs to be dynamically modified in the configuration center.

Datakit supports multiple configuration centers, such as etcdv3 consul redis zookeeper file, and can work together with multiple configuration centers at the same time. 
When the configuration center data changes, datakit can automatically change the configuration, add or delete collectors, and relevant collectors are restarted as necessary.

## Introducing Configuration Center {#Configuration-Center-Import}

=== "datakit.conf introduced"

	Datakit introduces resources in the configuration center by modifying `/datakit/conf.d/datakit.conf '. For example: 
	
	```
	# Other existing configuration information...
	[[confds]]
	  enable = true
	  backend = "zookeeper"
	  nodes = ["IP address:2181","IP address2:2181"...]
	[[confds]]
	  enable = true
	  backend = "etcdv3"
	  nodes = ["IP address:2379","IP address2:2379"...]
	  # client_cert = "optional"
	  # client_key = "optional"
	  # client_ca_keys = "optional"
	  # basic_auth = "optional"
	  # username = "optional"
	  # password = "optional"
	[[confds]]
	  enable = true
	  backend = "redis"
	  nodes = ["IP address:6379","IP address2:6379"...]
	  # client_key = "optional"
	  # separator = "optional|0 by default"
	[[confds]]
	  enable = true
	  backend = "consul"
	  nodes = ["IP address:8500","IP address 2:8500"...]
	  # scheme = "optional"
	  # client_cert = "optional"
	  # client_key = "optional"
	  # client_ca_keys = "optional"
	  # basic_auth = "optional"
	  # username = "optional"
	  # password = "optional"
	# Not recommended  
	[[confds]]
	  enable = false
	  backend = "file"
	  file = ["/file1access/file1","/file2路径/文件2"...]
	# Other existing configuration information...
	```
	
	Multiple datacenter backends can be configured at the same time, and the data configuration information of each backend is merged and injected into datakit. Any back-end information changes will be detected by datakit, and datakit will automatically update the relevant configuration and restart the corresponding collector.

=== "Kubernates introduced"

	Because of the particularity of Kubernates environment, the installation/configuration mode with environment variable passing is the simplest.
	
	When installing in Kubernates, you need to set the following environment variables to bring Confd configuration information into it:
	
	See [Kubernetes document](datakit-daemonset-deploy.md#env-confd) for more details.

=== "Introduced during program installation"

	If you need to define some DataKit configuration during the installation phase, you can add environment variables to the installation command, just append them before DK_DATAWAY. Such as:
	
	```shell
	# Linux/Mac
	DK_CONFD_BACKEND="etcdv3" DK_CONFD_BACKEND_NODES="[127.0.0.1:2379]" DK_DATAWAY="https://openway.guance.com?token=<TOKEN>" bash -c "$	(curl -L https://static.guance.com/datakit/install.sh)"
	
	# Windows
	$env:DK_CONFD_BACKEND="etcdv3";$env:DK_CONFD_BACKEND_NODES="[127.0.0.1:2379]"; $env:DK_DATAWAY="https://openway.guance.com?	token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://	static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
	```
	
	The two environment variables are formatted as:
	
	```shell
	# Windows: multiple environment variables are divided by semicolons
	$env:NAME1="value1"; $env:Name2="value2"
	
	# Linux/Mac: multiple environment variables are divided by spaces
	NAME1="value1" NAME2="value2"
	```
	
	See [host installation documentation](datakit-install.md#env-confd) for more information.

## Collector Turned on by Default {#default-enabled-inputs}
After DataKit is installed, a batch of host-related collectors will be turned on by default without manual configuration, such as:

cpu, disk, diskio, memand so on. See [Collector Configuration](datakit-input-conf.md#default-enabled-inputs) for details.

Configuration Center can modify the configuration of these collectors, but cannot delete or stop these collectors.

If you want to delete the default collector, you can open the datakit.conf file in the DataKit conf.d directory and delete the collector in default_enabled_inputs.

Self can neither delete, stop, nor modify the configuration.

## Collector Singleton Run Control {#input-singleton}

Some collectors only need to run singletons, such as all default open collectors, netstat, etc. Some can be run in multiple instances, such as nginx, nvidia_smi... and so on.

In the collector configuration of single case operation, only the data ranked first is accepted, and the latter is automatically discarded.

## Data Format {#data-format}

Datakit configuration information is stored in the data center as a Key-Value.

The prefix of Key must be `/datakit/`, such as  `/datakit/XXX` , `XXX` is not duplicated. It is recommended to use the corresponding collector name, such as `/datakit/netstat`.

The contents of Value are the full contents of the various configuration files in the conf. d subdirectory. For example:
```
`
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
```
file mode: the contents of the. conf file are the contents of the original. conf file.

## How the Configuration Center Updates the Configuration(Take golang as an Example) {#update-config}

### zookeeper {#update-zookeeper}

```
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
	// Create a first-level directory node
	add(conn, "/datakit/confd", "")
	// Create a node
	key := "/datakit/confd/netstat"
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

// add
func add(conn *zk.Conn, path, value string) {
	if path == "" {
		return
	}

	var data = []byte(value)
	var flags int32 = 0
	acls := zk.WorldACL(zk.PermAll)
	s, err := conn.Create(path, data, flags, acls)
	if err != nil {
		fmt.Println("creat error: ", err)
		modify(conn, path, value)
		return
	}
	fmt.Println("successfully created", s)
}

// modify
func modify(conn *zk.Conn, path, value string) {
	if path == "" {
		return
	}
	var data = []byte(value)
	_, sate, _ := conn.Get(path)
	s, err := conn.Set(path, data, sate.Version)
	if err != nil {
		fmt.Println("modify error: ", err)
		return
	}
	fmt.Println("successfully modified", s)
}

```

### etcdv3 {#update-etcdv3}

```
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
	key := "/datakit/confd/netstat"
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

### redis {#update-redis}

```
import (
	"github.com/go-redis/redis/v8"
)

func redisDo(index int) {
	// initialize context
	ctx := context.Background()

	// initialize redis client end
	rdb := redis.NewClient(&redis.Options{
		Addr:     ip + ":6379",
		Password: "654123", // no password set
		DB:       0,        // use default DB
	})

	// operate redis
	key := "/datakit/confd/netstat"
	value := `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

	// write
	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		panic(err)
	}

	// publish Subscription
	n, err := rdb.Publish(ctx, "__keyspace@0__:/datakit*", "set").Result()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%d clients received the message\n", n)
}
```

### consul {#update-consul}

```
import (
	"github.com/hashicorp/consul/api"
)

func consulDo(index int) {
	// create terminal
	client, err := api.NewClient(&api.Config{
		Address: "http://" + ip + ":8500",
	})
	if err != nil {
		fmt.Println(" error: ", err)
	}

	// get a KV handle
	kv := client.KV()
  
    // note that datakit is not preceded by /
	key := "datakit/confd/netstat"
	value := `
[[inputs.netstat]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

	// write data
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

### aws secrets manager  {#update-aws}

```
import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go"
)

func consulDo(index int) {
	// creat terminal
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

	// obtain a KV handle
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

	// write in data
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

### nacos {#update-nacos}

    1. Log in to the `nacos` management page through the URL.
    2. Create two spaces: `/datakit/confd` and `/datakit/pipeline`.
    3. Group names are created in the style of `datakit/conf.d` and `datakit/pipeline`.
    4. `dataID` is created according to the rules of `.conf` and `.p` files. (The suffix cannot be omitted).
    5. Add/delete/change `dataID` through the management page.

## Updating Pipeline in Configuration Center  {#update-config-pipeline}

Refer to [how Configuration Center updates configuration](#update-config)

Change the key name `datakit/confd` to `datakit/pipeline`, plus the `type/file name`.

For example,  `datakit/pipeline/logging/nginx.p`.

The key value is the text of the pipeline.

Update Pipeline supports etcdv3 consul redis zookeeper, not file backend.

## Backend Data Source Software Version Description {#backend-version}

In the process of development and testing, the back-end data source software uses the following version.
- REDIS: v6.0.16
- ETCD: v3.3.0 
- CONSUL: v1.13.2
- ZOOKEEPER: v3.7.0