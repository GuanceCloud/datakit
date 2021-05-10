package mongodb

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestGatherMongoDb(t *testing.T) {
	input := &Input{
		Interval:               datakit.Duration{Duration: 3 * time.Second},
		Servers:                []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStatus: false,
		GatherClusterStatus:    false,
		GatherPerdbStats:       false,
		GatherColStats:         false,
		GatherTopStat:          false,
		ColStatsDbs:            []string{"local"},
		EnableTls:              false,
		mongos:                 make(map[string]*Server),
	}
	err := input.Gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host}, 1)
			data.AddDefaultStats()
			data.AddDbStats()
			data.AddColStats()
			data.AddShardHostStats()
			data.AddTopStats()
			data.append()
			// data.flush()
			for _, data := range data.collectCache {
				if point, err := data.LineProto(); err != nil {
					l.Error(err.Error())
				} else {
					l.Info(point.String())
				}
			}
		}
	}
}

func TestGatherMongoDbPerDbStat(t *testing.T) {
	input := &Input{
		Interval:               datakit.Duration{Duration: 3 * time.Second},
		Servers:                []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStatus: false,
		GatherClusterStatus:    false,
		GatherPerdbStats:       false,
		GatherColStats:         true,
		GatherTopStat:          false,
		ColStatsDbs:            []string{"local", "config"},
		EnableTls:              false,
		mongos:                 make(map[string]*Server),
	}
	err := input.Gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host}, 1)
			data.AddDefaultStats()
			data.AddDbStats()
			data.AddColStats()
			data.AddShardHostStats()
			data.AddTopStats()
			data.append()
			// data.flush()
			for _, data := range data.collectCache {
				if point, err := data.LineProto(); err != nil {
					l.Error(err.Error())
				} else {
					l.Info(point.String())
				}
			}
		}
	}
}
