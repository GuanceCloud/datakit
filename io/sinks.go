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
		log.Fatalf("sink %s exist(from datakit)", creatorID)
	}
	SinkImplCreator[creatorID] = creator
}

func AddImpl(sink ISink) {
	SinkImpls = append(SinkImpls, sink)
}

//----------------------------------------------------------------------

var (
	SinkImplCreator = make(map[string]SinkCreator)
	SinkImpls       = []ISink{}
	SinkCategoryMap = make(map[string][]ISink)

	onceInit        sync.Once
	isInitSucceeded bool
)

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
		categoriesArray, ok := v["categories"].([]string)
		if !ok {
			return fmt.Errorf("invalid categories: not []string")
		}

		mCategory := make(map[string]struct{})
		for _, category := range categoriesArray {
			mCategory[category] = struct{}{}
		}

		for category := range mCategory {
			for _, impl := range SinkImpls {
				id, err := getAssertString("id", v)
				if err != nil {
					return err
				}
				if id == impl.GetID() {
					SinkCategoryMap[category] = append(SinkCategoryMap[category], impl)
				}
			}
		}
	}
	return nil
}

func buildSinkImpls(sincfg []map[string]interface{}) error {
	for _, v := range sincfg {
		target, err := getAssertString("target", v)
		if err != nil {
			return err
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
		id, err := getAssertString("id", v)
		if err != nil {
			return err
		}
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

func getAssertString(name string, mSingle map[string]interface{}) (string, error) {
	str, ok := mSingle[name].(string)
	if !ok {
		return "", getAssertStringError(name)
	}
	return str, nil
}

func getAssertStringError(name string) error {
	return fmt.Errorf("invalid %s: not string", name)
}

//----------------------------------------------------------------------
