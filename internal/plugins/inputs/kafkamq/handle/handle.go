// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package handle send msg to remote.
package handle

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/Shopify/sarama"
)

var (
	log       = logger.DefaultSLogger("kafkamq.handle")
	transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	categorys = map[string]point.Category{
		point.SMetric:            point.Metric,
		point.SMetricDeprecated:  point.MetricDeprecated,
		point.SNetwork:           point.Network,
		point.SKeyEvent:          point.KeyEvent,
		point.SObject:            point.Object,
		point.SCustomObject:      point.CustomObject,
		point.SLogging:           point.Logging,
		point.STracing:           point.Tracing,
		point.SRUM:               point.RUM,
		point.SSecurity:          point.Security,
		point.SProfiling:         point.Profiling,
		point.SUnknownCategory:   point.UnknownCategory,
		point.SDynamicDWCategory: point.DynamicDWCategory,
	}
)

type Handle struct {
	Topics           []string `toml:"topics"`
	Endpoint         string   `toml:"endpoint"`
	SendMessageCount int      `toml:"send_message_count"`
	Debug            bool     `toml:"debug"`
	IsResponsePoint  bool     `toml:"is_response_point"`
	Pipeline         string   `toml:"pipeline"`
	HeaderCheck      bool     `toml:"header_check"`

	localCache chan *KafkaMessage
	feeder     dkio.Feeder
}

type KafkaMessage struct {
	Topic string `json:"topic"`
	Value []byte `json:"value"`
}

func (h *Handle) SetFeeder(feeder dkio.Feeder) {
	h.feeder = feeder
}

func (h *Handle) Init() error {
	log = logger.SLogger("kafkamq.handle")
	if h.SendMessageCount <= 0 {
		h.SendMessageCount = 1
	}
	h.localCache = make(chan *KafkaMessage, h.SendMessageCount)
	go h.assignMsg()

	return nil
}

func (h *Handle) GetTopics() []string {
	return h.Topics
}

func (h *Handle) Process(msg *sarama.ConsumerMessage) error {
	if h.HeaderCheck {
		buf := msg.Value
		if len(buf) < 4 {
			return nil
		}
		if buf[0] != 0xEF && buf[2] != 0 {
			log.Debugf("header check buf[0]:%d , buf[2]=%d", buf[0], buf[2])
			return nil
		}
		if buf[3] != 40 && buf[3] != 70 {
			log.Debugf("header check not 40 or 70, buf[3]=%d", buf[3])
			return nil
		}
	}

	// debug 模式 直接发送
	if h.Debug {
		log.Debugf("debug mode, send message to remote:%s", h.Endpoint)
		h.sendToRemote(msg.Value)
		return nil
	}

	h.localCache <- &KafkaMessage{Topic: msg.Topic, Value: msg.Value}
	return nil
}

func (h *Handle) assignMsg() {
	tiker := time.NewTicker(time.Second * 5)
	cache := make([]*KafkaMessage, 0, h.SendMessageCount)
	for {
		select {
		case <-tiker.C:
			if len(cache) > 0 {
				log.Debugf("time ticker 5s,send %d msg to remote", len(cache))
				h.sendToRemote(marshal(cache))
				cache = make([]*KafkaMessage, 0, h.SendMessageCount)
			}
		case m := <-h.localCache:
			cache = append(cache, m)
			if len(cache) >= h.SendMessageCount {
				log.Debugf("send %d msg to remote", len(cache))
				h.sendToRemote(marshal(cache))
				cache = make([]*KafkaMessage, 0, h.SendMessageCount)
			}
		case <-datakit.Exit.Wait():
			log.Info("kafkamq handle exit")
			return
		}
	}
}

func marshal(msgs []*KafkaMessage) []byte {
	data, err := json.Marshal(msgs)
	if err != nil {
		log.Errorf("json marshal err=%v", err)
	}
	return data
}

func (h *Handle) sendToRemote(data []byte) {
	r, err := http.NewRequest(http.MethodPost, h.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		log.Errorf("new request err=%v", err)
		return
	}
	r.Header.Add("Content-Type", "application/json")
	resp, err := transport.RoundTrip(r)
	if err != nil {
		log.Errorf("send to remote %s ,err=%v", h.Endpoint, err)
		count := 0
		for {
			resp, err = transport.RoundTrip(r)
			if err != nil {
				count++
				log.Errorf("wait 5 econd ,try %d send to remote %s ,err=%v", count, h.Endpoint, err)
				time.Sleep(time.Second * 5)
			} else {
				break
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("status code is %d ,not 200", resp.StatusCode)
		return
	}
	defer resp.Body.Close() //nolint
	if h.IsResponsePoint {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("read body err=%v", err)
			return
		}
		isjson := strings.Contains(resp.Header.Get("Content-Type"), "application/json")
		category := point.Tracing
		if c, ok := categorys[resp.Header.Get("X-category")]; ok {
			category = c
		}
		pts, err := httpapi.HandleWriteBody(body, isjson, point.WithPrecision(point.NS))
		if err != nil {
			log.Errorf("from response body decode to point err=%v", err)
			return
		}
		err = h.feeder.Feed("kafkamq_handle", category, pts, &dkio.Option{})
		if err != nil {
			log.Warnf("feed io err=%v", err)
		}
		log.Debugf("IsResponsePoint=true, send %d point to dataway", len(pts))
	}
}
