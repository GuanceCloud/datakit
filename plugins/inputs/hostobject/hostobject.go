package hostobject

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var moduleLogger *logger.Logger

type (
	Collector struct {
		Name     string
		Class    string
		Desc     string `toml:"description,omitempty"`
		Interval datakit.Duration
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

		var objs []map[string]interface{}

		obj := map[string]interface{}{
			`__name`: c.Name,
		}
		if c.Desc != "" {
			obj[`__description`] = c.Desc
		}

		tags := map[string]string{
			"uuid":    datakit.Cfg.MainCfg.UUID,
			"__class": c.Class,
		}

		tags["host"] = datakit.Cfg.MainCfg.Hostname

		ipval := getIP()
		if mac, err := getMacAddr(ipval); err == nil && mac != "" {
			tags["mac"] = mac
		}
		tags["ip"] = ipval

		oi := getOSInfo()
		tags["os_type"] = oi.OSType
		tags["os"] = oi.Release

		for k, v := range c.Tags {
			tags[k] = v
		}

		obj[`__tags`] = tags

		switch c.Name {
		case "__mac":
			obj[`__name`] = tags["mac"]
		case "__ip":
			obj[`__name`] = tags["ip"]
		case "__uuid":
			obj[`__name`] = tags["uuid"]
		case "__host":
			obj[`__name`] = tags["host"]
		case "__os":
			obj[`__name`] = tags["os"]
		case "__os_type":
			obj[`__name`] = tags["os_type"]
		}

		objs = append(objs, obj)

		data, err := json.Marshal(&objs)
		if err == nil {
			io.NamedFeed(data, io.Object, inputName)
		} else {
			moduleLogger.Errorf("%s", err)
		}

		datakit.SleepContext(ctx, c.Interval.Duration)
	}

}

func (c *Collector) initialize() error {

	if c.Class == "" {
		c.Class = "Servers"
	}

	if c.Name == "" {
		c.Name = datakit.Cfg.MainCfg.Hostname
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
