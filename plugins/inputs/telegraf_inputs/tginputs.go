package telegraf_inputs

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
	"github.com/influxdata/telegraf/plugins/inputs/elasticsearch"
	"github.com/influxdata/telegraf/plugins/inputs/exec"
	"github.com/influxdata/telegraf/plugins/inputs/fluentd"
	"github.com/influxdata/telegraf/plugins/inputs/github"
	"github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/inputs/http_response"
	"github.com/influxdata/telegraf/plugins/inputs/httpjson"
	"github.com/influxdata/telegraf/plugins/inputs/influxdb"
	"github.com/influxdata/telegraf/plugins/inputs/jenkins"
	"github.com/influxdata/telegraf/plugins/inputs/jolokia2"
	"github.com/influxdata/telegraf/plugins/inputs/kafka_consumer"
	"github.com/influxdata/telegraf/plugins/inputs/kapacitor"
	"github.com/influxdata/telegraf/plugins/inputs/kibana"
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
	"github.com/influxdata/telegraf/plugins/inputs/nginx_plus"
	"github.com/influxdata/telegraf/plugins/inputs/nginx_plus_api"
	"github.com/influxdata/telegraf/plugins/inputs/nginx_upstream_check"
	"github.com/influxdata/telegraf/plugins/inputs/nginx_vts"
	"github.com/influxdata/telegraf/plugins/inputs/nsq"
	"github.com/influxdata/telegraf/plugins/inputs/nsq_consumer"
	"github.com/influxdata/telegraf/plugins/inputs/ntpq"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi"
	"github.com/influxdata/telegraf/plugins/inputs/openldap"
	"github.com/influxdata/telegraf/plugins/inputs/openntpd"
	"github.com/influxdata/telegraf/plugins/inputs/ping"
	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
	"github.com/influxdata/telegraf/plugins/inputs/processes"
	"github.com/influxdata/telegraf/plugins/inputs/procstat"
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
	//"github.com/influxdata/telegraf/plugins/inputs/consul" // get ambiguous import error
	//"github.com/influxdata/telegraf/plugins/inputs/phpfpm" // not exported
	//"github.com/influxdata/telegraf/plugins/inputs/haproxy" // not exported
	//"github.com/influxdata/telegraf/plugins/inputs/kube_inventory" // runtime crash
)

type TelegrafInput struct {
	input   telegraf.Input
	name    string
	Catalog string

	// For some telegraf inputs, there are reasons to use specific Sample:
	//  - for telegraf source code that can't build with exists datakit (package conflict)
	//  - for telegraf source code that not exported
	//  - some of telegraf inputs use the same telegraf input, i.e., jolokia2_agent, win_perf_counters
	// so we can't get sample from input.SampleConfig(), we just return the Sample field
	//
	// Bad news: it's hard to sync upstream telegraf updates on these inputs
	Sample string
}

func (ti *TelegrafInput) SampleConfig() string {

	if ti.Sample != "" { // prefer specified sample
		return ti.Sample
	}

	// telegraf not exported inputs, return sample directly(if configured in init())
	if ti.input == nil {
		panic(fmt.Errorf("%s should have a specific config-sample", ti.name))
	}

	return fmt.Sprintf("[[inputs.%s]]\n%s", ti.name, ti.input.SampleConfig())
}

var (
	TelegrafInputs = map[string]*TelegrafInput{ // Name: Catalog

		"disk":      {name: "disk", Catalog: "host", input: &disk.DiskStats{}},
		"diskio":    {name: "diskio", Catalog: "host", input: &diskio.DiskIO{}},
		"mem":       {name: "mem", Catalog: "host", input: &mem.MemStats{}},
		"swap":      {name: "swap", Catalog: "host", input: &swap.SwapStats{}},
		"system":    {name: "system", Catalog: "host", input: &system.SystemStats{}},
		"cpu":       {name: "cpu", Catalog: "host", input: &cpu.CPUStats{}},
		"procstat":  {name: "procstat", Catalog: "host", input: &procstat.Procstat{}},
		"processes": {name: "processes", Catalog: "host", input: &processes.Processes{}},

		"internal": {name: "internal", Catalog: "internal", Sample: samples["internal"], input: nil}, // import internal package not allowed

		"ping":            {name: "ping", Catalog: "network", input: &ping.Ping{}},
		"net":             {name: "net", Catalog: "network", input: &net.NetStats{}},
		"netstat":         {name: "netstat", Catalog: "network", input: &net.NetIOStats{}},
		"net_response":    {name: "net_response", Catalog: "network", input: &net_response.NetResponse{}},
		"http":            {name: "http", Catalog: "network", input: &http.HTTP{}},
		"http_response":   {name: "http_response", Catalog: "network", input: &http_response.HTTPResponse{}},
		"httpjson":        {name: "httpjson", Catalog: "network", input: &httpjson.HttpJson{}},
		"socket_listener": {name: "socket_listener", Catalog: "network", input: &socket_listener.SocketListener{}},

		// collectd use socket_listener to gather data
		"collectd": {name: "socket_listener", Catalog: "collectd", input: &socket_listener.SocketListener{}},

		"nginx":                {name: "nginx", Catalog: "nginx", input: &nginx.Nginx{}},
		"nginx_upstream_check": {name: "nginx_upstream_check", Catalog: "nginx", input: &nginx_upstream_check.NginxUpstreamCheck{}},
		"nginx_plus_api":       {name: "nginx_plus_api", Catalog: "nginx", input: &nginx_plus_api.NginxPlusApi{}},
		"nginx_plus":           {name: "nginx_plus", Catalog: "nginx", input: &nginx_plus.NginxPlus{}},
		"nginx_vts":            {name: "nginx_vts", Catalog: "nginx", input: &nginx_vts.NginxVTS{}},

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
		`goruntime`:     {name: "influxdb", Catalog: "golang", input: &influxdb.InfluxDB{}},

		"openldap": {name: "openldap", Catalog: "openldap", input: &openldap.Openldap{}},

		"zookeeper": {name: "zookeeper", Catalog: "zookeeper", input: &zookeeper.Zookeeper{}},
		"ceph":      {name: "ceph", Catalog: "ceph", input: &ceph.Ceph{}},
		"dns_query": {name: "dns_query", Catalog: "dns_query", input: &dns_query.DnsQuery{}},

		"docker": {name: "docker", Catalog: "docker", input: &docker.Docker{}},

		"activemq":       {name: "activemq", Catalog: "activemq", input: &activemq.ActiveMQ{}},
		"rabbitmq":       {name: "rabbitmq", Catalog: "rabbitmq", input: &rabbitmq.RabbitMQ{}},
		"nsq":            {name: "nsq", Catalog: "nsq", input: &nsq.NSQ{}},
		"nsq_consumer":   {name: "nsq_consumer", Catalog: "nsq", input: &nsq_consumer.NSQConsumer{}},
		"kafka_consumer": {name: "kafka_consumer", Catalog: "kafka", input: &kafka_consumer.KafkaConsumer{}},
		"mqtt_consumer":  {name: "mqtt_consumer", Catalog: "mqtt", input: &mqtt_consumer.MQTTConsumer{}},

		"fluentd":    {name: "fluentd", Catalog: "fluentd", input: &fluentd.Fluentd{}},
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

		"nvidia_smi":    {name: "nvidia_smi", Catalog: "nvidia", input: &nvidia_smi.NvidiaSMI{}},
		"kubernetes":    {name: "kubernetes", Catalog: "k8s", input: &kubernetes.Kubernetes{}},
		"amqp_consumer": {name: "amqp_consumer", Catalog: "amqp", input: &amqp_consumer.AMQPConsumer{}},
		"github":        {name: "github", Catalog: "github", input: &github.GitHub{}},
		"uwsgi":         {name: "uwsgi", Catalog: "uwsgi", input: &uwsgi.Uwsgi{}},
		`kibana`:        {name: "kibana", Catalog: "kibana", input: &kibana.Kibana{}},
		`modbus`:        {name: "modbus", Catalog: "modbus", input: &modbus.Modbus{}},

		// jolokia2 related
		`weblogic`:       {name: "jolokia2_agent", Catalog: "weblogic", input: &jolokia2.JolokiaAgent{}},
		`jvm`:            {name: "jolokia2_agent", Catalog: "jvm", input: &jolokia2.JolokiaAgent{}},
		`hadoop_hdfs`:    {name: "jolokia2_agent", Catalog: "hadoop_hdfs", input: &jolokia2.JolokiaAgent{}},
		"jolokia2_agent": {name: "jolokia2_agent", Catalog: "jolokia2_agent", input: &jolokia2.JolokiaAgent{}},
		"jboss":          {name: "jolokia2_agent", Catalog: "jboss", input: &jolokia2.JolokiaAgent{}},
		"cassandra":      {name: "jolokia2_agent", Catalog: "cassandra", input: &jolokia2.JolokiaAgent{}},
		"bitbucket":      {name: "jolokia2_agent", Catalog: "bitbucket", input: &jolokia2.JolokiaAgent{}},
		"kafka":          {name: "jolokia2_agent", Catalog: "kafka", input: &jolokia2.JolokiaAgent{}},

		// ambiguous import
		`consul`: {name: "consul", Catalog: "consul", Sample: samples["consul"], input: nil},

		// get panic:
		//   panic: mismatching message name: got k8s.io.kubernetes.pkg.watch.versioned.Event,
		//          want github.com/ericchiang.k8s.watch.versioned.Event
		"kube_inventory": {name: "kube_inventory", Catalog: "k8s", Sample: samples["kube_inventory"], input: nil},

		// telegraf not exported
		"phpfpm":  {name: "phpfpm", Catalog: "phpfpm", Sample: samples["phpfpm"], input: nil},
		"haproxy": {name: "haproxy", Catalog: "haproxy", Sample: samples["haproxy"], input: nil},
	}
)

func CheckTelegrafToml(name string, tomlcfg []byte) error {

	ti, ok := TelegrafInputs[name]

	if !ok {
		return fmt.Errorf("input not found")
	}

	if ti.input == nil {
		return fmt.Errorf("input check unavailable")
	}

	if err := toml.Unmarshal(tomlcfg, ti.input); err != nil {
		l.Errorf("toml.Unmarshal: %s", err.Error())
		return err
	}

	l.Debugf("toml %+#v", ti.input)
	return nil
}
