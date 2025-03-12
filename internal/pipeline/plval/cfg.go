// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"regexp"

	"github.com/GuanceCloud/grok"
	"github.com/GuanceCloud/pipeline-go/manager"
	"github.com/GuanceCloud/pipeline-go/offload"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb/geoip"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb/iploc"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
)

var pipelineDefaultCfg = &PipelineCfg{
	IPdbType: "iploc",
	IPdbAttr: map[string]string{
		"iploc_file": "iploc.bin",
		"isp_file":   "ip2isp.txt",
	},
}

const (
	TypIPLOC    = "iploc"
	TypGeoLite2 = "geolite2"
)

type PipelineCfg struct {
	IPdbAttr               map[string]string      `toml:"ipdb_attr"`
	IPdbType               string                 `toml:"ipdb_type"`
	RemotePullInterval     string                 `toml:"remote_pull_interval"`
	ReferTableURL          string                 `toml:"refer_table_url"`
	ReferTablePullInterval string                 `toml:"refer_table_pull_interval"`
	UseSQLite              bool                   `toml:"use_sqlite"`
	SQLiteMemMode          bool                   `toml:"sqlite_mem_mode"`
	Offload                *offload.OffloadConfig `toml:"offload"`
	EnableDebugFields      bool                   `toml:"-"`
	DefaultPipeline        map[string]string      `toml:"default_pipeline"`

	DisableHTTPRequestFunc        bool     `toml:"disable_http_request_func"`
	HTTPRequestHostWhitelist      []string `toml:"http_request_host_whitelist"`
	HTTPRequestCIDRWhitelist      []string `toml:"http_request_cidr_whiltelist"`
	HTTPRequestDisableInternalNet bool     `toml:"http_request_disable_internal_net"`

	DeprecatedDisableAppendRunInfo bool `toml:"disable_append_run_info"`
}

// InitIPdb init ipdb instance.
func InitIPdb(dataDir string, pipelineCfg *PipelineCfg) (ipdb.IPdb, error) {
	var ipdbInstance ipdb.IPdb // get ip location and isp

	if pipelineCfg == nil {
		pipelineCfg = pipelineDefaultCfg
	}

	switch pipelineCfg.IPdbType {
	case TypIPLOC:
		ipdbInstance = &iploc.IPloc{}
		ipdbInstance.Init(dataDir, pipelineCfg.IPdbAttr)
	case TypGeoLite2:
		ipdbInstance = &geoip.Geoip{}
		ipdbInstance.Init(dataDir, pipelineCfg.IPdbAttr)
	default:
		l.Warnf("invalid ipdb_type %s, use default iploc", pipelineCfg.IPdbType)
		ipdbInstance = &iploc.IPloc{}
		ipdbInstance.Init(dataDir, pipelineCfg.IPdbAttr)
	}

	return ipdbInstance, nil
}

func LoadPatterns(patternDir string) error {
	// 从文件加载 pattern
	loadedPatterns, err := grok.LoadPatternsFromPath(patternDir)
	if err != nil {
		return err
	}

	// 使用内置的 pattern，可能覆盖文件中的 pattern
	for k, v := range manager.CopyDefalutPatterns() {
		loadedPatterns[k] = v
	}

	denormalizedGlobalPatterns, invalidPatterns := grok.DenormalizePatternsFromMap(loadedPatterns)

	for k, v := range denormalizedGlobalPatterns {
		if _, err := regexp.Compile(v.Denormalized()); err != nil {
			invalidPatterns[k] = err.Error()
		}
	}

	if len(invalidPatterns) != 0 {
		for k, v := range invalidPatterns {
			delete(denormalizedGlobalPatterns, k)
			l.Errorf("load pattern '%s', err: '%s'", k, v)
		}
	}

	// 替换 ppl runtime 中的 patterns
	runtime.DenormalizedGlobalPatterns = denormalizedGlobalPatterns

	return nil
}
