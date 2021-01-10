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

		mode string

		testResult *inputs.TestResult
		testError  error
	}

	osInfo struct {
		Arch    string
		OSType  string
		Release string
	}
)

func (c *Collector) isTest() bool {
	return c.mode == "test"
}

func (_ *Collector) Catalog() string {
	return "hostobject"
}

func (_ *Collector) SampleConfig() string {
	return sampleConfig
}

func (c *Collector) Test() (*inputs.TestResult, error) {
	c.mode = "test"
	c.testResult = &inputs.TestResult{}
	c.Run()
	return c.testResult, c.testError
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
			`*name`:  c.Name,
			"*class": c.Class,
		}
		if c.Desc != "" {
			obj[`description`] = c.Desc
		}

		content := map[string]string{
			"uuid": datakit.Cfg.MainCfg.UUID,
		}

		content["host"] = datakit.Cfg.MainCfg.Hostname

		ipval := getIP()
		if mac, err := getMacAddr(ipval); err == nil && mac != "" {
			content["mac"] = mac
		}
		content["ip"] = ipval

		oi := getOSInfo()
		content["os_type"] = oi.OSType
		content["os"] = oi.Release

		for k, v := range c.Tags {
			content[k] = v
		}

		data, err := json.Marshal(content)

		obj[`content`] = string(data)

		switch c.Name {
		case "__mac":
			obj[`*name`] = content["mac"]
		case "__ip":
			obj[`*name`] = content["ip"]
		case "__uuid":
			obj[`*name`] = content["uuid"]
		case "__host":
			obj[`*name`] = content["host"]
		case "__os":
			obj[`*name`] = content["os"]
		case "__os_type":
			obj[`*name`] = content["os_type"]
		}

		objs = append(objs, obj)

		data, err = json.Marshal(&objs)
		if err == nil {
			if c.isTest() {
				c.testResult.Result = append(c.testResult.Result, data...)
			} else {
				io.NamedFeed(data, io.Object, inputName)
			}
		} else {
			if c.isTest() {
				c.testError = err
			}
			moduleLogger.Errorf("%s", err)
		}

		if c.isTest() {
			break
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
