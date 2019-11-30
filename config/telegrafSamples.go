package config

var (
	telegrafCfgSamples = make(map[string]string)

	supportsTelegrafMetraicNames = []string{
		`nginx`,
		`apache`,
		`disk`,
		`cpu`,
		`diskio`,
		`mysql`,
		`mongodb`,
		`redis`,
		`elasticsearch`,
		`openldap`,
		`sqlserver`,
		`phpfpm`,
		`memory`,
		`activemq`,
		`zookeeper`,
		`ceph`,
		`dns_query`,
		`docker`,
		`docker_log`,
		`fluentd`,
		`haproxy`,
		`memcached`,
		`iptables`,
		`jenkins`,
		`kapacitor`,
		`ntpq`,
		`openntpd`,
		`ping_check`,
		`postgresql`,
		`rabbitmq`,
		`swap`,
		`system`,
		`tengine`,
		`network`,
		`processes`,
		`tcp_udp_check`,
		`http_check`,
		`kernel`,
		`tls`,
		`nats`,
		`win_services`,
	}

	metricsEnablesFlags = make([]bool, len(supportsTelegrafMetraicNames))
)

func Init() {

	telegrafCfgSamples["win_services"] = `
## Reports information about Windows service status.
## Monitoring some services may require running Telegraf with administrator privileges.
## Names of the services to monitor. Leave empty to monitor all the available services on the host
#service_names = [
#	"LanmanServer",
#	"TermService",
#]
`

	telegrafCfgSamples["tls"] = `
# # Reads metrics from a SSL certificate
# [[inputs.x509_cert]]
#   ## List certificate sources
#   sources = ["/etc/ssl/certs/ssl-cert-snakeoil.pem", "tcp://example.org:443"]
#
#   ## Timeout for SSL connection
#   # timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
`

	telegrafCfgSamples["tengine"] = `
# # Read Tengine's basic status information (ngx_http_reqstat_module)
# [[inputs.tengine]]
#   # An array of Tengine reqstat module URI to gather stats.
#   urls = ["http://127.0.0.1/us"]
#
#   # HTTP response timeout (default: 5s)
#   # response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.cer"
#   # tls_key = "/etc/telegraf/key.key"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`
	telegrafCfgSamples["rabbitmq"] = `
# # Reads metrics from RabbitMQ servers via the Management Plugin
# [[inputs.rabbitmq]]
#   ## Management Plugin url. (default: http://localhost:15672)
#   # url = "http://localhost:15672"
#   ## Tag added to rabbitmq_overview series; deprecated: use tags
#   # name = "rmq-server-1"
#   ## Credentials
#   # username = "guest"
#   # password = "guest"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
#
#   ## Optional request timeouts
#   ##
#   ## ResponseHeaderTimeout, if non-zero, specifies the amount of time to wait
#   ## for a server's response headers after fully writing the request.
#   # header_timeout = "3s"
#   ##
#   ## client_timeout specifies a time limit for requests made by this client.
#   ## Includes connection time, any redirects, and reading the response body.
#   # client_timeout = "4s"
#
#   ## A list of nodes to gather as the rabbitmq_node measurement. If not
#   ## specified, metrics for all nodes are gathered.
#   # nodes = ["rabbit@node1", "rabbit@node2"]
#
#   ## A list of queues to gather as the rabbitmq_queue measurement. If not
#   ## specified, metrics for all queues are gathered.
#   # queues = ["telegraf"]
#
#   ## A list of exchanges to gather as the rabbitmq_exchange measurement. If not
#   ## specified, metrics for all exchanges are gathered.
#   # exchanges = ["telegraf"]
#
#   ## Queues to include and exclude. Globs accepted.
#   ## Note that an empty array for both will include all queues
#   queue_name_include = []
#   queue_name_exclude = []
`

	telegrafCfgSamples["openntpd"] = `
# # Get standard NTP query metrics from OpenNTPD.
# [[inputs.openntpd]]
#   ## Run ntpctl binary with sudo.
#   # use_sudo = false
#
#   ## Location of the ntpctl binary.
#   # binary = "/usr/sbin/ntpctl"
#
#   ## Maximum time the ntpctl binary is allowed to run.
#   # timeout = "5ms"
`

	telegrafCfgSamples["ntpq"] = `
# # Get standard NTP query metrics, requires ntpq executable.
# [[inputs.ntpq]]
#   ## If false, set the -n ntpq flag. Can reduce metric gather time.
#   dns_lookup = true
`
	telegrafCfgSamples["jenkins"] = `
# # Read jobs and cluster metrics from Jenkins instances
# [[inputs.jenkins]]
#   ## The Jenkins URL
#   url = "http://my-jenkins-instance:8080"
#   # username = "admin"
#   # password = "admin"
#
#   ## Set response_timeout
#   response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use SSL but skip chain & host verification
#   # insecure_skip_verify = false
#
#   ## Optional Max Job Build Age filter
#   ## Default 1 hour, ignore builds older than max_build_age
#   # max_build_age = "1h"
#
#   ## Optional Sub Job Depth filter
#   ## Jenkins can have unlimited layer of sub jobs
#   ## This config will limit the layers of pulling, default value 0 means
#   ## unlimited pulling until no more sub jobs
#   # max_subjob_depth = 0
#
#   ## Optional Sub Job Per Layer
#   ## In workflow-multibranch-plugin, each branch will be created as a sub job.
#   ## This config will limit to call only the lasted branches in each layer,
#   ## empty will use default value 10
#   # max_subjob_per_layer = 10
#
#   ## Jobs to exclude from gathering
#   # job_exclude = [ "job1", "job2/subjob1/subjob2", "job3/*"]
#
#   ## Nodes to exclude from gathering
#   # node_exclude = [ "node1", "node2" ]
#
#   ## Worker pool for jenkins plugin only
#   ## Empty this field will use default value 5
#   # max_connections = 5
`

	telegrafCfgSamples["ceph"] = `
# # Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster.
# [[inputs.ceph]]
#   ## This is the recommended interval to poll.  Too frequent and you will lose
#   ## data points due to timeouts during rebalancing and recovery
#   interval = '1m'
#
#   ## All configuration values are optional, defaults are shown below
#
#   ## location of ceph binary
#   ceph_binary = "/usr/bin/ceph"
#
#   ## directory in which to look for socket files
#   socket_dir = "/var/run/ceph"
#
#   ## prefix of MON and OSD socket files, used to determine socket type
#   mon_prefix = "ceph-mon"
#   osd_prefix = "ceph-osd"
#
#   ## suffix used to identify socket files
#   socket_suffix = "asok"
#
#   ## Ceph user to authenticate as
#   ceph_user = "client.admin"
#
#   ## Ceph configuration to use to locate the cluster
#   ceph_config = "/etc/ceph/ceph.conf"
#
#   ## Whether to gather statistics via the admin socket
#   gather_admin_socket_stats = true
#
#   ## Whether to gather statistics via ceph commands
#   gather_cluster_stats = false
`

	telegrafCfgSamples["zookeeper"] = `
# # Reads 'mntr' stats from one or many zookeeper servers
# [[inputs.zookeeper]]
#   ## An array of address to gather stats about. Specify an ip or hostname
#   ## with port. ie localhost:2181, 10.0.0.1:2181, etc.
#
#   ## If no servers are specified, then localhost is used as the host.
#   ## If no port is specified, 2181 is used
#   servers = [":2181"]
#
#   ## Timeout for metric collections from all servers.  Minimum timeout is "1s".
#   # timeout = "5s"
#
#   ## Optional TLS Config
#   # enable_tls = true
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## If false, skip chain & host verification
#   # insecure_skip_verify = true
`

	telegrafCfgSamples["haproxy"] = `
# # Read metrics of haproxy, via socket or csv stats page
# [[inputs.haproxy]]
#   ## An array of address to gather stats about. Specify an ip on hostname
#   ## with optional port. ie localhost, 10.10.3.33:1936, etc.
#   ## Make sure you specify the complete path to the stats endpoint
#   ## including the protocol, ie http://10.10.3.33:1936/haproxy?stats
#
#   ## If no servers are specified, then default to 127.0.0.1:1936/haproxy?stats
#   servers = ["http://myhaproxy.com:1936/haproxy?stats"]
#
#   ## Credentials for basic HTTP authentication
#   # username = "admin"
#   # password = "admin"
#
#   ## You can also use local socket with standard wildcard globbing.
#   ## Server address not starting with 'http' will be treated as a possible
#   ## socket, so both examples below are valid.
#   # servers = ["socket:/run/haproxy/admin.sock", "/run/haproxy/*.sock"]
#
#   ## By default, some of the fields are renamed from what haproxy calls them.
#   ## Setting this option to true results in the plugin keeping the original
#   ## field names.
#   # keep_field_names = false
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples["fluentd"] = `
# # Read metrics exposed by fluentd in_monitor plugin
# [[inputs.fluentd]]
#   ## This plugin reads information exposed by fluentd (using /api/plugins.json endpoint).
#   ##
#   ## Endpoint:
#   ## - only one URI is allowed
#   ## - https is not supported
#   endpoint = "http://localhost:24220/api/plugins.json"
#
#   ## Define which plugins have to be excluded (based on "type" field - e.g. monitor_agent)
#   exclude = [
# 	  "monitor_agent",
# 	  "dummy",
#   ]
`

	telegrafCfgSamples["cpu"] = `
#[[inputs.cpu]]
#  ## Whether to report per-cpu stats or not
#  percpu = true
	## Whether to report total system cpu stats or not
#  totalcpu = true
#  ## If true, collect raw CPU time metrics.
#  collect_cpu_time = false
	## If true, compute and report the sum of all non-idle CPU states.
#  report_active = false
`

	telegrafCfgSamples[`disk`] = `
# Read metrics about disk usage by mount point
#[[inputs.disk]]
  ## By default stats will be gathered for all mount points.
  ## Set mount_points will restrict the stats to only the specified mount points.
  # mount_points = ["/"]

  ## Ignore mount points by filesystem type.
#  ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]
`

	telegrafCfgSamples[`diskio`] = `
# Read metrics about disk IO by device
#[[inputs.diskio]]
  ## By default, telegraf will gather stats for all devices including
  ## disk partitions.
  ## Setting devices will restrict the stats to the specified devices.
  # devices = ["sda", "sdb", "vd*"]
  ## Uncomment the following line if you need disk serial numbers.
  # skip_serial_number = false
  #
  ## On systems which support it, device metadata can be added in the form of
  ## tags.
  ## Currently only Linux is supported via udev properties. You can view
  ## available properties for a device by running:
  ## 'udevadm info -q property -n /dev/sda'
  ## Note: Most, but not all, udev properties can be accessed this way. Properties
  ## that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
  # device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]
  #
  ## Using the same metadata source as device_tags, you can also customize the
  ## name of the device via templates.
  ## The 'name_templates' parameter is a list of templates to try and apply to
  ## the device. The template may contain variables in the form of '$PROPERTY' or
  ## '${PROPERTY}'. The first template which does not contain any variables not
  ## present for the device is used as the device name tag.
  ## The typical use case is for LVM volumes, to get the VG/LV name instead of
  ## the near-meaningless DM-0 name.
  # name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME"]
`

	telegrafCfgSamples[`kernel`] = `
# Get kernel statistics from /proc/stat
#[[inputs.kernel]]
  # no configuration

# # Get kernel statistics from /proc/vmstat
# [[inputs.kernel_vmstat]]
#   # no configuration
`

	telegrafCfgSamples[`memory`] = `
# Read metrics about memory usage
#[[inputs.mem]]
  # no configuration
`

	telegrafCfgSamples[`processes`] = `
# Get the number of processes and group them by status
#[[inputs.processes]]
  # no configuration
`

	telegrafCfgSamples[`swap`] = `
# Read metrics about swap memory usage
#[[inputs.swap]]
  # no configuration
`

	telegrafCfgSamples[`system`] = `
# Read metrics about system load & uptime
#[[inputs.system]]
  ## Uncomment to remove deprecated metrics.
  # fielddrop = ["uptime_format"]
`

	telegrafCfgSamples[`activemq`] = `
# # Gather ActiveMQ metrics
# [[inputs.activemq]]
#   ## ActiveMQ WebConsole URL
#   url = "http://127.0.0.1:8161"
#
#   ## Required ActiveMQ Endpoint
#   ##   deprecated in 1.11; use the url option
#   # server = "127.0.0.1"
#   # port = 8161
#
#   ## Credentials for basic HTTP authentication
#   # username = "admin"
#   # password = "admin"
#
#   ## Required ActiveMQ webadmin root path
#   # webadmin = "admin"
#
#   ## Maximum time to receive response.
#   # response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`dns_query`] = `
# # Query given DNS server and gives statistics
# [[inputs.dns_query]]
#   ## servers to query
#   servers = ["8.8.8.8"]
#
#   ## Network is the network protocol name.
#   # network = "udp"
#
#   ## Domains or subdomains to query.
#   # domains = ["."]
#
#   ## Query record type.
#   ## Posible values: A, AAAA, CNAME, MX, NS, PTR, TXT, SOA, SPF, SRV.
#   # record_type = "A"
#
#   ## Dns server port.
#   # port = 53
#
#   ## Query timeout in seconds.
#   # timeout = 2	
`

	telegrafCfgSamples[`docker`] = `
# # Read metrics about docker containers
# [[inputs.docker]]
#   ## Docker Endpoint
#   ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
#   ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
#   endpoint = "unix:///var/run/docker.sock"
#
#   ## Set to true to collect Swarm metrics(desired_replicas, running_replicas)
#   gather_services = false
#
#   ## Only collect metrics for these containers, collect all if empty
#   container_names = []
#
#   ## Containers to include and exclude. Globs accepted.
#   ## Note that an empty array for both will include all containers
#   container_name_include = []
#   container_name_exclude = []
#
#   ## Container states to include and exclude. Globs accepted.
#   ## When empty only containers in the "running" state will be captured.
#   ## example: container_state_include = ["created", "restarting", "running", "removing", "paused", "exited", "dead"]
#   ## example: container_state_exclude = ["created", "restarting", "running", "removing", "paused", "exited", "dead"]
#   # container_state_include = []
#   # container_state_exclude = []
#
#   ## Timeout for docker list, info, and stats commands
#   timeout = "5s"
#
#   ## Whether to report for each container per-device blkio (8:0, 8:1...) and
#   ## network (eth0, eth1, ...) stats or not
#   perdevice = true
#   ## Whether to report for each container total blkio and network stats or not
#   total = false
#   ## Which environment variables should we use as a tag
#   ##tag_env = ["JAVA_HOME", "HEAP_SIZE"]
#
#   ## docker labels to include and exclude as tags.  Globs accepted.
#   ## Note that an empty array for both will include all labels as tags
#   docker_label_include = []
#   docker_label_exclude = []
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`docker_log`] = `
# # Read logging output from the Docker engine
# [[inputs.docker_log]]
#   ## Docker Endpoint
#   ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
#   ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
#   # endpoint = "unix:///var/run/docker.sock"
#
#   ## When true, container logs are read from the beginning; otherwise
#   ## reading begins at the end of the log.
#   # from_beginning = false
#
#   ## Timeout for Docker API calls.
#   # timeout = "5s"
#
#   ## Containers to include and exclude. Globs accepted.
#   ## Note that an empty array for both will include all containers
#   # container_name_include = []
#   # container_name_exclude = []
#
#   ## Container states to include and exclude. Globs accepted.
#   ## When empty only containers in the "running" state will be captured.
#   # container_state_include = []
#   # container_state_exclude = []
#
#   ## docker labels to include and exclude as tags.  Globs accepted.
#   ## Note that an empty array for both will include all labels as tags
#   # docker_label_include = []
#   # docker_label_exclude = []
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`elasticsearch`] = `
# # Read stats from one or more Elasticsearch servers or clusters
# [[inputs.elasticsearch]]
#   ## specify a list of one or more Elasticsearch servers
#   # you can add username and password to your url to use basic authentication:
#   # servers = ["http://user:pass@localhost:9200"]
#   servers = ["http://localhost:9200"]
#
#   ## Timeout for HTTP requests to the elastic search server(s)
#   http_timeout = "5s"
#
#   ## When local is true (the default), the node will read only its own stats.
#   ## Set local to false when you want to read the node stats from all nodes
#   ## of the cluster.
#   local = true
#
#   ## Set cluster_health to true when you want to also obtain cluster health stats
#   cluster_health = false
#
#   ## Adjust cluster_health_level when you want to also obtain detailed health stats
#   ## The options are
#   ##  - indices (default)
#   ##  - cluster
#   # cluster_health_level = "indices"
#
#   ## Set cluster_stats to true when you want to also obtain cluster stats.
#   cluster_stats = false
#
#   ## Only gather cluster_stats from the master node. To work this require local = true
#   cluster_stats_only_from_master = true
#
#   ## Indices to collect; can be one or more indices names or _all
#   indices_include = ["_all"]
#
#   ## One of "shards", "cluster", "indices"
#   indices_level = "shards"
#
#   ## node_stats is a list of sub-stats that you want to have gathered. Valid options
#   ## are "indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http",
#   ## "breaker". Per default, all stats are gathered.
#   # node_stats = ["jvm", "http"]
#
#   ## HTTP Basic Authentication username and password.
#   # username = ""
#   # password = ""
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`http_check`] = `
# # Read formatted metrics from one or more HTTP endpoints
# [[inputs.http]]
#   ## One or more URLs from which to read formatted metrics
#   urls = [
#     "http://localhost/metrics"
#   ]
#
#   ## HTTP method
#   # method = "GET"
#
#   ## Optional HTTP headers
#   # headers = {"X-Special-Header" = "Special-Value"}
#
#   ## Optional HTTP Basic Auth Credentials
#   # username = "username"
#   # password = "pa$$word"
#
#   ## HTTP entity-body to send with POST/PUT requests.
#   # body = ""
#
#   ## HTTP Content-Encoding for write request body, can be set to "gzip" to
#   ## compress body or "identity" to apply no encoding.
#   # content_encoding = "identity"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
#
#   ## Amount of time allowed to complete the HTTP request
#   # timeout = "5s"
#
#   ## Data format to consume.
#   ## Each data format has its own unique set of configuration options, read
#   ## more about them here:
#   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
#   # data_format = "influx"


# # HTTP/HTTPS request given an address a method and a timeout
# [[inputs.http_response]]
#   ## Deprecated in 1.12, use 'urls'
#   ## Server address (default http://localhost)
#   # address = "http://localhost"
#
#   ## List of urls to query.
#   # urls = ["http://localhost"]
#
#   ## Set http_proxy (telegraf uses the system wide proxy settings if it's is not set)
#   # http_proxy = "http://localhost:8888"
#
#   ## Set response_timeout (default 5 seconds)
#   # response_timeout = "5s"
#
#   ## HTTP Request Method
#   # method = "GET"
#
#   ## Whether to follow redirects from the server (defaults to false)
#   # follow_redirects = false
#
#   ## Optional HTTP Request Body
#   # body = '''
#   # {'fake':'data'}
#   # '''
#
#   ## Optional substring or regex match in body of the response
#   # response_string_match = "\"service_status\": \"up\""
#   # response_string_match = "ok"
#   # response_string_match = "\".*_status\".?:.?\"up\""
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
#
#   ## HTTP Request Headers (all values must be strings)
#   # [inputs.http_response.headers]
#   #   Host = "github.com"
#
#   ## Interface to use when dialing an address
#   # interface = "eth0"


# # Read flattened metrics from one or more JSON HTTP endpoints
# [[inputs.httpjson]]
#   ## NOTE This plugin only reads numerical measurements, strings and booleans
#   ## will be ignored.
#
#   ## Name for the service being polled.  Will be appended to the name of the
#   ## measurement e.g. httpjson_webserver_stats
#   ##
#   ## Deprecated (1.3.0): Use name_override, name_suffix, name_prefix instead.
#   name = "webserver_stats"
#
#   ## URL of each server in the service's cluster
#   servers = [
#     "http://localhost:9999/stats/",
#     "http://localhost:9998/stats/",
#   ]
#   ## Set response_timeout (default 5 seconds)
#   response_timeout = "5s"
#
#   ## HTTP method to use: GET or POST (case-sensitive)
#   method = "GET"
#
#   ## List of tag names to extract from top-level of JSON server response
#   # tag_keys = [
#   #   "my_tag_1",
#   #   "my_tag_2"
#   # ]
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
#
#   ## HTTP parameters (all values must be strings).  For "GET" requests, data
#   ## will be included in the query.  For "POST" requests, data will be included
#   ## in the request body as "x-www-form-urlencoded".
#   # [inputs.httpjson.parameters]
#   #   event_type = "cpu_spike"
#   #   threshold = "0.75"
#
#   ## HTTP Headers (all values must be strings)
#   # [inputs.httpjson.headers]
#   #   X-Auth-Token = "my-xauth-token"
#   #   apiVersion = "v1"
`

	telegrafCfgSamples[`iptables`] = `
# # Gather packets and bytes throughput from iptables
# [[inputs.iptables]]
#   ## iptables require root access on most systems.
#   ## Setting 'use_sudo' to true will make use of sudo to run iptables.
#   ## Users must configure sudo to allow telegraf user to run iptables with no password.
#   ## iptables can be restricted to only list command "iptables -nvL".
#   use_sudo = false
#   ## Setting 'use_lock' to true runs iptables with the "-w" option.
#   ## Adjust your sudo settings appropriately if using this option ("iptables -w 5 -nvl")
#   use_lock = false
#   ## Define an alternate executable, such as "ip6tables". Default is "iptables".
#   # binary = "ip6tables"
#   ## defines the table to monitor:
#   table = "filter"
#   ## defines the chains to monitor.
#   ## NOTE: iptables rules without a comment will not be monitored.
#   ## Read the plugin documentation for more information.
#   chains = [ "INPUT" ]
`

	telegrafCfgSamples[`redis`] = `
# # Read metrics from one or many redis servers
# [[inputs.redis]]
#   ## specify servers via a url matching:
#   ##  [protocol://][:password]@address[:port]
#   ##  e.g.
#   ##    tcp://localhost:6379
#   ##    tcp://:password@192.168.99.100
#   ##    unix:///var/run/redis.sock
#   ##
#   ## If no servers are specified, then localhost is used as the host.
#   ## If no port is specified, 6379 is used
#   servers = ["tcp://localhost:6379"]
#
#   ## specify server password
#   # password = "s#cr@t%"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = true
`

	telegrafCfgSamples[`mongodb`] = `
# # Read metrics from one or many MongoDB servers
# [[inputs.mongodb]]
#   ## An array of URLs of the form:
#   ##   "mongodb://" [user ":" pass "@"] host [ ":" port]
#   ## For example:
#   ##   mongodb://user:auth_key@10.10.3.30:27017,
#   ##   mongodb://10.10.3.33:18832,
#   servers = ["mongodb://127.0.0.1:27017"]
#
#   ## When true, collect per database stats
#   # gather_perdb_stats = false
#   ## When true, collect per collection stats
#   # gather_col_stats = false
#   ## List of db where collections stats are collected
#   ## If empty, all db are concerned
#   # col_stats_dbs = ["local"]
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`mysql`] = `
# # Read metrics from one or many mysql servers
# [[inputs.mysql]]
#   ## specify servers via a url matching:
#   ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify|custom]]
#   ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
#   ##  e.g.
#   ##    servers = ["user:passwd@tcp(127.0.0.1:3306)/?tls=false"]
#   ##    servers = ["user@tcp(127.0.0.1:3306)/?tls=false"]
#   #
#   ## If no servers are specified, then localhost is used as the host.
#   servers = ["tcp(127.0.0.1:3306)/"]
#
#   ## Selects the metric output format.
#   ##
#   ## This option exists to maintain backwards compatibility, if you have
#   ## existing metrics do not set or change this value until you are ready to
#   ## migrate to the new format.
#   ##
#   ## If you do not have existing metrics from this plugin set to the latest
#   ## version.
#   ##
#   ## Telegraf >=1.6: metric_version = 2
#   ##           <1.6: metric_version = 1 (or unset)
#   metric_version = 2
#
#   ## the limits for metrics form perf_events_statements
#   perf_events_statements_digest_text_limit  = 120
#   perf_events_statements_limit              = 250
#   perf_events_statements_time_limit         = 86400
#   #
#   ## if the list is empty, then metrics are gathered from all databasee tables
#   table_schema_databases                    = []
#   #
#   ## gather metrics from INFORMATION_SCHEMA.TABLES for databases provided above list
#   gather_table_schema                       = false
#   #
#   ## gather thread state counts from INFORMATION_SCHEMA.PROCESSLIST
#   gather_process_list                       = true
#   #
#   ## gather user statistics from INFORMATION_SCHEMA.USER_STATISTICS
#   gather_user_statistics                    = true
#   #
#   ## gather auto_increment columns and max values from information schema
#   gather_info_schema_auto_inc               = true
#   #
#   ## gather metrics from INFORMATION_SCHEMA.INNODB_METRICS
#   gather_innodb_metrics                     = true
#   #
#   ## gather metrics from SHOW SLAVE STATUS command output
#   gather_slave_status                       = true
#   #
#   ## gather metrics from SHOW BINARY LOGS command output
#   gather_binary_logs                        = false
#   #
#   ## gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_TABLE
#   gather_table_io_waits                     = false
#   #
#   ## gather metrics from PERFORMANCE_SCHEMA.TABLE_LOCK_WAITS
#   gather_table_lock_waits                   = false
#   #
#   ## gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_INDEX_USAGE
#   gather_index_io_waits                     = false
#   #
#   ## gather metrics from PERFORMANCE_SCHEMA.EVENT_WAITS
#   gather_event_waits                        = false
#   #
#   ## gather metrics from PERFORMANCE_SCHEMA.FILE_SUMMARY_BY_EVENT_NAME
#   gather_file_events_stats                  = false
#   #
#   ## gather metrics from PERFORMANCE_SCHEMA.EVENTS_STATEMENTS_SUMMARY_BY_DIGEST
#   gather_perf_events_statements             = false
#   #
#   ## Some queries we may want to run less often (such as SHOW GLOBAL VARIABLES)
#   interval_slow                   = "30m"
#
#   ## Optional TLS Config (will be used if tls=custom parameter specified in server uri)
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`nats`] = `
# # Provides metrics about the state of a NATS server
# [[inputs.nats]]
#   ## The address of the monitoring endpoint of the NATS server
#   server = "http://localhost:8222"
#
#   ## Maximum time to receive response
#   # response_timeout = "5s"
`

	telegrafCfgSamples[`network`] = `
# # Read metrics about network interface usage
# [[inputs.net]]
#   ## By default, telegraf gathers stats from any up interface (excluding loopback)
#   ## Setting interfaces will tell it to gather these explicit interfaces,
#   ## regardless of status.
#   ##
#   # interfaces = ["eth0"]
#   ##
#   ## On linux systems telegraf also collects protocol stats.
#   ## Setting ignore_protocol_stats to true will skip reporting of protocol metrics.
#   ##
#   # ignore_protocol_stats = false
#   ##
`

	telegrafCfgSamples[`tcp_udp_check`] = `
# # Collect response time of a TCP or UDP connection
# [[inputs.net_response]]
#   ## Protocol, must be "tcp" or "udp"
#   ## NOTE: because the "udp" protocol does not respond to requests, it requires
#   ## a send/expect string pair (see below).
#   protocol = "tcp"
#   ## Server address (default localhost)
#   address = "localhost:80"
#
#   ## Set timeout
#   # timeout = "1s"
#
#   ## Set read timeout (only used if expecting a response)
#   # read_timeout = "1s"
#
#   ## The following options are required for UDP checks. For TCP, they are
#   ## optional. The plugin will send the given string to the server and then
#   ## expect to receive the given 'expect' string back.
#   ## string sent to the server
#   # send = "ssh"
#   ## expected string in answer
#   # expect = "ssh"
#
#   ## Uncomment to remove deprecated fields
#   # fielddrop = ["result_type", "string_found"]


# # Read TCP metrics such as established, time wait and sockets counts.
# [[inputs.netstat]]
#   # no configuration
`

	telegrafCfgSamples[`apache`] = `
# # Read Apache status information (mod_status)
# [[inputs.apache]]
#   ## An array of URLs to gather from, must be directed at the machine
#   ## readable version of the mod_status page including the auto query string.
#   ## Default is "http://localhost/server-status?auto".
#   urls = ["http://localhost/server-status?auto"]
#
#   ## Credentials for basic HTTP authentication.
#   # username = "myuser"
#   # password = "mypassword"
#
#   ## Maximum time to receive response.
#   # response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`nginx`] = `
# # Read Nginx's basic status information (ngx_http_stub_status_module)
# [[inputs.nginx]]
#   # An array of Nginx stub_status URI to gather stats.
#   urls = ["http://localhost/server_status"]
#
#   ## Optional TLS Config
#   tls_ca = "/etc/telegraf/ca.pem"
#   tls_cert = "/etc/telegraf/cert.cer"
#   tls_key = "/etc/telegraf/key.key"
#   ## Use TLS but skip chain & host verification
#   insecure_skip_verify = false
#
#   # HTTP response timeout (default: 5s)
#   response_timeout = "5s"


# # Read Nginx Plus' full status information (ngx_http_status_module)
# [[inputs.nginx_plus]]
#   ## An array of ngx_http_status_module or status URI to gather stats.
#   urls = ["http://localhost/status"]
#
#   # HTTP response timeout (default: 5s)
#   response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false


# # Read Nginx Plus Api documentation
# [[inputs.nginx_plus_api]]
#   ## An array of API URI to gather stats.
#   urls = ["http://localhost/api"]
#
#   # Nginx API version, default: 3
#   # api_version = 3
#
#   # HTTP response timeout (default: 5s)
#   response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false


# # Read nginx_upstream_check module status information (https://github.com/yaoweibin/nginx_upstream_check_module)
# [[inputs.nginx_upstream_check]]
#   ## An URL where Nginx Upstream check module is enabled
#   ## It should be set to return a JSON formatted response
#   url = "http://127.0.0.1/status?format=json"
#
#   ## HTTP method
#   # method = "GET"
#
#   ## Optional HTTP headers
#   # headers = {"X-Special-Header" = "Special-Value"}
#
#   ## Override HTTP "Host" header
#   # host_header = "check.example.com"
#
#   ## Timeout for HTTP requests
#   timeout = "5s"
#
#   ## Optional HTTP Basic Auth credentials
#   # username = "username"
#   # password = "pa$$word"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false


# # Read Nginx virtual host traffic status module information (nginx-module-vts)
# [[inputs.nginx_vts]]
#   ## An array of ngx_http_status_module or status URI to gather stats.
#   urls = ["http://localhost/status"]
#
#   ## HTTP response timeout (default: 5s)
#   response_timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`postgresql`] = `
# # Read metrics from one or many postgresql servers
# [[inputs.postgresql]]
#   ## specify address via a url matching:
#   ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
#   ##       ?sslmode=[disable|verify-ca|verify-full]
#   ## or a simple string:
#   ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
#   ##
#   ## All connection parameters are optional.
#   ##
#   ## Without the dbname parameter, the driver will default to a database
#   ## with the same name as the user. This dbname is just for instantiating a
#   ## connection with the server and doesn't restrict the databases we are trying
#   ## to grab metrics for.
#   ##
#   address = "host=localhost user=postgres sslmode=disable"
#   ## A custom name for the database that will be used as the "server" tag in the
#   ## measurement output. If not specified, a default one generated from
#   ## the connection address is used.
#   # outputaddress = "db01"
#
#   ## connection configuration.
#   ## maxlifetime - specify the maximum lifetime of a connection.
#   ## default is forever (0s)
#   max_lifetime = "0s"
#
#   ## A  list of databases to explicitly ignore.  If not specified, metrics for all
#   ## databases are gathered.  Do NOT use with the 'databases' option.
#   # ignored_databases = ["postgres", "template0", "template1"]
#
#   ## A list of databases to pull metrics about. If not specified, metrics for all
#   ## databases are gathered.  Do NOT use with the 'ignored_databases' option.
#   # databases = ["app_production", "testing"]


# # Read metrics from one or many postgresql servers
# [[inputs.postgresql_extensible]]
#   ## specify address via a url matching:
#   ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
#   ##       ?sslmode=[disable|verify-ca|verify-full]
#   ## or a simple string:
#   ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
#   #
#   ## All connection parameters are optional.  #
#   ## Without the dbname parameter, the driver will default to a database
#   ## with the same name as the user. This dbname is just for instantiating a
#   ## connection with the server and doesn't restrict the databases we are trying
#   ## to grab metrics for.
#   #
#   address = "host=localhost user=postgres sslmode=disable"
#
#   ## connection configuration.
#   ## maxlifetime - specify the maximum lifetime of a connection.
#   ## default is forever (0s)
#   max_lifetime = "0s"
#
#   ## A list of databases to pull metrics about. If not specified, metrics for all
#   ## databases are gathered.
#   ## databases = ["app_production", "testing"]
#   #
#   ## A custom name for the database that will be used as the "server" tag in the
#   ## measurement output. If not specified, a default one generated from
#   ## the connection address is used.
#   # outputaddress = "db01"
#   #
#   ## Define the toml config where the sql queries are stored
#   ## New queries can be added, if the withdbname is set to true and there is no
#   ## databases defined in the 'databases field', the sql query is ended by a
#   ## 'is not null' in order to make the query succeed.
#   ## Example :
#   ## The sqlquery : "SELECT * FROM pg_stat_database where datname" become
#   ## "SELECT * FROM pg_stat_database where datname IN ('postgres', 'pgbench')"
#   ## because the databases variable was set to ['postgres', 'pgbench' ] and the
#   ## withdbname was true. Be careful that if the withdbname is set to false you
#   ## don't have to define the where clause (aka with the dbname) the tagvalue
#   ## field is used to define custom tags (separated by commas)
#   ## The optional "measurement" value can be used to override the default
#   ## output measurement name ("postgresql").
#   #
#   ## Structure :
#   ## [[inputs.postgresql_extensible.query]]
#   ##   sqlquery string
#   ##   version string
#   ##   withdbname boolean
#   ##   tagvalue string (comma separated)
#   ##   measurement string
#   [[inputs.postgresql_extensible.query]]
#     sqlquery="SELECT * FROM pg_stat_database"
#     version=901
#     withdbname=false
#     tagvalue=""
#     measurement=""
#   [[inputs.postgresql_extensible.query]]
#     sqlquery="SELECT * FROM pg_stat_bgwriter"
#     version=901
#     withdbname=false
#     tagvalue="postgresql.stats"
`

	telegrafCfgSamples[`openldap`] = `
# # OpenLDAP cn=Monitor plugin
# [[inputs.openldap]]
#   host = "localhost"
#   port = 389
#
#   # ldaps, starttls, or no encryption. default is an empty string, disabling all encryption.
#   # note that port will likely need to be changed to 636 for ldaps
#   # valid options: "" | "starttls" | "ldaps"
#   tls = ""
#
#   # skip peer certificate verification. Default is false.
#   insecure_skip_verify = false
#
#   # Path to PEM-encoded Root certificate to use to verify server certificate
#   tls_ca = "/etc/ssl/certs.pem"
#
#   # dn/password to bind with. If bind_dn is empty, an anonymous bind is performed.
#   bind_dn = ""
#   bind_password = ""
#
#   # Reverse metric names so they sort more naturally. Recommended.
#   # This defaults to false if unset, but is set to true when generating a new config
#   reverse_metric_names = true
`

	telegrafCfgSamples[`kapacitor`] = `
# # Read Kapacitor-formatted JSON metrics from one or more HTTP endpoints
# [[inputs.kapacitor]]
#   ## Multiple URLs from which to read Kapacitor-formatted JSON
#   ## Default is "http://localhost:9092/kapacitor/v1/debug/vars".
#   urls = [
#     "http://localhost:9092/kapacitor/v1/debug/vars"
#   ]
#
#   ## Time limit for http requests
#   timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`phpfpm`] = `
# # Read metrics of phpfpm, via HTTP status page or socket
# [[inputs.phpfpm]]
#   ## An array of addresses to gather stats about. Specify an ip or hostname
#   ## with optional port and path
#   ##
#   ## Plugin can be configured in three modes (either can be used):
#   ##   - http: the URL must start with http:// or https://, ie:
#   ##       "http://localhost/status"
#   ##       "http://192.168.130.1/status?full"
#   ##
#   ##   - unixsocket: path to fpm socket, ie:
#   ##       "/var/run/php5-fpm.sock"
#   ##      or using a custom fpm status path:
#   ##       "/var/run/php5-fpm.sock:fpm-custom-status-path"
#   ##
#   ##   - fcgi: the URL must start with fcgi:// or cgi://, and port must be present, ie:
#   ##       "fcgi://10.0.0.12:9000/status"
#   ##       "cgi://10.0.10.12:9001/status"
#   ##
#   ## Example of multiple gathering from local socket and remote host
#   ## urls = ["http://192.168.1.20/status", "/tmp/fpm.sock"]
#   urls = ["http://localhost/status"]
#
#   ## Duration allowed to complete HTTP requests.
#   # timeout = "5s"
#
#   ## Optional TLS Config
#   # tls_ca = "/etc/telegraf/ca.pem"
#   # tls_cert = "/etc/telegraf/cert.pem"
#   # tls_key = "/etc/telegraf/key.pem"
#   ## Use TLS but skip chain & host verification
#   # insecure_skip_verify = false
`

	telegrafCfgSamples[`sqlserver`] = `
# # Read metrics from Microsoft SQL Server
# [[inputs.sqlserver]]
#   ## Specify instances to monitor with a list of connection strings.
#   ## All connection parameters are optional.
#   ## By default, the host is localhost, listening on default port, TCP 1433.
#   ##   for Windows, the user is the currently running AD user (SSO).
#   ##   See https://github.com/denisenkom/go-mssqldb for detailed connection
#   ##   parameters.
#   # servers = [
#   #  "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;app name=telegraf;log=1;",
#   # ]
#
#   ## Optional parameter, setting this to 2 will use a new version
#   ## of the collection queries that break compatibility with the original
#   ## dashboards.
#   query_version = 2
#
#   ## If you are using AzureDB, setting this to true will gather resource utilization metrics
#   # azuredb = false
#
#   ## If you would like to exclude some of the metrics queries, list them here
#   ## Possible choices:
#   ## - PerformanceCounters
#   ## - WaitStatsCategorized
#   ## - DatabaseIO
#   ## - DatabaseProperties
#   ## - CPUHistory
#   ## - DatabaseSize
#   ## - DatabaseStats
#   ## - MemoryClerk
#   ## - VolumeSpace
#   ## - PerformanceMetrics
#   ## - Schedulers
#   ## - AzureDBResourceStats
#   ## - AzureDBResourceGovernance
#   ## - SqlRequests
#   exclude_query = [ 'Schedulers' ]
`

	telegrafCfgSamples[`memcached`] = `
# # Read metrics from one or many memcached servers
# [[inputs.memcached]]
#   ## An array of address to gather stats about. Specify an ip on hostname
#   ## with optional port. ie localhost, 10.0.0.1:11211, etc.
#   servers = ["localhost:11211"]
#   # unix_sockets = ["/var/run/memcached.sock"]
`

	telegrafCfgSamples[`ping_check`] = `
# # Ping given url(s) and return statistics
# [[inputs.ping]]
#   ## List of urls to ping
#   urls = ["example.org"]
#
#   ## Number of pings to send per collection (ping -c <COUNT>)
#   # count = 1
#
#   ## Interval, in s, at which to ping. 0 == default (ping -i <PING_INTERVAL>)
#   # ping_interval = 1.0
#
#   ## Per-ping timeout, in s. 0 == no timeout (ping -W <TIMEOUT>)
#   # timeout = 1.0
#
#   ## Total-ping deadline, in s. 0 == no deadline (ping -w <DEADLINE>)
#   # deadline = 10
#
#   ## Interface or source address to send ping from (ping -I[-S] <INTERFACE/SRC_ADDR>)
#   # interface = ""
#
#   ## How to ping. "native" doesn't have external dependencies, while "exec" depends on 'ping'.
#   # method = "exec"
#
#   ## Specify the ping executable binary, default is "ping"
# 	# binary = "ping"
#
#   ## Arguments for ping command. When arguments is not empty, system binary will be used and
#   ## other options (ping_interval, timeout, etc) will be ignored.
#   # arguments = ["-c", "3"]
#
#   ## Use only ipv6 addresses when resolving hostnames.
#   # ipv6 = false
`
}
