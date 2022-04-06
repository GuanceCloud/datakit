// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sink contains sink implement
package sink

import (
	"fmt"
	"strings"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkinfluxdb"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkm3db"
)

//----------------------------------------------------------------------

func Write(category string, pts []sinkcommon.ISinkPoint) error {
	if !isInitSucceeded {
		return fmt.Errorf("not inited")
	}

	if impls, ok := sinkcommon.SinkCategoryMap[category]; ok {
		var errKeep error
		for _, v := range impls {
			if err := v.Write(pts); err != nil {
				errKeep = err
			}
		}
		return errKeep
	} else if defaultCallPtr != nil {
		return defaultCallPtr(category, pts)
	}

	return fmt.Errorf("unsupport category")
}

func Init(sincfg []map[string]interface{}, defCall func(string, []sinkcommon.ISinkPoint) error) error {
	var err error
	onceInit.Do(func() {
		l = logger.SLogger(packageName)

		err = func() error {
			if isInitSucceeded {
				return fmt.Errorf("init twice")
			}

			// check sink config
			if err := checkSinkConfig(sincfg); err != nil {
				return err
			}

			l.Debugf("SinkImplCreator = %#v", sinkcommon.SinkImplCreator)

			if err := buildSinkImpls(sincfg); err != nil {
				return err
			}

			l.Debugf("SinkImpls = %#v", sinkcommon.SinkImpls)

			if err := aggregationCategorys(sincfg); err != nil {
				return err
			}

			l.Debugf("SinkCategoryMap = %#v", sinkcommon.SinkCategoryMap)

			defaultCallPtr = defCall

			isInitSucceeded = true
			return nil
		}()
	})
	return err
}

//----------------------------------------------------------------------

const packageName = "sink"

var (
	l               = logger.DefaultSLogger(packageName)
	onceInit        sync.Once
	isInitSucceeded bool
	defaultCallPtr  func(string, []sinkcommon.ISinkPoint) error
)

func aggregationCategorys(sincfg []map[string]interface{}) error {
	for _, v := range sincfg {
		if len(v) == 0 {
			continue // empty
		}
		val, ok := v["categories"]
		if !ok {
			return fmt.Errorf("invalid categories: not found")
		}
		categoriesArray, ok := val.([]interface{})
		if !ok {
			return fmt.Errorf("invalid categories: not []interface{}: %#v", val)
		}
		if len(categoriesArray) == 0 {
			return fmt.Errorf("invalid categories: empty")
		}

		mCategory := make(map[string]struct{})
		for _, categoryLine := range categoriesArray {
			category, ok := categoryLine.(string)
			if !ok {
				return fmt.Errorf("invalid categories: not string")
			}
			mCategory[category] = struct{}{}
		}

		for category := range mCategory {
			for _, impl := range sinkcommon.SinkImpls {
				id, err := dkstring.GetMapAssertString("id", v)
				if err != nil {
					return err
				}
				if id == impl.GetID() {
					newCategory, err := getMapCategory(category)
					if err != nil {
						return err
					}
					sinkcommon.SinkCategoryMap[newCategory] = append(sinkcommon.SinkCategoryMap[newCategory], impl)
				}
			}
		}
	}
	return nil
}

func getMapCategory(originCategory string) (string, error) {
	var newCategory string

	tmpCategory := dkstring.TrimString(originCategory)
	category := strings.ToUpper(tmpCategory)

	switch category {
	case datakit.SinkCategoryMetric:
		newCategory = datakit.Metric
	case datakit.SinkCategoryNetwork:
		newCategory = datakit.Network
	case datakit.SinkCategoryKeyEvent:
		newCategory = datakit.KeyEvent
	case datakit.SinkCategoryObject:
		newCategory = datakit.Object
	case datakit.SinkCategoryCustomObject:
		newCategory = datakit.CustomObject
	case datakit.SinkCategoryLogging:
		newCategory = datakit.Logging
	case datakit.SinkCategoryTracing:
		newCategory = datakit.Tracing
	case datakit.SinkCategoryRUM:
		newCategory = datakit.RUM
	case datakit.SinkCategorySecurity:
		newCategory = datakit.Security
	default:
		return "", fmt.Errorf("unrecognized category")
	}

	return newCategory, nil
}

func buildSinkImpls(sincfg []map[string]interface{}) error {
	for _, v := range sincfg {
		if len(v) == 0 {
			continue // empty
		}
		target, err := dkstring.GetMapAssertString("target", v)
		if err != nil {
			return err
		}
		if target == datakit.SinkTargetExample {
			continue // ignore example
		}
		if ins := getSinkInstanceFromTarget(target); ins != nil {
			if err := ins.LoadConfig(v); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%s not implemented yet", target)
		}
	}
	return nil
}

func getSinkInstanceFromTarget(target string) sinkcommon.ISink {
	for k, v := range sinkcommon.SinkImplCreator {
		if k == target {
			return v()
		}
	}
	return nil
}

func checkSinkConfig(sincfg []map[string]interface{}) error {
	// check id unique
	mSinkID := make(map[string]struct{})
	for _, v := range sincfg {
		if len(v) == 0 {
			continue // empty
		}
		id, err := dkstring.GetMapAssertString("id", v)
		if err != nil {
			return err
		}
		if _, err := dkstring.CheckNotEmpty(id, "id"); err != nil {
			return err
		}
		if _, ok := mSinkID[id]; ok {
			return fmt.Errorf("invalid sink config: id not unique")
		} else {
			mSinkID[id] = struct{}{}
		}
	}
	return nil
}

//----------------------------------------------------------------------
