// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalking handle SkyWalking tracing metrics.
package skywalking

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/Shopify/sarama"
)

const (
	tracePath     = "/v3/trace"
	metricPath    = "/v3/metric"
	loggingPath   = "/v3/logging"
	profilingPath = "/v3/profiling"
)

var (
	log         = logger.DefaultSLogger("kafkamq")
	metric      = "skywalking-metrics"
	profilings  = "skywalking-profilings"
	segments    = "skywalking-segments"
	managements = "skywalking-managements"
	meters      = "skywalking-meters"
	logging     = "skywalking-logging"

	skyAPIPaths = map[string]string{
		metric:      metricPath,
		logging:     loggingPath,
		profilings:  profilingPath,
		segments:    tracePath,
		managements: "", // todo
		meters:      "",
	}
)

type SkyConsumer struct {
	DKEndpoint string   `toml:"dk_endpoint"`
	Topics     []string `toml:"topics"`
	Namespace  string   `toml:"namespace"`

	topics []string
	client http.RoundTripper
	dkURLs map[string]string
}

func (sky *SkyConsumer) Init() error {
	log = logger.SLogger("kafkamq.skywalking")
	//	formatTopics := make(map[string]string)
	topics := make([]string, 0)
	dkURLs := make(map[string]string)
	for _, topic := range sky.Topics {
		baseURL, err := url.Parse(sky.DKEndpoint)
		if err != nil {
			log.Errorf("DKEndpoint is not url,err=%v", err)
			return err
		}
		if sky.Namespace != "" {
			t := sky.Namespace + "-" + topic
			topics = append(topics, t)
			dkURLs[t] = baseURL.ResolveReference(&url.URL{Path: skyAPIPaths[topic]}).String()
		} else {
			topics = append(topics, topic)
			dkURLs[topic] = baseURL.ResolveReference(&url.URL{Path: skyAPIPaths[topic]}).String()
		}
	}
	log.Debugf("topic is %v", topics)

	sky.topics = topics
	sky.client = dkhttp.DefTransport()
	sky.dkURLs = dkURLs
	if !sky.checkURL() {
		return fmt.Errorf("checkURL error")
	}
	return nil
}

func (sky *SkyConsumer) checkURL() (c bool) {
	c = true
	for key, dkurl := range sky.dkURLs {
		_, err := url.Parse(dkurl)
		if err != nil {
			log.Errorf("parse topic:[%s] url err=%v", key, err)
			c = false
		}
	}
	return
}

func (sky *SkyConsumer) GetTopics() []string {
	return sky.topics
}

func (sky *SkyConsumer) Process(msg *sarama.ConsumerMessage) error {
	u, ok := sky.dkURLs[msg.Topic]
	if !ok {
		log.Warnf("can not find Topic:[%s] url", msg.Topic)
	}
	r, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(msg.Value))
	if err != nil {
		log.Errorf("new request err=%v", err)
		return err
	}
	r.Header.Add("Content-Type", "application/x-protobuf")
	resp, err := sky.client.RoundTrip(r)
	if err != nil {
		log.Warnf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != 200 {
		err = fmt.Errorf("url=%s  statusCode not 200, code=%d", u, resp.StatusCode)
		log.Warn(err)
		return err
	} else {
		log.Debugf("Response status code: %d", resp.StatusCode)
	}
	return nil
}
