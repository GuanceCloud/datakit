package httpstat

import "time"

const (
	description = `stat http protocol request time, contain dnsLookup, tcpConnection, tlsHandshake,
	serverProcessing, contentTransfer, and total time`
	httpstatConfigSample = `
#[[inputs.httpstat]]
#  ## default is 10s
#  interval = '10s'
#  [[inputs.httpstat.action]]
#    url = ""
#    method = ""
#    playload = ""
#    kAlive = true
#    tlsSkipVerify = true
#    compress = true
`
)

type Httpstat struct {
	MetricName       string    `toml:"metricName"`
	Interval         string    `toml:"interval"`
	Actions          []*Action `toml:"action"`
	httpPing         []*httpPing
	IntervalDuration time.Duration `json:"-" toml:"-"`
	test             bool          `toml:"-"`
	resData          []byte        `toml:"-"`
}

type Action struct {
	Url           string `toml:"url"`
	Method        string `toml:"method"`
	Playload      string `toml:"playload"`
	KAlive        bool   `toml:"kAlive"`
	TLSSkipVerify bool   `toml:"tlsSkipVerify"`
	Compress      bool   `toml:"compress"`
}
