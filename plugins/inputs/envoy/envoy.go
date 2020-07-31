package envoy

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "envoy"

	defaultMeasurement = "envoy"

	sampleCfg = `
# [[inputs.envoy]]
#	# required
# 	host = "127.0.0.1"
#
#	# required
# 	port = 9901
#
# 	# valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
#	# required
# 	interval = "10s"
#
# 	# use HTTPS TLS
# 	tls_open = false
#
# 	# CA
# 	tls_cacert_file = "ca.crt"
#
# 	# client
# 	tls_cert_file = "peer.crt"
#
# 	# key
# 	tls_key_file = "peer.key"
#
# 	# [inputs.envoy.tags]
# 	# tags1 = "value1"
`
)

var (
	l          *logger.Logger
	testAssert bool
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Envoy{}
	})
}

type Envoy struct {
	Host       string            `toml:"host"`
	Port       int               `toml:"port"`
	Interval   string            `toml:"interval"`
	TLSOpen    bool              `toml:"tls_open"`
	CacertFile string            `toml:"tls_cacert_file"`
	CertFile   string            `toml:"tls_cert_file"`
	KeyFile    string            `toml:"tls_key_file"`
	Tags       map[string]string `toml:"tags"`

	address  string
	duration time.Duration

	// HTTPS TLS config
	tlsConfig *tls.Config
}

func (_ *Envoy) SampleConfig() string {
	return sampleCfg
}

func (_ *Envoy) Catalog() string {
	return inputName
}

func (e *Envoy) Run() {
	l = logger.SLogger(inputName)

	if e.loadcfg() {
		return
	}
	ticker := time.NewTicker(e.duration)
	defer ticker.Stop()

	l.Infof("envoy input started.")

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
			if testAssert {
				l.Debugf("date: %s", string(data))
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

func (e *Envoy) loadcfg() bool {

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
			e.address = fmt.Sprintf("https://%s:%d/stats/prometheus", e.Host, e.Port)
			tc, err := TLSConfig(e.CacertFile, e.CertFile, e.KeyFile)
			if err != nil {
				l.Errorf("failed to TLS, err: %s", err.Error())
				time.Sleep(time.Second)
			} else {
				e.tlsConfig = tc
				break
			}
		} else {
			e.address = fmt.Sprintf("http://%s:%d/stats/prometheus", e.Host, e.Port)
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

func (e *Envoy) getMetrics() ([]byte, error) {

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
			if _, ok := collectList[k]; ok {
				tags[k] = v
			}
		}

		for k, v := range metric.Fields() {
			if _, ok := collectList[k]; ok {
				fields[k] = v
			}
		}
	}

	for k, v := range e.Tags {
		tags[k] = v
	}

	return io.MakeMetric(defaultMeasurement, tags, fields, time.Now())
}

func TLSConfig(caFile, certFile, keyFile string) (*tls.Config, error) {

	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append certs from PEM.")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	tlsConfig.BuildNameToCertificate()

	return tlsConfig, nil
}
