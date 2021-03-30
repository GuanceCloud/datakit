package redis

import (
	"time"
	"github.com/go-redis/redis"
)

const (
	configSample = `
[[inputs.redis]]
	## @param host - string - required
    ## Enter the host to connect to.
    #
  	host = "localhost"

  	## @param port - integer - required
    ## Enter the port of the host to connect to.
    #
    port = 6379

    ## @param unix_socket_path - string - optional
    ## Connect through a unix socket instead of using a host and port.
    #
    # unix_socket_path = "/var/run/redis/redis.sock"

    ## @param db - integer - optional - default: 0
    ## The index of the database (keyspace) to use.
    ## The default is index 0. Any other index results in a SELECT command sent upon connection
    ## to choose the desired database. For redis.clone, this can also be NA in which case the same database is
    ## used as in rc.
    #
    # db = 0

    ## @param username - string - optional
    ## The username for the connection. Redis 6+ only.
    #
    # username = "<USERNAME>"

    ## @param password - string - optional
    ## The password for the connection.
    #
    # password = "<PASSWORD>"

	## @param service - string - optional
    ## Attach the tag service:<SERVICE> to every metric
    ##
    # service = "<SERVICE>"

    ## @param interval - number - optional - default: 15
    ## This changes the collection interval of the check. For more information, see:
    #
    # interval = "15s"

    ## @param collect_client_metrics - boolean - optional - default: false
    ## Collects metrics using the CLIENT command.
    ## This requires the Redis CLIENT command to be available on your servers.
    #
    # collect_client_metrics = false

    ## @param ssl - boolean - optional - default: false
    ## Enable SSL/TLS encryption for the check.
    #
    # ssl: false

    ## @param ssl_keyfile - string - optional
    ## The path to the client-side private keyfile.
    #
    # ssl_keyfile = <CERT_KEY_PATH>

    ## @param ssl_certfile - string - optional
    ## The path to the client-side certificate file.
    #
    # ssl_certfile = <CERT_PEM_PATH>

    ## @param ssl_ca_certs - string - optional
    ## The path to the ca_certs file.
    #
    # ssl_ca_certs = <CERT_PATH>

    ## @param keys - list of strings - optional
    ## Enter the list of keys to collect the lengths from.
    ## The length is 1 for strings.
    ## The length is zero for keys that have a type other than list, set, hash, or sorted set.
    ## Note: Keys can be expressed as patterns, see https://redis.io/commands/keys.
    #
    # keys:
    #   - <KEY_1>
    #   - <KEY_PATTERN>

    ## @param warn_on_missing_keys - boolean - optional - default: true
    ## If you provide a list of 'keys', set this to true to have the Agent log a warning
    ## when keys are missing.
    #
    # warn_on_missing_keys = true

    ## @param slowlog-max-len - integer - optional - default: 128
    ## Set the maximum number of entries to fetch from the slow query log.
    ## By default, the check reads this value from the redis config, but is limited to 128.
    ##
    ## Set a custom value here if you need to get more than 128 slowlog entries every 15 seconds.
    ## Warning: Higher values may impact the performance of your Redis instance.
    #
    # slowlog-max-len = 128

    ## @param command_stats - boolean - optional - default: false
    ## Collect INFO COMMANDSTATS output as metrics.
    #
    # command_stats = false

    ## @param disable_connection_cache - boolean - optional - default: false
    ## Enable the connections cache so the check attempts to reuse the same Redis connections
    ## at every collection cycle. If disabled, this prevents stale connections.
    #
    # disable_connection_cache = false

    ## @param tags - list of strings - optional
    ## A list of tags to attach to every metric and service check emitted by this instance.
    ##
    # tags:
    #   - <KEY_1> = <VALUE_1>
    #   - <KEY_2> = <VALUE_2>

    ## @param empty_default_hostname - boolean - optional - default: false
    ## This forces the check to send metrics with no hostname.
    ##
    ## This is useful for cluster-level checks.
    #
    # empty_default_hostname = false
`
)

// Redis
type Redis struct {
	Host              string
	Port              int
	UnixSocketPath    string        `toml:"unix_socket_path"`
	DB                int
	Password          string
	MetricName        string
	Service           string		`toml:"service"`
	SocketTimeout     int           `toml:"socket_timeout"`
	Interval          string        `toml:"interval"`
	IntervalDuration  time.Duration `toml:"-"`
	Keys              []string
	WarnOnMissingKeys bool          `toml:"warn_on_missing_keys"`
	SlowlogMaxLen     float64       `toml:"slowlog-max-len"`
	Tags              map[string]string `toml:"tags"`
	client           *redis.Client
	//lastTimestampSeen map[instance]int64
	resData          map[string]interface{}
}

