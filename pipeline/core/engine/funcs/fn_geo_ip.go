// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb"
)

var ipdbInstance ipdb.IPdb

var geoDefaultVal = "unknown"

func Geo(ip string) (*ipdb.IPdbRecord, error) {
	if ipdbInstance != nil {
		return ipdbInstance.Geo(ip)
	} else {
		return &ipdb.IPdbRecord{}, nil
	}
}

func InitIPdb(instance ipdb.IPdb) {
	ipdbInstance = instance
}

func GeoIPChecking(ng *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	return nil
}

func GeoIP(ng *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	if len(funcExpr.Param) != 1 {
		return fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	cont, err := ng.GetKeyConv2Str(key)
	if err != nil {
		l.Debugf("key `%v' not exist, ignored", key)
		return nil //nolint:nilerr
	}

	if dic, err := GeoIPHandle(cont); err != nil {
		l.Debugf("GeoIPHandle: %s, ignored", err)
		return err
	} else {
		for k, v := range dic {
			if err := ng.AddKey2PtWithVal(k, v, ast.String, runtime.KindPtDefault); err != nil {
				l.Debug(err)
				return nil
			}
		}
	}

	return nil
}
