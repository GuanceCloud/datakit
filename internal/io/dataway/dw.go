// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dataway implement API request to dataway.
package dataway

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/compact"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/endpoint"
	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

const (
	HeaderXGlobalTags   = "X-Global-Tags"
	HeaderXGlobalTagsV2 = "X-Global-Tags-V2" // with header key/value URL encoded

	HeaderXStorageIndexName = "X-Storage-Index-Name"

	// DeprecatedDefaultMaxRawBodySize will cause too many memory, we set it to
	// 1MB. Set 1MB because the max-log length(message) is 1MB at storage side.
	DeprecatedDefaultMaxRawBodySize = 10 * (1 << 20)  // 10MB
	DefaultMaxRawBodySize           = (1 << 20)       // 1MB
	MinimalRawBodySize              = 100 * (1 << 10) // 100KB
)

type IDataway interface {
	Write(...compact.WriteOption) error
	Pull(what string) ([]byte, error)
}

var (
	datawayAPIs = []string{
		point.MetricDeprecated.URL(),
		point.Metric.URL(),
		point.Network.URL(),
		point.KeyEvent.URL(),
		point.Object.URL(),
		point.ObjectChange.URL(), // Deprecated.
		point.CustomObject.URL(),
		point.Logging.URL(),
		point.Tracing.URL(),
		point.RUM.URL(),
		point.Security.URL(),
		point.Profiling.URL(),
		datakit.BrowserScreenshotUpload,

		datakit.DatakitPull,
		datakit.LogFilter,
		datakit.SessionReplayUpload,
		datakit.SessionReplayAssetUpload,
		datakit.HeartBeat,
		datakit.Election,
		datakit.ElectionHeartbeat,
		datakit.QueryRaw,
		datakit.Workspace,
		datakit.ListDataWay,
		datakit.ObjectLabel,
		datakit.LogUpload,
		datakit.BugReportUpload,
		datakit.PipelinePull,
		datakit.ProfilingUpload,
		datakit.TokenCheck,
		datakit.SessionReplayAssetCheck,
		datakit.SessionReplayAssetGet,
		datakit.UsageTrace,
		datakit.NTPSync,
		datakit.RemoteJob,
		datakit.EnvVariable,
		datakit.Aggregate,
		datakit.TailSamplingConfig,
		datakit.TailSampling,
	}

	AvailableDataways          = []string{}
	l                          = logger.DefaultSLogger("dataway")
	datawayListIntervalDefault = 60
)

func NewDefaultDataway(opts ...DWOption) *Dataway {
	dw := &Dataway{
		URLs:                 []string{},
		HTTPTimeout:          30 * time.Second,
		IdleTimeout:          90 * time.Second,
		DropExpiredPackageAt: 12 * time.Hour,
		MaxRawBodySize:       DefaultMaxRawBodySize,
		GlobalCustomerKeys:   []string{},
		ContentEncoding:      "v2",
		GZip:                 true,

		MaxRetryCount:     endpoint.DefaultRetryCount,
		RetryDelay:        endpoint.DefaultRetryDelay,
		IPFamilyPolicy:    dnet.IPFamilyPolicyPreferIPv6,
		IPv4FallbackDelay: dnet.DefaultIPv4FallbackDelay,

		EnableHTTPTrace: true,

		NTP: &ntp{
			Enable:     true,
			Interval:   time.Minute * 5,
			SyncOnDiff: time.Second * 30,
		},

		walq: map[point.Category]*WALQueue{},
		WAL: &WALConf{
			MaxCapacityGB: 2.0,

			NoPos:           false,
			PosDumpAt:       100,
			PosDumpInterval: 100 * time.Millisecond,

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
	Enable     bool          `toml:"enable"`
	Interval   time.Duration `toml:"interval"`
	SyncOnDiff time.Duration `toml:"diff"`
}

type Dataway struct {
	URLs []string `toml:"urls"`

	DeprecatedHTTPTimeout string        `toml:"timeout,omitempty"`
	HTTPTimeout           time.Duration `toml:"timeout_v2"`
	MaxRetryCount         int           `toml:"max_retry_count"`
	RetryDelay            time.Duration `toml:"retry_delay"`
	IPFamilyPolicy        string        `toml:"ip_family_policy"`
	IPv4FallbackDelay     time.Duration `toml:"ipv4_fallback_delay"`

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

	IdleTimeout          time.Duration `toml:"idle_timeout"`
	DropExpiredPackageAt time.Duration `toml:"drop_expired_package_at"`

	GZip bool `toml:"gzip"`

	PayloadObfuscation string `toml:"payload_obfuscation"`

	EnableHTTPTrace bool `toml:"enable_httptrace"`
	EnableSinker    bool `toml:"enable_sinker"`

	InsecureSkipVerify bool `toml:"tls_insecure"`

	SinkerHeaderVersion string   `toml:"sinker_header_version"`
	GlobalCustomerKeys  []string `toml:"global_customer_keys"`
	WAL                 *WALConf `toml:"wal"`

	eps []*endpoint.EndPoint

	walq    map[point.Category]*WALQueue
	walFail *WALQueue

	locker     sync.RWMutex
	dnsCachers []*endpoint.DNSCacher

	globalTags                map[string]string
	globalTagsHTTPHeaderValue string

	NTP *ntp `toml:"ntp"`

	Token string `toml:"-"` // fast path to get main token
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
		dw.globalTagsHTTPHeaderValue = dw.sinkHeaderValueFromGlobalTags()
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
		for k, v := range x.CategoryURL {
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
		if ep.Token != "" {
			arr = append(arr, ep.Token)
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

func TagHeaderValueV2(tags map[string]string) string {
	var arr []string
	for k, v := range tags {
		arr = append(arr, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
	}
	sort.Strings(arr)
	return strings.Join(arr, ",")
}

func (dw *Dataway) sinkHeaderValueFromGlobalTags() string {
	if dw.SinkerHeaderVersion == "v2" {
		return TagHeaderValueV2(dw.globalTags)
	}
	return TagHeaderValue(dw.globalTags)
}

func (dw *Dataway) sinkHeaderKey() string {
	if dw.SinkerHeaderVersion == "v2" {
		return HeaderXGlobalTagsV2
	}

	return HeaderXGlobalTags
}

// SinkHeaderKey returns the sinker global tags header name.
func (dw *Dataway) SinkHeaderKey() string {
	return dw.sinkHeaderKey()
}

// SinkHeaderValueFromTags returns the sinker header value from point tags.
func (dw *Dataway) SinkHeaderValueFromTags(tags map[string]string) string {
	xGlobalTag := sinkHeaderValueFromTags(tags,
		dw.GlobalTags(),
		dw.CustomTagKeys(),
		dw.SinkerHeaderVersion == "v2")
	if xGlobalTag == "" {
		xGlobalTag = dw.GlobalTagsHTTPHeaderValue()
	}

	return xGlobalTag
}

var defaultInvalidDatawayURL = "https://guance.openway.com?token=YOUR-WORKSPACE-TOKEN"

func (dw *Dataway) doInit() error {
	l = logger.SLogger("dataway")
	// 如果 env 已传入了 dataway 配置, 则不再追加老的 dataway 配置,
	// 避免俩边配置了同样的 dataway, 造成数据混乱
	if dw.DeprecatedURL != "" && len(dw.URLs) == 0 {
		dw.URLs = []string{dw.DeprecatedURL}
	}

	dw.contentEncoding = point.EncodingStr(dw.ContentEncoding)

	if dw.PayloadObfuscation != "" {
		switch {
		case dw.PayloadObfuscation != uhttp.PayloadObfuscationGzipCaesarV1:
			l.Warnf("unsupported dataway payload obfuscation %q, disabled", dw.PayloadObfuscation)
			dw.PayloadObfuscation = ""
		case !dw.GZip:
			l.Warn("dataway payload obfuscation requires gzip, disabled")
			dw.PayloadObfuscation = ""
		}
	}

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

	if dw.IPFamilyPolicy == "" {
		dw.IPFamilyPolicy = dnet.IPFamilyPolicyPreferIPv6
	}
	if err := dnet.ValidateIPFamilyPolicy(dw.IPFamilyPolicy); err != nil {
		return err
	}
	if dw.IPv4FallbackDelay <= 0 {
		dw.IPv4FallbackDelay = dnet.DefaultIPv4FallbackDelay
	}
	l.Infof("dataway IP family policy: %s, IPv4 fallback delay: %s",
		dw.IPFamilyPolicy, dw.IPv4FallbackDelay)

	l.Infof("set %d global tags to dataway", len(dw.globalTags))
	if len(dw.globalTags) > 0 && dw.EnableSinker {
		dw.globalTagsHTTPHeaderValue = dw.sinkHeaderValueFromGlobalTags()
	}

	for _, u := range dw.URLs {
		var connectionLogged atomic.Bool
		ep, err := endpoint.NewEndpoint(u,
			endpoint.WithOwner("dataway"),
			endpoint.WithProxy(dw.HTTPProxy),
			endpoint.WithInsecureSkipVerify(dw.InsecureSkipVerify),
			endpoint.WithAPIs(datawayAPIs),
			endpoint.WithHTTPHeaders(map[string]string{
				// HeaderXGlobalTags: dw.globalTagsHTTPHeaderValue,

				// DatakitUserAgent define HTTP User-Agent header.
				// user-agent format. See
				// 	 https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
				"User-Agent": fmt.Sprintf("datakit-%s-%s/%s/%s",
					runtime.GOOS, runtime.GOARCH, git.Version, datakit.DKHost),
				"Referer": "DataKit",
			}),
			endpoint.WithHTTPTimeout(dw.HTTPTimeout),
			endpoint.WithHTTPTrace(dw.EnableHTTPTrace),
			endpoint.WithMaxHTTPIdleConnectionPerHost(dw.MaxIdleConnsPerHost),
			endpoint.WithMaxHTTPConnections(dw.MaxIdleConns),
			endpoint.WithHTTPIdleTimeout(dw.IdleTimeout),
			endpoint.WithMaxRetryCount(dw.MaxRetryCount),
			endpoint.WithRetryDelay(dw.RetryDelay),
			endpoint.WithDNSDialOptions(dnet.DNSCacheDialOptions{
				IPFamilyPolicy: dw.IPFamilyPolicy,
				FallbackDelay:  dw.IPv4FallbackDelay,
				DialDone: func(family, result string, elapsed time.Duration) {
					datawayDialTotal.WithLabelValues(family, result).Inc()
					datawayDialSeconds.WithLabelValues(family).Observe(elapsed.Seconds())
				},
				FallbackStarted: func() {
					datawayIPv4FallbackTotal.WithLabelValues().Inc()
				},
				ConnectionOpened: func(family, remote string) {
					datawayActiveConnections.WithLabelValues(family).Inc()
					if connectionLogged.CompareAndSwap(false, true) {
						l.Infof("dataway connection established: policy=%s, family=%s, remote=%s",
							dw.IPFamilyPolicy, family, remote)
					}
				},
				ConnectionClosed: func(family string) {
					datawayActiveConnections.WithLabelValues(family).Dec()
				},
			}),
		)
		if err != nil {
			l.Errorf("init dataway url %s failed: %s", u, err.Error())
			return err
		}

		dw.eps = append(dw.eps, ep)

		dw.addDNSCache(ep.Hostname)
	}

	// set main token
	if len(dw.eps) > 0 {
		dw.Token = dw.eps[0].Token
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

	dnsCache := &endpoint.DNSCacher{}
	dnsCache.InitDNSCache(host, dw.initEndpoints)

	dw.dnsCachers = append(dw.dnsCachers, dnsCache)
}

func (dw *Dataway) initEndpoints() error {
	dw.locker.Lock()
	defer dw.locker.Unlock()

	for _, ep := range dw.eps {
		if err := ep.SetupHTTP(); err != nil {
			return err
		}
	}

	return nil
}

func (dw *Dataway) doGroupPoints(ptg *ptGrouper, cat point.Category, points []*point.Point) {
	for _, pt := range points {
		// clear kvs for current pt
		ptg.kvarr = ptg.kvarr[:0]
		ptg.extKVs = ptg.extKVs[:0]

		ptg.pt = pt
		ptg.cat = cat

		tv := ptg.sinkHeaderValue(dw.globalTags, dw.GlobalCustomerKeys)

		l.Debugf("add point to group %q", tv)

		ptg.groupedPts[tv] = append(ptg.groupedPts[tv], pt)
	}
}

func (dw *Dataway) groupPoints(ptg *ptGrouper,
	cat point.Category,
	points []*point.Point,
) {
	dw.doGroupPoints(ptg, cat, points)
	groupedRequestVec.WithLabelValues(cat.String()).Observe(float64(len(ptg.groupedPts)))
}

// SinkPointGroup groups points that should share the same sinker header.
type SinkPointGroup struct {
	HeaderKey   string
	HeaderValue string
	Points      []*point.Point
}

// GroupPointsBySinkHeader groups points using the same sinker rules as normal
// writes. If sinker is disabled, it returns a single group without headers.
func (dw *Dataway) GroupPointsBySinkHeader(cat point.Category, points []*point.Point) []SinkPointGroup {
	if dw == nil || !dw.sinkEnabled() {
		return []SinkPointGroup{{Points: points}}
	}

	ptg := getGrouper()
	defer putGrouper(ptg)
	ptg.safe = dw.SinkerHeaderVersion == "v2"

	dw.doGroupPoints(ptg, cat, points)

	headerKey := dw.sinkHeaderKey()
	groups := make([]SinkPointGroup, 0, len(ptg.groupedPts))
	for headerValue, pts := range ptg.groupedPts {
		groups = append(groups, SinkPointGroup{
			HeaderKey:   headerKey,
			HeaderValue: headerValue,
			Points:      []*point.Point(pts),
		})
	}

	return groups
}

func (dw *Dataway) Write(opts ...compact.WriteOption) error {
	gzOn := compact.GzipNotSet
	if dw.GZip {
		gzOn = compact.GzipSet
	}

	w := compact.GetWriter(
		// set content encoding(protobuf/line-protocol/json)
		compact.WithHTTPEncoding(dw.contentEncoding),
		// setup gzip on or off
		compact.WithGzip(gzOn),
		// set raw body size limit
		compact.WithMaxBodyCap(dw.MaxRawBodySize),
	)

	defer compact.PutWriter(w)

	// Append extra wirte options from caller
	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}

	// apply index name to HTTP header.
	if w.IndexName != "" {
		compact.WithHTTPHeader(HeaderXStorageIndexName, w.IndexName)(w)
	}

	if w.Callback == nil { // set default callback
		if w.NoWAL {
			if len(dw.eps) == 0 {
				return fmt.Errorf("no endpoints on dataway, should not been here")
			}

			// NOTE: only send to 1st dataway endpoint.
			w.Callback = func(w *compact.Writer, b *compact.Body) error {
				return dw.writePointData(dw.eps[0], w, b)
			}
		} else {
			w.Callback = dw.enqueueBody // enqueu to WAL
		}
	}

	// split single point array into multiple part according to
	// different X-Global-Tags.
	if dw.sinkEnabled() {
		l.Debugf("under sinker...")

		ptg := getGrouper()
		defer putGrouper(ptg)
		ptg.safe = dw.SinkerHeaderVersion == "v2"

		dw.groupPoints(ptg, w.Category, w.Points)

		headerKey := dw.sinkHeaderKey()
		for k, points := range ptg.groupedPts {
			compact.WithHTTPHeader(headerKey, k)(w)
			compact.WithPoints(points)(w)

			if err := w.BuildPointsBody(); err != nil {
				return err
			}
		}
	} else {
		if err := w.BuildPointsBody(); err != nil {
			return err
		}
	}

	return nil
}

func (dw *Dataway) writePointData(ep *endpoint.EndPoint, w *compact.Writer, b *compact.Body) error {
	mode := dw.PayloadObfuscation
	if mode == "" || w.Gzip != compact.GzipSet {
		return ep.WritePointData(w, b)
	}

	cat := b.Cat()
	path := cat.URL()
	if cat == point.DynamicDWCategory {
		u, err := url.ParseRequestURI(w.DynamicURL)
		if err != nil {
			return ep.WritePointData(w, b)
		}
		path = u.Path
		cat = point.CatURL(path)
		if cat == point.UnknownCategory {
			return ep.WritePointData(w, b)
		}
	}
	if cat == point.UnknownCategory ||
		cat == point.ObjectChange ||
		cat == point.LLM ||
		cat == point.AgentLLM ||
		!strings.HasPrefix(path, "/v1/write/") {
		return ep.WritePointData(w, b)
	}

	if err := uhttp.ObfuscatePayload(mode, b.Buf()); err != nil {
		l.Warnf("obfuscate dataway payload: %s, send standard gzip", err)
		return ep.WritePointData(w, b)
	}

	previousHeader, hadHeader := w.HTTPHeaders[uhttp.PayloadObfuscationHeader]
	compact.WithHTTPHeader(uhttp.PayloadObfuscationHeader, mode)(w)
	defer func() {
		if hadHeader {
			w.HTTPHeaders[uhttp.PayloadObfuscationHeader] = previousHeader
		} else {
			delete(w.HTTPHeaders, uhttp.PayloadObfuscationHeader)
		}
	}()
	defer func() {
		if err := uhttp.DeobfuscatePayload(mode, b.Buf()); err != nil {
			l.Errorf("restore dataway payload: %s", err)
		}
	}()

	return ep.WritePointData(w, b)
}

func (dw *Dataway) sinkEnabled() bool {
	return dw.EnableSinker &&
		(len(dw.globalTags) > 0 || len(dw.GlobalCustomerKeys) > 0) &&
		len(dw.eps) > 0
}
