package cloudprober

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/common/expfmt"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

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
			n.getMetric()
			if n.lastErr != nil {
				iod.FeedLastError(inputName, n.lastErr.Error())
			}
		case <-datakit.Exit.Wait():

			l.Info("cloudprober exit")
			return
		}
	}
}

func (n *Input) getMetric() {
	resp, err := n.client.Get(n.URL)
	if err != nil {
		l.Errorf("error making HTTP request to %s: %s", n.URL, err)
		n.lastErr = err
		return
	}
	defer resp.Body.Close()

	collector, err := parse(resp.Body)
	if err != nil {
		n.lastErr = err
		l.Error(err.Error())
		return
	}
	if err := inputs.FeedMeasurement(inputName, datakit.Metric, collector, &iod.Option{CollectCost: time.Since(n.start)}); err != nil {
		l.Error(err.Error())
		n.lastErr = err
	}

}

func parse(reader io.Reader) ([]inputs.Measurement, error) {
	var (
		parse     expfmt.TextParser
		collector []inputs.Measurement
	)
	Family, err := parse.TextToMetricFamilies(reader)
	if err != nil {
		return collector, err
	}
	for metricName, family := range Family {
		for _, metric := range family.Metric {
			measurement := &Measurement{
				tags:   map[string]string{},
				fields: map[string]interface{}{},
				ts:     datakit.TimestampMsToTime(metric.GetTimestampMs()),
			}

			for _, label := range metric.Label {
				if label.GetName() == "ptype" {
					measurement.name = fmt.Sprintf("probe_%s", label.GetValue())
					continue
				}
				measurement.tags[label.GetName()] = label.GetValue()
			}
			switch family.GetType().String() {
			case "COUNTER":
				measurement.fields[metricName] = metric.Counter.GetValue()
			case "GAUGE":
				measurement.fields[metricName] = metric.Gauge.GetValue()
			case "SUMMARY":
				measurement.fields[metricName] = metric.Summary.GetSampleCount()
			case "UNTYPED":
				measurement.fields[metricName] = metric.Untyped.GetValue()
			case "HISTOGRAM":
				measurement.fields[metricName] = metric.Histogram.GetSampleCount()

			}
			collector = append(collector, measurement)
		}
	}
	return collector, nil
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
		&Measurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 5},
		}
		return s
	})
}
