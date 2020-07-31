// +build ignore

// +build linux,amd64

package tcpdump

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/google/gopacket/layers"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// Packet holds all layers information
type Tcpdump struct {
	Device string // interface
	Filter string
	Type   string // filter type

	SrcHost string
	DstHost string

	// packet layers data
	TCP *layers.TCP
	UDP *layers.UDP
}

const (
	configSample = `
#[[inputs.tcpdump]]
#  ## 采集的频度，最小粒度5m
#  interval = '10s'
#  libPath = ''
#  ## 指标集名称，默认值tcpdump
#  metricName = ''
#  ## 网卡信息
#  device  = ""
#  ## 协议过滤
#  filter = ['tcp', 'udp']
#`
)

var (
	l *logger.Logger
)

func (_ *Tcpdump) Catalog() string {
	return "network"
}

func (_ *Tcpdump) SampleConfig() string {
	return configSample
}

func (o *Tcpdump) Run() {
	l = logger.SLogger("tcpdump")
	l.Info("tcpdump started")

	l.Info("starting external tcpdump...")

	bin := filepath.Join(datakit.InstallDir, "externals", "tcpdump")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	if _, err := os.Stat(bin); err != nil {
		l.Error("check %s failed: %s", bin, err.Error())
		return
	}

	cfg, err := json.Marshal(o)
	if err != nil {
		l.Errorf("toml marshal failed: %v", err)
		return
	}

	b64cfg := base64.StdEncoding.EncodeToString(cfg)

	args := []string{
		"-cfg", b64cfg,
		"-rpc-server", "unix://" + datakit.GRPCDomainSock,
		"-desc", o.Desc,
		"-log", filepath.Join(datakit.InstallDir, "externals", "tcpdump.log"),
		"-log-level", config.Cfg.MainCfg.LogLevel,
	}

	cmd := exec.Command(bin, args...)
	cmd.Env = []string{ // we need libcap  lib
		fmt.Sprintf("LD_LIBRARY_PATH=%s:$LD_LIBRARY_PATH", o.LibPath),
	}

	l.Infof("starting process %+#v", cmd)
	if err := cmd.Start(); err != nil {
		l.Error(err)
		return
	}

	l.Infof("tcpdump PID: %d", cmd.Process.Pid)
	datakit.MonitProc(cmd.Process, "tcpdump") // blocking
}

func init() {
	inputs.Add("tcpdump", func() inputs.Input {
		return &Tcpdump{}
	})
}
