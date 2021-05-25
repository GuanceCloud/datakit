package cshark

import (
	"bufio"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"strconv"
	"time"
	// "github.com/gcla/termshark/v2"
	"encoding/json"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/cshark/protocol"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/cshark/util"
	"strings"
)

const (
	SEPARATOR = "~"
)

var (
	l         *logger.Logger
	inputName = "cshark"
	optChan   = make(chan *Params)
	params    *Params
	duration  int64
)

func (_ *Shark) SampleConfig() string {
	return sharkConfigSample
}

func (_ *Shark) Catalog() string {
	return "network"
}

func (_ *Shark) Description() string {
	return "datakit online capture netpacket"
}

func (_ *Shark) Gather() error {
	return nil
}

func SendCmdOpt(opt string) error {
	if err := parseParam(opt); err != nil {
		return fmt.Errorf("command param err %v", err)
	}

	// check config
	if err := checkParam(); err != nil {
		return err
	}

	if params.Sync {
		select {
		case optChan <- params:
			fmt.Println("send success!")
			params.Fin = make(chan error)

			err := <-params.Fin

			if err != nil {
				return err
			}

			return nil
		default:
			return fmt.Errorf("busy!")
		}
	} else {
		select {
		case optChan <- params:
			fmt.Println("send success!")
			return nil
		default:
			return fmt.Errorf("busy!")
		}
	}
}

func (s *Shark) Run() {
	l = logger.SLogger("cshark")

	l.Info("cshark input started...")
	if s.MetricName == "" {
		s.MetricName = "cshark"
	}

	if s.Interval == "" {
		s.Interval = "10s"
	}

	interval, err := time.ParseDuration(s.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if _, err := TSharkVersion(s.TsharkPath); err != nil {
				l.Errorf("tshark not install or Env path config error %v", err)
			} else {
				goto lable
			}
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}

lable:
	for {
		select {
		case opt := <-optChan:
			if err := s.Exec(); err != nil {
				l.Errorf("exec error %v", err)
			}

			if opt.Sync {
				opt.Fin <- err
			}

		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

// 参数解析
func parseParam(option string) error {
	params = new(Params)
	if err := json.Unmarshal([]byte(option), &params); err != nil {
		return fmt.Errorf("parsse option error:%v", err)
	}

	return nil
}

// 参数校验
func checkParam() error {
	// 协议check
	if !util.IsSupport(params.Stream.Protocol) {
		return fmt.Errorf("not support this protocol %s", params.Stream.Protocol)
	}

	if len(params.Stream.Duration) == 0 {
		params.Stream.Duration = "1m"
	}

	// 时间check(todo)
	du, err := time.ParseDuration(params.Stream.Duration)
	if err != nil {
		duration = 60
		l.Error(err)
	}

	duration = du.Nanoseconds() / 1e9

	// src ip check
	for _, ip := range params.Stream.SrcIPs {
		if !util.IsIP(ip) {
			return fmt.Errorf("source ip is not right %s", ip)
		}
	}

	// dst ip check
	for _, ip := range params.Stream.DstIPs {
		if !util.IsIP(ip) {
			return fmt.Errorf("destination ip is not right %s", ip)
		}
	}

	// port
	for _, port := range params.Stream.Ports {
		portN, _ := strconv.ParseInt(port, 10, 64)
		if int(portN) > 65535 || int(portN) < 0 {
			return fmt.Errorf("port ip is not right %s", port)
		}
	}

	// pfb校验(todo)

	return nil
}

// 构建抓包命令行
func (s *Shark) buildCommand() string {
	args := make([]string, 0)
	portFilterStr := ""
	srcIPFilterStr := ""
	dstIPFilterStr := ""

	args = append(args, s.TsharkPath)

	// 控制参数
	args = append(args, "-l")
	if len(params.Device) > 0 {
		args = append(args, "-i", params.Device)
	}

	if len(params.Device) == 0 {
		args = append(args, "-i", "any")
	}

	if params.Stream.Count != 0 {
		count := fmt.Sprintf("%d", params.Stream.Count)
		args = append(args, "-c", count)
	}

	// 时常控制
	du := fmt.Sprintf("duration:%d", duration)
	args = append(args, "-a", du)

	// 过滤器 (todo)
	if params.Stream.Filter != "" {
		filter := fmt.Sprintf("'%s'", params.Stream.Filter)
		args = append(args, "-f", filter)
	}

	// 端口
	// if len(params.Stream.Ports) > 0 {
	// 	for _, port := range params.Stream.Ports {
	// 		portFilterStr += "port " + port + " or "
	// 	}
	// 	portFilterStr = strings.Trim(portFilterStr, "or ")
	// 	portFilterStr = fmt.Sprintf("'%s'", portFilterStr)
	// 	args = append(args, "-f", portFilterStr)
	// }

	// ip
	if (len(params.Stream.SrcIPs) > 0) && (len(params.Stream.DstIPs) == 0) && (len(params.Stream.Ports) == 0) {
		var filterStr string

		for _, srcIP := range params.Stream.SrcIPs {
			srcIPFilterStr += "src host " + srcIP + " or "
		}
		srcIPFilterStr = strings.Trim(srcIPFilterStr, "or ")

		if strings.ToUpper(params.Stream.Protocol) == "TCP" || strings.ToUpper(params.Stream.Protocol) == "UDP" {
			filterStr = fmt.Sprintf("'%s and (%s)'", params.Stream.Protocol, srcIPFilterStr)
		} else {
			filterStr = fmt.Sprintf("'%s'", srcIPFilterStr)
		}

		args = append(args, "-f", filterStr)
	}

	if (len(params.Stream.DstIPs) > 0) && (len(params.Stream.SrcIPs) == 0) && (len(params.Stream.Ports) == 0) {
		var filterStr string

		for _, dstIP := range params.Stream.DstIPs {
			dstIPFilterStr += "dst host " + dstIP + " or "
		}
		dstIPFilterStr = strings.Trim(dstIPFilterStr, "or ")

		if strings.ToUpper(params.Stream.Protocol) == "TCP" || strings.ToUpper(params.Stream.Protocol) == "UDP" {
			filterStr = fmt.Sprintf("'%s and (%s)'", params.Stream.Protocol, dstIPFilterStr)
		} else {
			filterStr = fmt.Sprintf("'%s'", dstIPFilterStr)
		}

		args = append(args, "-f", filterStr)
	}

	if (len(params.Stream.SrcIPs) > 0) && (len(params.Stream.DstIPs) > 0) && (len(params.Stream.Ports) == 0) {
		var filterStr string
		for _, srcIP := range params.Stream.SrcIPs {
			srcIPFilterStr += "src host " + srcIP + " or "
		}
		srcIPFilterStr = strings.Trim(srcIPFilterStr, "or ")

		for _, dstIP := range params.Stream.DstIPs {
			dstIPFilterStr += "dst host " + dstIP + " or "
		}

		dstIPFilterStr = strings.Trim(dstIPFilterStr, "or ")

		if strings.ToUpper(params.Stream.Protocol) == "TCP" || strings.ToUpper(params.Stream.Protocol) == "UDP" {
			filterStr = fmt.Sprintf("'%s and (%s) and (%s)'", params.Stream.Protocol, srcIPFilterStr, dstIPFilterStr)
		} else {
			filterStr = fmt.Sprintf("'(%s) and (%s)'", srcIPFilterStr, dstIPFilterStr)
		}

		args = append(args, "-f", filterStr)
	}

	if (len(params.Stream.DstIPs) > 0) && (len(params.Stream.SrcIPs) > 0) && (len(params.Stream.Ports) > 0) {
		var filterStr string

		for _, srcIP := range params.Stream.SrcIPs {
			srcIPFilterStr += "src host " + srcIP + " or "
		}
		srcIPFilterStr = strings.Trim(srcIPFilterStr, "or ")

		for _, dstIP := range params.Stream.DstIPs {
			dstIPFilterStr += "dst host " + dstIP + " or "
		}

		dstIPFilterStr = strings.Trim(dstIPFilterStr, "or ")

		for _, port := range params.Stream.Ports {
			portFilterStr += "port " + port + " or "
		}
		portFilterStr = strings.Trim(portFilterStr, "or ")

		if strings.ToUpper(params.Stream.Protocol) == "TCP" || strings.ToUpper(params.Stream.Protocol) == "UDP" {
			filterStr = fmt.Sprintf("'%s and (%s) and (%s) and (%s)'", params.Stream.Protocol, portFilterStr, srcIPFilterStr, dstIPFilterStr)
		} else {
			filterStr = fmt.Sprintf("'(%s) and (%s) and (%s)'", portFilterStr, srcIPFilterStr, dstIPFilterStr)
		}

		args = append(args, "-f", filterStr)
	}

	if len(params.Stream.Protocol) > 0 {
		args = append(args, "-Y", params.Stream.Protocol)

		// 协议分发
		switch strings.ToUpper(params.Stream.Protocol) {
		case "HTTP":
			protocol.CommonItems = append(protocol.CommonItems, protocol.HttpItems...)
		case "MYSQL":
			protocol.CommonItems = append(protocol.CommonItems, protocol.MysqlItems...)
		case "DNS":
			protocol.CommonItems = append(protocol.CommonItems, protocol.DnslItems...)
		}
	}

	// 输出控制
	separator := fmt.Sprintf("separator=%s", SEPARATOR)
	args = append(args, "-T", "fields", "-E", separator)

	// 输出field
	fileds := protocol.GetFiled()

	args = append(args, fileds...)

	cmdStr := strings.Join(args, " ")

	return cmdStr
}

func (s *Shark) Exec() error {
	// 构造命令
	var streamCmdStr string
	if params.Stream != nil {
		streamCmdStr = s.buildCommand()
		l.Info("stream cmd ====>", streamCmdStr)
	}

	fmt.Println("streamCmd ========>", streamCmdStr)

	// 构造统计命令(todo)
	return s.streamExec(streamCmdStr)
}

func (s *Shark) streamExec(cmdStr string) error {
	cmd := RunCommand(cmdStr)
	out, err := cmd.StdoutReader()
	defer cmd.Close()

	if err != nil {
		// print err info
		l.Errorf("exec set pipline error %v", err)
		return err
	}

	if err = cmd.Start(); err != nil {
		l.Errorf("exec start error %v", err)
		return err
	}

	scan := bufio.NewScanner(out)
	for scan.Scan() {
		line := scan.Text()
		// build influxdb point line data
		pt := s.parseLine(line)
		if err != nil {
			l.Errorf("build point line data error %v", err)
			continue
		}

		fmt.Println("point =====>", string(pt))

		// io output
		err = io.NamedFeed(pt, datakit.Metric, inputName)
		if err != nil {
			l.Errorf("push metric point error %s", err)
		}
	}

	if err = cmd.Wait(); err != nil {
		l.Errorf("exec wait error %v", err)
		return err
	}

	fmt.Println("exec complete, exit...")

	return nil
}

func (s *Shark) parseLine(line string) []byte {
	var (
		tm     time.Time
		tags   = map[string]string{}
		fields = map[string]interface{}{}
	)

	items := strings.Split(line, SEPARATOR)
	if len(items) == 1 {
		return nil
	}

	for idx, item := range items {
		if idx < len(protocol.CommonItems) {
			field := protocol.CommonItems[idx]

			if idx > 0 {
				if field.Tag {
					tags[field.Header] = item
				} else {
					if field.Type == "Int" {
						if val, err := strconv.ParseInt(item, 10, 64); err == nil {
							fields[field.Header] = val
						}
					} else {
						fields[field.Header] = item
					}
				}
			} else {
				if timestamp, err := strconv.ParseInt(item, 10, 64); err != nil {
					tm = time.Now()
				} else {
					tm = time.Unix(timestamp, 0)
				}
			}
		}
	}

	pt, err := io.MakeMetric(s.MetricName, tags, fields, tm)
	if err != nil {
		l.Errorf("make metric point error %s", err)
	}

	return pt
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Shark{}
	})
}
