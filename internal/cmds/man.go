// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/man"
	plfuncs "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func runDocFlags() error {
	inputs.TODO = *flagDocTODO

	opt := &man.ExportOption{
		ManVersion:    *flagDocVersion,
		Path:          *flagDocExportDocs,
		Skips:         *flagDocIgnore,
		IgnoreMissing: true,
	}

	return exportMan(opt)
}

// buildAll used to build all exported data based on template.
//
// These data include:
//  - docs: markdown doc
//  - pipeline: markdown doc
//  - dashboard: JSON
//  - monitor: JSON
func buildAll(lang inputs.I18n, opt *man.ExportOption) (map[string][]byte, error) {
	res := map[string][]byte{}

	// build all inputs docs
	inputDocs, err := man.AllDocs.ReadDir(filepath.Join("doc", lang.String(), "inputs"))
	if err != nil {
		return nil, err
	}

	for _, f := range inputDocs {
		if f.IsDir() {
			continue // ignore dir
		}

		if !strings.HasSuffix(f.Name(), ".md") {
			continue // ignore non-markdown
		}

		name := strings.Split(f.Name(), ".")[0] // cpu.md -> cpu

		if strings.Contains(opt.Skips, name) {
			cp.Infof("skip build exports for input %q, skip list: %q\n", name, opt.Skips)
			continue
		}

		md, err := man.AllDocs.ReadFile(filepath.Join("doc", lang.String(), "inputs", f.Name()))
		if err != nil {
			cp.Warnf("read doc on input %q failed: %s, ignored\n", name, err)
			continue
		}

		doc, err := man.BuildInputDoc(name, md, opt)
		if err != nil {
			cp.Errorf("man.BuildInputDoc(%q): %s\n", name, err)
			os.Exit(-1) // failed to build markdown, fail ASAP.
		}

		res[filepath.Join("inputs", f.Name())] = doc

		// build monitor
		monitors, err := man.BuildMonitor(name, lang)
		if err != nil {
			cp.Errorf("man.BuildMonitor(%q): %s\n", name, err)
			os.Exit(-1) // failed to build markdown, fail ASAP.
		} else {
			if len(monitors) > 0 {
				for k, v := range monitors {
					res[filepath.Join("monitor", k+".json")] = v
				}
			} else {
				cp.Warnf("[W] no monitor for %s(%s), ignored\n", name, lang)
			}
		}

		// build dashboard: one input may got multiple dashboards
		dashboards, err := man.BuildDashboard(name, lang)
		if err != nil {
			cp.Errorf("man.BuildDashboard(%q): %s\n", name, err)
			os.Exit(-1) // failed to build markdown, fail ASAP.
		} else {
			if len(dashboards) > 0 {
				for k, v := range dashboards {
					res[filepath.Join("dashboard", k+".json")] = v
				}
			} else {
				cp.Warnf("[W] no dashboard for %s(%s), ignored\n", name, lang)
			}
		}
	}

	// build all pipeline docs
	plDocs, err := man.AllDocs.ReadDir(filepath.Join("doc", lang.String(), "pipeline"))
	if err != nil {
		return nil, err
	}

	for _, f := range plDocs {
		if f.IsDir() {
			continue
		}

		if !strings.HasSuffix(f.Name(), ".md") {
			continue // ignore non-markdown
		}

		md, err := man.AllDocs.ReadFile(filepath.Join("doc", lang.String(), "pipeline", f.Name()))
		if err != nil {
			continue
		}

		if f.Name() == "pipeline-built-in-function.md" {
			var fndocs map[string]*plfuncs.PLDoc
			switch lang {
			case inputs.I18nZh:
				fndocs = plfuncs.PipelineFunctionDocs
			case inputs.I18nEn:
				fndocs = plfuncs.PipelineFunctionDocsEN
			}

			doc, err := man.BuildPipelineDocs(md, fndocs, opt)
			if err != nil {
				return nil, err
			} else {
				if _, ok := res["pipeline/pipeline-built-in-function.md"]; ok {
					return nil, fmt.Errorf("resource pipeline/pipeline-built-in-function.md exists")
				}
				res["pipeline/pipeline-built-in-function.md"] = doc
			}
		} else {
			res[filepath.Join("pipeline", f.Name())] = md
		}
	}

	// build other docs
	otherDocs, err := man.AllDocs.ReadDir(filepath.Join("doc", lang.String()))
	if err != nil {
		return nil, err
	}

	for _, f := range otherDocs {
		if f.IsDir() {
			continue
		}

		if !strings.HasSuffix(f.Name(), ".md") {
			continue
		}

		if f.Name() == "pipeline.md" {
			continue // ignore pipeline docs
		}

		md, err := man.AllDocs.ReadFile(filepath.Join("doc", lang.String(), f.Name()))
		if err != nil {
			continue
		}

		doc, err := man.BuildNonInputDocs(md, opt)
		if err != nil {
			cp.Errorf("man.BuildInputDoc(%q): %s\n", f.Name(), err)
			os.Exit(-1) // failed to build markdown, fail ASAP.
		}

		res[f.Name()] = doc
	}

	return res, nil
}

func exportMan(opt *man.ExportOption) error {
	if err := os.MkdirAll(opt.Path, os.ModePerm); err != nil {
		return err
	}

	for _, x := range []inputs.I18n{inputs.I18nZh, inputs.I18nEn} {
		// create dir to hold the docs
		for _, dir := range []string{
			filepath.Join(opt.Path, x.String()),
			filepath.Join(opt.Path, x.String(), "pipeline"),
			filepath.Join(opt.Path, x.String(), "inputs"),
			filepath.Join(opt.Path, x.String(), "dashboard"),
			filepath.Join(opt.Path, x.String(), "monitor"),
		} {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}

		docs, err := buildAll(x, opt)
		if err != nil {
			return err
		}

		for f, data := range docs {
			if err := ioutil.WriteFile(filepath.Join(opt.Path, x.String(), f), data, os.ModePerm); err != nil {
				return fmt.Errorf("ioutil.WriteFile: %w", err)
			}
		}
	}

	return nil
}
