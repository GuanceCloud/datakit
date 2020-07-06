package etcd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "etcd"

	defaultMeasurement = "etcd"

	configSample = `
# [[etcd]]
#       ## etcd 地址
#	host = "127.0.0.1"
#
#       ## etcd 端口
#	port = 2379
#
#       ## 是否开启 HTTPS TLS，如果开启则需要同时配置下面的3个路径
#       tls_open = false
#
#       ## CA 证书路径
#       tls_cacert_file = "ca.crt"
#
#       ## 客户端证书文件路径
#	tls_cert_file = "peer.crt"
#
#	## 私钥文件路径
#	tls_key_file = "peer.key"
#
#	## 采集周期，时间单位是秒
#	collect_cycle = 60
`
)

var l *zap.SugaredLogger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Etcd{}
	})
}

type (
	Etcd struct {
		C []Impl `toml:"etcd"`
	}
	Impl struct {
		Host       string        `toml:"host"`
		Port       int           `toml:"port"`
		TLSOpen    bool          `toml:"tls_open"`
		CacertFile string        `toml:"tls_cacert_file"`
		CertFile   string        `toml:"tls_cert_file"`
		KeyFile    string        `toml:"tls_key_file"`
		Cycle      time.Duration `toml:"collect_cycle"`
		address    string
		// HTTPS TLS
		tlsConfig *tls.Config
	}
)

func (_ *Etcd) SampleConfig() string {
	return configSample
}

func (_ *Etcd) Catalog() string {
	return inputName
}

func (e *Etcd) Run() {
	l = logger.SLogger(inputName)

	for _, c := range e.C {
		c.start()
	}
}

func (i *Impl) start() {
	// "https" or "http"
	if i.TLSOpen {
		i.address = fmt.Sprintf("https://%s:%d/metrics", i.Host, i.Port)
		tc, err := TLSConfig(i.CacertFile, i.CertFile, i.KeyFile)
		if err != nil {
			l.Error(err)
			return
		}
		i.tlsConfig = tc

	} else {
		i.address = fmt.Sprintf("http://%s:%d/metrics", i.Host, i.Port)
	}

	ticker := time.NewTicker(time.Second * i.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			pt, err := i.getMetrics()
			if err != nil {
				l.Error(err)
				continue
			}
			io.Feed([]byte(pt.String()), io.Metric)
		}
	}
}

func (i *Impl) getMetrics() (*influxdb.Point, error) {

	client := &http.Client{}
	client.Timeout = time.Second * 5
	defer client.CloseIdleConnections()

	if i.tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: i.tlsConfig,
		}
	}

	resp, err := client.Get(i.address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	metrics, err := ParseV2(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return nil, errors.New("metrics is empty")
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

	return influxdb.NewPoint(defaultMeasurement, tags, fields, time.Now())
}
