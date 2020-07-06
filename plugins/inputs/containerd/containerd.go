// +build linux

package containerd

import (
	"bytes"
	"time"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "containerd"

	configSample = `
# [[containerd]]
#       ## containerd 在本机的 sock 地址，一般使用默认即可
#       host_path = "/run/containerd/containerd.sock"
#
#       ## 需要采集的 containerd namespace
#       ## 可以使 'ps -ef | grep containerd | grep containerd-shim' 查看详情
#       namespace = "moby"
#
#       ## 需要采集的 containerd ID 列表，ID 是一串长度为 64 的字符串
#       ## 如果该值是 "*" ，会默认采集所有
#       ID_list = ["*"]
#
#	## 采集周期，时间单位秒
#	collect_cycle = 60
`
)

var l *zap.SugaredLogger

type (
	Containerd struct {
		C []Impl `toml:"containerd"`
	}

	Impl struct {
		HostPath  string        `toml:"host_path"`
		Namespace string        `toml:"namespace"`
		IDList    []string      `toml:"ID_list"`
		Cycle     time.Duration `toml:"collect_cycle"`
		// get all ids metrics
		isAll bool
		// id cache
		ids map[string]byte
	}
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Containerd{}
	})
}

func (_ *Containerd) Catalog() string {
	return inputName
}

func (_ *Containerd) SampleConfig() string {
	return configSample
}

func (c *Containerd) Run() {
	l = logger.SLogger(inputName)

	for _, i := range c.C {
		go i.start()
	}
}

func (i *Impl) start() {
	i.isAll = len(i.IDList) == 1 && i.IDList[0] == "*"
	i.ids = func() map[string]byte {
		m := make(map[string]byte)
		for _, v := range i.IDList {
			m[v] = '0'
		}
		return m
	}()

	ticker := time.NewTicker(time.Second * i.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")

		case <-ticker.C:
			pts, err := i.collectContainerd()
			if err != nil {
				l.Error(err)
				continue
			}

			var data bytes.Buffer

			for _, pt := range pts {
				if _, err := data.WriteString(pt.String()); err != nil {
					l.Error(err)
					continue
				}
				data.WriteString("\n")
			}

			io.Feed(data.Bytes(), io.Metric)
		}
	}
}
