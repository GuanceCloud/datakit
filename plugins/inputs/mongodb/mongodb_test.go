package mongodb

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestGatherServerStats(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    false,
		GatherPerDbStats:      false,
		GatherPerColStats:     false,
		ColStatsDbs:           []string{"local"},
		GatherTopStat:         false,
		EnableTls:             false,
		mongos:                make(map[string]*Server),
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

func TestGatherCluster(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    true,
		GatherPerDbStats:      false,
		GatherPerColStats:     false,
		ColStatsDbs:           []string{""},
		GatherTopStat:         false,
		EnableTls:             false,
		mongos:                make(map[string]*Server),
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

func TestGatherPerDbStats(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    false,
		GatherPerDbStats:      true,
		GatherPerColStats:     false,
		ColStatsDbs:           []string{},
		GatherTopStat:         false,
		EnableTls:             false,
		mongos:                make(map[string]*Server),
	}
	err := input.Gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host}, 1)
			// data.AddDefaultStats()
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

// TODO: add testing gathering sharded conn pool
func TestGathertShard(t *testing.T) {
}

func TestGatherCollection(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    false,
		GatherPerDbStats:      false,
		GatherPerColStats:     true,
		ColStatsDbs:           []string{"admin", "local", "config"},
		GatherTopStat:         false,
		EnableTls:             false,
		mongos:                make(map[string]*Server),
	}
	err := input.Gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host}, 1)
			// data.AddDefaultStats()
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

func TestGatherTop(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    false,
		GatherPerDbStats:      false,
		GatherPerColStats:     false,
		ColStatsDbs:           []string{"admin", "local", "config"},
		GatherTopStat:         true,
		EnableTls:             false,
		mongos:                make(map[string]*Server),
	}
	err := input.Gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host}, 1)
			// data.AddDefaultStats()
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

// TODO: add tls dial and connection test
func TestTlsConnectCollect(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    false,
		GatherPerDbStats:      false,
		GatherPerColStats:     false,
		ColStatsDbs:           []string{"admin"},
		GatherTopStat:         false,
		EnableTls:             true,
		TlsConf: &TlsClientConfig{
			CaCerts:            []string{},
			Cert:               "",
			CertKey:            "",
			InsecureSkipVerify: false,
			ServerName:         "",
		},
		mongos: make(map[string]*Server),
	}
	err := input.Gather()
	if err != nil {
		l.Panic(err.Error())
	}
}
