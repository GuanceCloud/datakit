package hostobject

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type (
	Collector struct {
		Name  string
		Class string
		Desc  string `toml:"description"`
		Tags  map[string]string
	}
)

func (_ *Collector) Catalog() string {
	return "object"
}

func (_ *Collector) SampleConfig() string {
	return sampleConfig
}

func (_ *Collector) Description() string {
	return "Collect host info and send to Dataflux as object data format."
}

func (c *Collector) Gather(acc telegraf.Accumulator) error {

	obj := map[string]interface{}{
		"$name":        c.Name,
		"$class":       c.Class,
		"$description": c.Desc,
	}

	tags := map[string]string{}

	hostname, err := os.Hostname()
	if err == nil {
		tags["host"] = hostname
	}

	ipval := getIP()
	if mac, err := getMacAddr(ipval); err == nil && mac != "" {
		tags["mac"] = mac
	}
	tags["ip"] = ipval
	tags["os_type"] = runtime.GOOS
	tags["cpu_total"] = fmt.Sprintf("%d", runtime.NumCPU())

	for k, v := range c.Tags {
		tags[k] = v
	}

	obj["$tags"] = tags

	data, err := json.Marshal(&obj)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"object": string(data),
	}

	acc.AddFields(inputName, fields, nil, time.Now().UTC())

	return nil
}

func (c *Collector) Init() error {
	if c.Class == "" {
		c.Class = "Servers"
	}
	if c.Name == "" {
		name, err := os.Hostname()
		if err != nil {
			return err
		}
		c.Name = name
	}
	return nil
}

func getMacAddr(targetIP string) (string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, ifs := range ifas {
		addrs, err := ifs.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
				if ip.To4() != nil {
					if ip.To4().String() == targetIP {
						return ifs.HardwareAddr.String(), nil
					}
				}
			}
		}
	}
	return "", nil
}

func getIP() string {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		conn, err = net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return ""
		}
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		ac := &Collector{}
		return ac
	})
}
