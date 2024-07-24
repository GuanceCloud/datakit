// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import "strings"

type ENVInfo struct {
	// Is the Env comes from input(collector) plugins or not?
	DocType string

	// We can convert FieldName(Such as DisableHostTag) to ENVName(by specify some prefix
	// such as ENV_INPUT_XXX_) and ConfField:
	//  - ENV_INPUT_XXX_DISABLE_HOST_TAG
	//  - disable_host_tag
	//
	// Most of the time, we just need to set Fieldname, for some reason, if exist ENVName and
	// ConfField not convertable from FieldName, we can specify them manually, such as:
	//
	//   FieldName: "DisableHostTag"
	//   ENVName: "ENV_INPUT_XXX_DISABLE_HOST_TAGS", // with `S' suffix to the env key
	//   ConfField: "disable_host_tags" // also with `s' suffix
	//
	// Or we can convert ENVName(without prefix) to ConfField, such as:
	//
	//   ENVName: "DISABLE_HOST_TAGS",
	//   ConfField: "disable_host_tags"
	//
	// If there is no ConfField exist, we should manually set it to doc.NoField.
	FieldName,
	ENVName,
	ConfField string

	// Env value type
	Type string

	// Env example value
	Example string

	// Default value for the Env
	Default string

	// Is the env required?
	Required string

	// Env description in English and Chinese.
	Desc,
	DescZh string
}

// GetENVSample build docs on ENVs.
func GetENVSample(infos []*ENVInfo, zh bool) string {
	result := ""
	if len(infos) == 0 {
		return ""
	}

	for _, info := range infos {
		s := []string{}
		s = append(s, "- **"+info.ENVName+"**\n\n")

		if zh && info.DescZh != "" {
			s = append(s, "    "+info.DescZh+"\n\n")
		} else if info.Desc != "" {
			s = append(s, "    "+info.Desc+"\n\n")
		}

		if info.Type != "" {
			if zh {
				s = append(s, "    "+"**字段类型**: "+info.Type+"\n\n")
			} else {
				s = append(s, "    "+"**Type**: "+info.Type+"\n\n")
			}
		}

		// info.DocType == "" --> input doc
		if info.DocType == "" && info.ConfField != "" {
			if zh {
				s = append(s, "    "+"**采集器配置字段**: `"+info.ConfField+"`\n\n")
			} else {
				s = append(s, "    "+"**input.conf**: `"+info.ConfField+"`\n\n")
			}
		}

		if info.Example != "" {
			if zh {
				s = append(s, "    "+"**示例**: "+info.Example+"\n\n")
			} else {
				s = append(s, "    "+"**Example**: "+info.Example+"\n\n")
			}
		}

		if info.Default != "" {
			if zh {
				s = append(s, "    "+"**默认值**: "+info.Default+"\n\n")
			} else {
				s = append(s, "    "+"**Default**: "+info.Default+"\n\n")
			}
		}

		if info.Required != "" {
			if zh {
				s = append(s, "    "+"**必填**: "+info.Required+"\n\n")
			} else {
				s = append(s, "    "+"**Required**: "+info.Required+"\n\n")
			}
		}

		for _, v := range s {
			result += v
		}
	}
	result = strings.TrimSuffix(result, "\n\n")

	return result
}
