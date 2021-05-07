package cloudprober

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"time"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"net/http"
)

var (
	inputName   = `cloudprober`
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
[[inputs.cloudprober]]
	url = "http://localhost/server_status"
	# ##(optional) collection interval, default is 30s
	# interval = "30s"
	use_vts = false
	## Optional TLS Config
	# tls_ca = "/xxx/ca.pem"
	# tls_cert = "/xxx/cert.cer"
	# tls_key = "/xxx/key.key"
	## Use TLS but skip chain & host verification
	insecure_skip_verify = false

	[inputs.cloudprober.tags]
	# a = "b"`

)

func (_ *Input) SampleConfig() string {
	return sample
}

func (_ *Input) Catalog() string {
	return inputName
}


func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("cloudprober start")
	n.Interval.Duration = datakit.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] cloudprober init client err:%s", err.Error())
		return
	}
	n.client = client

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:

		case <-datakit.Exit.Wait():

			l.Info("cloudprober exit")
			return
		}
	}
}


func (n *Input) createHttpClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
	}

	return client, nil
}

func (_ *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{

	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
		return s
	})
}
