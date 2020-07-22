package hostobject

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var moduleLogger *logger.Logger

type (
	Collector struct {
		Name     string
		Class    string
		Desc     string `toml:"description"`
		Interval internal.Duration
		Tags     map[string]string `toml:"tags,omitempty"`
	}

	osInfo struct {
		Arch    string
		OSType  string
		Release string
	}
)

func (_ *Collector) Catalog() string {
	return "hostobject"
}

func (_ *Collector) SampleConfig() string {
	return sampleConfig
}

// func (_ *Collector) Description() string {
// 	return "Collect host info and send to Dataflux as object data format."
// }

func (c *Collector) Run() {

	moduleLogger = logger.SLogger(inputName)

	defer func() {
		if e := recover(); e != nil {
			moduleLogger.Errorf("panic error, %v", e)
		}
	}()

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		if err := c.initialize(); err == nil {
			break
		} else {
			moduleLogger.Errorf("%s", err)
			time.Sleep(time.Second)
		}
	}

	ctx, cancelFun := context.WithCancel(context.Background())

	go func() {
		<-datakit.Exit.Wait()
		cancelFun()
	}()

	for {

		select {
		case <-ctx.Done():
			return
		default:
		}

		var objs []*internal.ObjectData

		obj := &internal.ObjectData{
			Name:        c.Name,
			Description: c.Desc,
		}

		tags := map[string]string{
			"uuid":    config.Cfg.MainCfg.UUID,
			"__class": c.Class,
		}

		hostname, err := os.Hostname()
		if err == nil {
			tags["host"] = hostname
		}

		ipval := getIP()
		if mac, err := getMacAddr(ipval); err == nil && mac != "" {
			tags["mac"] = mac
		}
		tags["ip"] = ipval

		oi := getOSInfo()
		tags["os_type"] = oi.OSType
		tags["os"] = oi.Release

		//tags["cpu_total"] = fmt.Sprintf("%d", runtime.NumCPU())

		//meminfo, _ := mem.VirtualMemory()
		//tags["memory_total"] = fmt.Sprintf("%v", meminfo.Total/uint64(1024*1024*1024))

		for k, v := range c.Tags {
			tags[k] = v
		}

		obj.Tags = tags

		switch c.Name {
		case "__mac":
			obj.Name = tags["mac"]
		case "__ip":
			obj.Name = tags["ip"]
		case "__uuid":
			obj.Name = tags["uuid"]
		case "__host":
			obj.Name = tags["host"]
		case "__os":
			obj.Name = tags["os"]
		case "__os_type":
			obj.Name = tags["os_type"]
		}

		objs = append(objs, obj)

		data, err := json.Marshal(&objs)
		if err == nil {
			io.NamedFeed(data, io.Object, inputName)
		} else {
			moduleLogger.Errorf("%s", err)
		}

		internal.SleepContext(ctx, c.Interval.Duration)
	}

}

func (c *Collector) initialize() error {

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
	if c.Interval.Duration == 0 {
		c.Interval.Duration = 3 * time.Minute
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
