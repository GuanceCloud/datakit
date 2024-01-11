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
	"sync"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq/worker"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/IBM/sarama"
)

var (
	log        = logger.DefaultSLogger("jaeger")
	jaegerPath = "/apis/traces"
)

type Consumer struct {
	DKEndpoint string   `toml:"dk_endpoint"`
	Topics     []string `toml:"topics"`
	Source     string   `toml:"source"`
	Thread     int      `toml:"thread"`

	dkURL     string
	feeder    dkio.Feeder
	ptsLock   sync.Mutex
	ptsCache  []*point.Point
	transport http.RoundTripper

	wp *worker.WorkerPool
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
	mq.wp = worker.NewWorkerPool(mq.DoMsg, mq.Thread)
	mq.ptsCache = make([]*point.Point, 0)
	itrace.SetLogger(log)
	return nil
}

func (mq *Consumer) GetTopics() []string {
	return mq.Topics
}

func (mq *Consumer) DoMsg(msg *sarama.ConsumerMessage) error {
	if mq.Source == "daoke" {
		mq.DaoKeOTELMsg(msg)
		return nil
	}
	r, err := http.NewRequest(http.MethodPost, mq.dkURL, bytes.NewBuffer(msg.Value))
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

func (mq *Consumer) DaoKeOTELMsg(msg *sarama.ConsumerMessage) {
	pt, err := itrace.ParseToJaeger(msg.Value)
	if err != nil {
		log.Warnf("err=%v", err)
		return
	}
	mq.ptsLock.Lock()
	defer mq.ptsLock.Unlock()
	mq.ptsCache = append(mq.ptsCache, pt)
	if len(mq.ptsCache) >= 100 {
		pts := make([]*point.Point, 0, 100)
		pts = append(pts, mq.ptsCache...)
		err = mq.feeder.Feed("otel_jaeger", point.Tracing, pts)
		if err != nil {
			log.Errorf("err=%v", err)
		}
		mq.ptsCache = make([]*point.Point, 0)
	}
}

func (mq *Consumer) Process(msg *sarama.ConsumerMessage) error {
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

func (mq *Consumer) SetFeeder(feeder dkio.Feeder) {
	mq.feeder = feeder
}
