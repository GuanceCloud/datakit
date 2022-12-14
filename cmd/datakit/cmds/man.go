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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func runDocFlags() error {
	inputs.TODO = *flagDocTODO

	opt := &man.Option{
		ManVersion:    *flagDocVersion,
		Path:          *flagDocExportDocs,
		Skips:         *flagDocIgnore,
		IgnoreMissing: true,
	}

	return exportMan(opt)
}

func exportMan(opt *man.Option) error {
	if err := os.MkdirAll(opt.Path, os.ModePerm); err != nil {
		return err
	}

	for k := range inputs.Inputs {
		if strings.Contains(opt.Skips, k) {
			l.Warnf("skip %s", k)
			continue
		}

		l.Debugf("build doc %s.md...", k)
		datas, err := man.BuildMarkdownManual(k, opt)
		if err != nil {
			return fmt.Errorf("man.BuildMarkdownManual: %w", err)
		}

		if len(datas) == 0 {
			l.Warnf("no data, skip %s", k)
			continue
		}

		for i18n, data := range datas {
			if err := ioutil.WriteFile(filepath.Join(opt.Path, i18n.String(), k+".md"), data, os.ModePerm); err != nil {
				return fmt.Errorf("ioutil.WriteFile: %w", err)
			}
		}

		l.Infof("export %s to %s ok", k+".md", opt.Path)
	}

	for k := range man.OtherDocs {
		if strings.Contains(opt.Skips, k) {
			l.Warnf("skip %s", k)
			continue
		}

		datas, err := man.BuildMarkdownManual(k, opt)
		if err != nil {
			return fmt.Errorf("man.BuildMarkdownManual: %w", err)
		}

		if len(datas) == 0 {
			l.Warnf("no data in %s, ignored", k)
			continue
		}

		for i18n, data := range datas {
			if err := ioutil.WriteFile(filepath.Join(opt.Path, i18n.String(), k+".md"), data, os.ModePerm); err != nil {
				return fmt.Errorf("ioutil.WriteFile: %w", err)
			}
		}

		l.Infof("export %s to %s ok", k+".md", opt.Path)
	}

	return nil
}
