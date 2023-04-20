
# Use of Various Other Tools
---

DataKit has built-in many different gadgets, which are convenient for everyone to use everyday. Command-line help for DataKit can be viewed with the following command:

```shell
datakit help
```

>Note: The specific help content will be different due to the differences of different platforms.

## DataKit Automatic Command Completion {#completion}

> DataKit 1.2. 12 supported this completion, and only two Linux distributions, Ubuntu and CentOS, were tested. Other Windows and Mac are not supported.

In the process of using DataKit command line, because there are many command line parameters, we added command prompt and completion functions here.

Mainstream Linux basically has command completion support. Take Ubuntu and CentOS as examples. If you want to use command completion function, you can install the following additional software packages:

- Ubuntu：`apt install bash-completion`
- CentOS: `yum install bash-completion bash-completion-extras`

If the software is already installed before the DataKit is installed, the DataKit is automatically installed with command completion. If these packages are updated after the DataKit installation, do the following to install the DataKit Command Completion feature:

```shell
datakit tool --setup-completer-script
```

Examples of completion use:

```shell
$ datakit <tab> # Enter \tab to prompt the following command
dql       help      install   monitor   pipeline  run       service   tool

$ datakit dql <tab> # Enter \tab to prompt the following options
--auto-json   --csv         -F,--force    --host        -J,--json     --log         -R,--run      -T,--token    -V,--verbose
```

All the commands mentioned below can be operated in this way.

### Get Auto-completion Script {#get-completion}

If your Linux system is not Ubuntu and CentOS, you can get the completion script through the following command, and then add it one by one according to the shell completion method of the corresponding platform.

```shell
# Export the completion script to the local datakit-completer.sh file
datakit tool --completer-script > datakit-completer.sh
```

## View DataKit Running {#using-monitor}

> Current monitor viewing has been deprecated (still available and will be deprecated soon), new monitor functionality [see here](datakit-monitor.md).

You can view the running status of DataKit on the terminal, and its effect is similar to that of the monitor page on the browser side:

DataKit's new monitor usage [see here](datakit-monitor.md).

## Check Whether the Collector is Configured Correctly {#check-conf}

After editing the collector's configuration file, there may be some configuration errors (such as the configuration file format error), which can be checked by the following command:

```shell
datakit tool --check-config
------------------------
checked 13 conf, all passing, cost 22.27455ms
```

### Collect SNMP Configuration Once {#check-snmp}

After editing the configuration file of the SNMP collector, there may be some configuration errors (such as the configuration file format error). You can collect the SNMP device once to check whether it is correct by the following command:

```shell
datakit tool --test-snmp /usr/local/datakit/conf.d/snmp/snmp.conf
# The collected information will be printed below...
......
```

## View Workspace Information {#workspace-info}

To facilitate you to view workspace information on the server side, DataKit provides the following commands:

```shell
datakit tool --workspace-info
{
  "token": {
    "ws_uuid": "wksp_2dc431d6693711eb8ff97aeee04b54af",
    "bill_state": "normal",
    "ver_type": "pay",
    "token": "tkn_2dc438b6693711eb8ff97aeee04b54af",
    "db_uuid": "ifdb_c0fss9qc8kg4gj9bjjag",
    "status": 0,
    "creator": "",
    "expire_at": -1,
    "create_at": 0,
    "update_at": 0,
    "delete_at": 0
  },
  "data_usage": {
    "data_metric": 96966,
    "data_logging": 3253,
    "data_tracing": 2868,
    "data_rum": 0,
    "is_over_usage": false
  }
}
```

## View DataKit Related Events {#event}

During the running of DataKit, some key events will be reported in the form of logs, such as the startup of DataKit and the running errors of collector. You can query through dql at the command line terminal.

```shell
datakit dql

dql > L::datakit limit 10;

-----------------[ r1.datakit.s1 ]-----------------
    __docid 'L_c6vvetpaahl15ivd7vng'
   category 'input'
create_time 1639970679664
    date_ns 835000
       host 'demo'
    message 'elasticsearch Get "http://myweb:9200/_nodes/_local/name": dial tcp 150.158.54.252:9200: connect: connection refused'
     source 'datakit'
     status 'warning'
       time 2021-12-20 11:24:34 +0800 CST
-----------------[ r2.datakit.s1 ]-----------------
    __docid 'L_c6vvetpaahl15ivd7vn0'
   category 'input'
create_time 1639970679664
    date_ns 67000
       host 'demo'
    message 'postgresql pq: password authentication failed for user "postgres"'
     source 'datakit'
     status 'warning'
       time 2021-12-20 11:24:32 +0800 CST
-----------------[ r3.datakit.s1 ]-----------------
    __docid 'L_c6tish1aahlf03dqas00'
   category 'default'
create_time 1639657028706
    date_ns 246000
       host 'zhengs-MacBook-Pro.local'
    message 'datakit start ok, ready for collecting metrics.'
     source 'datakit'
     status 'info'
       time 2021-12-20 11:16:58 +0800 CST       
          
          ...       
```

**Partial field description**
 - category: default to `default`, or an alternative value of `input`, indicating that it is associated with a collector (`input`)
 - status: Event level, and the desirable values are `info`, `warning` and `error`

## DataKit Update IP Database File {#install-ipdb}

=== "Host Installation"

    - You can install/update the IP Geographic Repository directly using the following command (here you can select another IP Address Repository `geolite2` by simply replacing  `iploc` with `geolite2`):
    
    ```shell
    datakit install --ipdb iploc
    ```
    
    - Modify the datakit.conf configuration after updating the IP geo-repository:
    
    ``` toml
    [pipeline]
      ipdb_type = "iploc"
    ```
    
    - Restart DataKit to take effect
    
    - Test the IP library for effectiveness
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```
    
    If the installation fails, the output is as follows:
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province: 
       country: 
    ```

=== "Kubernetes(yaml)"

    - Modify *datakit.yaml* and open the following highlighted content commented out:
    
    ```yaml hl_lines="2 3"
        # ---iploc-start  
        #- name: ENV_IPDB
        #  value: iploc        
        # ---iploc-end      
    ```
    
    - Restart DataKit：
    
    ```shell
    $ kubectl apply -f datakit.yaml
    
    # Make sure the DataKit container starts
    $ kubectl get pod -n datakit
    ```
    
    - Enter the container and test whether the IP library is effective
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```
    
    If the installation fails, the output is as follows:
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province:
       country:
    ```

=== "Kubernetes(helm)"

    - helm deploy add `--set iploc.enable`
    
    ```shell
    $ helm install datakit datakit/datakit -n datakit \
    --set datakit.dataway_url="https://openway.guance.com?token=<YOUR-TOKEN>" \
    --set iploc.enable true \
    --create-namespace 
    ```
    
    For helm deployment, see [here](datakit-daemonset-deploy.md/#__tabbed_1_2).
    
    - Enter the container and test whether the IP library is effective
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```
    
    If the installation fails, the output is as follows:
    
    ```shell
    $ datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province:
       country:
    ```

## DataKit Installing Third-party Software {#extras}

### Telegraf Integration {#telegraf}

> Note: It is recommended that you make sure that DataKit satisfies the desired data collection before using Telegraf. If DataKit is already supported, Telegraf is not recommended for collection, which may lead to data conflicts and cause problems in use.

Installing Telegraf integration

```shell
datakit install --telegraf
```

Start Telegraf

```shell
cd /etc/telegraf
cp telegraf.conf.sample telegraf.conf
telegraf --config telegraf.conf
```

See [here](telegraf.md) for the use of Telegraf.

### Security Checker Integration {#scheck}

Installing Security Checker

```shell
datakit install --scheck
```

It will run automatically after successful installation, and Security Checker is used in [here](../scheck/scheck-install.md).

### DataKit eBPF Integration {#ebpf}

The DataKit eBPF collector currently only supports `linux/amd64 | linux/arm64` platform. See [DataKit eBPF collector](ebpf.md) for instructions on how to use the collector.

```shell
datakit install --ebpf
```

If you are prompted `open /usr/local/datakit/externals/datakit-ebpf: text file busy`, stop the DataKit service before executing the command.

???+ warning

    The install command has been remove in [:octicons-tag-24: Version-1.5.6](changelog.md#cl-1.5.6-brk).

## Upload DataKit Run Log {#upload-log}

When troubleshooting DataKit problems, it is usually necessary to check the DataKit running log. To simplify the log collection process, DataKit supports one-click uploading of log files:

```shell
datakit tool --upload-log
log info: path/to/tkn_xxxxx/your-hostname/datakit-log-2021-11-08-1636340937.zip # Just send this path information to our engineers
```

After running the command, all log files in the log directory are packaged and compressed, and then uploaded to the specified store. Our engineers will find the corresponding file according to the hostname and Token of the uploaded log, and then troubleshoot the DataKit problem.

## Collect DataKit Running information {#bug-report}

[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9) · [:octicons-beaker-24: Experimental](index.md#experimental)

When troubleshooting issues with DataKit, it is necessary to manually collect various relevant information such as logs, configuration files, and monitoring data. This process can be cumbersome. To simplify this process, DataKit provides a command that can retrieve all the relevant information at once and package it into a file. Usage is as follows:

```shell
datakit tool --bug-report
```

After successful execution, a zip file will be generated in the current directory with the naming format of `info-<timestamp in milliseconds>.zip`。

The list of files is as follows:

```shell

├── config
│   ├── container
│   │   └── container.conf
│   ├── datakit.conf
│   ├── db
│   │   ├── kafka.conf
│   │   ├── mysql.conf
│   │   └── sqlserver.conf
│   ├── host
│   │   ├── cpu.conf
│   │   ├── disk.conf
│   │   └── system.conf
│   ├── network
│   │   └── dialtesting.conf
│   ├── profile
│   │   └── profile.conf
│   ├── pythond
│   │   └── pythond.conf
│   └── rum
│       └── rum.conf
├── env.txt
├── metrics 
│   ├── metric-1680513455403 
│   ├── metric-1680513460410
│   └── metric-1680513465416 
├── log
│   ├── gin.log
│   └── log
└── profile
    ├── allocs
    ├── heap
    └── profile

```

Document Explanation

| name      | dir  | description                                                                                            |
| ---:      | ---: | ---:                                                                                                   |
| `config`  | yes  | Configuration file, including the main configuration and the configuration of the enabled collectors.  |
| `env.txt` | no   | The environment variables of the runtime.                                                              |
| `log`     | yes  | Latest log files, such as log and gin log, not supporting `stdout` currently                           |
| `profile` | yes  | When pprof is enabled, it will collect profile data.                                                   |
| `metrics` | yes  | The data returned by the `/metrics` API is named in the format of `metric-<timestamp in milliseconds>` |

**Mask sensitive information**

When collecting information, sensitive information (such as tokens, passwords, etc.) will be automatically filtered and replaced. The specific rules are as follows:

- Environment variables

Only retrieve environment variables starting with `ENV_`, and mask environment variables containing `password`, `token`, `key`, `key_pw`, `secret` in their names by replacing them with `******`.

- Configuration files 

Perform the following regular expression replacement on the contents of the configuration file, for example:

```
https://openway.guance.com?token=tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` => `https://openway.guance.com?token=******
pass = "1111111"` => `pass = "******"
postgres://postgres:123456@localhost/test` => `postgres://postgres:******@localhost/test
```

After the above treatment, most sensitive information can be removed. Nevertheless, if there is still some sensitive information in the exported file, you can manually remove it.

## View Cloud Property Data {#cloudinfo}

If the DataKit is installed on a cloud server (currently supports `aliyun/tencent/aws/hwcloud/azure`), you can view some of the cloud attribute data with the following commands, such as (marked `-` to indicate that the field is invalid):

```shell
datakit tool --show-cloud-info aws

           cloud_provider: aws
              description: -
     instance_charge_type: -
              instance_id: i-09b37dc1xxxxxxxxx
            instance_name: -
    instance_network_type: -
          instance_status: -
            instance_type: t2.nano
               private_ip: 172.31.22.123
                   region: cn-northwest-1
        security_group_id: launch-wizard-1
                  zone_id: cnnw1-az2
```

## Parse Line Protocols {#parse-lp}

[:octicons-tag-24: Version-1.5.6](changelog.md#cl-1.5.6)

You can run the following command to parse the line protocol data:

```shell
datakit tool --parse-lp /path/to/file
Parse 201 points OK, with 2 measurements and 201 time series
```

It can be output in JSON:

```shell
datakit tool --parse-lp /path/to/file --json
{
  "measurements": {  # Measurement list
    "testing": {
      "points": 7,
      "time_series": 6
    },
    "testing_module": {
      "points": 195,
      "time_series": 195
    }
  },
  "point": 202,      # Total points
  "time_serial": 201 # Total time series
}
```
