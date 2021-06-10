package cmds

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type integration struct {
	Title  string   `yaml:"title"`
	Tags   []string `yaml:"tags"`
	DocURL string   `yaml:"url"`
}

var (
	titles = map[string]string{
		"cpu":            "CPU",
		"ddtrace":        "DDTrace",
		"dialtesting":    "网络拨测",
		"disk":           "Disk",
		"diskio":         "DiskIO",
		"docker":         "Docker",
		"elasticsearch":  "ElasticSearch",
		"host_processes": "Process",
		"hostobject":     "主机",
		"jvm":            "JVM",
		"kafka":          "Kafka",
		"logging":        "日志",
		"mem":            "Memory",
		"mysql":          "MySQL",
		"net":            "Net",
		"nginx":          "Nginx",
		"oracle":         "Oracle",
		"rabbitmq":       "RabbitMQ",
		"redis":          "Redis",
		"swap":           "Swap",
		"system":         "System",
		"rum":            "RUM",
		"dataway":        "DataWay",
		"proxy":          "DataKit代理",
		"mongodb":        "MongoDB",
		"kubernetes":     "Kubernetes",
		"sqlserver":      "SQLServer",
		"postgresql":     "PostgreSQL",
		"statsd":         "Statsd",
		"container":      "容器",
	}
)

func ExportIntegration(to, ignore string) error {
	if err := os.MkdirAll(to, os.ModePerm); err != nil {
		return err
	}

	arr := strings.Split(ignore, ",")
	skip := map[string]bool{}
	for _, x := range arr {
		skip[x] = true
	}

	for k, i := range inputs.Inputs {
		if _, ok := skip[k]; ok {
			continue
		}

		switch i().(type) {
		case inputs.InputV2:

			title := titles[k]
			if title == "" {
				title = k
			}

			x := integration{
				Title:  title,
				Tags:   []string{"IT运维", "指标采集"},
				DocURL: "https://www.yuque.com/dataflux/datakit/" + k,
			}

			if err := os.MkdirAll(filepath.Join(to, k), os.ModePerm); err != nil {
				return err
			}

			mf, err := yaml.Marshal(x)
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(filepath.Join(to, k, "manifest.yaml"), mf, os.ModePerm); err != nil {
				return err
			}

		default:
			l.Debugf("ignore %s", k)
		}
	}
	return nil
}
