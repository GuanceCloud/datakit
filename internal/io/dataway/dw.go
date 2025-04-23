// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dataway implement API request to dataway.
package dataway

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
)

const (
	HeaderXGlobalTags = "X-Global-Tags"
	DefaultRetryCount = 1
	DefaultRetryDelay = time.Second

	// DeprecatedDefaultMaxRawBodySize will cause too many memory, we set it to
	// 1MB. Set 1MB because the max-log length(message) is 1MB at storage side.
	DeprecatedDefaultMaxRawBodySize = 10 * (1 << 20)  // 10MB
	DefaultMaxRawBodySize           = (1 << 20)       // 1MB
	MinimalRawBodySize              = 100 * (1 << 10) // 100KB
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
		point.ObjectChange.URL(),
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
		datakit.TokenCheck,
		datakit.UsageTrace,
		datakit.NTPSync,
		datakit.RemoteJob,
	}

	AvailableDataways          = []string{}
	l                          = logger.DefaultSLogger("dataway")
	datawayListIntervalDefault = 60
)

func NewDefaultDataway(opts ...DWOption) *Dataway {
	dw := &Dataway{
		URLs:               []string{},
		HTTPTimeout:        30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxRawBodySize:     DefaultMaxRawBodySize,
		GlobalCustomerKeys: []string{},
		ContentEncoding:    "v2",
		GZip:               true,
		MaxRetryCount:      DefaultRetryCount,
		RetryDelay:         DefaultRetryDelay,
		NTP: &ntp{
			Interval:   time.Minute * 5,
			SyncOnDiff: time.Second * 30,
		},

		walq: map[point.Category]*WALQueue{},
		WAL: &WALConf{
			MaxCapacityGB:          2.0,
			Path:                   filepath.Join(datakit.CacheDir, "dw-wal"),
			FailCacheCleanInterval: time.Second * 30,
		},
	}

	for _, opt := range opts {
		opt(dw)
	}

	return dw
}

type ntp struct {
	Interval   time.Duration `toml:"interval"`
	SyncOnDiff time.Duration `toml:"diff"`
}

type Dataway struct {
	URLs []string `toml:"urls"`

	DeprecatedHTTPTimeout string        `toml:"timeout,omitempty"`
	HTTPTimeout           time.Duration `toml:"timeout_v2"`
	MaxRetryCount         int           `toml:"max_retry_count"`
	RetryDelay            time.Duration `toml:"retry_delay"`

	HTTPProxy string `toml:"http_proxy"`

	Hostname string `toml:"-"`

	// Deprecated
	DeprecatedHost   string `toml:"host,omitempty"`
	DeprecatedScheme string `toml:"scheme,omitempty"`
	DeprecatedToken  string `toml:"token,omitempty"`
	DeprecatedURL    string `toml:"url,omitempty"`

	// limit HTTP underlying TCP connections.
	MaxIdleConnsPerHost int `toml:"max_idle_conns_per_host,omitempty"`
	MaxIdleConns        int `toml:"max_idle_conns"`

	// limit body size before gzip.
	MaxRawBodySize int `toml:"max_raw_body_size"`

	ContentEncoding string `toml:"content_encoding"`
	contentEncoding point.Encoding

	IdleTimeout time.Duration `toml:"idle_timeout"`

	GZip bool `toml:"gzip"`

	EnableHTTPTrace    bool `toml:"enable_httptrace"`
	EnableSinker       bool `toml:"enable_sinker"`
	InsecureSkipVerify bool `toml:"tls_insecure"`

	GlobalCustomerKeys []string `toml:"global_customer_keys"`
	WAL                *WALConf `toml:"wal"`

	eps []*endPoint

	walq    map[point.Category]*WALQueue
	walFail *WALQueue

	locker     sync.RWMutex
	dnsCachers []*dnsCacher

	globalTags                map[string]string
	globalTagsHTTPHeaderValue string

	NTP *ntp `toml:"ntp"`
}

// ParseGlobalCustomerKeys parse custom tag keys used for sinker.
func ParseGlobalCustomerKeys(v string) (arr []string) {
	for _, elem := range strings.Split(v, ",") { // remove white space
		if x := strings.TrimSpace(elem); len(x) > 0 {
			arr = append(arr, x)
		}
	}
	return
}

// UpdateGlobalTags hot-update dataway's global tags.
func (dw *Dataway) UpdateGlobalTags(tags map[string]string) {
	dw.locker.Lock()
	defer dw.locker.Unlock()
	dw.globalTags = tags
	l.Infof("set %d global tags to dataway", len(dw.globalTags))
	if len(dw.globalTags) > 0 && dw.EnableSinker {
		dw.globalTagsHTTPHeaderValue = TagHeaderValue(dw.globalTags)
	}
}

// Init setup current dataway.
//
// During Init(), we also accept options to update dataway's field after NewDefaultDataway().
func (dw *Dataway) Init(opts ...DWOption) error {
	l = logger.SLogger("dataway")

	for _, opt := range opts {
		if opt != nil {
			opt(dw)
		}
	}

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

	arr = append(arr, fmt.Sprintf("wal: %s, cap: %fGB", dw.WAL.Path, dw.WAL.MaxCapacityGB))

	return strings.Join(arr, "\n")
}

func (dw *Dataway) ClientsCount() int {
	return len(dw.eps)
}

// GetTokens list all dataway's tokens.
func (dw *Dataway) GetTokens() []string {
	var arr []string
	for _, ep := range dw.eps {
		if ep.token != "" {
			arr = append(arr, ep.token)
		}
	}

	return arr
}

// TagHeaderValue create X-Global-Tags header value in the
// form of key=val,key=val with ASC sorted.
func TagHeaderValue(tags map[string]string) string {
	var arr []string
	for k, v := range tags {
		arr = append(arr, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(arr)
	return strings.Join(arr, ",")
}

var defaultInvalidDatawayURL = "https://guance.openway.com?token=YOUR-WORKSPACE-TOKEN"

func (dw *Dataway) doInit() error {
	// 如果 env 已传入了 dataway 配置, 则不再追加老的 dataway 配置,
	// 避免俩边配置了同样的 dataway, 造成数据混乱
	if dw.DeprecatedURL != "" && len(dw.URLs) == 0 {
		dw.URLs = []string{dw.DeprecatedURL}
	}

	dw.contentEncoding = point.EncodingStr(dw.ContentEncoding)

	// set default raw body size to 10MB
	if dw.MaxRawBodySize == 0 {
		dw.MaxRawBodySize = DefaultMaxRawBodySize
	}

	if len(dw.URLs) == 0 {
		l.Warnf("dataway not set: urls is empty, set to %q", defaultInvalidDatawayURL)
		dw.URLs = append(dw.URLs, defaultInvalidDatawayURL)
	}

	if dw.HTTPTimeout <= time.Duration(0) {
		dw.HTTPTimeout = time.Second * 30
	}

	if dw.MaxIdleConnsPerHost == 0 {
		dw.MaxIdleConnsPerHost = 64
	}

	if dw.MaxRetryCount <= 0 {
		dw.MaxRetryCount = 1
	}

	if dw.MaxRetryCount > 10 {
		dw.MaxRetryCount = 10
	}

	l.Infof("set %d global tags to dataway", len(dw.globalTags))
	if len(dw.globalTags) > 0 && dw.EnableSinker {
		dw.globalTagsHTTPHeaderValue = TagHeaderValue(dw.globalTags)
	}

	for _, u := range dw.URLs {
		ep, err := newEndpoint(u,
			withProxy(dw.HTTPProxy),
			withInsecureSkipVerify(dw.InsecureSkipVerify),
			withAPIs(dwAPIs),
			withHTTPHeaders(map[string]string{
				// HeaderXGlobalTags: dw.globalTagsHTTPHeaderValue,

				// DatakitUserAgent define HTTP User-Agent header.
				// user-agent format. See
				// 	 https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
				"User-Agent": fmt.Sprintf("datakit-%s-%s/%s/%s",
					runtime.GOOS, runtime.GOARCH, git.Version, datakit.DatakitHostName),
			}),
			withHTTPTimeout(dw.HTTPTimeout),
			withHTTPTrace(dw.EnableHTTPTrace),
			withMaxHTTPIdleConnectionPerHost(dw.MaxIdleConnsPerHost),
			withMaxHTTPConnections(dw.MaxIdleConns),
			withHTTPIdleTimeout(dw.IdleTimeout),
			withMaxRetryCount(dw.MaxRetryCount),
			withRetryDelay(dw.RetryDelay),
		)
		if err != nil {
			l.Errorf("init dataway url %s failed: %s", u, err.Error())
			return err
		}

		if dw.EnableSinker {
			ep.httpHeaders[HeaderXGlobalTags] = dw.globalTagsHTTPHeaderValue
		}

		dw.eps = append(dw.eps, ep)

		dw.addDNSCache(ep.host)
	}

	return nil
}

// GlobalTags list all global tags of the dataway.
func (dw *Dataway) GlobalTags() map[string]string {
	return dw.globalTags
}

// CustomTagKeys list all custome keys of the dataway.
func (dw *Dataway) CustomTagKeys() []string {
	return dw.GlobalCustomerKeys
}

func (dw *Dataway) GlobalTagsHTTPHeaderValue() string {
	return dw.globalTagsHTTPHeaderValue
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
