// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package man manages all datakit documents
package man

import (
	"bytes"
	"fmt"
	"sort"

	// nolint:typecheck
	"strings"
	"text/template"

	packr "github.com/gobuffalo/packr/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	ManualBox = packr.New("manuals", "./manuals")
	OtherDocs = map[string]interface{}{
		// value not used, just document the markdown relative path
		// all manuals under man/manuals/
		"datakit-sink-guide":       "man/manuals/datakit-sink-guide.md",
		"datakit-sink-dev":         "man/manuals/datakit-sink-dev.md",
		"datakit-sink-influxdb":    "man/manuals/datakit-sink-influxdb.md",
		"datakit-sink-logstash":    "man/manuals/datakit-sink-logstash.md",
		"datakit-sink-m3db":        "man/manuals/datakit-sink-m3db.md",
		"datakit-sink-otel-jaeger": "man/manuals/datakit-sink-otel-jaeger.md",

		"apis":                 "man/manuals/apis.md",
		"changelog":            "man/manuals/changelog.md",
		"datakit-arch":         "man/manuals/datakit-arch.md",
		"datakit-batch-deploy": "man/manuals/datakit-batch-deploy.md",

		"datakit-conf":       "man/manuals/datakit-conf.md",
		"datakit-input-conf": "man/manuals/datakit-input-conf.md",

		"datakit-daemonset-deploy": "man/manuals/datakit-daemonset-deploy.md",
		"datakit-daemonset-update": "man/manuals/datakit-daemonset-update.md",
		"datakit-daemonset-bp":     "man/manuals/datakit-daemonset-bp.md",
		"datakit-dql-how-to":       "man/manuals/datakit-dql-how-to.md",
		"datakit-filter":           "man/manuals/datakit-filter.md",
		"datakit-logging-how":      "man/manuals/datakit-logging-how.md",
		"datakit-install":          "man/manuals/datakit-install.md",
		"datakit-logging":          "man/manuals/datakit-logging.md",
		"datakit-monitor":          "man/manuals/datakit-monitor.md",
		"datakit-offline-install":  "man/manuals/datakit-offline-install.md",
		"datakit-on-public":        "man/manuals/datakit-on-public.md",
		"datakit-pl-how-to":        "man/manuals/datakit-pl-how-to.md",
		"datakit-pl-global":        "man/manuals/datakit-pl-global.md",
		"datakit-service-how-to":   "man/manuals/datakit-service-how-to.md",
		"datakit-tools-how-to":     "man/manuals/datakit-tools-how-to.md",
		"datakit-tracing":          "man/manuals/datakit-tracing.md",
		"datakit-tracing-struct":   "man/manuals/datakit-tracing-struct.md",
		//"datakit-tracing-pl":       "man/manuals/datakit-tracing-pl.md",
		"datakit-update":         "man/manuals/datakit-update.md",
		"datatypes":              "man/manuals/datatypes.md",
		"dataway":                "man/manuals/dataway.md",
		"dca":                    "man/manuals/dca.md",
		"ddtrace-golang":         "man/manuals/ddtrace-golang.md",
		"ddtrace-java":           "man/manuals/ddtrace-java.md",
		"ddtrace-python":         "man/manuals/ddtrace-python.md",
		"ddtrace-php":            "man/manuals/ddtrace-php.md",
		"ddtrace-nodejs":         "man/manuals/ddtrace-nodejs.md",
		"ddtrace-cpp":            "man/manuals/ddtrace-cpp.md",
		"ddtrace-ruby":           "man/manuals/ddtrace-ruby.md",
		"development":            "man/manuals/development.md",
		"dialtesting_json":       "man/manuals/dialtesting_json.md",
		"election":               "man/manuals/election.md",
		"k8s-config-how-to":      "man/manuals/k8s-config-how-to.md",
		"kubernetes-prom":        "man/manuals/kubernetes-prom.md",
		"kubernetes-x":           "man/manuals/kubernetes-x.md",
		"logfwd":                 "man/manuals/logfwd.md",
		"logging-pipeline-bench": "man/manuals/logging-pipeline-bench.md",
		"logging_socket":         "man/manuals/logging_socket.md",
		"opentelemetry-go":       "man/manuals/opentelemetry-go.md",
		"opentelemetry-java":     "man/manuals/opentelemetry-java.md",
		"pipeline":               "man/manuals/pipeline.md",
		"prometheus":             "man/manuals/prometheus.md",
		"rum":                    "man/manuals/rum.md",
		"sec-checker":            "man/manuals/sec-checker.md",
		"telegraf":               "man/manuals/telegraf.md",
		"why-no-data":            "man/manuals/why-no-data.md",
		"git-config-how-to":      "man/manuals/git-config-how-to.md",
	}
	l = logger.DefaultSLogger("man")
)

type Params struct {
	InputName      string
	Catalog        string
	InputSample    string
	Version        string
	ReleaseDate    string
	Measurements   []*inputs.MeasurementInfo
	CSS            string
	AvailableArchs string
	PipelineFuncs  string
}

func Get(name string) (string, error) {
	return ManualBox.FindString(name + ".md")
}

type Option struct {
	WithCSS                       bool
	IgnoreMissing                 bool
	DisableMonofontOnTagFieldName bool
	ManVersion                    string
}

func BuildMarkdownManual(name string, opt *Option) ([]byte, error) {
	var p *Params

	css := MarkdownCSS
	ver := datakit.Version

	if !opt.WithCSS {
		css = ""
	}

	if opt.ManVersion != "" {
		ver = opt.ManVersion
	}

	if opt.DisableMonofontOnTagFieldName {
		inputs.MonofontOnTagFieldName = false
	}

	if _, ok := OtherDocs[name]; ok {
		p = &Params{
			Version:     ver,
			ReleaseDate: git.BuildAt,
			CSS:         css,
		}

		// Add pipeline functions doc.
		if name == "pipeline" {
			sb := strings.Builder{}
			names := []string{}
			for k := range funcs.PipelineFunctionDocs {
				// order by name
				names = append(names, k)
			}
			sort.Strings(names)

			for _, name := range names {
				sb.WriteString(funcs.PipelineFunctionDocs[name].Doc)
			}

			p.PipelineFuncs = sb.String()
		}
	} else {
		c, ok := inputs.Inputs[name]
		if !ok {
			return nil, fmt.Errorf("input %s not found", name)
		}

		input := c()
		switch i := input.(type) {
		case inputs.InputV2:
			sampleMeasurements := i.SampleMeasurement()
			p = &Params{
				InputName:      name,
				InputSample:    i.SampleConfig(),
				Catalog:        i.Catalog(),
				Version:        ver,
				ReleaseDate:    git.BuildAt,
				CSS:            css,
				AvailableArchs: strings.Join(i.AvailableArchs(), ","),
			}
			for _, m := range sampleMeasurements {
				p.Measurements = append(p.Measurements, m.Info())
			}

		default:
			l.Warnf("incomplete input: %s", name)

			return nil, nil
		}
	}

	md, err := Get(name)
	if err != nil {
		if !opt.IgnoreMissing {
			return nil, err
		} else {
			l.Warn(err)
			return nil, nil
		}
	}
	temp, err := template.New(name).Parse(md)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
