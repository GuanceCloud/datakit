package simple

// simple.go

import (
	"fmt"
	"net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
	# trace ip addr/domain
	addr = "www.baidu.com"
`

type TraceRouter struct {
	Addr string
}

func (t *TraceRouter) Description() string {
	return "trace router"
}

func (t *TraceRouter) SampleConfig() string {
	return sampleConfig
}

func (t *TraceRouter) Init() error {
	return nil
}

func (t *TraceRouter) Gather(acc telegraf.Accumulator) error {

	return nil
}

// NewUUID 创建UUID
func NewUUID() string {
	v, err := guuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return v.String()
}

func printHop(domain, ipAddr string, hop traceroute.TracerouteHop) {
	globalCfg := config.GetGlobalConfig()
	dataWay := globalCfg.DataWay
	tracerouterCfg := globalCfg.TraceRouter

	addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
	hostOrAddr := addr
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}

	measurement := &influxdb.Measurement{}
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	traceId := NewUUID()
	if hop.Success {

		tags["distAddr"] = domain
		tags["distAddrIp"] = ipAddr
		tags["traceId"] = traceId
		fields["seq"] = fmt.Sprintf("%d", hop.TTL)
		fields["addr"] = fmt.Sprintf("\"%s\"", addr)
		fields["ttl"] = hop.ElapsedTime.Microseconds()

		fmt.Printf("%-3d %v (%v)  %v\n", hop.TTL, hostOrAddr, addr, hop.ElapsedTime)
	} else {
		fmt.Printf("%-3d *\n", hop.TTL)
	}

	acc.AddFields("filesize", fields, tags)
}

func address(address [4]byte) string {
	return fmt.Sprintf("%v.%v.%v.%v", address[0], address[1], address[2], address[3])
}

func (t *TraceRouter) exec(acc telegraf.Accumulator) {
	host := t.Addr
	options := traceroute.TracerouteOptions{}
	options.SetMaxHops(traceroute.DEFAULT_MAX_HOPS + 1)
	options.SetFirstHop(traceroute.DEFAULT_FIRST_HOP)

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return
	}

	fmt.Printf("traceroute to %v (%v), %v hops max, %v byte packets\n", host, ipAddr, options.MaxHops(), options.PacketSize())

	c := make(chan traceroute.TracerouteHop, 0)
	go func() {
		for {
			hop, ok := <-c
			if !ok {
				fmt.Println()
				return
			}
			printHop(domain, ipAddr, hop)
		}
	}()

	_, err = traceroute.Traceroute(host, &options, c)
	if err != nil {
		fmt.Printf("Error: ", err)
	}
}

func init() {
	inputs.Add("tracerouter", func() telegraf.Input { return &TraceRouter{} })
}
