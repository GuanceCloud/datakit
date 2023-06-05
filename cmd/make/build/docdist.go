// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
)

func generateMetaInfo() error {
	return cmds.ExportMetaInfo("measurements-meta.json")
}

func generatePipelineDocEN() error {
	encoding := base64.StdEncoding
	protoPrefix, descPrefix := "Function prototype: ", "Function description: "
	// Write function description & prototype.
	for _, plDoc := range funcs.PipelineFunctionDocsEN {
		lines := strings.Split(plDoc.Doc, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, protoPrefix) {
				proto := strings.TrimPrefix(line, protoPrefix)
				// Prototype line contains starting and trailing ` only.
				if len(proto) >= 2 && strings.Index(proto, "`") == 0 && strings.Index(proto[1:], "`") == len(proto[1:])-1 {
					proto = proto[1 : len(proto)-1]
				}
				plDoc.Prototype = proto
			} else if strings.HasPrefix(line, descPrefix) {
				plDoc.Description = strings.TrimPrefix(line, descPrefix)
			}
		}
	}
	// Encode Markdown docs with base64.
	for _, plDoc := range funcs.PipelineFunctionDocsEN {
		plDoc.Doc = encoding.EncodeToString([]byte(plDoc.Doc))
		plDoc.Prototype = encoding.EncodeToString([]byte(plDoc.Prototype))
		plDoc.Description = encoding.EncodeToString([]byte(plDoc.Description))
	}
	exportPLDocs := struct {
		Version   string                  `json:"version"`
		Docs      string                  `json:"docs"`
		Functions map[string]*funcs.PLDoc `json:"functions"`
	}{
		Version:   git.Version,
		Docs:      "Base64-encoded pipeline function documentation, including function prototypes, function descriptions, and usage examples.",
		Functions: funcs.PipelineFunctionDocsEN,
	}
	data, err := json.Marshal(exportPLDocs)
	if err != nil {
		return err
	}
	f, err := os.Create("pipeline-docs.en.json")
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func generatePipelineDoc() error {
	encoding := base64.StdEncoding
	protoPrefix, descPrefix := "函数原型：", "函数说明："
	// Write function description & prototype.
	for _, plDoc := range funcs.PipelineFunctionDocs {
		lines := strings.Split(plDoc.Doc, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, protoPrefix) {
				proto := strings.TrimPrefix(line, protoPrefix)
				// Prototype line contains starting and trailing ` only.
				if len(proto) >= 2 && strings.Index(proto, "`") == 0 && strings.Index(proto[1:], "`") == len(proto[1:])-1 {
					proto = proto[1 : len(proto)-1]
				}
				plDoc.Prototype = proto
			} else if strings.HasPrefix(line, descPrefix) {
				plDoc.Description = strings.TrimPrefix(line, descPrefix)
			}
		}
	}
	// Encode Markdown docs with base64.
	for _, plDoc := range funcs.PipelineFunctionDocs {
		plDoc.Doc = encoding.EncodeToString([]byte(plDoc.Doc))
		plDoc.Prototype = encoding.EncodeToString([]byte(plDoc.Prototype))
		plDoc.Description = encoding.EncodeToString([]byte(plDoc.Description))
	}
	exportPLDocs := struct {
		Version   string                  `json:"version"`
		Docs      string                  `json:"docs"`
		Functions map[string]*funcs.PLDoc `json:"functions"`
	}{
		Version:   git.Version,
		Docs:      "经过 base64 编码的 pipeline 函数文档，包括各函数原型、函数说明、使用示例",
		Functions: funcs.PipelineFunctionDocs,
	}
	data, err := json.Marshal(exportPLDocs)
	if err != nil {
		return err
	}
	f, err := os.Create("pipeline-docs.json")
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func generatePipelineScripts() error {
	encoding := base64.StdEncoding
	demoMap, err := config.GetPipelineDemoMap()
	if err != nil {
		return err
	}

	// Encode script and log examples with base64.
	for scriptName, demo := range demoMap {
		demo.Pipeline = encoding.EncodeToString([]byte(demo.Pipeline))
		for n, e := range demo.Examples {
			demo.Examples[n] = encoding.EncodeToString([]byte(e))
		}
		demoMap[scriptName] = demo
	}

	data, err := json.Marshal(demoMap)
	if err != nil {
		return err
	}
	f, err := os.Create("internal-pipelines.json")
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}
