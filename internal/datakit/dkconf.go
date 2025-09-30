// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

import "fmt"

// MainConfSample get default sample of datakit.conf.
func MainConfSample(domain string) string {
	return fmt.Sprintf(datakitConfSample, domain)
}

var (
	BrandDomainTemplate = `<<<custom_key.brand_main_domain>>>`
	BrandDomain         = "not-set" // we should set it during compile time
	datakitConfSample   = `

################################################
# Global configures
################################################
# Default enabled input list.
default_enabled_inputs = [
  "cpu",
  "disk",
  "diskio",
  "host_processes",
  "hostobject",
  "mem",
  "net",
  "swap",
  "system",
]

# enable_pprof: bool
# If pprof enabled, we can profiling the running datakit
enable_pprof = true
pprof_listen = "localhost:6060" # pprof listen

# protect_mode: bool, default false
# When protect_mode eanbled, we can set radical collect parameters, these may cause Datakit
# collect data more frequently.
protect_mode = true

# The user name running datakit. Generally for audit purpose. Default is root.
datakit_user = "root"

################################################
# ulimit: set max open-files limit(Linux only)
################################################
ulimit = 64000

################################################
# point_pool: use point pool for better memory usage
################################################
[point_pool]
  enable = false
  reserved_capacity = 4096

################################################
# DCA configure
################################################
[dca]
  # Enable or disable DCA
  enable = false

  # DCA websocket server address
  websocket_server = "ws://localhost:8000/ws"

################################################
# Upgrader 
################################################
[dk_upgrader]
  # host address
  host = "0.0.0.0"

  # port number
  port = 9542 

################################################
# Pipeline
################################################
[pipeline]
  # IP database type, support iploc and geolite2
  ipdb_type = "iploc"

  # How often to sync remote pipeline
  remote_pull_interval = "1m"

  #
  # reftab configures
  #
  # Reftab remote HTTP URL(https/http)
  refer_table_url = ""

  # How often reftab sync the remote
  refer_table_pull_interval = "5m"

  # use sqlite to store reftab data to release memory usage
  use_sqlite = false
  # or use pure memory to cache the reftab data
  sqlite_mem_mode = false

  # append run info
  disable_append_run_info = false

  # default pipeline
  [pipeline.default_pipeline]
    # logging = "<your_script.p>"
    # metric  = "<your_script.p>"
    # tracing = "<your_script.p>"

  # Offload data processing tasks to post-level data processors.
  [pipeline.offload]
    receiver = "datakit-http"
    addresses = [
      # "http://<ip>:<port>"
    ]

################################################
# HTTP server(9529)
################################################
[http_api]

  # HTTP server address
  listen = "localhost:9529"

  # Disable 404 page to hide detailed Datakit info
  disable_404page = false

  # only enable these APIs. If list empty, all APIs are enabled.
  public_apis = []

  # Datakit server-side timeout
  timeout = "30s"
  close_idle_connection = false

  # API rate limit(QPS)
  request_rate_limit       = 100.0
  request_rate_limit_ttl   = "60s"
  request_rate_limit_burst = 500

  #
  # RUM related: we should port these configures to RUM inputs(TODO)
  #
  # When serving RUM(/v1/write/rum), extract the IP address from this HTTP header
  rum_origin_ip_header = "X-Forwarded-For"
  # When serving RUM(/v1/write/rum), only accept requests from these app-id.
  # If the list empty, all app's requests accepted.
  rum_app_id_white_list = []

  # only these domains enable CORS. If list empty, all domains are enabled.
  allowed_cors_origins = []

  # Start Datakit web server with HTTPS
  [http_api.tls]
    # cert = "path/to/certificate/file"
    # privkey = "path/to/private_key/file"

################################################
# io configures
################################################
[io]
  # How often Datakit flush data to dataway.
  # Datakit will upload data points if cached(in memory) points
  #  reached(>=) the max_cache_count or the flush_interval triggered.
  max_cache_count = 1000
  flush_workers   = 0 # default to (cpu_core * 2)
  flush_interval  = "10s"

  # Queue size of feed.
  feed_chan_size = 1

  # Set blocking if queue is full.
  # NOTE: Global blocking mode may consume more memory on large metric points.
  global_blocking = false

  # auto correct point's time.
  auto_timestamp_correction = true

  # Data point filter configures.
  # NOTE: Most of the time, you should use web-side filter, it's a debug helper for developers.
  #[io.filters]
  #  logging = [
  #   "{ source = 'datakit' or f1 IN [ 1, 2, 3] }"
  #  ]
  #  metric = [
  #    "{ measurement IN ['datakit', 'disk'] }",
  #    "{ measurement CONTAIN ['host.*', 'swap'] }",
  #  ]
  #  object = [
  #    { class CONTAIN ['host_.*'] }",
  #  ]
  #  tracing = [
  #    "{ service = re("abc.*") AND some_tag CONTAIN ['def_.*'] }",
  #  ]

[recorder]
  enabled = false
  #path = "/path/to/point-data/dir"
  encoding = "v2"  # use protobuf-json format
  duration = "30m" # record for 30 minutes

  # only record these inputs, if empty, record all
  inputs = [
    #"cpu",
    #"mem",
  ]

  # only record these categoris, if empty, record all
  category = [
    #"logging",
    #"object",
  ]

################################################
# Dataway configure
################################################
[dataway]
  # urls: Dataway URL list
  # NOTE: do not configure multiple URLs here, it's a deprecated feature.
  urls = [
    # "https://openway.%s?token=<YOUR-WORKSPACE-TOKEN>"
  ]

  # Dataway HTTP timeout
  timeout_v2 = "30s"

  # max_retry_count specifies at most how many times will be tried when dataway API fails(not 4xx),
  # default value(and minimal) is 1 and maximum value is 10.
  #
  # The default set to 1 to makes the API fails ASAP to release memroy.
  max_retry_count = 1

  # The interval between two retry operation, valid time units are "ns", "us", "ms", "s", "m", "h"
  retry_delay = "1s"

  # HTTP Proxy
  # Format: "http(s)://IP:Port"
  http_proxy = ""

  max_idle_conns   = 0       # limit idle TCP connections for HTTP request to Dataway
  enable_httptrace = false   # enable trace HTTP metrics(connection/NDS/TLS and so on)
  idle_timeout     = "90s"   # not-set, default 90s

  # HTTP body content type, other candidates are(case insensitive):
  #  - v1: line-protocol
  #  - v2: protobuf
  content_encoding = "v2"

  # Enable GZip to upload point data.
  #
  # do NOT disable gzip or your get large network payload.
  gzip = true

  max_raw_body_size = 1048576 # max body size(before gizp) in bytes

  # Customer tag or field keys that will extract from exist points
  # to build the X-Global-Tags HTTP header value.
  global_customer_keys = []
  enable_sinker        = false # disable sinker

  # use dataway as NTP server
  [dataway.ntp]
    enable   = true
    interval = "5m"  # sync dataway time each 5min(minimal 1min)

    # if abs(datakit time - dataway time) >= diff, datakit will adjust data point
    # time with dataway time.
    diff     = "30s"  # minimal 5s

  # WAL queue for uploading points
  [dataway.wal]
    max_capacity_gb = 2.0 # 2GB reserved disk space for each category(M/L/O/T/...)
    #workers = 4          # flush workers on WAL(default to CPU limited cores)
    #mem_cap = 4          # in-memory queue capacity(default to CPU limited cores)
    #fail_cache_clean_interval = "30s" # duration for clean fail uploaded data
    #no_drop_categories = ["L"]        # category list that disable drop data when disk cache full


################################################
# Datakit logging configure
################################################
[logging]

  # log path
  log = "/var/log/datakit/log"

  # HTTP access log
  gin_log = "/var/log/datakit/gin.log"

  # level level(info/debug)
  level = "info"

  # Disable log color
  disable_color = false

  # log rotate size(in MB)
  # DataKit will always keep at most n+1(n backup log and 1 writing log) splited log files on disk.
  rotate = 32

  # Upper limit count of backup log
  rotate_backups = 5

################################################
# Global tags
################################################
# We will try to add these tags to every collected data point if these
# tags do not exist in orignal data.
#
# NOTE: we can get the real IP of current note, we just need
# to set "$datakit_ip" or "__datakit_ip" here. Same for the hostname.
[global_host_tags]
  ip   = "$datakit_ip"
  host = "$datakit_hostname"

[election]
  # Enable election
  enable = false

  # Election whitelist
  # NOTE: Empty to disable whitelist
  node_whitelist = []

  # Election namespace.
  # NOTE: for single workspace, there can be multiple election namespace.
  namespace = "default"
  
  # If enabled, every data point will add a tag with election_namespace = <your-election-namespace>
  enable_namespace_tag = false

  # Like global_host_tags, but only for data points that are remotely collected(such as MySQL/Nginx).
  [election.tags]
    #  project = "my-project"
    #  cluster = "my-cluster"

###################################################
# Tricky: we can rename the default hostname here
###################################################
[environments]
  ENV_HOSTNAME = ""

################################################
# resource limit configures
################################################
[resource_limit]

  # enable or disable resource limit
  enable = true

  # Linux only, cgroup path
  path = "/datakit"

  # Limit CPU cores
  cpu_cores = 2.0

  # set max memory usage(MB)
  mem_max_mb = 4096

################################################
# git_repos configures
################################################

# We can hosting all input configures on git server
[git_repos]
  # git pull interval
  pull_interval = "1m"

  # git repository settings
  [[git_repos.repo]]
    # enable the repository or not
    enable = false

    # the branch name to pull
    branch = "master"

    # git repository URL. There are 3 formats here:
    #   - HTTP(s): such as "https://github.datakit.com/path/to/datakit-conf.git"
    #   - Git: such as "git@github.com:path/to/datakit.git"
    #   - SSH: such as "ssh://git@github.com:9000/path/to/repository.git"
    url = ""

    # For formats Git and SSH, we need extra configures:
    ssh_private_key_path = ""
    ssh_private_key_password = ""

################################################
# crypto key or key filePath.
################################################
[crypto]
  aes_key = ""
  aes_Key_file = ""

[remote_job]
  enable=false
  envs = ["OSS_BUCKET_HOST=host","OSS_ACCESS_KEY_ID=key","OSS_ACCESS_KEY_SECRET=secret","OSS_BUCKET_NAME=bucket"]
  interval = "30s"
  java_home=""
`
)
