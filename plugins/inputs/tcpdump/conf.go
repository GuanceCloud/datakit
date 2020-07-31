package tcpdump

const (
	configSample = `
#[[inputs.tcpdump]]
#  ## 网卡
#  device = '10m'
#  ## 协议类型
#  protocol = ['tcp', 'udp']
#  ## 指标集名称，默认值tcpdump
#  metricName = ''
#
#`
)

// Packet holds all layers information
type Tcpdump struct {
	Device     string            `toml:"device"`
	Protocol   []string          `toml:"protocol"`
	MetricName string            `toml:"metricName"`
	Tags       map[string]string `toml:"tags"`
}
