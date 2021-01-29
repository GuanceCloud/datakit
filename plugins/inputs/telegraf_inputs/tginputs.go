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
	"github.com/influxdata/telegraf/plugins/inputs/procstat"
	"github.com/influxdata/telegraf/plugins/inputs/rabbitmq"
	"github.com/influxdata/telegraf/plugins/inputs/redis"
	"github.com/influxdata/telegraf/plugins/inputs/smart"
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
	Input   telegraf.Input
	name    string
	Catalog string

	// For some telegraf inputs, there are reasons to use specific Sample:
	//  - for telegraf source code that can't build with exists datakit (package conflict)
	//  - for telegraf source code that not exported
	//  - some of telegraf inputs use the same telegraf input, i.e., jolokia2_agent, win_perf_counters
	// so we can't get sample from input.SampleConfig(), we just return the Sample field
	//
	// Bad news: it's hard to sync upstream telegraf updates on these inputs
	Sample   string
	Pipeline string
}

func (ti *TelegrafInput) SampleConfig() string {

	if ti.Sample != "" { // prefer specified sample
		return ti.Sample
	}

	// telegraf not exported inputs, return sample directly(if configured in init())
	if ti.Input == nil {
		l.Fatal("%s should have a specific config-sample", ti.name)
	}

	s := ti.Input.SampleConfig()
	if s == "" {
		s = "# no sample need here, just open the input"
	}

	return fmt.Sprintf("[[inputs.%s]]\n%s", ti.name, s)
}

var (
	TelegrafInputs = map[string]*TelegrafInput{ // Name: Catalog
		"disk":   {name: "disk", Catalog: "host", Input: &disk.DiskStats{}},
		"diskio": {name: "diskio", Catalog: "host", Input: &diskio.DiskIO{}},
		"mem":    {name: "mem", Catalog: "host", Input: &mem.MemStats{}},
		"swap":   {name: "swap", Catalog: "host", Input: &swap.SwapStats{}},
		"system": {name: "system", Catalog: "host", Input: &system.SystemStats{}},
		//"cpu":      {name: "cpu", Catalog: "host", input: &cpu.CPUStats{}},
		"cpu":      {name: "cpu", Catalog: "host", Sample: samples["cpu"], Input: nil},
		"procstat": {name: "procstat", Catalog: "host", Input: &procstat.Procstat{}},
		"smart":    {name: "smart", Catalog: "host", Input: &smart.Smart{}},

		"internal": {name: "internal", Catalog: "internal", Sample: samples["internal"], Input: nil}, // import internal package not allowed

		"ping":            {name: "ping", Catalog: "network", Input: &ping.Ping{}},
		"net":             {name: "net", Catalog: "host", Input: &net.NetIOStats{}},
		"netstat":         {name: "netstat", Catalog: "network", Input: &net.NetStats{}},
		"net_response":    {name: "net_response", Catalog: "network", Input: &net_response.NetResponse{}},
		"http":            {name: "http", Catalog: "network", Input: &http.HTTP{}},
		"http_response":   {name: "http_response", Catalog: "network", Input: &http_response.HTTPResponse{}},
		"httpjson":        {name: "httpjson", Catalog: "network", Input: &httpjson.HttpJson{}},
		"socket_listener": {name: "socket_listener", Catalog: "network", Input: &socket_listener.SocketListener{}},

		// collectd use socket_listener to gather data
		"collectd": {name: "socket_listener", Catalog: "collectd", Input: &socket_listener.SocketListener{}},

		"nginx":                {name: "nginx", Catalog: "nginx", Sample: samples["nginx"], Input: &nginx.Nginx{}},
		"nginx_upstream_check": {name: "nginx_upstream_check", Catalog: "nginx", Input: &nginx_upstream_check.NginxUpstreamCheck{}},
		"nginx_plus_api":       {name: "nginx_plus_api", Catalog: "nginx", Input: &nginx_plus_api.NginxPlusApi{}},
		"nginx_plus":           {name: "nginx_plus", Catalog: "nginx", Input: &nginx_plus.NginxPlus{}},
		"nginx_vts":            {name: "nginx_vts", Catalog: "nginx", Input: &nginx_vts.NginxVTS{}},

		"tengine": {name: "tengine", Catalog: "tengine", Input: &tengine.Tengine{}},
		"apache":  {name: "apache", Catalog: "apache", Input: &apache.Apache{}},

		"postgresql":    {name: "postgresql", Catalog: "db", Input: &postgresql.Postgresql{}},
		"mongodb":       {name: "mongodb", Catalog: "db", Input: &mongodb.MongoDB{}},
		"redis":         {name: "redis", Catalog: "db", Input: &redis.Redis{}},
		"elasticsearch": {name: "elasticsearch", Catalog: "db", Input: &elasticsearch.Elasticsearch{}},
		"sqlserver":     {name: "sqlserver", Catalog: "db", Input: &sqlserver.SQLServer{}},
		"memcached":     {name: "memcached", Catalog: "db", Input: &memcached.Memcached{}},
		"solr":          {name: "solr", Catalog: "db", Input: &solr.Solr{}},
		"clickhouse":    {name: "clickhouse", Catalog: "db", Input: &clickhouse.ClickHouse{}},
		`influxdb`:      {name: "influxdb", Catalog: "db", Input: &influxdb.InfluxDB{}},

		"openldap": {name: "openldap", Catalog: "openldap", Input: &openldap.Openldap{}},

		"zookeeper": {name: "zookeeper", Catalog: "zookeeper", Input: &zookeeper.Zookeeper{}},
		"ceph":      {name: "ceph", Catalog: "ceph", Input: &ceph.Ceph{}},
		"dns_query": {name: "dns_query", Catalog: "dns_query", Input: &dns_query.DnsQuery{}},

		"docker": {name: "docker", Catalog: "docker", Input: &docker.Docker{}},

		"activemq":       {name: "activemq", Catalog: "activemq", Input: &activemq.ActiveMQ{}},
		"rabbitmq":       {name: "rabbitmq", Catalog: "rabbitmq", Input: &rabbitmq.RabbitMQ{}},
		"nsq":            {name: "nsq", Catalog: "nsq", Input: &nsq.NSQ{}},
		"nsq_consumer":   {name: "nsq_consumer", Catalog: "nsq", Input: &nsq_consumer.NSQConsumer{}},
		"kafka_consumer": {name: "kafka_consumer", Catalog: "kafka", Input: &kafka_consumer.KafkaConsumer{}},
		"mqtt_consumer":  {name: "mqtt_consumer", Catalog: "mqtt", Input: &mqtt_consumer.MQTTConsumer{}},

		"fluentd":    {name: "fluentd", Catalog: "fluentd", Input: &fluentd.Fluentd{}},
		"jenkins":    {name: "jenkins", Catalog: "jenkins", Input: &jenkins.Jenkins{}},
		"kapacitor":  {name: "kapacitor", Catalog: "kapacitor", Input: &kapacitor.Kapacitor{}},
		"ntpq":       {name: "ntpq", Catalog: "ntpq", Input: &ntpq.NTPQ{}},
		"openntpd":   {name: "openntpd", Catalog: "openntpd", Input: &openntpd.Openntpd{}},
		"x509_cert":  {name: "x509_cert", Catalog: "tls", Input: &x509_cert.X509Cert{}},
		"nats":       {name: "nats", Catalog: "nats", Input: &nats.Nats{}},
		"cloudwatch": {name: "cloudwatch", Catalog: "aws", Input: &cloudwatch.CloudWatch{}},
		"vsphere":    {name: "vsphere", Catalog: "vmware", Input: &vsphere.VSphere{}},
		"snmp":       {name: "snmp", Catalog: "snmp", Input: &snmp.Snmp{}},
		"exec":       {name: "exec", Catalog: "exec", Input: &exec.Exec{}},
		"syslog":     {name: "syslog", Catalog: "syslog", Input: &syslog.Syslog{}},

		"nvidia_smi":    {name: "nvidia_smi", Catalog: "nvidia", Input: &nvidia_smi.NvidiaSMI{}},
		"kubernetes":    {name: "kubernetes", Catalog: "k8s", Sample: samples["kubernetes"], Input: &kubernetes.Kubernetes{}},
		"amqp_consumer": {name: "amqp_consumer", Catalog: "amqp", Input: &amqp_consumer.AMQPConsumer{}},
		"github":        {name: "github", Catalog: "github", Input: &github.GitHub{}},
		"uwsgi":         {name: "uwsgi", Catalog: "uwsgi", Input: &uwsgi.Uwsgi{}},
		`kibana`:        {name: "kibana", Catalog: "kibana", Input: &kibana.Kibana{}},
		`modbus`:        {name: "modbus", Catalog: "modbus", Input: &modbus.Modbus{}},

		// jolokia2 related
		`weblogic`:       {name: "jolokia2_agent", Catalog: "weblogic", Input: &jolokia2.JolokiaAgent{}},
		`jvm`:            {name: "jolokia2_agent", Catalog: "jvm", Input: &jolokia2.JolokiaAgent{}},
		`hadoop_hdfs`:    {name: "jolokia2_agent", Catalog: "hadoop_hdfs", Input: &jolokia2.JolokiaAgent{}},
		"jolokia2_agent": {name: "jolokia2_agent", Catalog: "jolokia2_agent", Input: &jolokia2.JolokiaAgent{}},
		"jboss":          {name: "jolokia2_agent", Catalog: "jboss", Input: &jolokia2.JolokiaAgent{}},
		"cassandra":      {name: "jolokia2_agent", Catalog: "cassandra", Input: &jolokia2.JolokiaAgent{}},
		"bitbucket":      {name: "jolokia2_agent", Catalog: "bitbucket", Input: &jolokia2.JolokiaAgent{}},
		"kafka":          {name: "jolokia2_agent", Catalog: "kafka", Input: &jolokia2.JolokiaAgent{}},

		// ambiguous import
		`consul`: {name: "consul", Catalog: "consul", Sample: samples["consul"], Input: nil},

		// get panic:
		//   panic: mismatching message name: got k8s.io.kubernetes.pkg.watch.versioned.Event,
		//          want github.com/ericchiang.k8s.watch.versioned.Event
		"kube_inventory": {name: "kube_inventory", Catalog: "k8s", Sample: samples["kube_inventory"], Input: nil},

		// telegraf not exported
		"phpfpm":  {name: "phpfpm", Catalog: "phpfpm", Sample: samples["phpfpm"], Input: nil},
		"haproxy": {name: "haproxy", Catalog: "haproxy", Sample: samples["haproxy"], Input: nil},
	}
)

func CheckTelegrafToml(name string, tomlcfg []byte) error {

	ti, ok := TelegrafInputs[name]

	if !ok {
		return fmt.Errorf("input not found")
	}

	if ti.Input == nil {
		return fmt.Errorf("input check unavailable")
	}

	if err := toml.Unmarshal(tomlcfg, ti.Input); err != nil {
		l.Errorf("toml.Unmarshal: %s", err.Error())
		return err
	}

	l.Debugf("toml %+#v", ti.Input)
	return nil
}
