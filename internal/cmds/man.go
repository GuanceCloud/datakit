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

type i18n int

const (
	i18nZh = iota
	i18nEn
)

func (x i18n) String() string {
	switch x {
	case i18nZh:
		return "zh"
	case i18nEn:
		return "en"
	default:
		panic(fmt.Sprintf("should not been here: unsupport language: %s", x.String()))
	}
}

func buildAllDocs(lang i18n, opt *man.ExportOption) (map[string][]byte, error) {
	res := map[string][]byte{}

	// build all inputs docs
	inputDocs, err := man.AllDocs.ReadDir(filepath.Join("docs", lang.String(), "inputs"))
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

		name := strings.Split(f.Name(), ".")[0] // get cpu.md -> cpu
		md, err := man.AllDocs.ReadFile(filepath.Join("docs", lang.String(), "inputs", f.Name()))
		if err != nil {
			continue
		}

		doc, err := man.BuildInputDoc(name, md, opt)
		if err != nil {
			cp.Errorf("man.BuildInputDoc(%q): %s", name, err)
			continue
		}

		res[f.Name()] = doc
	}

	// build other docs
	otherDocs, err := man.AllDocs.ReadDir(filepath.Join("docs", lang.String()))
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

		md, err := man.AllDocs.ReadFile(filepath.Join("docs", lang.String(), f.Name()))
		if err != nil {
			continue
		}

		doc, err := man.BuildNonInputDocs(md, opt)
		if err != nil {
			cp.Errorf("man.BuildInputDoc(%q): %s", f.Name(), err)
			continue
		}

		res[f.Name()] = doc
	}

	// build pipeline docs
	plmd, err := man.AllDocs.ReadFile(filepath.Join("docs", lang.String(), "pipeline.md"))
	if err != nil {
		return nil, err
	}

	var fndocs map[string]*plfuncs.PLDoc
	switch lang {
	case i18nZh:
		fndocs = plfuncs.PipelineFunctionDocs
	case i18nEn:
		fndocs = plfuncs.PipelineFunctionDocsEN
	}

	doc, err := man.BuildPipelineDocs(plmd, fndocs, opt)
	if err != nil {
		return nil, err
	} else {
		res["pipeline.md"] = doc
	}

	return res, nil
}

func exportMan(opt *man.ExportOption) error {
	if err := os.MkdirAll(opt.Path, os.ModePerm); err != nil {
		return err
	}

	for _, x := range []i18n{i18nZh, i18nEn} {
		docs, err := buildAllDocs(x, opt)
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
