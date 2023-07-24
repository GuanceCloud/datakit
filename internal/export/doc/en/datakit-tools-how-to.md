
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
datakit check --config
------------------------
checked 13 conf, all passing, cost 22.27455ms
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

## DataKit Debugging Commands {#debugging}

### Using Glob Rules to Retrieve File Paths {#glob-conf}
[:octicons-tag-24: Version-1.8.0](changelog.md#cl-1.8.0)

In logging collection, [glob rules can be used to configure log paths](logging.md#glob-rules).

By using the DataKit debugging glob rule, a configuration file must be provided where each line of the file is a glob statement.

Config Example:

```shell
$ cat glob-config
/tmp/log-test/*.log
/tmp/log-test/**/*.log
```

Command Example:

```shell
$ datakit debug --glob-conf glob-config
============= glob paths ============
/tmp/log-test/*.log
/tmp/log-test/**/*.log

========== found the files ==========
/tmp/log-test/1.log
/tmp/log-test/logfwd.log
/tmp/log-test/123/1.log
/tmp/log-test/123/2.log
```

### Matching Text with Regular Expressions {#regex-conf}
[:octicons-tag-24: Version-1.8.0](changelog.md#cl-1.8.0)

In log collection, regular expressions can be used to configure [multiline log collection](logging.md#multiline).

By using the DataKit debugging regular expression rule, a configuration file must be provided where the first line of the file is the regular expression statement and the remaining contents are the matched text.

Config Example:

```shell
$ cat regex-config
^\d{4}-\d{2}-\d{2}
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

Command Example:

```shell
$ datakit debug --regex-conf regex-config
============= regex rule ============
^\d{4}-\d{2}-\d{2}

========== matching results ==========
  Ok:  2020-10-23 06:41:56,688 INFO demo.py 1.0
  Ok:  2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Fail:  Traceback (most recent call last):
Fail:    File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
Fail:      response = self.full_dispatch_request()
Fail:  ZeroDivisionError: division by zero
  Ok:  2020-10-23 06:41:56,688 INFO demo.py 5.0
```
