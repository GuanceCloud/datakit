package cshark

const (
	sharkConfigSample = `
#[[inputs.cshark]]
## tshark check频度
#interval = "10s"
## tshark 安装路径（$type tshark)
#tsharkPath = "/usr/bin/tshark"
`
)

type Shark struct {
	MetricName string
	Interval   string `toml:"interval"`
	TsharkPath string `toml:"tsharkPath"`
}

type Params struct {
	Device     string      `json:"device"`
	Stream     *Stream     `json:"stream"`
	Statistics *Statistics `json:"statistics"`
	Sync       bool        `json:"sync"`
	Fin        chan error  `json:"-"`
}

type Stream struct {
	Duration string   `json:"duration"`
	Count    int      `json:"count"`
	Protocol string   `json:"protocol"`
	Ports    []string `json:"port"`
	SrcIPs   []string `json:"srcIP"`
	DstIPs   []string `json:"dstIP"`
	Filter   string   `json:"filter"`
}

type Statistics struct {
	Condition string
}
