// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dataway implement all dataway API request.
package dataway

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var (
	apis = []string{
		datakit.MetricDeprecated,
		datakit.DatakitPull,
		datakit.Metric,
		datakit.Network,
		datakit.KeyEvent,
		datakit.Object,
		datakit.CustomObject,
		datakit.Logging,
		datakit.LogFilter,
		datakit.Tracing,
		datakit.RUM,
		datakit.SessionReplayUpload,
		datakit.Security,
		datakit.HeartBeat,
		datakit.Election,
		datakit.ElectionHeartbeat,
		datakit.QueryRaw,
		datakit.Workspace,
		datakit.ListDataWay,
		datakit.ObjectLabel,
		datakit.LogUpload,
		datakit.PipelinePull,
		datakit.Profiling,
		datakit.ProfilingUpload,
	}

	ExtraHeaders               = map[string]string{}
	AvailableDataways          = []string{}
	log                        = logger.DefaultSLogger("dataway")
	datawayListIntervalDefault = 60
	heartBeatIntervalDefault   = 30
)

type DataWayDefault struct {
	*DataWayCfg
	endPoints  []*endPoint
	ontest     bool
	httpCli    *retryablehttp.Client
	locker     sync.RWMutex
	dnsCachers []*dnsCacher
}

type endPoint struct {
	url         string
	host        string
	scheme      string
	proxy       string
	urlValues   url.Values
	categoryURL map[string]string
	ontest      bool
	fails       int
	dw          *DataWayDefault // reference
}

func (dw *DataWayDefault) Init(cfg *DataWayCfg) error {
	if cfg == nil {
		return fmt.Errorf("init dataway error: empty dataway config")
	}

	dw.DataWayCfg = cfg

	if err := dw.Apply(); err != nil {
		return err
	}

	return nil
}

func (dw *DataWayDefault) String() string {
	arr := []string{fmt.Sprintf("dataways: [%s]", strings.Join(dw.URLs, ","))}

	for _, x := range dw.endPoints {
		arr = append(arr, "---------------------------------")
		for k, v := range x.categoryURL {
			arr = append(arr, fmt.Sprintf("% 24s: %s", k, v))
		}
	}

	return strings.Join(arr, "\n")
}

func (dw *DataWayDefault) ClientsCount() int {
	return len(dw.endPoints)
}

func (dw *DataWayDefault) IsLogFilter() bool {
	return len(dw.endPoints) == 1
}

func (dw *DataWayDefault) GetTokens() []string {
	resToken := []string{}
	for _, ep := range dw.endPoints {
		if ep.urlValues != nil {
			token := ep.urlValues.Get("token")
			if token != "" {
				resToken = append(resToken, token)
			}
		}
	}

	return resToken
}

var tokenFormatMap = map[string]int{
	"token_": 32,
	"tkn_":   32,
	"tokn_":  24,
}

func (dw *DataWayDefault) CheckToken(token string) (err error) {
	err = fmt.Errorf("token invalid format")

	parts := strings.Split(token, "_")

	if len(parts) == 2 {
		prefix := parts[0] + "_"
		tokenVal := parts[1]

		if tokenLen, ok := tokenFormatMap[prefix]; ok {
			if len(tokenVal) == tokenLen {
				err = nil
			}
		}
	}

	return
}

func (dw *DataWayDefault) Apply() error {
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

	if dw.MaxFails == 0 {
		dw.MaxFails = 20
	}

	timeout, err := time.ParseDuration(dw.HTTPTimeout)
	if err != nil {
		return err
	}

	dw.TimeoutDuration = timeout

	if err := dw.initHTTP(); err != nil {
		return err
	}

	dw.endPoints = dw.endPoints[:0]

	for _, httpurl := range dw.URLs {
		ep, err := dw.initEndpoint(httpurl)
		if err != nil {
			log.Errorf("init dataway url %s failed: %s", httpurl, err.Error())
			return err
		}

		dw.endPoints = append(dw.endPoints, ep)

		dw.addDNSCache(ep.host)
	}

	return nil
}

func (dw *DataWayDefault) addDNSCache(host string) {
	for _, v := range dw.dnsCachers {
		if v.GetDomain() == host {
			return // avoid repeat add same domain
		}
	}

	dnsCache := &dnsCacher{}
	dnsCache.initDNSCache(host, dw.initHTTP)

	dw.dnsCachers = append(dw.dnsCachers, dnsCache)
}

func (dw *DataWayDefault) initEndpoint(httpurl string) (*endPoint, error) {
	u, err := url.ParseRequestURI(httpurl)
	if err != nil {
		log.Errorf("parse dataway url %s failed: %s", httpurl, err.Error())
		return nil, err
	}

	cli := &endPoint{
		url:         httpurl,
		scheme:      u.Scheme,
		urlValues:   u.Query(),
		host:        u.Host,
		categoryURL: map[string]string{},
		ontest:      dw.ontest,
		proxy:       dw.HTTPProxy,
		dw:          dw, // reference
	}

	for _, api := range apis {
		if cli.urlValues.Encode() != "" {
			cli.categoryURL[api] = fmt.Sprintf("%s://%s%s?%s",
				cli.scheme,
				cli.host,
				api,
				cli.urlValues.Encode())
		} else {
			cli.categoryURL[api] = fmt.Sprintf("%s://%s%s",
				cli.scheme,
				cli.host,
				api)
		}
	}

	return cli, nil
}

const (
	defaultDNSCacheFreq          = time.Minute
	defaultDNSCacheLookUpTimeout = 10 * time.Second
)

func (dw *DataWayDefault) initHTTP() error {
	dw.locker.Lock()
	defer dw.locker.Unlock()

	dialContext, err := dnet.GetDNSCacheDialContext(defaultDNSCacheFreq, defaultDNSCacheLookUpTimeout)
	if err != nil {
		log.Warnf("GetDNSCacheDialContext failed: %v", err)
		dialContext = nil // if failed, then not use dns cache.
	}

	cliopts := &ihttp.Options{
		DialTimeout:         dw.TimeoutDuration,
		MaxIdleConnsPerHost: dw.MaxIdleConnsPerHost,
		DialContext:         dialContext,
	}

	if dw.HTTPProxy != "" { // set proxy
		if u, err := url.ParseRequestURI(dw.HTTPProxy); err != nil {
			log.Warnf("parse http proxy failed err: %s, ignored", err.Error())
		} else {
			cliopts.ProxyURL = u
			log.Infof("set dataway proxy to %s ok", dw.HTTPProxy)
		}
	}

	dw.httpCli = newRetryCli(cliopts, dw.TimeoutDuration)

	return nil
}

func (dw *DataWayDefault) DatakitPull(args string) ([]byte, error) {
	if dw.ClientsCount() == 0 {
		return nil, fmt.Errorf("dataway URL not set")
	}

	return dw.endPoints[0].datakitPull(args)
}

type DatawayWriteOpt struct {
	URLs     []string
	Proxy    string
	Category string
	Points   []*point.Point
}
