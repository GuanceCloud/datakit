package cshark

import (
	"fmt"
	"time"
	"bufio"
    "strconv"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	// "github.com/gcla/termshark/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/cshark/util"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/cshark/protocol"
	"encoding/json"
	"strings"
)

const (
	SEPARATOR = "#"
)

var (
	l          *logger.Logger
	inputName  = "cshark"
	optChan = make(chan *Params)
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

func (s *Shark) SendOpt(opt string) error {
	if err := s.parseParam(opt); err != nil {
		return fmt.Errorf("command param err %v", err)
	}

	// check config
	if err := s.checkParam(); err != nil {
		return err
	}

	select {
	case optChan <- s.Params:
		fmt.Println("send success!")
		return nil
	default:
		return fmt.Errorf("busy!")
	}

	// optChan <- s.Params

	return nil
}

func (s *Shark) Run() {
	l = logger.SLogger("cshark")

	l.Info("cshark input started...")
	if s.MetricName == "" {
		s.MetricName = "cshark"
	}

	go func () {
		for {
			select {
			case <- optChan:
				fmt.Println("receive....")
				s.Exec()
			case <-datakit.Exit.Wait():
				l.Info("exit")
				return
			}
		}
	}()
}

// 参数解析
func (s *Shark) parseParam(option string) error {
	if err := json.Unmarshal([]byte(option), &s.Params); err != nil {
		// l.Errorf("parsse option error:%v", err)
		return fmt.Errorf("parsse option error:%v", err)
	}

	return nil
}

// 参数校验
func (s *Shark) checkParam() error {
	// 协议check
	if !util.IsSupport(s.Params.Stream.Protocol) {
		return fmt.Errorf("not support this protocol %s", s.Params.Stream.Protocol)
	}

	// 时间check(todo)
	duration, err := time.ParseDuration(s.Params.Stream.Duration)
	if err != nil {
		duration = 60
		l.Error(err)
	}

	s.Duration = duration.Nanoseconds()/1e9

	// src ip check
	for _, ip := range s.Params.Stream.SrcIPs {
		if !util.IsIP(ip) {
			return fmt.Errorf("source ip is not right %s", ip)
		}
	}

	// dst ip check
	for _, ip := range s.Params.Stream.DstIPs {
		if !util.IsIP(ip) {
			return fmt.Errorf("destination ip is not right %s", ip)
		}
	}

	// port
	for _, port := range s.Params.Stream.Ports {
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

	args = append(args, "tshark")

	// 控制参数
	args = append(args,"-l")
	for _, iface := range s.Params.Device {
		args = append(args, "-i", iface)
	}

	// 时常控制
	du := fmt.Sprintf("duration:%d", s.Duration)
	args = append(args, "-a", du)

	// 过滤器 (todo)
	if s.Params.Stream.Filter != "" {
		args = append(args, "-f", s.Params.Stream.Filter)
	}

	args = append(args, "-Y", s.Params.Stream.Protocol)

	// 输出控制
	separator := fmt.Sprintf("separator=%s", SEPARATOR)
	// args = append(args, "-T", "fields", "-E", separator, "-E", "quote=d")
	args = append(args, "-T", "fields", "-E", separator)

	// 输出field
	fileds := protocol.GetFiled()

	args = append(args, fileds...)

	cmdStr := strings.Join(args, " ")

	return cmdStr
}

func (s *Shark) Exec() {
	// 构造命令
	var streamCmdStr string
	if s.Params.Stream != nil {
		streamCmdStr = s.buildCommand()
		l.Info("stream cmd ====>", streamCmdStr)
	}

	fmt.Println("streamCmd ========>", streamCmdStr)

	// 构造统计命令(todo)
	s.streamExec(streamCmdStr)
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
		err = io.NamedFeed(pt, io.Metric, inputName)
		if err != nil {
			l.Errorf("push metric point error %s", err)
		}
	}

	if err = cmd.Wait(); err != nil {
		l.Errorf("exec wait error %v", err)
		return err
	}

	return nil
}

func (s *Shark) parseLine(line string) []byte {
	var (
		tm time.Time
		tags = map[string]string{}
		fields = map[string]interface{}{}
	)

	items := strings.Split(line, SEPARATOR)
	if len(items) == 1 {
		return nil
	}

	for idx, item := range items {
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
