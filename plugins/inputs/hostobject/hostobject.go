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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

var moduleLogger *logger.Logger

type (
	Collector struct {
		Name     string
		Class    string
		Desc     string `toml:"description,omitempty"`
		Interval datakit.Duration
		Pipeline string            `toml:"pipeline"`
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

func (r *Collector) PipelineConfig() map[string]string {
	return map[string]string{
		"hostobject": pipelineSample,
	}
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

	pp, err := pipeline.NewPipeline(c.Pipeline)
	if err != nil {
		moduleLogger.Errorf("%s", err)
	}

	for {

		select {
		case <-ctx.Done():
			return
		default:
		}

		className := c.Class

		tags := map[string]string{
			"name": c.Name,
		}

		obj := map[string]string{
			"uuid": datakit.Cfg.MainCfg.UUID,
		}

		obj["host"] = datakit.Cfg.MainCfg.Hostname

		ipval := getIP()
		if mac, err := getMacAddr(ipval); err == nil && mac != "" {
			obj["mac"] = mac
		}
		obj["ip"] = ipval

		oi := getOSInfo()
		obj["os_type"] = oi.OSType
		obj["os"] = oi.Release

		for k, v := range c.Tags {
			obj[k] = v
		}

		data, err := json.Marshal(obj)
		if err != nil {
			moduleLogger.Errorf("json marshal err:%s", err.Error())
			datakit.SleepContext(ctx, c.Interval.Duration)
			continue
		}

		fields := map[string]interface{}{}
		if pp != nil {
			if result, err := pp.Run(string(data)).Result(); err == nil {
				fields = result
			} else {
				moduleLogger.Errorf("%s", err)
			}
		}

		fields["message"] = string(data)
		if c.Desc != "" {
			fields[`description`] = c.Desc
		}

		name := tags["name"]
		switch c.Name {
		case "__mac":
			name = obj["mac"]
		case "__ip":
			name = obj["ip"]
		case "__uuid":
			name = obj["uuid"]
		case "__host":
			name = obj["host"]
		case "__os":
			name = obj["os"]
		case "__os_type":
			name = obj["os_type"]
		}
		tags["name"] = name

		io.NamedFeedEx(inputName, io.Object, className, tags, fields, time.Now().UTC())

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
