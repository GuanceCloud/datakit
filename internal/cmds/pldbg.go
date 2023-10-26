// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/point"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
	plremote "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/targzutil"
)

func runPLFlags() error {
	var txt string

	if *flagPLTxtFile != "" {
		txtBytes, err := os.ReadFile(*flagPLTxtFile)
		if err != nil {
			return fmt.Errorf("os.ReadFile: %w", err)
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
		cp.Warnf("[E] txt has suffix EOL\n")
	}

	var cat point.Category
	switch {
	case point.CatString(*flagPLCategory) != point.UnknownCategory:
		cat = point.CatString(*flagPLCategory)
	case point.CatAlias(*flagPLCategory) != point.UnknownCategory:
		cat = point.CatAlias(*flagPLCategory)
	default:
		return fmt.Errorf("unsupported category: %s", *flagPLCategory)
	}

	return pipelineDebugger(cat, *flagPLName, *flagPLNS, txt, false)
}

func pipelineDebugger(cat point.Category, plname, ns, txt string, isPt bool) error {
	if err := pipeline.InitPipeline(config.Cfg.Pipeline, nil, datakit.GlobalHostTags(),
		datakit.InstallDir); err != nil {
		return err
	}

	if config.Cfg.Pipeline != nil &&
		config.Cfg.Pipeline.ReferTableURL != "" {
		if reft, ok := plval.GetRefTb(); ok && reft != nil {
			if ok := reft.InitFinished(time.Second * 20); ok {
				cp.Infof("Initialize Reference Table: Done")
			} else {
				cp.Errorf("Initialize Reference Table: Timeout")
			}
		}
	}

	scriptTmpStore, errScripts := plScriptTmpStore(cat)

	if m, ok := errScripts[ns]; ok {
		if e, ok := m[plname]; ok {
			return e
		}
	}

	plScript, ok := scriptTmpStore.GetWithNs(plname, ns)

	if !ok {
		return fmt.Errorf("get pipeline failed: name:%s namespace:%s", plname, ns)
	}

	var (
		start = time.Now()

		name = "default"
		pt   *point.Point
		dec  *point.Decoder
	)

	switch cat { //nolint:exhaustive
	case point.Logging:
		fieldsSrc := map[string]interface{}{pipeline.FieldMessage: txt}
		kvs := point.NewKVs(fieldsSrc)
		opt := append(point.DefaultLoggingOptions(), point.WithTime(time.Now()))
		newPt := point.NewPointV2(name, kvs, opt...)
		pt = newPt

	case point.Metric:
		dec = point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
		defer point.PutDecoder(dec)
		pts, err := dec.Decode([]byte(txt), point.DefaultMetricOptions()...)
		if err != nil {
			return err
		}
		pt = pts[0]

	default:
		dec = point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
		defer point.PutDecoder(dec)
		pts, err := dec.Decode([]byte(txt), point.CommonLoggingOptions()...)
		if err != nil {
			return err
		}

		pt = pts[0]
	}

	plRes, err := (&pipeline.Pipeline{
		Script: plScript,
	}).Run(cat, pt, nil, nil)
	if err != nil {
		return fmt.Errorf("run pipeline failed: %w", err)
	}
	cost := time.Since(start)

	if plRes == nil {
		cp.Errorf("[E] No data extracted from pipeline\n")
		return nil
	}

	res := plRes.Point()

	fields := res.InfluxFields()
	tags := res.MapTags()

	if len(fields) == 0 && len(tags) == 0 {
		cp.Errorf("[E] No data extracted from pipeline\n")
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

	name = res.Name()

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
		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", defaultJSONIndent)
		if err := encoder.Encode(result); err != nil {
			return err
		}
		fmt.Printf("%s\n", buf.String())
	}

	cp.Infof("---------------\n")
	cp.Infof("Extracted %d fields, %d tags; measurement(M)<source(L),class(O)...>: %s, drop: %v, cost: %v\n",
		len(fields), len(tags), name, plRes.Dropped(), cost)

	return nil
}

func plScriptTmpStore(category point.Category) (*manager.ScriptStore, map[string]map[string]error) {
	store := manager.NewScriptStore(category, manager.NewManagerCfg(nil, nil))

	errs := map[string]map[string]error{}

	{ // default
		ns := manager.DefaultScriptNS
		plPath := filepath.Join(datakit.InstallDir, "pipeline")
		scripts, scriptsPath := manager.ReadPlScriptFromPlStructPath(plPath)
		errs[ns] = store.UpdateScriptsWithNS(ns, scripts[category], scriptsPath[category])
	}
	{ // gitrepo
		ns := manager.GitRepoScriptNS
		plPath := filepath.Join(datakit.GitReposRepoFullPath, "pipeline")
		scripts, scriptsPath := manager.ReadPlScriptFromPlStructPath(plPath)
		errs[ns] = store.UpdateScriptsWithNS(ns, scripts[category], scriptsPath[category])
	}
	{ // remote
		ns := manager.RemoteScriptNS
		plPath := filepath.Join(datakit.PipelineRemoteDir, plremote.GetConentFileName())
		if tarMap, err := targzutil.ReadTarToMap(plPath); err == nil {
			allCategory := plremote.ConvertContentMapToThreeMap(tarMap)
			scripts := allCategory[datakit.CategoryDirName()[category.String()]]
			scriptsPath := map[string]string{}
			for k := range scripts {
				scriptsPath[k] = filepath.Join(plPath, category.String(), k)
			}
			errs[ns] = store.UpdateScriptsWithNS(ns, scripts, scriptsPath)
		}
	}

	return store, errs
}
