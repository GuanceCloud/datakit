package host_process

import (
	"encoding/json"
	"fmt"
	pr "github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var l = logger.DefaultSLogger(inputName)

func (_ *Processes) Catalog() string {
	return category
}

func (_ *Processes) SampleConfig() string {
	return sampleConfig
}

func (p *Processes) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

func (p *Processes) Test() (*inputs.TestResult, error) {
	p.isTest = true
	p.WriteObject()
	return p.result, nil
}

func (p *Processes) Run() {
	l = logger.SLogger(inputName)

	l.Info("process start...")

	if p.ProcessName != nil {
		re := strings.Join(p.ProcessName, "|")
		if regexp.MustCompile(re) == nil {
			l.Error("[error] regexp err")
			return
		}
		p.re = re
	}
	if p.ObjectInterval.Duration == 0 {
		p.ObjectInterval.Duration = 5 * time.Minute
	}
	if p.RunTime.Duration == 0 || p.RunTime.Duration < 10*time.Minute {
		p.RunTime.Duration = 10 * time.Minute
	}
	tick := time.NewTicker(p.ObjectInterval.Duration)
	defer tick.Stop()
	if p.OpenMetric {
		go func() {
			if p.MetricInterval.Duration == 0 {
				p.MetricInterval.Duration = 5 * time.Minute
			}
			tick := time.NewTicker(p.MetricInterval.Duration)
			defer tick.Stop()
			for {
				p.WriteMetric()
				select {
				case <-tick.C:
				case <-datakit.Exit.Wait():
					l.Info("process exit")
					return
				}
			}

		}()
	}

	for {
		p.WriteObject()
		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("process write object exit")
			return
		}
	}

}

func (p *Processes) getProcesses() (processList []*pr.Process) {
	pses, err := pr.Processes()
	if err != nil {
		l.Errorf("[error] get process err:%s", err.Error())
		return
	}

	for _, ps := range pses {
		name, _ := ps.Name()
		if ok, _ := regexp.Match(p.re, []byte(name)); !ok {
			continue
		}
		t, _ := ps.CreateTime()
		tm := time.Unix(0, t*1000000) // 转纳秒
		if time.Now().Sub(tm) > p.RunTime.Duration {
			processList = append(processList, ps)
		}
	}
	return processList
}

func (p *Processes) Parse(ps *pr.Process) (username, state, name string, fields, message map[string]interface{}) {
	fields = map[string]interface{}{}
	message = map[string]interface{}{}
	name, err := ps.Name()
	if err != nil {
		l.Warnf("[warning] process get name err:%s", err.Error())
	}
	username, err = ps.Username()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get username err:%s", name, ps.Pid, err.Error())
	}
	status, err := ps.Status()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get state err:%s", name, ps.Pid, err.Error())
		state = ""
	} else {
		state = status[0]
	}
	mem, err := ps.MemoryInfo()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get memoryinfo err:%s", name, ps.Pid, err.Error())
	} else {
		message["memory"] = mem
		fields["rss"] = mem.RSS
	}
	memPercent, err := ps.MemoryPercent()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get mempercent err:%s", name, ps.Pid, err.Error())
	} else {
		fields["mem"] = memPercent
	}
	cpu, err := ps.Times()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get cpu err:%s", name, ps.Pid, err.Error())
	} else {
		message["cpu"] = cpu
	}
	cpuPercent, _ := ps.CPUPercent()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get cpupercent err:%s", name, ps.Pid, err.Error())
	} else {
		fields["cpu"] = cpuPercent
	}
	Threads, err := ps.NumThreads()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get threads err:%s", name, ps.Pid, err.Error())
	} else {
		fields["threads"] = Threads
	}
	if runtime.GOOS == "linux" {
		OpenFiles, err := ps.OpenFiles()
		if err != nil {
			l.Warnf("[warning] process:%s,pid:%d get openfile err:%s", name, ps.Pid, err.Error())
		} else {
			fields["open_files"] = len(OpenFiles)
			message["open_files"] = OpenFiles
		}
	}

	return username, state, name, fields, message

}

func (p *Processes) WriteObject() {
	times := time.Now().UTC()
	var points []string
	for _, ps := range p.getProcesses() {
		t, _ := ps.CreateTime()
		username, state, name, fields, message := p.Parse(ps)
		tags := map[string]string{
			"username":     username,
			"state":        state,
			"name":         fmt.Sprintf("%s_%d_%d", name, ps.Pid, t),
			"process_name": name,
		}
		m, _ := json.Marshal(message)
		if p.Pipeline != "" {
			pipe, err := pipeline.NewPipelineByScriptPath(p.Pipeline)
			if err == nil {
				pipeMap, err := pipe.Run(string(m)).Result()
				if err == nil {
					for k, v := range pipeMap {
						fields[k] = v
					}
				} else {
					l.Errorf("[error] process run pipeline err:%s", err.Error())
				}

			} else {
				l.Errorf("[error] process new pipeline err:%s", err.Error())
			}
		}

		fields["message"] = string(m)
		fields["pid"] = ps.Pid
		fields["start_time"] = t
		if runtime.GOOS == "linux" {
			dir, err := ps.Cwd()
			if err != nil {
				l.Warnf("[warning] process:%s,pid:%d get work_directory err:%s", name, ps.Pid, err.Error())
			} else {
				fields["work_directory"] = dir
			}
		}
		cmd, err := ps.Cmdline()
		if err != nil {
			l.Warnf("[warning] process:%s,pid:%d get cmd err:%s", name, ps.Pid, err.Error())
		} else {
			fields["cmdline"] = cmd
		}
		if p.isTest {
			point, err := io.MakeMetric("host_processes", tags, fields, times)
			if err != nil {
				l.Errorf("make metric err:%s", err.Error())
				p.result.Result = []byte(err.Error())
			} else {
				p.result.Result = point
			}
			return
		}
		point, err := io.MakeMetric("host_processes", tags, fields, times)
		if err != nil {
			l.Errorf("[error] make metric err:%s", err.Error())
			continue
		}
		points = append(points, string(point))
	}
	io.NamedFeed([]byte(strings.Join(points, "\n")), io.Object, inputName)
}

func (p *Processes) WriteMetric() {
	for _, ps := range p.getProcesses() {
		username, _, name, fields, _ := p.Parse(ps)
		tags := map[string]string{
			"username":     username,
			"pid":          fmt.Sprintf("%d", ps.Pid),
			"process_name": name,
		}
		io.NamedFeedEx(inputName, io.Metric, "host_processes", tags, fields, time.Now().UTC())
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Processes{}
	})
}
