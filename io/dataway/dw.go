// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dataway implement API request to dataway.
package dataway

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type IDataway interface {
	Write(...WriteOption) error
	Pull(what string) ([]byte, error)
}

var (
	dwAPIs = []string{
		point.MetricDeprecated.URL(),
		point.Metric.URL(),
		point.Network.URL(),
		point.KeyEvent.URL(),
		point.Object.URL(),
		point.CustomObject.URL(),
		point.Logging.URL(),
		point.Tracing.URL(),
		point.RUM.URL(),
		point.Security.URL(),
		point.Profiling.URL(),

		datakit.DatakitPull,
		datakit.LogFilter,
		datakit.SessionReplayUpload,
		datakit.HeartBeat,
		datakit.Election,
		datakit.ElectionHeartbeat,
		datakit.QueryRaw,
		datakit.Workspace,
		datakit.ListDataWay,
		datakit.ObjectLabel,
		datakit.LogUpload,
		datakit.PipelinePull,
		datakit.ProfilingUpload,
	}

	AvailableDataways          = []string{}
	log                        = logger.DefaultSLogger("dataway")
	datawayListIntervalDefault = 60
)

type Dataway struct {
	URLs []string `toml:"urls"`

	HTTPTimeout string `toml:"timeout"`
	HTTPProxy   string `toml:"http_proxy"`

	Hostname string `toml:"-"`

	Sinkers []*Sinker `toml:"sinkers,omitempty"`

	// Deprecated
	DeprecatedHost   string `toml:"host,omitempty"`
	DeprecatedScheme string `toml:"scheme,omitempty"`
	DeprecatedToken  string `toml:"token,omitempty"`
	DeprecatedURL    string `toml:"url,omitempty"`

	MaxIdleConnsPerHost int `toml:"max_idle_conns_per_host,omitempty"`

	httpTimeout time.Duration

	Proxy bool `toml:"proxy,omitempty"`

	EnableHTTPTrace bool `toml:"enable_httptrace,omitempty"`

	eps        []*endPoint
	locker     sync.RWMutex
	dnsCachers []*dnsCacher

	// metrics
}

func (dw *Dataway) Init() error {
	if err := dw.doInit(); err != nil {
		return err
	}

	return nil
}

func (dw *Dataway) String() string {
	arr := []string{fmt.Sprintf("dataways: [%s]", strings.Join(dw.URLs, ","))}

	for _, x := range dw.eps {
		arr = append(arr, "---------------------------------")
		for k, v := range x.categoryURL {
			arr = append(arr, fmt.Sprintf("% 24s: %s", k, v))
		}
	}

	return strings.Join(arr, "\n")
}

func (dw *Dataway) ClientsCount() int {
	return len(dw.eps)
}

func (dw *Dataway) IsLogFilter() bool {
	return len(dw.eps) == 1
}

func (dw *Dataway) GetTokens() []string {
	var arr []string
	for _, ep := range dw.eps {
		if ep.token != "" {
			arr = append(arr, ep.token)
		}
	}

	return arr
}

func (dw *Dataway) doInit() error {
	log = logger.SLogger("dataway")

	// 如果 env 已传入了 dataway 配置, 则不再追加老的 dataway 配置,
	// 避免俩边配置了同样的 dataway, 造成数据混乱
	if dw.DeprecatedURL != "" && len(dw.URLs) == 0 {
		dw.URLs = []string{dw.DeprecatedURL}
	}

	if len(dw.URLs) == 0 {
		return fmt.Errorf("dataway not set: urls is empty")
	}

	if dw.HTTPTimeout == "" {
		dw.HTTPTimeout = "30s"
	}

	if dw.MaxIdleConnsPerHost == 0 {
		dw.MaxIdleConnsPerHost = 64
	}

	du, err := time.ParseDuration(dw.HTTPTimeout)
	if err != nil {
		return err
	}
	dw.httpTimeout = du

	for _, s := range dw.Sinkers {
		if err := s.Setup(); err != nil {
			log.Warnf("sinker %s setup failed: %s", s.String(), err.Error())
		}
	}

	for _, u := range dw.URLs {
		ep, err := newEndpoint(u,
			withProxy(dw.HTTPProxy),
			withAPIs(dwAPIs),
			withHTTPTimeout(dw.httpTimeout),
			withHTTPTrace(dw.EnableHTTPTrace),
			withMaxHTTPIdleConnectionPerHost(dw.MaxIdleConnsPerHost),
		)
		if err != nil {
			log.Errorf("init dataway url %s failed: %s", u, err.Error())
			return err
		}

		dw.eps = append(dw.eps, ep)

		dw.addDNSCache(ep.host)
	}

	return nil
}

func (dw *Dataway) addDNSCache(host string) {
	for _, v := range dw.dnsCachers {
		if v.GetDomain() == host {
			return // avoid repeat add same domain
		}
	}

	dnsCache := &dnsCacher{}
	dnsCache.initDNSCache(host, dw.initEndpoints)

	dw.dnsCachers = append(dw.dnsCachers, dnsCache)
}

func (dw *Dataway) initEndpoints() error {
	dw.locker.Lock()
	defer dw.locker.Unlock()

	for _, ep := range dw.eps {
		if err := ep.setupHTTP(); err != nil {
			return err
		}
	}

	return nil
}
