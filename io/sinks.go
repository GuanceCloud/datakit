package io

import (
	"fmt"
	"sync"
)

//----------------------------------------------------------------------

type ISink interface {
	LoadConfig(mConf map[string]interface{}) error
	Write(pts []*Point) error
	Metrics() map[string]interface{}
}

type SinkCreator func() ISink

func Add(name string, creator SinkCreator) {
	if _, ok := SinkImpls[name]; ok {
		l.Fatalf("sinks %s exist(from datakit)", name)
	}

	SinkImpls[name] = creator
}

//----------------------------------------------------------------------

const (
	sinksPackageName = "sinks"
)

var (
	SinkImpls           = map[string]SinkCreator{}
	ErrorNotImplemented = fmt.Errorf("not implemented yet")

	onceInit   sync.Once
	sinkConfig []map[string]interface{}
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
	if len(sinkConfig) == 0 {
		return fmt.Errorf("not inited")
	}
	switch category {
	// case SinkCategoryLog:
	// case SinkCategoryMetric:
	// case SinkCategoryTrace:
	}

	return ErrorNotImplemented
}

func InitSink(sincfg []map[string]interface{}) error {
	onceInit.Do(func() {
		sinkConfig = sincfg
	})
	return nil
}

//----------------------------------------------------------------------
