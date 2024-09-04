// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package manager for managing pipeline scripts
package manager

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

func ScriptName(relation *ScriptRelation, cat point.Category, pt *point.Point, scriptMap map[string]string) (string, bool) {
	if pt == nil {
		return "", false
	}

	var scriptName string

	// built-in rules last
	switch cat { //nolint:exhaustive
	case point.RUM:
		scriptName = _rumSName(pt)
	case point.Security:
		scriptName = _securitySName(pt)
	case point.Tracing, point.Profiling:
		scriptName = _apmSName(pt)
	default:
		scriptName = _defaultCatSName(pt)
	}

	if scriptName == "" {
		return "", false
	}

	// remote relation first
	if relation != nil {
		if sName, ok := relation.Query(cat, scriptName); ok {
			return sName, true
		}
	}

	// config rules second
	if sName, ok := scriptMap[scriptName]; ok {
		switch sName {
		case "-":
			return "", false
		case "":
		default:
			return sName, true
		}
	}

	// built-in rule last
	return scriptName + ".p", true
}
