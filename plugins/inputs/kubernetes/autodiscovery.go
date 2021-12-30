package kubernetes

import (
	"bufio"
	"context"

	//nolint:gosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type Discovery struct {
	list map[string]interface{}
	mu   sync.Mutex
}

func NewDiscovery() *Discovery {
	return &Discovery{list: make(map[string]interface{})}
}

func (d *Discovery) TryRun(name, cfg string) error {
	creator, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("invalid inputName")
	}

	existed, md5str := d.IsExist(cfg)
	if existed {
		return nil
	}

	inputList, err := config.LoadInputConfig(cfg, creator)
	if err != nil {
		return err
	}

	d.addList(md5str)

	l.Infof("discovery: add %s inputs, len %d", name, len(inputList))

	// input run() 不受全局 election 影响
	// election 模块运行在此之前，且其列表是固定的
	g := datakit.G("kubernetes-autodiscovery")
	for _, ii := range inputList {
		if ii == nil {
			l.Debugf("skip non-datakit-input %s", name)
			continue
		}

		func(name string, ii inputs.Input) {
			g.Go(func(ctx context.Context) error {
				time.Sleep(time.Duration(rand.Int63n(int64(10 * time.Second)))) //nolint:gosec
				l.Infof("discovery: starting input %s ...", name)
				ii.Run()
				l.Infof("discovery: input %s exited", name)
				return nil
			})
		}(name, ii)
	}

	return nil
}

func (d *Discovery) IsExist(config string) (bool, string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	h := md5.New() //nolint:gosec
	h.Write([]byte(config))
	md5Str := hex.EncodeToString(h.Sum(nil))
	_, exist := d.list[md5Str]
	return exist, md5Str
}

func (d *Discovery) addList(md5str string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.list[md5str]; ok {
		return
	}
	d.list[md5str] = nil
}

func shouldForkInput(nodeName string) bool {
	if !datakit.Docker {
		return true
	}
	// ENV NODE_NAME 在 daemonset.yaml 配置，是当前程序所在的 Node 名称
	// Node 名称匹配，表示运行在同一个 Node，此时才需要 fork

	// Node 名称为空属于 unreachable
	return datakit.GetEnv("NODE_NAME") == nodeName
}

// podlogging config example
/*
## your logging source, if it's empty, use 'default'
source = ""

## add service tag, if it's empty, use $source.
service = ""

## grok pipeline script path
pipeline = ""

## optional status:
##   "emerg","alert","critical","error","warning","info","debug","OK"
ignore_status = []

## removes ANSI escape codes from text strings
remove_ansi_escape_codes = false

[tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
*/

type podlogging struct {
	Source                string            `toml:"source"`
	Service               string            `toml:"service"`
	Pipeline              string            `toml:"pipeline"`
	IgnoreStatus          []string          `toml:"ignore_status"`
	RemoveAnsiEscapeCodes bool              `toml:"remove_ansi_escape_codes"`
	Tags                  map[string]string `toml:"tags"`

	pipe *pipeline.Pipeline
}

var tailLines int64 = 10

func (p *podlogging) run(client podclient, namespace, podName string) {
	if p.Source == "" {
		p.Source = "default"
	}
	if p.Service == "" {
		p.Service = p.Source
	}
	if p.Pipeline != "" {
		path, _ := config.GetPipelinePath(p.Pipeline)
		p.pipe, _ = pipeline.NewPipelineFromFile(path, false)
	}
	if len(p.Tags) == 0 {
		p.Tags = make(map[string]string)
	}
	p.Tags["service"] = p.Service

	req := client.getPods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow:    true,
		TailLines: &tailLines,
	})

	if err := p.ConsumeRequest(context.Background(), req); err != nil {
		l.Error(err)
	}
}

const loggingDisableAddStatus = false

func (p *podlogging) ConsumeRequest(ctx context.Context, request rest.ResponseWrapper) error {
	stream, err := request.Stream(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = stream.Close()
	}()

	r := bufio.NewReader(stream)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) != 0 {
			msg := string(line)
			// Remove a line break
			msg = strings.TrimSuffix(msg, "\n")

			if err := tailer.NewLogs(msg).
				RemoveAnsiEscapeCodesOfText(p.RemoveAnsiEscapeCodes).
				Pipeline(p.pipe).
				CheckFieldsLength().
				AddStatus(loggingDisableAddStatus).
				IgnoreStatus(p.IgnoreStatus).
				TakeTime().
				Point(p.Source, p.Tags).
				Feed(inputName).
				Err(); err != nil {
				l.Error("logging gather failed, container_id: %s, container_name:%s, err: %s", err.Error())
			}
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}
