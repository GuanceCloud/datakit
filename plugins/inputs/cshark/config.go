package cshark

const (
	sharkConfigSample = `
#[[inputs.cshark]]
`
)

type Shark struct {
	MetricName string
	Params  *Params
	Duration int64
}

type Params struct {
    Device  []string       `json:"device"`
    Stream  *Stream        `json:"stream"`
    Statistics *Statistics `json:"statistics"`
}

type Stream struct {
	Duration string        `json:"duration"`
	Count    int           `json:"count"`
	Protocol string	       `json:"protocol"`
    Ports     []string     `json:"port"`
    SrcIPs    []string     `json:"srcIP"`
    DstIPs    []string     `json:"dstIP"`
    Filter    string      `json:"filter"`
}

type Statistics struct {
	Condition  string
}