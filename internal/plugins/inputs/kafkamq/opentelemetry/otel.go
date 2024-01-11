// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry handle T/M from kafka.
package opentelemetry

import (
	"bytes"
	"net/http"
	"net/url"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq/worker"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/IBM/sarama"
)

var (
	log                 = logger.DefaultSLogger("kafkamq-otel")
	traceTopicEndpoint  = make(map[string]string)
	metricTopicEndpoint = make(map[string]string)
)

type OTELHandle struct {
	DKEndpoint   string   `toml:"dk_endpoint"`
	TraceAPI     string   `toml:"trace_api"`
	MetricAPI    string   `toml:"metric_api"`
	TraceTopics  []string `toml:"trace_topics"`
	MetricTopics []string `toml:"metric_topics"`
	Thread       int      `toml:"thread"`
	wp           *worker.WorkerPool
	transport    http.RoundTripper
}

func (mq *OTELHandle) Init() error {
	log = logger.SLogger("kafkamq-otel")
	baseURL, err := url.Parse(mq.DKEndpoint)
	if err != nil {
		log.Errorf("parse[jaeger.endpoint] err=%v", err)
		return err
	}
	if mq.TraceTopics != nil {
		traceURL := baseURL.ResolveReference(&url.URL{Path: mq.TraceAPI})
		for _, topic := range mq.TraceTopics {
			traceTopicEndpoint[topic] = traceURL.String()
		}
	}
	if mq.MetricTopics != nil {
		metricURL := baseURL.ResolveReference(&url.URL{Path: mq.MetricAPI})
		for _, topic := range mq.MetricTopics {
			metricTopicEndpoint[topic] = metricURL.String()
		}
	}

	mq.transport = http.DefaultTransport
	mq.wp = worker.NewWorkerPool(mq.DoMsg, mq.Thread)
	return nil
}

func (mq *OTELHandle) GetTopics() []string {
	return append(mq.TraceTopics, mq.MetricTopics...)
}

func (mq *OTELHandle) DoMsg(msg *sarama.ConsumerMessage) error {
	remoteURL := ""
	if u, ok := traceTopicEndpoint[msg.Topic]; ok {
		remoteURL = u
	}
	if u, ok := metricTopicEndpoint[msg.Topic]; ok {
		remoteURL = u
	}
	r, err := http.NewRequest(http.MethodPost, remoteURL, bytes.NewBuffer(msg.Value))
	if err != nil {
		log.Errorf("new request err=%v", err)
		return err
	}
	r.Header.Add("Content-Type", "application/x-protobuf")
	resp, err := mq.transport.RoundTrip(r)
	if err != nil {
		log.Warnf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	log.Debugf("Response status code: %d", resp.StatusCode)
	return nil
}

func (mq *OTELHandle) Process(msg *sarama.ConsumerMessage) error {
	if mq.wp != nil {
		f := mq.wp.GetWorker()
		go func(message *sarama.ConsumerMessage) {
			_ = f(message)
			mq.wp.PutWorker(f)
		}(msg)
		return nil
	} else {
		return mq.DoMsg(msg)
	}
}
