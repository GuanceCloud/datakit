package self

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = datakit.DatakitInputName
	l         = logger.DefaultSLogger(inputName)
)

type SelfInfo struct {
	stat *ClientStat

	semStop *cliutils.Sem // start stop signal
}

func (*SelfInfo) Catalog() string {
	return inputName
}

func (*SelfInfo) SampleConfig() string {
	return ``
}

func (si *SelfInfo) Run() {
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	l = logger.SLogger(inputName)
	l.Info("self input started...")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("self exit")
			return

		case <-si.semStop.Wait():
			l.Info("self return")
			return

		case <-tick.C:
			si.stat.Update()
			pt := si.stat.ToMetric()
			_ = io.Feed(inputName, datakit.Metric, []*io.Point{pt}, nil)
		}
	}
}

func (si *SelfInfo) Terminate() {
	if si.semStop != nil {
		si.semStop.Close()
	}
}

func init() { //nolint:gochecknoinits
	StartTime = time.Now()
	inputs.Add(inputName, func() inputs.Input {
		return &SelfInfo{
			stat: &ClientStat{
				OS:       runtime.GOOS,
				OSDetail: OSDetail(),
				Arch:     runtime.GOARCH,
				PID:      os.Getpid(),
			},
			semStop: cliutils.NewSem(),
		}
	})
}

func OSDetail() string {
	switch runtime.GOOS {
	case `darwin`:
		return macVersion()
	case `windows`:
		return windowsVersion()
	case `linux`:
		return linuxVersion()
	default:
		return "unknown"
	}
}

func linuxVersion() string {
	linux := `linux`
	fp := "/etc/os-release"
	_, err := os.Stat(fp)
	if err != nil {
		return linux
	}
	f, err := os.Open(fp)
	if err != nil {
		return linux
	}
	defer f.Close() //nolint:errcheck,gosec
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		// exit when error occurs or EOF encountered
		if err != nil {
			return linux
		}
		if strings.HasPrefix(line, "PRETTY_NAME") {
			ss := strings.Split(line, "=")
			if len(ss) < 2 {
				return linux
			}
			ret := strings.TrimSuffix(ss[1], "\n")
			if strings.HasPrefix(ret, "\"") && strings.HasSuffix(ret, "\"") {
				ret = ret[1 : len(ret)-1]
			}
			return ret
		}
	}
}

func macVersion() string {
	ver, err := getKernelRelease()
	if err != nil {
		return "macOS"
	}
	kernelVersion, err := strconv.Atoi(ver[:strings.Index(ver, ".")])
	if err != nil {
		return "macOS"
	}
	var ret string
	switch kernelVersion {
	case 5:
		ret = "Mac OS X 10.1 Puma"
	case 6:
		ret = "Mac OS X 10.2 Jaguar"
	case 7:
		ret = "Mac OS X 10.3 Panther"
	case 8:
		ret = "Mac OS X 10.4 Tiger"
	case 9:
		ret = "Mac OS X 10.5 Leopard"
	case 10:
		ret = "Mac OS X 10.6 Snow Leopard"
	case 11:
		ret = "Mac OS X 10.7 Lion"
	case 12:
		ret = "Mac OS X 10.8 Mountain Lion"
	case 13:
		ret = "Mac OS X 10.9 Mavericks"
	case 14:
		ret = "Mac OS X 10.10 Yosemite"
	case 15:
		ret = "Mac OS X 10.11 El Capitan"
	case 16:
		ret = "macOS 10.12 Sierra"
	case 17:
		ret = "macOS 10.13 High Sierra"
	case 18:
		ret = "macOS 10.14 Mojave"
	case 19:
		ret = "macOS 10.15 Catalina"
	case 20:
		ret = "macOS 11.0 Big Sur"
	case 21:
		ret = "macOS 12.0 Monterey"
	default:
		ret = "macOS"
	}
	return ret
}

func getKernelRelease() (string, error) {
	out, err := exec.Command("sysctl", "-n", "kern.osrelease").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func windowsVersion() string {
	display := `Windows`

	if version := getVersion(); version != "" {
		display += " " + version
	}

	return display
}

func getVersion() string {
	version := getEdition()
	parts := strings.Split(version, ".")
	majormin := parts[0] + "." + parts[1]

	var edition string

	switch majormin {
	case "10.0": // 10 Server
		edition = "10"
	case "6.3": // Server 2012 R2
		edition = "8.1"
	case "6.2": // Server 2012
		edition = "8"
	case "6.1":
		edition = "7"
	case "6.0":
		edition = "Vista"
	case "5.2":
		edition = "Server 2003"
	case "5.1":
		edition = "XP"
	case "5.0":
		edition = "2000"
	}

	return edition
}

func getEdition() string {
	cmd := exec.Command("cmd")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return ""
	}

	raw := out.String()
	j := strings.Index(raw, "]")
	i := j - 1
	for i >= 0 && (raw[i] == '.' || '0' <= raw[i] && raw[i] <= '9') {
		i--
	}
	i++
	var ver string

	if i == -1 || j == -1 {
		ver = ""
	} else {
		ver = raw[i:j]
	}

	return ver
}
