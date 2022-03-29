// Package sinklogstash contains Logstash sink implement
package sinklogstash

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

const (
	creatorID = "logstash"
)

var (
	_             sinkcommon.ISink = new(SinkLogstash)
	initSucceeded                  = false
)

type SinkLogstash struct {
	ID string // sink config identity, unique, automatically generated.
}

func (s *SinkLogstash) LoadConfig(mConf map[string]interface{}) error {
	//

	initSucceeded = true
	sinkcommon.AddImpl(s)
	return nil
}

func (s *SinkLogstash) Write(pts []sinkcommon.ISinkPoint) error {
	if !initSucceeded {
		return fmt.Errorf("not_init")
	}

	return s.writeLogstash(pts)
}

func (s *SinkLogstash) GetInfo() *sinkcommon.SinkInfo {
	return &sinkcommon.SinkInfo{
		ID:         s.ID,
		CreateID:   creatorID,
		Categories: []string{datakit.SinkCategoryLogging},
	}
}

func (s *SinkLogstash) writeLogstash(pts []sinkcommon.ISinkPoint) error {
	return nil
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(creatorID, func() sinkcommon.ISink {
		return &SinkLogstash{}
	})
}
