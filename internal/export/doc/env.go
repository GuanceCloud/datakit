// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package doc be doc helper.
package doc

import (
	"strings"
	"unicode"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	Int          = "Int"
	Float        = "Float"
	String       = "String"
	Map          = "Map"
	Boolean      = "Boolean"
	URL          = "URL"
	JSON         = "JSON"
	List         = "List"
	TimeDuration = "TimeDuration"
)

// Help add default info.
// nolint:lll
var defaultInfos = map[string]*inputs.ENVInfo{
	"Interval": {Type: TimeDuration, Default: "10s", Desc: "Collect interval", DescZh: "采集器重复间隔时长"},
	"Tags":     {Type: Map, Example: `tag1=value1,tag2=value2`, Desc: "Customize tags. If there is a tag with the same name in the configuration file, it will be overwritten", DescZh: "自定义标签。如果配置文件有同名标签，将会覆盖它"},
	"Election": {Type: Boolean, Default: "true", Desc: "Enable election", DescZh: "开启选举"},
	"Timeout":  {Type: TimeDuration, Default: "30s", Desc: "Timeout", DescZh: "超时时长"},
}

func SetENVDoc(prefix string, infos []*inputs.ENVInfo) []*inputs.ENVInfo {
	for _, info := range infos {
		setENVAndToml(prefix, info)

		// Try add default info
		defaultInfo, ok := defaultInfos[info.FieldName]
		if ok {
			if info.Type == "" {
				info.Type = defaultInfo.Type
			}
			if info.Default == "" {
				info.Default = defaultInfo.Default
			}
			if info.Example == "" {
				info.Example = defaultInfo.Example
			}
			if info.Desc == "" {
				info.Desc = defaultInfo.Desc
			}
			if info.DescZh == "" {
				info.DescZh = defaultInfo.DescZh
			}
		}
	}

	return infos
}

func setENVAndToml(prefix string, info *inputs.ENVInfo) {
	// "TagsIgnore" --> "TAGS_IGNORE"
	if info.ENVName == "" {
		info.ENVName = camelToSnake(info.FieldName)
	}

	// "TAGS_IGNORE" --> "tags_ignore"
	if info.ConfField == "" {
		info.ConfField = strings.ToLower(info.ENVName)
	}

	// --> "ENV_INPUT_NEO4J_TAGS_IGNORE"
	info.ENVName = prefix + info.ENVName
}

// Help FieldName --> ENVName automatic.
var abbreviation = map[string]string{
	"HTTPS": "Https",
	"HTTP":  "Http",
	"HTML":  "Html",
	"TCP":   "Tcp",
	"UDP":   "Udp",
	"IP":    "Ip",
	"TLS":   "Tls",
	"URL":   "Url",
	"ID":    "Id",
	"KV":    "Kv",
	"CPU":   "Cpu",
	"GPU":   "Gpu",
}

func camelToSnake(s string) string {
	for k, v := range abbreviation {
		s = strings.ReplaceAll(s, k, v)
	}
	var output []rune
	for i, r := range s {
		if i == 0 {
			output = append(output, unicode.ToUpper(r))
		} else {
			if unicode.IsUpper(r) {
				output = append(output, '_')
			}

			output = append(output, unicode.ToUpper(r))
		}
	}
	return string(output)
}
