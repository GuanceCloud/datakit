package jvm

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	JvmConfigSample = `[[inputs.jvm]]
  # default_tag_prefix      = ""
  # default_field_prefix    = ""
  # default_field_separator = "."
 
  # username = ""
  # password = ""
  # response_timeout = "5s"

  ## Optional TLS config
  # tls_ca   = "/var/private/ca.pem"
  # tls_cert = "/var/private/client.pem"
  # tls_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ## Monitor Intreval
  # interval   = "60s"

  # Add agents URLs to query
  urls = ["http://localhost:8080/jolokia"]

  ## Add metrics to read
  [[inputs.jvm.metric]]
    name  = "java_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]

  [[inputs.jvm.metric]]
    name  = "java_memory"
    mbean = "java.lang:type=Memory"
    paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

  [[inputs.jvm.metric]]
    name     = "java_garbage_collector"
    mbean    = "java.lang:name=*,type=GarbageCollector"
    paths    = ["CollectionTime", "CollectionCount"]
    tag_keys = ["name"]

  [[inputs.jvm.metric]]
    name  = "java_threading"
    mbean = "java.lang:type=Threading"
    paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]

  [[inputs.jvm.metric]]
    name  = "java_class_loading"
    mbean = "java.lang:type=ClassLoading"
    paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]

  [[inputs.jvm.metric]]
    name     = "java_memory_pool"
    mbean    = "java.lang:name=*,type=MemoryPool"
    paths    = ["Usage", "PeakUsage", "CollectionUsage"]
    tag_keys = ["name"]
`
)

type JolokiaAgent struct {
	DefaultFieldPrefix    string
	DefaultFieldSeparator string
	DefaultTagPrefix      string

	URLs            []string `toml:"urls"`
	Username        string
	Password        string
	ResponseTimeout time.Duration `toml:"response_timeout"`
	Interval        string

	tls.ClientConfig

	Metrics  []MetricConfig `toml:"metric"`
	gatherer *Gatherer
	clients  []*Client

	collectCache []inputs.Measurement
	PluginName   string `toml:"-"`
	l            *logger.Logger

	Tags  map[string]string `toml:"-"`
	Types map[string]string `toml:"-"`
}

func (ja *JolokiaAgent) Gather() error {
	if ja.gatherer == nil {
		ja.gatherer = NewGatherer(ja.createMetrics())
	}

	// Initialize clients once
	if ja.clients == nil {
		ja.clients = make([]*Client, 0, len(ja.URLs))
		for _, url := range ja.URLs {
			client, err := ja.createClient(url)
			if err != nil {
				return err
			}
			ja.clients = append(ja.clients, client)
		}
	}

	var wg sync.WaitGroup
	for _, client := range ja.clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()

			err := ja.gatherer.Gather(client, ja)
			if err != nil {
				ja.l.Errorf("unable to gather metrics for %s: %v", client.URL, err)
			}
		}(client)
	}

	wg.Wait()

	return nil
}

func (ja *JolokiaAgent) createMetrics() []Metric {
	var metrics []Metric

	for _, config := range ja.Metrics {
		metrics = append(metrics, NewMetric(config,
			ja.DefaultFieldPrefix, ja.DefaultFieldSeparator, ja.DefaultTagPrefix))
	}

	return metrics
}

func (ja *JolokiaAgent) createClient(url string) (*Client, error) {
	return NewClient(url, &ClientConfig{
		Username:        ja.Username,
		Password:        ja.Password,
		ResponseTimeout: ja.ResponseTimeout,
		ClientConfig:    ja.ClientConfig,
	})
}
