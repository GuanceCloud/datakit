package io

import (
	"testing"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
)

var (
	debugFilterRules []byte
	debugPoints      []byte
)

type debugLogFilterMock struct{}

func (*debugLogFilterMock) getLogFilter() ([]byte, error) {
	return debugFilterRules, nil
}

func (*debugLogFilterMock) preparePoints(pts []*Point) []*Point {
	influxPts, err := lp.ParsePoints(debugPoints, nil)
	if err != nil {
		l.Error(err)
	}

	newpts := []*Point{}
	for _, pt := range influxPts {
		newpts = append(newpts, &Point{Point: pt})
	}

	return newpts
}

func TestLogFilter(t *testing.T) {
	debugFilterRules = []byte(`{"content": ["{component = 'NETWORK' || context = 'conn1'}", "{a >= 123}"]}`)
	debugPoints = []byte(`
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="NETWORK",context="listener",message="{\"t\":{\"$date\":\"2021-06-17T02:35:48.279+00:00\"},\"s\":\"I\",  \"c\":\"NETWORK\",  \"id\":22943,   \"ctx\":\"listener\",\"msg\":\"Connection accepted\",\"attr\":{\"remote\":\"172.17.0.1:58008\",\"connectionId\":1,\"connectionCount\":1}}",msg="Connection accepted",status="info" 1623897348290344000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="COMMAND",context="conn1",message="{\"t\":{\"$date\":\"2021-06-17T02:35:48.297+00:00\"},\"s\":\"W\",  \"c\":\"COMMAND\",  \"id\":4718707, \"ctx\":\"conn1\",\"msg\":\"The shardConnPoolStats command is deprecated. Use instead the connPoolStats command.\"}",msg="The shardConnPoolStats command is deprecated. Use instead the connPoolStats command.",status="warning" 1623897378234503000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="NETWORK",context="listener",message="{\"t\":{\"$date\":\"2021-06-17T02:36:18.242+00:00\"},\"s\":\"I\",  \"c\":\"NETWORK\",  \"id\":22943,   \"ctx\":\"listener\",\"msg\":\"Connection accepted\",\"attr\":{\"remote\":\"172.17.0.1:58012\",\"connectionId\":2,\"connectionCount\":2}}",msg="Connection accepted",status="info" 1623897386561770000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="STORAGE",context="WTCheckpointThread",message="{\"t\":{\"$date\":\"2021-06-17T02:36:26.579+00:00\"},\"s\":\"I\",  \"c\":\"STORAGE\",  \"id\":22430,   \"ctx\":\"WTCheckpointThread\",\"msg\":\"WiredTiger message\",\"attr\":{\"message\":\"[1623897386:579717][1:0x7fb2db747700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 15, snapshot max: 15 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)\"}}",msg="WiredTiger message",status="info" 1623897446578633000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="STORAGE",context="WTCheckpointThread",message="{\"t\":{\"$date\":\"2021-06-17T02:37:26.596+00:00\"},\"s\":\"I\",  \"c\":\"STORAGE\",  \"id\":22430,   \"ctx\":\"WTCheckpointThread\",\"msg\":\"WiredTiger message\",\"attr\":{\"message\":\"[1623897446:596429][1:0x7fb2db747700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 16, snapshot max: 16 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)\"}}",msg="WiredTiger message",status="info" 1623897506592459000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb a=123i,component="STORAGE",context="WTCheckpointThread",message="{\"t\":{\"$date\":\"2021-06-17T02:38:26.610+00:00\"},\"s\":\"I\",  \"c\":\"STORAGE\",  \"id\":22430,   \"ctx\":\"WTCheckpointThread\",\"msg\":\"WiredTiger message\",\"attr\":{\"message\":\"[1623897506:610362][1:0x7fb2db747700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 17, snapshot max: 17 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)\"}}",msg="WiredTiger message",status="info" 1623897566610844000
	`)

	time.Sleep(3 * time.Second)

	switch defLogfilter.status {
	case filterReleased:
		l.Info("log filter released")
	case filterRefreshed:
		l.Info("log filter refreshed")
	default:
		l.Info("log filter status unknow")
	}

	l.Infof("log filter current rules: %q", defLogfilter.rules)

	after := defLogfilter.filter(nil)
	for _, pt := range after {
		l.Info(pt)
	}
}

func TestEmptyRule(t *testing.T) {
	debugFilterRules = []byte(`{"content": ["{component = 'NETWORK'}"]}`)
	debugPoints = []byte(`
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="NETWORK",context="listener",message="{\"t\":{\"$date\":\"2021-06-17T02:35:48.279+00:00\"},\"s\":\"I\",  \"c\":\"NETWORK\",  \"id\":22943,   \"ctx\":\"listener\",\"msg\":\"Connection accepted\",\"attr\":{\"remote\":\"172.17.0.1:58008\",\"connectionId\":1,\"connectionCount\":1}}",msg="Connection accepted",status="info" 1623897348290344000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="COMMAND",context="conn1",message="{\"t\":{\"$date\":\"2021-06-17T02:35:48.297+00:00\"},\"s\":\"W\",  \"c\":\"COMMAND\",  \"id\":4718707, \"ctx\":\"conn1\",\"msg\":\"The shardConnPoolStats command is deprecated. Use instead the connPoolStats command.\"}",msg="The shardConnPoolStats command is deprecated. Use instead the connPoolStats command.",status="warning" 1623897378234503000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="NETWORK",context="listener",message="{\"t\":{\"$date\":\"2021-06-17T02:36:18.242+00:00\"},\"s\":\"I\",  \"c\":\"NETWORK\",  \"id\":22943,   \"ctx\":\"listener\",\"msg\":\"Connection accepted\",\"attr\":{\"remote\":\"172.17.0.1:58012\",\"connectionId\":2,\"connectionCount\":2}}",msg="Connection accepted",status="info" 1623897386561770000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="STORAGE",context="WTCheckpointThread",message="{\"t\":{\"$date\":\"2021-06-17T02:36:26.579+00:00\"},\"s\":\"I\",  \"c\":\"STORAGE\",  \"id\":22430,   \"ctx\":\"WTCheckpointThread\",\"msg\":\"WiredTiger message\",\"attr\":{\"message\":\"[1623897386:579717][1:0x7fb2db747700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 15, snapshot max: 15 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)\"}}",msg="WiredTiger message",status="info" 1623897446578633000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="STORAGE",context="WTCheckpointThread",message="{\"t\":{\"$date\":\"2021-06-17T02:37:26.596+00:00\"},\"s\":\"I\",  \"c\":\"STORAGE\",  \"id\":22430,   \"ctx\":\"WTCheckpointThread\",\"msg\":\"WiredTiger message\",\"attr\":{\"message\":\"[1623897446:596429][1:0x7fb2db747700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 16, snapshot max: 16 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)\"}}",msg="WiredTiger message",status="info" 1623897506592459000
mongodb,filename=mongod.log,host=CodapeWilds-MacBook-Pro.local,service=mongodb component="STORAGE",context="WTCheckpointThread",message="{\"t\":{\"$date\":\"2021-06-17T02:38:26.610+00:00\"},\"s\":\"I\",  \"c\":\"STORAGE\",  \"id\":22430,   \"ctx\":\"WTCheckpointThread\",\"msg\":\"WiredTiger message\",\"attr\":{\"message\":\"[1623897506:610362][1:0x7fb2db747700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 17, snapshot max: 17 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)\"}}",msg="WiredTiger message",status="info" 1623897566610844000
	`)

	time.Sleep(3 * time.Second)

	switch defLogfilter.status {
	case filterReleased:
		l.Info("log filter released")
	case filterRefreshed:
		l.Info("log filter refreshed")
	default:
		l.Info("log filter status unknow")
	}

	l.Debug("log filter current rules: %q", defLogfilter.rules)

	after := defLogfilter.filter(nil)
	for _, pt := range after {
		l.Info(pt)
	}

	debugFilterRules = []byte(`{"content": []}`)
	time.Sleep(3 * time.Second)

	switch defLogfilter.status {
	case filterReleased:
		l.Info("log filter released")
	case filterRefreshed:
		l.Info("log filter refreshed")
	default:
		l.Info("log filter status unknow")
	}

	l.Infof("log filter current rules: %q", defLogfilter.rules)

	after = defLogfilter.filter(nil)
	for _, pt := range after {
		l.Info(pt)
	}
}

func init() { //nolint:gochecknoinits
	defIntervalDefault = 10
	defLogFilterMock = &debugLogFilterMock{}
}
