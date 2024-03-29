// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

// GuanceDocs export all markdown docs to docs.guance.com.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/GuanceCloud/cliutils/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type GuanceDocs struct {
	opt  *exportOptions
	docs map[string][]byte
}

func NewGuanceDodcs(opts ...option) *GuanceDocs {
	eo := &exportOptions{
		langs:   []inputs.I18n{inputs.I18nZh, inputs.I18nEn},
		version: "not-set",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(eo)
		}
	}

	return &GuanceDocs{
		opt:  eo,
		docs: map[string][]byte{},
	}
}

func (gd *GuanceDocs) Export() error {
	dirs := []string{
		"datakit",
		"integrations",
		"developers/pipeline",
	}

	for _, lang := range gd.opt.langs {
		if err := gd.exportPipelineDocs(lang); err != nil {
			return err
		}

		if err := gd.exportDatakitDocs(lang); err != nil {
			return err
		}

		if err := gd.exportInputDocs(lang); err != nil {
			return err
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(filepath.Join(gd.opt.topDir, lang.String(), dir), os.ModePerm); err != nil {
				return err
			}
		}
	}

	for k, v := range gd.docs {
		if err := os.WriteFile(k, v, 0o600); err != nil {
			return err
		}
		l.Debugf("write %q...", k)
	}

	return nil
}

func (gd *GuanceDocs) Check() error {
	// TODO: check if docs ok
	return nil
}

func (gd *GuanceDocs) exportPipelineDocs(lang inputs.I18n) error {
	// build all pipeline docs
	plDocs, err := AllDocs.ReadDir(filepath.Join("doc", lang.String(), "pipeline"))
	if err != nil {
		return err
	}

	var (
		n       = 0
		ignored = 0
	)

	for _, f := range plDocs {
		if f.IsDir() {
			continue
		}

		md, err := AllDocs.ReadFile(filepath.Join("doc", lang.String(), "pipeline", f.Name()))
		if err != nil {
			ignored++
			continue
		}

		if strings.Contains(f.Name(), ".pages") {
			gd.docs[filepath.Join(gd.opt.topDir, lang.String(), "developers", "pipeline", ".pages")] = md
			continue
		}

		if !strings.HasSuffix(f.Name(), ".md") {
			ignored++
			continue // ignore non-markdown
		}

		doc := md
		if f.Name() == "pipeline-built-in-function.md" {
			var fndocs map[string]*funcs.PLDoc
			switch lang {
			case inputs.I18nZh:
				fndocs = funcs.PipelineFunctionDocs
			case inputs.I18nEn:
				fndocs = funcs.PipelineFunctionDocsEN
			}

			doc, err = buildPipelineDocs(md, fndocs, gd.opt)
			if err != nil {
				return err
			}
		}

		l.Debugf("add doc %q to pipeline", f.Name())
		gd.docs[filepath.Join(gd.opt.topDir, lang.String(), "developers", "pipeline", f.Name())] = doc
		n++
	}

	l.Infof("exported %d pipeline docs, ignored: %d, total: %d", n, ignored, len(plDocs))
	return nil
}

// exportDatakitDocs export datakit itself docs, exclude inputs/pipeline docs.
func (gd *GuanceDocs) exportDatakitDocs(lang inputs.I18n) error {
	dkDocs, err := AllDocs.ReadDir(filepath.Join("doc", lang.String()))
	if err != nil {
		return err
	}

	var (
		ignored = 0
		n       = 0
	)
	for _, f := range dkDocs {
		if f.IsDir() {
			continue
		}

		md, err := AllDocs.ReadFile(filepath.Join("doc", lang.String(), f.Name()))
		if err != nil {
			ignored++
			continue
		}

		if f.Name() == "datakit.pages" {
			gd.docs[filepath.Join(gd.opt.topDir, lang.String(), "datakit", ".pages")] = md
			continue
		}

		if !strings.HasSuffix(f.Name(), ".md") {
			ignored++
			continue
		}

		doc, err := buildNonInputDocs("doc/"+f.Name(), md, gd.opt)
		if err != nil {
			l.Errorf("buildNonInputDocs(%q): %s", f.Name(), err)
			return err
		}

		l.Debugf("add doc %q to datakit", f.Name())
		gd.docs[filepath.Join(gd.opt.topDir, lang.String(), "datakit", f.Name())] = doc
		n++
	}

	l.Infof("exported %d datakit docs, ignored: %d, total: %d", n, ignored, len(dkDocs))

	return nil
}

func (gd *GuanceDocs) exportInputDocs(lang inputs.I18n) error {
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

		if strings.Contains(gd.opt.exclude, name) {
			l.Infof("skip build exports for input %q, skip list: %q", name, gd.opt.exclude)
			continue
		}

		md, err := AllDocs.ReadFile(filepath.Join("doc", lang.String(), "inputs", f.Name()))
		if err != nil {
			l.Warnf("read doc on input %q failed: %s, ignored", name, err)
			continue
		}

		var doc []byte
		if _, ok := inputs.Inputs[name]; ok {
			doc, err = buildInputDoc(name, md, gd.opt)
			if err != nil {
				return err
			}
		} else { // non-input docs, but they related to input, we put them to integrations subdir
			doc, err = buildNonInputDocs("doc/inputs/"+f.Name(), md, gd.opt)
			if err != nil {
				return err
			}
		}

		l.Debugf("add doc %q to integrations", f.Name())
		gd.docs[filepath.Join(gd.opt.topDir, lang.String(), "integrations", f.Name())] = doc
		n++
	}

	l.Infof("exported %d input docs, ignored: %d, total: %d", n, ignored, len(inputDocs))
	return nil
}
