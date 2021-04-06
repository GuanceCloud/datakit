package mysql

const (
	configSample = `
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
    user = "datakit"

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
    # sock = "<SOCK>"

    ## @param charset - string - optional
    ## Charset you want to use.
    #
    # charset = "utf8"

    ## @param connect_timeout - number - optional - default: 10
    ## Maximum number of seconds to wait before timing out when connecting to MySQL.
    #
    # connect_timeout = 10

   	## @param service - string - optional
    ## Attach the tag service:<SERVICE> to every metric
    ##
    # service = "<SERVICE>"

    ## @param ssl - mapping - optional
    ## Use this section to configure a TLS connection between the Agent and MySQL.
    ##
    ## The following fields are supported:
    ##
    ## key: Path to a key file.
    ## cert: Path to a cert file.
    ## ca: Path to a CA bundle file.
    #
    # ssl:
    #   key: <KEY_FILE_PATH>
    #   cert: <CERT_FILE_PATH>
    #   ca: <CA_PATH_FILE>


    ## Enable options to collect extra metrics from your MySQL integration.
    #
    options:
        ## @param disable_innodb_metrics - boolean - optional - default: false
        ## Set to `true` only if experiencing issues with older (unsupported) versions of MySQL
        ## that do not run or have InnoDB engine support.
        ##
        ## If this flag is enabled, you will only receive a small subset of metrics.
        ##
        ## see also the MySQL metrics listing: https://docs.datadoghq.com/integrations/mysql/#metrics
        #
        # disable_innodb_metrics = false

    ## @param tags - list of strings - optional
    ## A list of tags to attach to every metric and service check emitted by this instance.
    ##
    [inputs.httpProb.tags]
    # tag1 = val1
    # tag2 = val2
`
)



