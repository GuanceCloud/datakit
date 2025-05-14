// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	measurementNginx        = "nginx"
	measurementServerZone   = "nginx_server_zone"
	measurementUpstreamZone = "nginx_upstream_zone"
	measurementCacheZone    = "nginx_cache_zone"
	measurementLocationZone = "nginx_location_zone"

	NestGeneral       = "general"
	NestServerZone    = "server_zones"
	NestUpstreams     = "upstreams"
	NestCaches        = "caches"
	NestConnections   = "connections"
	NestLocationZones = "location_zones"
)

type PlusAPI struct {
	endpoint string
	nest     string
}

var PlusAPIEndpoints = []PlusAPI{
	{"nginx", NestGeneral},
	{"http/server_zones", NestServerZone},
	{"http/upstreams", NestUpstreams},
	{"http/caches", NestCaches},
	{"connections", NestConnections},
	{"http/location_zones", NestLocationZones},
}

type ngxlog struct {
	Files    []string `toml:"files"`
	Pipeline string   `toml:"pipeline"`
}

type Input struct {
	URLsDeprecated []string `toml:"urls,omitempty"`

	URL             string `toml:"url"`
	Ports           [2]int `toml:"ports"`
	host            string
	path            string
	PlusAPIURL      string            `toml:"plus_api_url"`
	Interval        time.Duration     `toml:"interval"`
	ResponseTimeout time.Duration     `toml:"response_timeout"`
	UseVts          bool              `toml:"use_vts"`
	UsePlusAPI      bool              `toml:"use_plus_api"`
	Log             *ngxlog           `toml:"log"`
	Tags            map[string]string `toml:"tags"`
	mergedTags      map[string]string

	UpState int

	Version            string
	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	tls.ClientConfig
	// HTTP client
	client *http.Client

	// alignTS   int64
	// startTime time.Time
	tail *tailer.Tailer

	lastErr error

	collectCache []*point.Point

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
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

type NginxPlusAPIResponse struct {
	General   Nginx
	Servers   map[string]ServerP
	Upstreams map[string]UpstreamsP
	Caches    map[string]CachesP
	Locations map[string]LocationP

	tags map[string]string
}
type Nginx struct {
	Version       string    `json:"version"`
	Build         string    `json:"build"`
	Address       string    `json:"address"`
	Generation    uint64    `json:"generation"`
	LoadTimestamp time.Time `json:"load_timestamp"`
	Timestamp     time.Time `json:"timestamp"`
	Pid           uint64    `json:"pid"`
	Ppid          uint64    `json:"ppid"`
}

type ServerP struct {
	Processing uint64 `json:"processing"`
	Requests   uint64 `json:"requests"`
	Responses  struct {
		OneXX   uint64 `json:"1xx"`
		TwoXX   uint64 `json:"2xx"`
		ThreeXX uint64 `json:"3xx"`
		FourXX  uint64 `json:"4xx"`
		FiveXX  uint64 `json:"5xx"`
		Codes   struct {
			Code200 uint64 `json:"200"`
			Code301 uint64 `json:"301"`
			Code404 uint64 `json:"404"`
			Code503 uint64 `json:"503"`
		} `json:"codes"`
		Total uint64 `json:"total"`
	} `json:"responses"`
	Discarded uint64 `json:"discarded"`
	Received  uint64 `json:"received"`
	Sent      uint64 `json:"sent"`
}

type UpstreamsP struct {
	Peers     []Peer `json:"peers"`
	Keepalive uint64 `json:"keepalive"`
	Zombies   uint64 `json:"zombies"`
	Zone      string `json:"zone"`
}

type Peer struct {
	ID           uint64 `json:"id"`
	Server       string `json:"server"`
	Name         string `json:"name"`
	Backup       bool   `json:"backup"`
	Weight       uint64 `json:"weight"`
	State        string `json:"state"`
	Active       uint64 `json:"active"`
	Requests     uint64 `json:"requests"`
	HeaderTime   uint64 `json:"header_time"`
	ResponseTime uint64 `json:"response_time"`
	Responses    struct {
		OneXX   uint64 `json:"1xx"`
		TwoXX   uint64 `json:"2xx"`
		ThreeXX uint64 `json:"3xx"`
		FourXX  uint64 `json:"4xx"`
		FiveXX  uint64 `json:"5xx"`
		Total   uint64 `json:"total"`
	} `json:"responses"`
	Sent         uint64 `json:"sent"`
	Received     uint64 `json:"received"`
	Fails        uint64 `json:"fails"`
	Unavail      uint64 `json:"unavail"`
	HealthChecks struct {
		Checks     uint64 `json:"checks"`
		Fails      uint64 `json:"fails"`
		Unhealthy  uint64 `json:"unhealthy"`
		LastPassed bool   `json:"last_passed"`
	} `json:"health_checks"`
	Downtime uint64 `json:"downtime"`
}

type CachesP struct {
	Size    uint64 `json:"size"`
	MaxSize uint64 `json:"max_size"`
	Cold    bool   `json:"cold"`
	Hit     struct {
		Responses uint64 `json:"responses"`
		Bytes     uint64 `json:"bytes"`
	} `json:"hit"`
	Stale struct {
		Responses uint64 `json:"responses"`
		Bytes     uint64 `json:"bytes"`
	} `json:"stale"`
	Updating struct {
		Responses uint64 `json:"responses"`
		Bytes     uint64 `json:"bytes"`
	} `json:"updating"`
	Revalidated struct {
		Responses uint64 `json:"responses"`
		Bytes     uint64 `json:"bytes"`
	} `json:"revalidated"`
	Miss struct {
		Responses uint64 `json:"responses"`
		Bytes     uint64 `json:"bytes"`
	} `json:"miss"`
	Expired struct {
		Responses        uint64 `json:"responses"`
		Bytes            uint64 `json:"bytes"`
		ResponsesWritten uint64 `json:"responses_written"`
		BytesWritten     uint64 `json:"bytes_written"`
	} `json:"expired"`
	Bypass struct {
		Responses        uint64 `json:"responses"`
		Bytes            uint64 `json:"bytes"`
		ResponsesWritten uint64 `json:"responses_written"`
		BytesWritten     uint64 `json:"bytes_written"`
	} `json:"bypass"`
}

type LocationP struct {
	Requests  uint64 `json:"requests"`
	Responses struct {
		OneXX   uint64 `json:"1xx"`
		TwoXX   uint64 `json:"2xx"`
		ThreeXX uint64 `json:"3xx"`
		FourXX  uint64 `json:"4xx"`
		FiveXX  uint64 `json:"5xx"`
		Codes   struct {
			Code200 uint64 `json:"200"`
			Code301 uint64 `json:"301"`
			Code404 uint64 `json:"404"`
			Code503 uint64 `json:"503"`
		} `json:"codes"`
		Total uint64 `json:"total"`
	} `json:"responses"`
	Discarded uint64 `json:"discarded"`
	Received  uint64 `json:"received"`
	Sent      uint64 `json:"sent"`
}
