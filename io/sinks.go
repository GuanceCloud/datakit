package io

import (
	"fmt"
	"strings"
	"sync"
)

//----------------------------------------------------------------------

type ISink interface {
	GetID() string
	LoadConfig(mConf map[string]interface{}) error
	Write(pts []*Point) error
	Metrics() map[string]interface{}
}

type SinkCreator func() ISink

func AddCreator(creatorID string, creator SinkCreator) {
	if _, ok := SinkImplCreator[creatorID]; ok {
		l.Fatalf("sinks %s exist(from datakit)", creatorID)
	}
	SinkImplCreator[creatorID] = creator
}

func AddImpl(sink ISink) {
	SinkImpls = append(SinkImpls, sink)
}

//----------------------------------------------------------------------

type SinkImplStruct struct {
	ID        string
	ISinkImpl ISink
}

var (
	SinkImplCreator = make(map[string]SinkCreator)
	SinkImpls       = []ISink{}
	SinkCategoryMap = make(map[string][]ISink)

	onceInit sync.Once
	// sinkConfig []map[string]interface{}
	isInitSucceeded bool
)

/*
type singleton struct{}

var ins *singleton
var once sync.Once

func GetIns() *singleton {
	once.Do(func() {
		ins = &singleton{}
	})
	return ins
}
*/

//----------------------------------------------------------------------

func Write(category string, pts []*Point) error {
	if !isInitSucceeded {
		return fmt.Errorf("not inited")
	}

	if impls, ok := SinkCategoryMap[category]; ok {
		var errKeep error
		for _, v := range impls {
			if err := v.Write(pts); err != nil {
				errKeep = err
			}
		}
		return errKeep
	} else {
		// default
		//
	}

	return fmt.Errorf("unsupport category")
}

func InitSink(sincfg []map[string]interface{}) error {
	var err error
	onceInit.Do(func() {
		err = func() error {
			if isInitSucceeded {
				return fmt.Errorf("init twice")
			}

			// check sinks config
			if err := checkSinksConfig(sincfg); err != nil {
				return err
			}

			log.Debugf("SinkImplCreator = %#v", SinkImplCreator)

			if err := buildSinkImpls(sincfg); err != nil {
				return err
			}

			log.Debugf("SinkImpls = %#v", SinkImpls)

			if err := aggregationCategorys(sincfg); err != nil {
				return err
			}

			log.Debugf("SinkCategoryMap = %v", SinkCategoryMap)

			isInitSucceeded = true
			return nil
		}()
	})
	return err
}

func aggregationCategorys(sincfg []map[string]interface{}) error {
	for _, v := range sincfg {
		categories := v["categories"]
		if categoriesArray, ok := categories.([]string); ok {
			mCategory := make(map[string]struct{})
			for _, category := range categoriesArray {
				mCategory[category] = struct{}{}
			}

			for category := range mCategory {
				for _, impl := range SinkImpls {
					id := v["id"].(string)
					if id == impl.GetID() {
						SinkCategoryMap[category] = append(SinkCategoryMap[category], impl)
					}
				}
			}

		} else {
			return fmt.Errorf("invalid categories")
		}
	}
	return nil
}

func buildSinkImpls(sincfg []map[string]interface{}) error {
	for _, v := range sincfg {
		target := v["target"].(string)
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

func getSinkInstanceFromTarget(target string) ISink {
	for k, v := range SinkImplCreator {
		if k == target {
			return v()
		}
	}
	return nil
}

func checkSinksConfig(sincfg []map[string]interface{}) error {
	// check id unique
	mSinkID := make(map[string]struct{})
	for _, v := range sincfg {
		id := v["id"].(string)
		idNew := strings.TrimSpace(id)
		if idNew == "" {
			return fmt.Errorf("invalid id: empty")
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
