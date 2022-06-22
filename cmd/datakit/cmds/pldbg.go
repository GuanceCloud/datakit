// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/convertutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/targzutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	plremote "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/remote"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
)

func runPLFlags() error {
	var txt string

	if *flagPLTxtFile != "" {
		txtBytes, err := ioutil.ReadFile(*flagPLTxtFile)
		if err != nil {
			return fmt.Errorf("ioutil.ReadFile: %w", err)
		}
		txt = string(txtBytes)
		txt = strings.TrimSuffix(txt, "\n")
	}

	if txt == "" {
		if *flagPLTxtData != "" {
			txt = *flagPLTxtData
		}
	}

	if txt == "" {
		return fmt.Errorf("no testing string")
	}

	if strings.HasSuffix(txt, "\n") {
		warnf("[E] txt has suffix EOL\n")
	}

	// TODO
	return pipelineDebugger(*flagPLCategory, debugPipelineName, *flagPLNS, txt, false)
}

func pipelineDebugger(category, plname, ns, txt string, isPt bool) error {
	category, err := convertutil.GetMapCategoryShortToFull(category)
	if err != nil {
		return err
	}

	if err := pipeline.Init(config.Cfg.Pipeline); err != nil {
		return err
	}

	scriptTmpStore, errScripts := plScriptTmpStore(category)

	if m, ok := errScripts[ns]; ok {
		if e, ok := m[plname]; ok {
			return e
		}
	}

	plScript, ok := scriptTmpStore.GetWithNs(plname, ns)

	if !ok {
		return fmt.Errorf("get pipeline failed: name:%s namespace:%s", plname, ns)
	}

	start := time.Now()

	opt := &io.PointOption{
		Category: category,
		Time:     time.Now(),
	}

	measurementName := "default"

	var pt *io.Point

	switch category {
	case datakit.Logging:
		fieldsSrc := map[string]interface{}{pipeline.FieldMessage: txt}
		newPt, err := io.NewPoint(measurementName, nil, fieldsSrc, opt)
		if err != nil {
			return err
		}
		pt = newPt
	default:
		pts, err := lp.ParsePoints([]byte(txt), nil)
		if err != nil {
			return err
		}
		ptsW := io.WrapPoint(pts)
		pt = ptsW[0]
	}

	res, dropFlag, err := (&pipeline.Pipeline{
		Script: plScript,
	}).Run(pt, nil, *opt)
	if err != nil {
		return fmt.Errorf("run pipeline failed: %w", err)
	}
	cost := time.Since(start)

	if res == nil {
		errorf("[E] No data extracted from pipeline\n")
		return nil
	}

	fields, _ := res.Fields()
	tags := res.Tags()
	if len(fields) == 0 && len(tags) == 0 {
		errorf("[E] No data extracted from pipeline\n")
		return nil
	}

	result := map[string]interface{}{}
	maxWidth := 0

	if *flagPLDate {
		result["time"] = res.Time()
	} else {
		result["time"] = res.Time().UnixNano()
	}

	for k, v := range fields {
		if len(k) > maxWidth {
			maxWidth = len(k)
		}
		result[k] = v
	}

	for k, v := range tags {
		result[k+"#"] = v
		if len(k)+1 > maxWidth {
			maxWidth = len(k) + 1
		}
	}

	measurementName = res.Name()

	if *flagPLTable {
		fmtStr := fmt.Sprintf("%% %ds: %%v", maxWidth)
		lines := []string{}
		for k, v := range result {
			lines = append(lines, fmt.Sprintf(fmtStr, k, v))
		}

		sort.Strings(lines)
		for _, l := range lines {
			fmt.Println(l)
		}
	} else {
		j, err := json.MarshalIndent(result, "", defaultJSONIndent)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(j))
	}

	infof("---------------\n")
	infof("Extracted %d fields, %d tags; measurement(M)<source(L),class(O)...>: %s, drop: %v, cost: %v\n",
		len(fields), len(tags), measurementName, dropFlag, cost)

	return nil
}

func plScriptTmpStore(category string) (*script.ScriptStore, map[string]map[string]error) {
	store := script.NewScriptStore(category)

	errs := map[string]map[string]error{}

	{ // default
		ns := script.DefaultScriptNS
		plPath := filepath.Join(datakit.InstallDir, "pipeline")
		scripts, scriptsPath := script.ReadPlScriptFromPlStructPath(plPath)
		errs[ns] = store.UpdateScriptsWithNS(ns, scripts[category], scriptsPath[category])
	}
	{ // gitrepo
		ns := script.GitRepoScriptNS
		plPath := filepath.Join(datakit.GitReposRepoFullPath, "pipeline")
		scripts, scriptsPath := script.ReadPlScriptFromPlStructPath(plPath)
		errs[ns] = store.UpdateScriptsWithNS(ns, scripts[category], scriptsPath[category])
	}
	{ // remote
		ns := script.RemoteScriptNS
		plPath := filepath.Join(datakit.PipelineRemoteDir, plremote.GetConentFileName())
		if tarMap, err := targzutil.ReadTarToMap(plPath); err == nil {
			allCategory := plremote.ConvertContentMapToThreeMap(tarMap)
			scripts := allCategory[datakit.CategoryDirName()[category]]
			scriptsPath := map[string]string{}
			for k := range scripts {
				scriptsPath[k] = filepath.Join(plPath, category, k)
			}
			errs[ns] = store.UpdateScriptsWithNS(ns, scripts, scriptsPath)
		}
	}

	return store, errs
}
