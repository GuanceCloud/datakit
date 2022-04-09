package nginx

import (
	"net/http"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	nginx        = "nginx"
	ServerZone   = "nginx_server_zone"
	UpstreamZone = "nginx_upstream_zone"
	CacheZone    = "nginx_cache_zone"
)

type ngxlog struct {
	Files    []string `toml:"files"`
	Pipeline string   `toml:"pipeline"`
}

type Input struct {
	URLsDeprecated []string `toml:"urls,omitempty"`

	URL             string            `toml:"url"`
	Interval        datakit.Duration  `toml:"interval"`
	ResponseTimeout datakit.Duration  `toml:"response_timeout"`
	UseVts          bool              `toml:"use_vts"`
	Log             *ngxlog           `toml:"log"`
	Tags            map[string]string `toml:"tags"`

	tls.ClientConfig
	// HTTP client
	client *http.Client

	start time.Time
	tail  *tailer.Tailer

	lastErr error

	collectCache []inputs.Measurement

	pause   bool
	pauseCh chan bool

	semStop *cliutils.Sem // start stop signal
}

func (n *Input) LogExample() map[string]string {
	return map[string]string{
		"Nginx error log1": `2021/04/21 09:24:04 [alert] 7#7: *168 write() to "/var/log/nginx/access.log" failed (28: No space left on device) while logging request, client: 120.204.196.129, server: localhost, request: "GET / HTTP/1.1", host: "47.98.103.73"`,
		"Nginx error log2": `2021/04/29 16:24:38 [emerg] 50102#0: unexpected ";" in /usr/local/etc/nginx/nginx.conf:23`,
		"Nginx access log": `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"`,
	}
}

type NginxVTSResponse struct {
	HostName      string `json:"hostName"`
	Version       string `json:"nginxVersion"`
	LoadTimestamp int64  `json:"loadMsec"`
	Now           int64  `json:"nowMsec"`
	Connections   struct {
		Active   uint64 `json:"active"`
		Reading  uint64 `json:"reading"`
		Writing  uint64 `json:"writing"`
		Waiting  uint64 `json:"waiting"`
		Accepted uint64 `json:"accepted"`
		Handled  uint64 `json:"handled"`
		Requests uint64 `json:"requests"`
	} `json:"connections"`
	ServerZones   map[string]Server            `json:"serverZones"`
	FilterZones   map[string]map[string]Server `json:"filterZones"`
	UpstreamZones map[string][]Upstream        `json:"upstreamZones"`
	CacheZones    map[string]Cache             `json:"cacheZones"`

	tags map[string]string
}

type Server struct {
	RequestCounter uint64 `json:"requestCounter"`
	InBytes        uint64 `json:"inBytes"`
	OutBytes       uint64 `json:"outBytes"`
	RequestMsec    uint64 `json:"requestMsec"`
	Responses      struct {
		OneXx       uint64 `json:"1xx"`
		TwoXx       uint64 `json:"2xx"`
		ThreeXx     uint64 `json:"3xx"`
		FourXx      uint64 `json:"4xx"`
		FiveXx      uint64 `json:"5xx"`
		Miss        uint64 `json:"miss"`
		Bypass      uint64 `json:"bypass"`
		Expired     uint64 `json:"expired"`
		Stale       uint64 `json:"stale"`
		Updating    uint64 `json:"updating"`
		Revalidated uint64 `json:"revalidated"`
		Hit         uint64 `json:"hit"`
		Scarce      uint64 `json:"scarce"`
	} `json:"responses"`
}

type Upstream struct {
	Server         string `json:"server"`
	RequestCounter uint64 `json:"requestCounter"`
	InBytes        uint64 `json:"inBytes"`
	OutBytes       uint64 `json:"outBytes"`
	Responses      struct {
		OneXx   uint64 `json:"1xx"`
		TwoXx   uint64 `json:"2xx"`
		ThreeXx uint64 `json:"3xx"`
		FourXx  uint64 `json:"4xx"`
		FiveXx  uint64 `json:"5xx"`
	} `json:"responses"`
	ResponseMsec uint64 `json:"responseMsec"`
	RequestMsec  uint64 `json:"requestMsec"`
	Weight       uint64 `json:"weight"`
	MaxFails     uint64 `json:"maxFails"`
	FailTimeout  uint64 `json:"failTimeout"`
	Backup       bool   `json:"backup"`
	Down         bool   `json:"down"`
}

type Cache struct {
	MaxSize   uint64 `json:"maxSize"`
	UsedSize  uint64 `json:"usedSize"`
	InBytes   uint64 `json:"inBytes"`
	OutBytes  uint64 `json:"outBytes"`
	Responses struct {
		Miss        uint64 `json:"miss"`
		Bypass      uint64 `json:"bypass"`
		Expired     uint64 `json:"expired"`
		Stale       uint64 `json:"stale"`
		Updating    uint64 `json:"updating"`
		Revalidated uint64 `json:"revalidated"`
		Hit         uint64 `json:"hit"`
		Scarce      uint64 `json:"scarce"`
	} `json:"responses"`
}
