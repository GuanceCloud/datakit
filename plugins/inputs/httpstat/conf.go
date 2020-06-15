package httpstat

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	pluginName = "httpstat"

	description = `stat http protocol request time, contain dnsLookup, tcpConnection, tlsHandshake,
	serverProcessing, contentTransfer, and total time`
	httpstatConfigSample = `
#    [[httpstat]]
#    ##if empty, use "httpstat"
#    metricName = ''
#    timeout = ''
#    ## default is 10s
#    interval = '10s'
#    [[httpstat.action]]
#    url = ""
#    method = ""
#    playload = ""
#    kAlive = true
#    tlsSkipVerify = true 
#    compress = true
`
)

type HttpstatCfg struct {
	MetricName string            `toml:"metricName"`
	Timeout    string            `toml:"timeout"`
	Interval   internal.Duration `toml:"interval"`
	Actions    []*Action         `toml:"action"`
}

type Action struct {
	Url           string `toml:"url"`
	Method        string `toml:"method"`
	Playload      string `toml:"playload"`
	KAlive        bool   `toml:"kAlive"`
	TLSSkipVerify bool   `toml:"tlsSkipVerify`
	Compress      bool   `toml:"compress`
}
