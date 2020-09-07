package cloudflare

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "cloudflare"

	defaultMeasurement = "cloudflare"

	cloudflareAPIURL = "https://api.cloudflare.com/client/v4"

	sampleCfg = `
[[inputs.cloudflare]]
    # cloudflare login email
    # required
    email = ""
    
    # service zone id
    # required
    zone_id = ""
    
    # api key
    # required
    api_key = ""
    
    # valid time units are "m", "h"
    # required
    interval = "24h"
    
    # [inputs.cloudflare.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Cloudflare{}
	})
}

type Cloudflare struct {
	Email    string            `toml:"email"`
	ZoneID   string            `toml:"zone_id"`
	APIKey   string            `toml:"api_key"`
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`

	requ     *http.Request
	duration time.Duration
}

func (*Cloudflare) SampleConfig() string {
	return sampleCfg
}

func (*Cloudflare) Catalog() string {
	return inputName
}

func (h *Cloudflare) Run() {
	l = logger.SLogger(inputName)

	if h.laodCfg() {
		return
	}

	ticker := time.NewTicker(h.duration)
	defer ticker.Stop()

	l.Infof("cloudflare input started.")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := h.getMetrics()
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

func (h *Cloudflare) laodCfg() bool {
	var err error
	var d time.Duration

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// next
		}

		d, err = time.ParseDuration(h.Interval)
		if err != nil || d <= 0 {
			l.Errorf("invalid interval, %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		h.duration = d

		h.requ, err = http.NewRequest("GET",
			fmt.Sprintf("%s/zones/%s/analytics/dashboard?since=-%d&continuous=true",
				cloudflareAPIURL,
				h.ZoneID,
				h.duration/time.Minute,
			), nil)
		if err != nil {
			l.Errorf("new request error: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		break
	}

	h.requ.Header.Add("X-Auth-Email", h.Email)
	h.requ.Header.Add("X-Auth-Key", h.APIKey)
	h.requ.Header.Add("Content-Type", "application/json")

	if _, ok := h.Tags["zone_id"]; !ok {
		h.Tags["zone_id"] = h.ZoneID
	}

	return false
}

func (h *Cloudflare) getMetrics() ([]byte, error) {
	resp, err := http.DefaultClient.Do(h.requ)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var buffer bytes.Buffer
	jsonMetrics := gjson.ParseBytes(body)

	if !jsonMetrics.Get("success").Bool() {
		for _, errMsg := range jsonMetrics.Get("errors").Array() {
			l.Errorf("err, code: %d, message: %s\n", errMsg.Get("code").Int(), errMsg.Get("message").String())
		}
		return nil, fmt.Errorf("failed to get cloudflare metrics")
	}

	timeseries := jsonMetrics.Get("result.timeseries").Array()

	for _, timeserie := range timeseries {
		t, err := time.Parse(time.RFC3339, timeserie.Get("until").String())
		if err != nil {
			continue
		}

		fields := make(map[string]interface{})
		fields["uniques"] = timeserie.Get("uniques.all").Int()
		parseRequest(fields, timeserie.Get("requests"))
		parseBandwidth(fields, timeserie.Get("bandwidth"))
		data, err := io.MakeMetric(defaultMeasurement, h.Tags, fields, t)
		if err != nil {
			continue
		}

		buffer.Write(data)
		buffer.WriteString("\n")
	}

	return buffer.Bytes(), nil
}

func parseRequest(fieldsMap map[string]interface{}, metrics gjson.Result) {

	fieldsMap["requests_all"] = metrics.Get("all").Int()
	fieldsMap["requests_cached"] = metrics.Get("cached").Int()
	fieldsMap["requests_uncached"] = metrics.Get("uncached").Int()
	fieldsMap["requests_encrypted"] = metrics.Get("ssl.encrypted").Int()
	fieldsMap["requests_unencrypted"] = metrics.Get("ssl.unencrypted").Int()

	for k, status := range metrics.Get("http_status").Map() {
		fieldsMap["requests_status_"+k] = status.Int()
	}

	for k, contentType := range metrics.Get("content_type").Map() {
		fieldsMap["requests_content_type_"+k] = contentType.Int()
	}

	for k, country := range metrics.Get("country").Map() {
		fieldsMap["requests_country_"+k] = country.Int()
	}

	for k, ipClass := range metrics.Get("ip_class").Map() {
		fieldsMap["requests_ip_class_"+k] = ipClass.Int()
	}
}

func parseBandwidth(fieldsMap map[string]interface{}, metrics gjson.Result) {

	fieldsMap["bandwidth_all"] = metrics.Get("all").Int()
	fieldsMap["bandwidth_cached"] = metrics.Get("cached").Int()
	fieldsMap["bandwidth_uncached"] = metrics.Get("uncached").Int()
	fieldsMap["bandwidth_encrypted"] = metrics.Get("ssl.encrypted").Int()
	fieldsMap["bandwidth_unencrypted"] = metrics.Get("ssl.unencrypted").Int()

	for k, contentType := range metrics.Get("content_type").Map() {
		fieldsMap["bandwidth_content_type_"+k] = contentType.Int()
	}

	for k, country := range metrics.Get("country").Map() {
		fieldsMap["bandwidth_country_"+k] = country.Int()
	}
}
