package tracerouter

import (
	"fmt"
	"net"

	guuid "github.com/google/uuid"

	"github.com/aeden/traceroute"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const sampleConfig = `
	# trace domain
	addr = "www.dataflux.cn"
`

type TraceRouter struct {
	Addr string
}

func (t *TraceRouter) Description() string {
	return "trace router"
}

func (t *TraceRouter) SampleConfig() string {
	return "trace router ip and ttl"
}

func (t *TraceRouter) Init() error {
	return nil
}

func (t *TraceRouter) Gather(acc telegraf.Accumulator) error {
	traceId := NewUUID()

	t.exec(traceId, acc)
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

func printHop(domain string, traceId string, hop traceroute.TracerouteHop, acc telegraf.Accumulator) {
	addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
	hostOrAddr := addr
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}

	tags := make(map[string]string)
	fields := make(map[string]interface{})
	// traceId := NewUUID()
	if hop.Success {
		tags["distAddr"] = domain
		tags["traceId"] = traceId
		fields["seq"] = fmt.Sprintf("%d", hop.TTL)
		fields["addr"] = fmt.Sprintf("\"%s\"", addr)
		fields["ttl"] = hop.ElapsedTime.Microseconds()

		fmt.Printf("%-3d %v (%v)  %v\n", hop.TTL, hostOrAddr, addr, hop.ElapsedTime)
	} else {
		fmt.Printf("%-3d *\n", hop.TTL)
	}

	acc.AddFields("tracerouter", fields, tags)
}

func address(address [4]byte) string {
	return fmt.Sprintf("%v.%v.%v.%v", address[0], address[1], address[2], address[3])
}

func (t *TraceRouter) exec(traceId string, acc telegraf.Accumulator) {
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

			printHop(t.Addr, traceId, hop, acc)
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
