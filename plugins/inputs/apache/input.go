package apache

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (_ *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"apache": pipeline,
	}
	return pipelineMap
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("apache start")
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	if n.Log != nil {
		go func() {
			inputs.JoinPipelinePath(n.Log, "apache.p")
			n.Log.Source = inputName
			n.Log.Match = `^\[\w+ \w+ \d+`
			n.Log.Tags = map[string]string{}
			for k, v := range n.Tags {
				n.Log.Tags[k] = v
			}
			tail, err := inputs.NewTailer(n.Log)
			if err != nil {
				l.Errorf("init tailf err:%s", err.Error())
				return
			}
			n.tail = tail
			tail.Run()
		}()
	}

	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] apache init client err:%s", err.Error())
		return
	}
	n.client = client

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if err := n.getMetric(); err != nil {
				iod.FeedLastError(inputName, err.Error())
			}
		case <-datakit.Exit.Wait():
			if n.tail != nil {
				n.tail.Close()
				l.Info("apache log exit")
			}
			l.Info("apache exit")
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
		Timeout: time.Second * 5,
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

func (n *Input) getMetric() error {
	n.start = time.Now()
	req, err := http.NewRequest("GET", n.Url, nil)
	if err != nil {
		return fmt.Errorf("error on new request to %s : %s", n.Url, err)
	}

	if len(n.Username) != 0 && len(n.Password) != 0 {
		req.SetBasicAuth(n.Username, n.Password)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s", n.Url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", n.Url, resp.Status)
	}
	return n.parse(resp.Body)
}

func (n *Input) parse(body io.Reader) error {
	sc := bufio.NewScanner(body)

	tags := map[string]string{
		"url": n.Url,
	}
	for k, v := range n.Tags {
		tags[k] = v

	}
	metric := Measurement{
		name:   inputName,
		fields: map[string]interface{}{},
		ts:     time.Now(),
	}

	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key, part := strings.Replace(parts[0], " ", "", -1), strings.TrimSpace(parts[1])
			if tagKey, ok := tagMap[key]; ok {
				tags[tagKey] = part
			}

			fieldKey, ok := filedMap[key]
			if !ok {
				continue
			}
			switch key {
			case "CPULoad":
				value, err := strconv.ParseFloat(part, 64)
				if err != nil {
					l.Error(err.Error())
					continue
				}
				metric.fields[fieldKey] = value

			default:
				value, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					l.Error(err.Error())
					continue
				}
				if fieldKey == "Total kBytes" {
					// kbyte to byte
					metric.fields[fieldKey] = value * 1024
					continue
				}
				metric.fields[fieldKey] = value
			}

		}
	}
	metric.tags = tags
	l.Debug(metric)
	return inputs.FeedMeasurement(inputName, datakit.Metric, []inputs.Measurement{&metric}, &iod.Option{CollectCost: time.Since(n.start)})
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 5},
		}
		return s
	})
}
