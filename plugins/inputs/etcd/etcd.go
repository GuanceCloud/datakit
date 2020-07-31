package etcd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "etcd"

	defaultMeasurement = "etcd"

	sampleCfg = `
[[inputs.etcd]]
	# required
	host = "127.0.0.1"

	# required
	port = 2379

	# valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
	# required
	interval = "10s"

	# use HTTPS TLS
	tls_open = false

	# CA
	tls_cacert_file = "ca.crt"

	# client
	tls_cert_file = "peer.crt"

	# key
	tls_key_file = "peer.key"

	# [inputs.etcd.tags]
	# tags1 = "value1"
`
)

var l *logger.Logger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Etcd{}
	})
}

type Etcd struct {
	Host       string            `toml:"host"`
	Port       int               `toml:"port"`
	Interval   string            `toml:"interval"`
	TLSOpen    bool              `toml:"tls_open"`
	CacertFile string            `toml:"tls_cacert_file"`
	CertFile   string            `toml:"tls_cert_file"`
	KeyFile    string            `toml:"tls_key_file"`
	Tags       map[string]string `toml:"tags"`

	// forward compatibility
	CollectCycle string `toml:"collect_cycle"`

	address  string
	duration time.Duration

	// HTTPS TLS config
	tlsConfig *tls.Config
}

func (_ *Etcd) SampleConfig() string {
	return sampleCfg
}

func (_ *Etcd) Catalog() string {
	return inputName
}

func (e *Etcd) Run() {
	l = logger.SLogger(inputName)

	if e.loadcfg() {
		return
	}
	ticker := time.NewTicker(e.duration)
	defer ticker.Stop()

	l.Infof("etcd input started.")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := e.getMetrics()
			if err != nil {
				l.Error(err)
				continue
			}
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (e *Etcd) loadcfg() bool {

	if e.Interval == "" && e.CollectCycle != "" {
		e.Interval = e.CollectCycle
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		d, err := time.ParseDuration(e.Interval)
		if err != nil || d <= 0 {
			l.Errorf("invalid interval, err %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		e.duration = d

		// "https" or "http"
		if e.TLSOpen {
			e.address = fmt.Sprintf("https://%s:%d/metrics", e.Host, e.Port)
			tc, err := TLSConfig(e.CacertFile, e.CertFile, e.KeyFile)
			if err != nil {
				l.Errorf("failed to TLS, err: %s", err.Error())
				time.Sleep(time.Second)
			} else {
				e.tlsConfig = tc
				break
			}
		} else {
			e.address = fmt.Sprintf("http://%s:%d/metrics", e.Host, e.Port)
			break
		}
	}

	if e.Tags == nil {
		e.Tags = make(map[string]string)
	}

	if _, ok := e.Tags["address"]; !ok {
		e.Tags["address"] = fmt.Sprintf("%s:%d", e.Host, e.Port)
	}

	return false
}

func (e *Etcd) getMetrics() ([]byte, error) {

	client := &http.Client{}
	client.Timeout = time.Second * 5
	defer client.CloseIdleConnections()

	if e.tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: e.tlsConfig,
		}
	}

	resp, err := client.Get(e.address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	metrics, err := ParseV2(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("metrics is empty")
	}

	var tags = make(map[string]string)
	var fields = make(map[string]interface{}, len(metrics))

	// prometheus to point
	for _, metric := range metrics {

		for k, v := range metric.Tags() {

			// Save 4 fields from prometheus.
			// It's STUPID!

			// prometheus,action=create etcd_debugging_store_writes_total=1 1586857769285050000
			// prometheus,action=set etcd_debugging_store_writes_total=4 1586857769285050000
			// prometheus,action=get etcd_debugging_store_reads_total=1 1586857769285050000
			// prometheus,action=getRecursive etcd_debugging_store_reads_total=2 1586857769285050000
			// ============== TO =============
			// etcd_debugging_store_writes_total_create=1
			// etcd_debugging_store_writes_total_set=4
			// etcd_debugging_store_reads_total_get=1
			// etcd_debugging_store_reads_total_getRecursive=2

			if _, ok := actionList[v]; ok {
				for _k, _v := range metric.Fields() {
					fields[_k+"_"+v] = _v
					goto END_ACTION
				}
			}

			if _, ok := collectList[k]; ok {
				tags[k] = v
			}
		}

		for k, v := range metric.Fields() {
			if _, ok := collectList[k]; ok {
				fields[k] = v
			}
		}

	END_ACTION:
	}

	for k, v := range e.Tags {
		tags[k] = v
	}

	return io.MakeMetric(defaultMeasurement, tags, fields, time.Now())
}
