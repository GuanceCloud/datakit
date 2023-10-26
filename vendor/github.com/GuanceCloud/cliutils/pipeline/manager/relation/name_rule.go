// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package relation

import (
	"github.com/GuanceCloud/cliutils/point"
)

func _rumSName(pt *point.Point) string {
	if id := pt.Get("app_id"); id != nil {
		if rumID, ok := id.(string); ok {
			return rumID + "_" + pt.Name()
		}
	}
	return ""
}

func _securitySName(pt *point.Point) string {
	if scheckCat := pt.Get("category"); scheckCat != nil {
		if cat, ok := scheckCat.(string); ok {
			return cat
		}
	}
	return ""
}

func _apmSName(pt *point.Point) string {
	if apmSvc := pt.Get("service"); apmSvc != nil {
		if svc, ok := apmSvc.(string); ok {
			return svc
		}
	}
	return ""
}

func _defaultCatSName(pt *point.Point) string {
	return pt.Name()
}
