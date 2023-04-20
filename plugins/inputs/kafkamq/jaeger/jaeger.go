// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package jaeger handle Jaeger Spans from kafka.
package jaeger

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/Shopify/sarama"
)

var (
	log        = logger.DefaultSLogger("jaeger")
	jaegerPath = "/apis/traces"
)

type Consumer struct {
	DKEndpoint string   `toml:"dk_endpoint"`
	Topics     []string `toml:"topics"`

	dkURL     string
	transport http.RoundTripper
}

func (mq *Consumer) Init() error {
	log = logger.SLogger("kafkamq.jaeger")
	baseURL, err := url.Parse(mq.DKEndpoint)
	if err != nil {
		log.Errorf("parse[jaeger.endpoint] err=%v", err)
		return err
	}

	newURL := baseURL.ResolveReference(&url.URL{Path: jaegerPath})
	mq.dkURL = newURL.String()
	mq.transport = http.DefaultTransport
	return nil
}

func (mq *Consumer) GetTopics() []string {
	return mq.Topics
}

func (mq *Consumer) Process(msg *sarama.ConsumerMessage) {
	r, err := http.NewRequest(http.MethodPost, mq.dkURL, bytes.NewBuffer(msg.Value))
	if err != nil {
		log.Errorf("new request err=%v", err)
		return
	}
	r.Header.Add("Content-Type", "application/x-protobuf")
	resp, err := mq.transport.RoundTrip(r)
	if err != nil {
		log.Warnf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	log.Debugf("Response status code: %d", resp.StatusCode)
}
