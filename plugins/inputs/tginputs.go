package inputs

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/toml"

	// telegraf plugins
	"github.com/influxdata/telegraf/plugins/inputs/activemq"
	"github.com/influxdata/telegraf/plugins/inputs/amqp_consumer"
	"github.com/influxdata/telegraf/plugins/inputs/apache"
	"github.com/influxdata/telegraf/plugins/inputs/ceph"
	"github.com/influxdata/telegraf/plugins/inputs/clickhouse"
	"github.com/influxdata/telegraf/plugins/inputs/cloudwatch"
	"github.com/influxdata/telegraf/plugins/inputs/cpu"
	"github.com/influxdata/telegraf/plugins/inputs/disk"
	"github.com/influxdata/telegraf/plugins/inputs/diskio"
	"github.com/influxdata/telegraf/plugins/inputs/dns_query"
	"github.com/influxdata/telegraf/plugins/inputs/docker"
	"github.com/influxdata/telegraf/plugins/inputs/docker_log"
	"github.com/influxdata/telegraf/plugins/inputs/elasticsearch"
	"github.com/influxdata/telegraf/plugins/inputs/exec"
	"github.com/influxdata/telegraf/plugins/inputs/fluentd"
	"github.com/influxdata/telegraf/plugins/inputs/github"
	"github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/inputs/influxdb"
	"github.com/influxdata/telegraf/plugins/inputs/jenkins"
	"github.com/influxdata/telegraf/plugins/inputs/jolokia2"
	"github.com/influxdata/telegraf/plugins/inputs/kafka_consumer"
	"github.com/influxdata/telegraf/plugins/inputs/kapacitor"
	"github.com/influxdata/telegraf/plugins/inputs/kibana"
	"github.com/influxdata/telegraf/plugins/inputs/kube_inventory"
	"github.com/influxdata/telegraf/plugins/inputs/kubernetes"
	"github.com/influxdata/telegraf/plugins/inputs/mem"
	"github.com/influxdata/telegraf/plugins/inputs/memcached"
	"github.com/influxdata/telegraf/plugins/inputs/modbus"
	"github.com/influxdata/telegraf/plugins/inputs/mongodb"
	"github.com/influxdata/telegraf/plugins/inputs/mqtt_consumer"
	"github.com/influxdata/telegraf/plugins/inputs/mysql"
	"github.com/influxdata/telegraf/plugins/inputs/nats"
	"github.com/influxdata/telegraf/plugins/inputs/net"
	"github.com/influxdata/telegraf/plugins/inputs/net_response"
	"github.com/influxdata/telegraf/plugins/inputs/nginx"
	"github.com/influxdata/telegraf/plugins/inputs/nsq"
	"github.com/influxdata/telegraf/plugins/inputs/nsq_consumer"
	"github.com/influxdata/telegraf/plugins/inputs/ntpq"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi"
	"github.com/influxdata/telegraf/plugins/inputs/openldap"
	"github.com/influxdata/telegraf/plugins/inputs/openntpd"
	"github.com/influxdata/telegraf/plugins/inputs/ping"
	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
	"github.com/influxdata/telegraf/plugins/inputs/rabbitmq"
	"github.com/influxdata/telegraf/plugins/inputs/redis"
	"github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/plugins/inputs/socket_listener"
	"github.com/influxdata/telegraf/plugins/inputs/solr"
	"github.com/influxdata/telegraf/plugins/inputs/sqlserver"
	"github.com/influxdata/telegraf/plugins/inputs/swap"
	"github.com/influxdata/telegraf/plugins/inputs/syslog"
	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/plugins/inputs/tengine"
	"github.com/influxdata/telegraf/plugins/inputs/uwsgi"
	"github.com/influxdata/telegraf/plugins/inputs/vsphere"
	"github.com/influxdata/telegraf/plugins/inputs/x509_cert"
	"github.com/influxdata/telegraf/plugins/inputs/zookeeper"
)

type TelegrafInput struct {
	input   telegraf.Input
	name    string
	Catalog string
	Sample  string
}

func (ti *TelegrafInput) SampleConfig() string {
	// telegraf not exported inputs, return sample directly(if configured in init())
	if ti.input == nil {
		return ti.Sample
	}

	return fmt.Sprintf("[[inputs.%s]]\n%s", ti.name, ti.input.SampleConfig())
}

func (ti *TelegrafInput) Enabled() (n int, cfgs []string) {

	mtx.RLock()
	defer mtx.RUnlock()

	arr, ok := inputInfos[ti.name]
	if !ok {
		return
	}

	for _, i := range arr {
		cfgs = append(cfgs, i.cfg)
	}
	n = len(arr)
	return
}

func CheckTelegrafToml(inputName string, tomlcfg []byte) error {
	switch inputName {
	case `cpu`:
		var i cpu.CPUStats
		if err := toml.Unmarshal(tomlcfg, &i); err != nil {
			return err
		}
	default:
		// TODO
	}

	return nil
}

var (
	TelegrafInputs = map[string]*TelegrafInput{ // Name: Catalog

		"disk":   {name: "disk", Catalog: "host", input: &disk.DiskStats{}},
		"diskio": {name: "diskio", Catalog: "host", input: &diskio.DiskIO{}},
		"mem":    {name: "mem", Catalog: "host", input: &mem.MemStats{}},
		"swap":   {name: "swap", Catalog: "host", input: &swap.SwapStats{}},
		"system": {name: "system", Catalog: "host", input: &system.SystemStats{}},
		"cpu":    {name: "cpu", Catalog: "host", input: &cpu.CPUStats{}},

		"nvidia_smi": {name: "nvidia_smi", Catalog: "nvidia", input: &nvidia_smi.NvidiaSMI{}},

		"ping":            {name: "ping", Catalog: "network", input: &ping.Ping{}},
		"net":             {name: "net", Catalog: "network", input: &net.NetIOStats{}},
		"net_response":    {name: "net_response", Catalog: "network", input: &net_response.NetResponse{}},
		"http":            {name: "http", Catalog: "network", input: &http.HTTP{}},
		"socket_listener": {name: "socket_listener", Catalog: "network", input: &socket_listener.SocketListener{}},

		"nginx":   {name: "nginx", Catalog: "nginx", input: &nginx.Nginx{}},
		"tengine": {name: "tengine", Catalog: "tengine", input: &tengine.Tengine{}},
		"apache":  {name: "apache", Catalog: "apache", input: &apache.Apache{}},

		"mysql":         {name: "mysql", Catalog: "db", input: &mysql.Mysql{}},
		"postgresql":    {name: "postgresql", Catalog: "db", input: &postgresql.Postgresql{}},
		"mongodb":       {name: "mongodb", Catalog: "db", input: &mongodb.MongoDB{}},
		"redis":         {name: "redis", Catalog: "db", input: &redis.Redis{}},
		"elasticsearch": {name: "elasticsearch", Catalog: "db", input: &elasticsearch.Elasticsearch{}},
		"sqlserver":     {name: "sqlserver", Catalog: "db", input: &sqlserver.SQLServer{}},
		"memcached":     {name: "memcached", Catalog: "db", input: &memcached.Memcached{}},
		"solr":          {name: "solr", Catalog: "db", input: &solr.Solr{}},
		"clickhouse":    {name: "clickhouse", Catalog: "db", input: &clickhouse.ClickHouse{}},
		`influxdb`:      {name: "influxdb", Catalog: "db", input: &influxdb.InfluxDB{}},

		"openldap":  {name: "openldap", Catalog: "openldap", input: &openldap.Openldap{}},
		"phpfpm":    {name: "phpfpm", Catalog: "phpfpm", input: nil}, // telegraf not exported
		"zookeeper": {name: "zookeeper", Catalog: "zookeeper", input: &zookeeper.Zookeeper{}},
		"ceph":      {name: "ceph", Catalog: "ceph", input: &ceph.Ceph{}},
		"dns_query": {name: "dns_query", Catalog: "dns_query", input: &dns_query.DnsQuery{}},

		"docker":     {name: "docker", Catalog: "docker", input: &docker.Docker{}},
		"docker_log": {name: "docker_log", Catalog: "docker", input: &docker_log.DockerLogs{}},

		"activemq":       {name: "activemq", Catalog: "activemq", input: &activemq.ActiveMQ{}},
		"rabbitmq":       {name: "rabbitmq", Catalog: "rabbitmq", input: &rabbitmq.RabbitMQ{}},
		"nsq":            {name: "nsq", Catalog: "nsq", input: &nsq.NSQ{}},
		"nsq_consumer":   {name: "nsq_consumer", Catalog: "nsq", input: &nsq_consumer.NSQConsumer{}},
		"kafka_consumer": {name: "kafka_consumer", Catalog: "kafka", input: &kafka_consumer.KafkaConsumer{}},
		"mqtt_consumer":  {name: "mqtt_consumer", Catalog: "mqtt", input: &mqtt_consumer.MQTTConsumer{}},

		"fluentd":    {name: "fluentd", Catalog: "fluentd", input: &fluentd.Fluentd{}},
		"haproxy":    {name: "haproxy", Catalog: "haproxy", input: nil}, // telegraf not exported
		"jenkins":    {name: "jenkins", Catalog: "jenkins", input: &jenkins.Jenkins{}},
		"kapacitor":  {name: "kapacitor", Catalog: "kapacitor", input: &kapacitor.Kapacitor{}},
		"ntpq":       {name: "ntpq", Catalog: "ntpq", input: &ntpq.NTPQ{}},
		"openntpd":   {name: "openntpd", Catalog: "openntpd", input: &openntpd.Openntpd{}},
		"x509_cert":  {name: "x509_cert", Catalog: "tls", input: &x509_cert.X509Cert{}},
		"nats":       {name: "nats", Catalog: "nats", input: &nats.Nats{}},
		"cloudwatch": {name: "cloudwatch", Catalog: "aws", input: &cloudwatch.CloudWatch{}},
		"vsphere":    {name: "vsphere", Catalog: "vmware", input: &vsphere.VSphere{}},
		"snmp":       {name: "snmp", Catalog: "snmp", input: &snmp.Snmp{}},
		"exec":       {name: "exec", Catalog: "exec", input: &exec.Exec{}},
		"syslog":     {name: "syslog", Catalog: "syslog", input: &syslog.Syslog{}},

		"kube_inventory": {name: "kube_inventory", Catalog: "k8s", input: &kube_inventory.KubernetesInventory{}},
		"kubernetes":     {name: "kubernetes", Catalog: "k8s", input: &kubernetes.Kubernetes{}},

		"jolokia2_agent": {name: "jolokia2_agent", Catalog: "jolokia2_agent", input: &jolokia2.JolokiaAgent{}},
		"amqp_consumer":  {name: "amqp_consumer", Catalog: "amqp", input: &amqp_consumer.AMQPConsumer{}},
		"github":         {name: "github", Catalog: "github", input: &github.GitHub{}},
		"uwsgi":          {name: "uwsgi", Catalog: "uwsgi", input: &uwsgi.Uwsgi{}},
		//`consul`:         {name: "consul", Catalog: "consul", input: &consul.Consul{}},
		`kibana`: {name: "kibana", Catalog: "kibana", input: &kibana.Kibana{}},
		`modbus`: {name: "modbus", Catalog: "modbus", input: &modbus.Modbus{}},
	}
)

func init() {
	TelegrafInputs["haproxy"].Sample = `
 # Read metrics of haproxy, via socket or csv stats page
 [[inputs.haproxy]]
   ## An array of address to gather stats about. Specify an ip on hostname
   ## with optional port. ie localhost, 10.10.3.33:1936, etc.
   ## Make sure you specify the complete path to the stats endpoint
   ## including the protocol, ie http://10.10.3.33:1936/haproxy?stats

   ## If no servers are specified, then default to 127.0.0.1:1936/haproxy?stats
   servers = ["http://myhaproxy.com:1936/haproxy?stats"]

   ## Credentials for basic HTTP authentication
   # username = "admin"
   # password = "admin"

   ## You can also use local socket with standard wildcard globbing.
   ## Server address not starting with 'http' will be treated as a possible
   ## socket, so both examples below are valid.
   # servers = ["socket:/run/haproxy/admin.sock", "/run/haproxy/*.sock"]

   ## By default, some of the fields are renamed from what haproxy calls them.
   ## Setting this option to true results in the plugin keeping the original
   ## field names.
   # keep_field_names = false

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs["phpfpm"].Sample = `
# Read metrics of phpfpm, via HTTP status page or socket
 [[inputs.phpfpm]]
   ## An array of addresses to gather stats about. Specify an ip or hostname
   ## with optional port and path
   ##
   ## Plugin can be configured in three modes (either can be used):
   ##   - http: the URL must start with http:// or https://, ie:
   ##       "http://localhost/status"
   ##       "http://192.168.130.1/status?full"
   ##
   ##   - unixsocket: path to fpm socket, ie:
   ##       "/var/run/php5-fpm.sock"
   ##      or using a custom fpm status path:
   ##       "/var/run/php5-fpm.sock:fpm-custom-status-path"
   ##
   ##   - fcgi: the URL must start with fcgi:// or cgi://, and port must be present, ie:
   ##       "fcgi://10.0.0.12:9000/status"
   ##       "cgi://10.0.10.12:9001/status"
   ##
   ## Example of multiple gathering from local socket and remote host
   ## urls = ["http://192.168.1.20/status", "/tmp/fpm.sock"]
   urls = ["http://localhost/status"]

   ## Duration allowed to complete HTTP requests.
   # timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs["consul"].Sample = `
# Gather health check statuses from services registered in Consul
	[[inputs.consul]]
		# Consul server address
		address = "localhost:8500"

		# URI scheme for the Consul server, one of "http", "https"
		scheme = "http"

		# ACL token used in every request
		token = ""

		# HTTP Basic Authentication username and password.
		username = ""
		password = ""

		# Data center to query the health checks from
		datacenter = ""

		# Optional TLS Config
		#tls_ca = "/etc/telegraf/ca.pem"
		#tls_cert = "/etc/telegraf/cert.pem"
		#tls_key = "/etc/telegraf/key.pem"
		# Use TLS but skip chain & host verification
		#insecure_skip_verify = true

		# Consul checks' tag splitting
		#When tags are formatted like "key:value" with ":" as a delimiter then
		#they will be splitted and reported as proper key:value in Telegraf
		#tag_delimiter = ":"`
}

func HaveTelegrafInputs() bool {

	mtx.RLock()
	defer mtx.RUnlock()

	for k := range TelegrafInputs {
		_, ok := inputInfos[k]
		if ok {
			return true
		}
	}

	return false
}
