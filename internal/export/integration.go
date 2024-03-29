// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

// Export docs/dashboards/monitors to integrations

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GuanceCloud/cliutils/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/all"
)

const (
	templateDashboardDir = "dashboard"
	templateMonitordDir  = "monitor"
)

type Integration struct {
	opt *exportOptions

	docs map[string][]byte
}

func NewIntegration(opts ...option) *Integration {
	eo := &exportOptions{
		langs:   []inputs.I18n{inputs.I18nZh, inputs.I18nEn},
		version: "not-set",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(eo)
		}
	}

	return &Integration{
		opt:  eo,
		docs: map[string][]byte{},
	}
}

func (i *Integration) Export() error {
	for _, lang := range i.opt.langs {
		l.Infof("exporting monitor(%s)...", lang)
		if err := i.exportTemplate(templateMonitordDir, lang); err != nil {
			return err
		}

		l.Infof("exporting dashboard(%s)...", lang)
		if err := i.exportTemplate(templateDashboardDir, lang); err != nil {
			return err
		}

		l.Infof("exporting integration(%s)...", lang)
		if err := i.exportIntegration(lang); err != nil {
			return err
		}

		l.Infof("exporting miscs(%s)...", lang)
		if err := i.exportMiscs(lang); err != nil {
			return err
		}
	}

	// prepare dirs
	dirs := map[string]bool{}
	for k := range i.docs {
		dirs[filepath.Dir(k)] = true
	}

	for dir := range dirs {
		l.Debugf("create dir %q", dir)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("MkdirAll: %w", err)
		}
	}

	for k, v := range i.docs {
		if err := os.WriteFile(k, v, 0o600); err != nil {
			return err
		}
		l.Debugf("write %q...", k)
	}

	return nil
}

func (i *Integration) Check() error { return nil }

// exportMisc export pipeline sample/docs(base64)/metric docs.
func (i *Integration) exportMiscs(lang inputs.I18n) error {
	if j, err := exportMetaInfo(inputs.Inputs); err != nil {
		return err
	} else {
		i.docs[filepath.Join(i.opt.topDir,
			"datakit",
			lang.String(),
			"measurements-meta.json")] = j
	}

	pe := newPLB64DocExporter(lang)
	if j, err := pe.export(); err != nil {
		return err
	} else {
		i.docs[filepath.Join(i.opt.topDir,
			"datakit",
			lang.String(),
			"pipeline-docs.json")] = j
	}

	if res, err := getPipelineDemoMap(); err != nil {
		return err
	} else {
		// Encode script and log examples with base64.
		for scriptName, demo := range res {
			demo.Pipeline = base64.StdEncoding.EncodeToString([]byte(demo.Pipeline))
			for n, e := range demo.Examples {
				demo.Examples[n] = base64.StdEncoding.EncodeToString([]byte(e))
			}
			res[scriptName] = demo
		}

		if j, err := json.Marshal(res); err != nil {
			return err
		} else {
			i.docs[filepath.Join(i.opt.topDir,
				"datakit",
				lang.String(),
				"internal-pipelines.json")] = j
			return nil
		}
	}
}

type pipelineDemo struct {
	Pipeline string            `json:"pipeline"`
	Examples map[string]string `json:"examples"`
}

func getPipelineDemoMap() (map[string]pipelineDemo, error) {
	demoMap := map[string]pipelineDemo{}
	for _, c := range inputs.Inputs {
		if v, ok := c().(inputs.PipelineInput); ok {
			for n, script := range v.PipelineConfig() {
				var d pipelineDemo
				// Ignore empty pipeline script.
				if script == "" {
					continue
				}
				name := n + ".p"
				if _, has := demoMap[name]; has {
					return nil, fmt.Errorf("duplicated pipeline script name: %s", name)
				}
				d.Pipeline = script
				if exampler, ok := c().(inputs.LogExampler); ok {
					if examples, has := exampler.LogExamples()[n]; has {
						d.Examples = examples
					}
				}
				demoMap[name] = d
			}
		}
	}
	return demoMap, nil
}

// export Pipeline docs in base64 format.
type plB64DocExporter struct {
	protoPrefix,
	descPrefix string

	Version   string                  `json:"version"`
	Docs      string                  `json:"docs"`
	Functions map[string]*funcs.PLDoc `json:"functions"`
}

func newPLB64DocExporter(lang inputs.I18n) *plB64DocExporter {
	// nolint: exhaustive
	switch lang {
	case inputs.I18nEn:
		return &plB64DocExporter{
			protoPrefix: "Function prototype: ",
			descPrefix:  "Function description: ",
			Version:     git.Version,
			Docs:        "Base64-encoded pipeline function documentation, including function prototypes, function descriptions, and usage examples.",
			Functions:   funcs.PipelineFunctionDocsEN,
		}

	default: // zh
		return &plB64DocExporter{
			protoPrefix: "函数原型：",
			descPrefix:  "函数说明：",
			Version:     git.Version,
			Docs:        "经过 base64 编码的 pipeline 函数文档，包括各函数原型、函数说明、使用示例",
			Functions:   funcs.PipelineFunctionDocs,
		}
	}
}

func (e *plB64DocExporter) export() ([]byte, error) {
	for _, plDoc := range e.Functions {
		lines := strings.Split(plDoc.Doc, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, e.protoPrefix) {
				proto := strings.TrimPrefix(line, e.protoPrefix)
				// Prototype line contains starting and trailing ` only.
				if len(proto) >= 2 && strings.Index(proto, "`") == 0 && strings.Index(proto[1:], "`") == len(proto[1:])-1 {
					proto = proto[1 : len(proto)-1]
				}
				plDoc.Prototype = proto
			} else if strings.HasPrefix(line, e.descPrefix) {
				plDoc.Description = strings.TrimPrefix(line, e.descPrefix)
			}
		}
	}

	// Encode Markdown docs with base64.
	for _, plDoc := range e.Functions {
		plDoc.Doc = base64.StdEncoding.EncodeToString([]byte(plDoc.Doc))
		plDoc.Prototype = base64.StdEncoding.EncodeToString([]byte(plDoc.Prototype))
		plDoc.Description = base64.StdEncoding.EncodeToString([]byte(plDoc.Description))
	}

	return json.Marshal(e)
}

func (i *Integration) exportIntegration(lang inputs.I18n) error {
	// build all inputs docs
	inputDocs, err := AllDocs.ReadDir(filepath.Join("doc", lang.String(), "inputs"))
	if err != nil {
		return err
	}

	var (
		ignored = 0
		n       = 0
	)

	for _, f := range inputDocs {
		if f.IsDir() {
			continue // ignore dir
		}

		if !strings.HasSuffix(f.Name(), ".md") {
			ignored++
			continue // ignore non-markdown
		}

		name := strings.Split(f.Name(), ".")[0] // cpu.md -> cpu

		if strings.Contains(i.opt.exclude, name) {
			l.Infof("skip build exports for input %q, skip list: %q", name, i.opt.exclude)
			continue
		}

		md, err := AllDocs.ReadFile(filepath.Join("doc", lang.String(), "inputs", f.Name()))
		if err != nil {
			l.Warnf("read doc on input %q failed: %s, ignored", name, err)
			continue
		}

		var doc []byte
		if _, ok := inputs.Inputs[name]; ok {
			l.Debugf("add non-input doc %q to integration", f.Name())
			doc, err = buildInputDoc(name, md, i.opt)
			if err != nil {
				return err
			}
		} else { // non-input docs, but they related to input, we put them to integrations subdir
			l.Debugf("add non-input doc %q to integration", f.Name())
			doc, err = buildNonInputDocs("doc/inputs/"+f.Name(), md, i.opt)
			if err != nil {
				return err
			}
		}

		i.docs[filepath.Join(i.opt.topDir, "integration", lang.String(), f.Name())] = doc
		n++
	}

	l.Infof("exported %d input docs, ignored: %d, total: %d", n, ignored, len(inputDocs))
	return nil
}

// exportTemplate export dashboard or monitor template.
func (i *Integration) exportTemplate(templateDir string, lang inputs.I18n) error {
	templateEntry, err := AllTemplates.ReadDir(templateDir)
	if err != nil {
		return err
	}

	for _, e := range templateEntry {
		if !e.IsDir() {
			l.Debugf("ignore non-dir %q under %s", e.Name(), templateDir)
			continue
		}
		if err := i.buildTemplate(templateDir, e.Name(), lang); err != nil {
			return err
		}
	}

	return nil
}

func (i *Integration) buildTemplate(templateDir, inputName string, lang inputs.I18n) error {
	var (
		templateMap = map[string][]byte{}
		p           = &Params{}
	)

	if templateDir == templateMonitordDir {
		// For monitor json, we have to escape jinja2 template(also the format {{ xx }}),
		// so we use customer delimeters for Go template.
		p.delims = [2]string{"<<", ">>"}
	}

	// inputs may specified its dashboard/monitor's specs.
	if c, ok := inputs.Inputs[inputName]; ok && c != nil {
		ipt := c()

		if i, ok := ipt.(inputs.Dashboard); ok {
			p.Dashboard = i.Dashboard(lang)
		}

		if i, ok := ipt.(inputs.Monitor); ok {
			p.Monitor = i.Monitor(lang)
		}
	} else {
		l.Warnf("input %s not exist", inputName)
	}

	// load all xx.json files under collector
	templateFileEntry, err := AllTemplates.ReadDir(filepath.Join(templateDir, inputName))
	if err != nil {
		return err
	}

	for _, e := range templateFileEntry {
		jsonFileName := e.Name()
		if e.IsDir() || !strings.HasSuffix(jsonFileName, ".json") {
			l.Debugf("ignore dir %s under collector %s", e.Name(), inputName)
			continue
		}

		// such as mysql.json/mysql_dbm.json
		fileParts := strings.Split(jsonFileName, ".")
		if len(fileParts) != 2 {
			l.Debugf("invalid file name %s for collector %s", jsonFileName, inputName)
			continue
		}

		if !strings.HasSuffix(fileParts[0], fmt.Sprintf("__%s", inputs.I18nZh.String())) &&
			!strings.HasSuffix(fileParts[0], fmt.Sprintf("__%s", inputs.I18nEn.String())) {
			if content, err := AllTemplates.ReadFile(filepath.Join(templateDir, inputName, jsonFileName)); err == nil {
				templateMap[fileParts[0]] = content
			} else {
				return err
			}
		} else if strings.HasSuffix(fileParts[0], fmt.Sprintf("__%s", lang.String())) {
			if content, err := AllTemplates.ReadFile(filepath.Join(templateDir, inputName, jsonFileName)); err == nil {
				fileName := strings.TrimSuffix(fileParts[0], fmt.Sprintf("__%s", lang.String()))
				i.docs[filepath.Join(i.opt.topDir, templateDir, lang.String(), fileName, "meta.json")] = content
			} else {
				return err
			}
		}
	}

	l.Infof("build %s: %d templates for %s...", templateDir, len(templateMap), inputName)

	for k, t := range templateMap {
		l.Infof("render %s...", k)
		buf, err := renderBuf(t, p)
		if err != nil {
			return fmt.Errorf("renderBuf: render on input %q[%q]: %w",
				inputName, k, err)
		} else {
			// check if JSON ok
			if !json.Valid(buf) {
				return fmt.Errorf("invalid JSON on input %q[%q]", inputName, k)
			}

			i.docs[filepath.Join(i.opt.topDir, templateDir, lang.String(), k, "meta.json")] = buf
		}
	}

	return nil
}
