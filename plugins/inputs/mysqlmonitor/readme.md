### 简介
mysql指标采集，参考datadog提供的指标，提供默认指标收集和用户自定义查询

### 配置
```
[[inputs.mysqlMonitor]]
    ## @param host - string - optional
    ## MySQL host to connect to.
    ## NOTE: Even if the host is "localhost", the agent connects to MySQL using TCP/IP, unless you also
    ## provide a value for the sock key (below).
    #
    host = "localhost"

    ## @param user - string - optional
    ## Username used to connect to MySQL.
    #
    user = "cc_monitor"

    ## @param pass - string - optional
    ## Password associated to the MySQL user.
    #
    pass = "<PASS>"

    ## @param port - number - optional - default: 3306
    ## Port to use when connecting to MySQL.
    #
    port = 3306

    ## @param sock - string - optional
    ## Path to a Unix Domain Socket to use when connecting to MySQL (instead of a TCP socket).
    ## If you specify a socket you dont need to specify a host.
    #
    # sock = "/tmp/mysql.sock"

    ## @param charset - string - optional
    ## Charset you want to use.
    #
    # charset = "utf8"

    ## @param connect_timeout - number - optional - default: 10
    ## Maximum number of seconds to wait before timing out when connecting to MySQL.
    #
    # connect_timeout = 10

    ## @param min_collection_interval - number - optional - default: 15
    ## This changes the collection interval of the check. For more information, see:
    #
    interval = 15

    ## @param ssl - mapping - optional
    ## Use this section to configure a TLS connection between the Agent and MySQL.
    ##
    ## The following fields are supported:
    ##
    ## key: Path to a key file.
    ## cert: Path to a cert file.
    ## ca: Path to a CA bundle file.
    #
    ## Optional TLS Config
    # [inputs.mysqlMonitor.tls]
    # tls_key = "/tmp/peer.key"
    # tls_cert = "/tmp/peer.crt"
    # tls_ca = "/tmp/ca.crt"

    ## @param service - string - optional
    ## Attach the tag `service:<SERVICE>` to every metric.
    #
    # service = "<SERVICE_NAME>""

    ## @param tags - list of strings - optional
    ## A list of tags to attach to every metric and service check emitted by this instance.
    #
    # [inputs.mysqlMonitor.tags]
    #   KEY_1 = "VALUE_1"
    #   KEY_2 = "VALUE_2"

    ## Enable options to collect extra metrics from your MySQL integration.
    #
    [inputs.mysqlMonitor.options]
        ## @param replication - boolean - optional - default: false
        ## Set to `true` to collect replication metrics.
        #
        replication = false

        ## @param replication_channel - string - optional
        ## If using multiple sources, set the channel name to monitor.
        #
        # replication_channel: <REPLICATION_CHANNEL>

        ## @param replication_non_blocking_status - boolean - optional - default: false
        ## Set to `true` to grab slave count in a non-blocking manner (requires `performance_schema`);
        #
        replication_non_blocking_status = false

        ## @param galera_cluster - boolean - optional - default: false
        ## Set to `true` to collect Galera cluster metrics.
        #
        galera_cluster = false

        ## @param extra_status_metrics - boolean - optional - default: false
        ## Set to `true` to enable extra status metrics.
        ##
        #
        extra_status_metrics = false

        ## @param extra_innodb_metrics - boolean - optional - default: false
        ## Set to `true` to enable extra InnoDB metrics.
        ##
        #
        extra_innodb_metrics = false

        ## @param disable_innodb_metrics - boolean - optional - default: false
        ## Set to `true` only if experiencing issues with older (unsupported) versions of MySQL
        ## that do not run or have InnoDB engine support.
        ##
        ## If this flag is enabled, you will only receive a small subset of metrics.
        ##
        #
        disable_innodb_metrics = false

        ## @param schema_size_metrics - boolean - optional - default: false
        ## Set to `true` to collect schema size metrics.
        ##
        ## Note that this runs a heavy query against your database to compute the relevant metrics
        ## for all your existing schemas. Due to the nature of these calls, if you have a
        ## high number of tables and schemas, this may have a negative impact on your database performance.
        ##
        #
        schema_size_metrics = false

        ## @param extra_performance_metrics - boolean - optional - default: false
        ## These metrics are reported if `performance_schema` is enabled in the MySQL instance
        ## and if the version for that instance is >= 5.6.0.
        ##
        ## Note that this runs a heavy query against your database to compute the relevant metrics
        ## for all your existing schemas. Due to the nature of these calls, if you have a
        ## high number of tables and schemas, this may have a negative impact on your database performance.
        ##
        ## Metrics provided by the options:
        ##   - mysql.info.schema.size (per schema)
        ##   - mysql.performance.query_run_time.avg (per schema)
        ##   - mysql.performance.digest_95th_percentile.avg_us
        ##
        ## Note that some of these require the `user` defined for this instance
        ## to have PROCESS and SELECT privileges. 
        #
        extra_performance_metrics = false 
```

###  收集指标

