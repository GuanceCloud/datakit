package huaweiyunces

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
	"unicode"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	moduleLogger *logger.Logger
	inputName    = "huaweiyunces"
)

func (*agent) SampleConfig() string {
	return sampleConfig
}

func (*agent) Catalog() string {
	return `huaweiyun`
}

func (ag *agent) Run() {

	moduleLogger = logger.SLogger(inputName)

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			moduleLogger.Errorf("panic: %s", err)
			moduleLogger.Errorf("%s", string(buf[:n]))
		}
	}()

	if ag.Delay.Duration == 0 {
		ag.Delay.Duration = time.Minute * 1
	}

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	if ag.ApiFrequency <= 0 {
		ag.ApiFrequency = 20
	}
	if ag.ApiFrequency > 1000 {
		ag.ApiFrequency = 1000
	}
	limit := rate.Every(time.Duration(1000/ag.ApiFrequency) * time.Millisecond)
	ag.limiter = rate.NewLimiter(limit, 1)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if len(ag.Namespace) > 0 {
		ag.runOld() //兼容老的配置格式
	} else {
		ag.run()
	}
}

func extendTags(to map[string]string, from map[string]string, override bool) {
	if from == nil {
		return
	}
	for k, v := range from {
		if !override {
			if _, exist := to[k]; exist {
				continue
			}
		}
		to[k] = v
	}
}

func formatMeasurement(project string) string {
	project = strings.Replace(project, "/", "_", -1)
	project = snakeCase(project)
	return fmt.Sprintf("%s_%s", inputName, project)
}

func snakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	s := strings.Replace(string(out), "__", "_", -1)

	return s
}

func newAgent(mode string) *agent {
	ac := &agent{}
	ac.mode = mode
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent("")
	})
}
