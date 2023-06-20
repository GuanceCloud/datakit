// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"

	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

func _rumSName(pt ptinput.PlInputPt) string {
	if id, _, err := pt.Get("app_id"); err == nil && id != nil {
		if rumID, ok := id.(string); ok {
			return rumID + "_" + pt.GetPtName()
		}
	}
	return ""
}

func _securitySName(pt ptinput.PlInputPt) string {
	if scheckCat, _, err := pt.Get("category"); err == nil && scheckCat != nil {
		if cat, ok := scheckCat.(string); ok {
			return cat
		}
	}
	return ""
}

func _apmSName(pt ptinput.PlInputPt) string {
	if apmSvc, _, err := pt.Get("service"); err == nil && apmSvc != nil {
		if svc, ok := apmSvc.(string); ok {
			return svc
		}
	}
	return ""
}

func _defaultCatSName(pt *dkpt.Point) string {
	return pt.Name()
}
