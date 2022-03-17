// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"

func JSONAllChecking(_ *parser.EngineData, _ parser.Node) error {
	l.Warnf("warning: json_all() is disabled")
	return nil
}

func JSONAll(_ *parser.EngineData, _ parser.Node) error {
	l.Warnf("warning: json_all() is disabled")
	return nil
}
