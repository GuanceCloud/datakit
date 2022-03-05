package sinkcommon

import (
	client "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

//----------------------------------------------------------------------

type ISinkPoint interface {
	ToPoint() *client.Point
}

type ISink interface {
	GetID() string
	LoadConfig(mConf map[string]interface{}) error
	Write(pts []ISinkPoint) error
	Metrics() map[string]interface{}
}

//----------------------------------------------------------------------

type SinkCreator func() ISink

func AddCreator(creatorID string, creator SinkCreator) {
	if _, ok := SinkImplCreator[creatorID]; ok {
		l.Fatalf("sink %s exist(from datakit)", creatorID)
	}
	SinkImplCreator[creatorID] = creator
}

func AddImpl(sink ISink) {
	SinkImpls = append(SinkImpls, sink)
}

//----------------------------------------------------------------------

const packageName = "sinkcommon"

var (
	SinkImplCreator = make(map[string]SinkCreator)
	SinkImpls       = []ISink{}
	SinkCategoryMap = make(map[string][]ISink)

	l = logger.DefaultSLogger(packageName)
)

func init() {
	l = logger.SLogger(packageName)
}

//----------------------------------------------------------------------
