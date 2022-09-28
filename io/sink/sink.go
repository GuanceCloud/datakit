// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sink contains sink implement
package sink

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/sinkfuncs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkdataway"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkinfluxdb"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinklogstash"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkm3db"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkotel"
)

//----------------------------------------------------------------------

func Write(category string, pts []*point.Point) (*point.Failed, error) {
	if len(pts) == 0 {
		l.Debug("sink entry empty")
		return nil, nil
	}

	if !isInitSucceeded {
		return nil, fmt.Errorf("not inited")
	}

	if impls, ok := sinkcommon.SinkCategoryMap[category]; ok {
		var errKeep error
		for _, v := range impls {
			if err := v.Write(category, pts); err != nil { // NOTE: sinker send fail not cached
				errKeep = err
			}
		}

		if errKeep != nil {
			l.Errorf("before remain, errKeep = %v", errKeep)
		}

		var remainPoints []*point.Point
		for _, v := range pts {
			if !v.GetWritten() {
				remainPoints = append(remainPoints, v)
			}
		}
		if defaultCallPtr != nil {
			_, errKeep = defaultCallPtr(category, remainPoints)
		}

		if errKeep != nil {
			l.Errorf("after remain, errKeep = %v", errKeep)
		}

		if datakit.LogSinkDetail {
			for _, v := range remainPoints {
				lineStr, _ := v.String()
				l.Infof("(sink_detail) remain point: (%s) (%s)", category, lineStr)
			}
		}

		return nil, errKeep // Note: in sink package, point.Failed always nil
	} else if defaultCallPtr != nil {
		if len(sinkcommon.SinkCategoryMap) == 0 {
			l.Debug("sink empty")
		}

		if datakit.LogSinkDetail {
			for _, v := range pts {
				line, _ := v.String()
				l.Infof("(sink_detail) default point: (%s) (%s)", category, line)
			}
		}

		return defaultCallPtr(category, pts)
	}

	l.Error("should not been here: no default dataway AND no sink.")

	return nil, &sinkcommon.SinkUnsupportError{} // Note: in sink package, point.Failed always nil
}

func Init(sinkcfg []map[string]interface{}, defCall func(string, []*point.Point) (*point.Failed, error)) error {
	var err error
	onceInit.Do(func() {
		l = logger.SLogger(packageName)

		err = func() error {
			if isInitSucceeded {
				return fmt.Errorf("init twice")
			}

			l.Infof("sinkcfg = %#v", sinkcfg)

			// check sink config
			if err := checkSinkConfig(sinkcfg); err != nil {
				return err
			}

			l.Debugf("SinkImplCreator = %#v", sinkcommon.SinkImplCreator)

			if err := buildSinkImpls(sinkcfg); err != nil {
				return err
			}

			l.Debugf("SinkImpls = %#v", sinkcommon.SinkImpls)

			if err := polymerizeCategorys(sinkcfg); err != nil {
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
	defaultCallPtr  func(string, []*point.Point) (*point.Failed, error)
)

func getCategories(val interface{}) (map[string]struct{}, error) {
	mCategory := make(map[string]struct{})

	switch categoriesArray := val.(type) {
	case []string:
		if len(categoriesArray) == 0 {
			return nil, fmt.Errorf("invalid categories: empty")
		}

		for _, categoryLine := range categoriesArray {
			mCategory[categoryLine] = struct{}{}
		}

	case []interface{}:
		if len(categoriesArray) == 0 {
			return nil, fmt.Errorf("invalid categories: empty")
		}

		for _, categoryLine := range categoriesArray {
			category, ok := categoryLine.(string)
			if !ok {
				return nil, fmt.Errorf("invalid categories: not string")
			}
			mCategory[category] = struct{}{}
		}

	default:
		return nil,
			fmt.Errorf("invalid categories: %s, %s: %#v",
				reflect.TypeOf(val).Name(),
				reflect.TypeOf(val).String(),
				val)
	}

	return mCategory, nil
}

func polymerizeCategorys(sinkcfg []map[string]interface{}) error {
	for _, v := range sinkcfg {
		if len(v) == 0 {
			continue // empty
		}
		val, ok := v["categories"]
		if !ok {
			return fmt.Errorf("invalid categories: not found")
		}
		mCategory, err := getCategories(val)
		if err != nil {
			return err
		}
		if len(mCategory) == 0 {
			return fmt.Errorf("categories empty")
		}

		l.Infof("sink impls length = %d, impls = %#v", len(sinkcommon.SinkImpls), sinkcommon.SinkImpls)

		for category := range mCategory {
			for _, impl := range sinkcommon.SinkImpls {
				isMatch, err := checkCategoryMatchImpl(v, category, impl)
				if err != nil {
					l.Errorf("checkCategoryMatchImpl failed: %v", err)
					return err
				}
				if isMatch {
					newCategory, err := getMapCategory(category)
					if err != nil {
						l.Errorf("getMapCategory failed: %v", err)
						return err
					}
					sinkcommon.SinkCategoryMap[newCategory] = append(sinkcommon.SinkCategoryMap[newCategory], impl)
				}
			}
		}
	}
	return nil
}

// checkCategoryMatchImpl returns whether [category, impl] matchable, AND error.
// onesinkcfg, ie, the sink config, is just for MD5 calculation here.
func checkCategoryMatchImpl(
	oneSinkCfg map[string]interface{},
	category string,
	impl sinkcommon.ISink,
) (bool, error) {
	cfgID, cfgOrigin, err := sinkfuncs.GetSinkCreatorID(oneSinkCfg)
	if err != nil {
		return false, err
	}

	// check whether support the category
	found := false
	supportCategories := impl.GetInfo().Categories
	for _, scs := range supportCategories {
		if category == scs {
			found = true
			break
		}
	}
	if !found {
		l.Warnf("%s not support category: %s", impl.GetInfo().CreateID, category)
		return false, nil
	}

	if cfgID != impl.GetInfo().ID {
		l.Infof("sink cfg %s ID not matched, cfgID = %s, implID = %s, cfgOrigin = %s, implOrigin = %s",
			impl.GetInfo().CreateID, cfgID, impl.GetInfo().ID, cfgOrigin, impl.GetInfo().IDStr)
		return false, nil
	}

	l.Infof("sink cfg %s ID matches, ID = %s, cfgOrigin = %s", impl.GetInfo().CreateID, cfgID, cfgOrigin)
	return true, nil
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
	case datakit.SinkCategoryProfiling:
		newCategory = datakit.Profiling
	default:
		return "", fmt.Errorf("unrecognized category")
	}

	return newCategory, nil
}

func buildSinkImpls(sinkcfg []map[string]interface{}) error {
	for _, v := range sinkcfg {
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

func checkSinkConfig(sinkcfg []map[string]interface{}) error {
	// check id unique
	mSinkID := make(map[string]struct{})
	for _, v := range sinkcfg {
		if len(v) == 0 {
			continue // empty
		}
		id, _, err := sinkfuncs.GetSinkCreatorID(v)
		if err != nil {
			return err
		}
		if _, err := dkstring.CheckNotEmpty(id, "sink md5sum"); err != nil {
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
