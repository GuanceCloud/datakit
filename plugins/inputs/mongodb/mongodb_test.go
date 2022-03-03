package mongodb

/* test: failed
func TestGatherServerStats(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    false,
		GatherPerDBStats:      false,
		GatherPerColStats:     false,
		ColStatsDBs:           []string{"local"},
		GatherTopStat:         false,
		mongos:                make(map[string]*Server),
	}
	err := input.gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host})
			data.AddDefaultStats()
			data.AddDBStats()
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
		GatherPerDBStats:      false,
		GatherPerColStats:     false,
		ColStatsDBs:           []string{""},
		GatherTopStat:         false,
		mongos:                make(map[string]*Server),
	}
	err := input.gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host})
			data.AddDefaultStats()
			data.AddDBStats()
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

func TestGatherPerDBStats(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    false,
		GatherPerDBStats:      true,
		GatherPerColStats:     false,
		ColStatsDBs:           []string{},
		GatherTopStat:         false,
		mongos:                make(map[string]*Server),
	}
	err := input.gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host})
			// data.AddDefaultStats()
			data.AddDBStats()
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
func TestGathertShard(t *testing.T) {}

func TestGatherCollection(t *testing.T) {
	input := &Input{
		Interval:              datakit.Duration{Duration: 3 * time.Second},
		Servers:               []string{"mongodb://127.0.0.1:27017"},
		GatherReplicaSetStats: false,
		GatherClusterStats:    false,
		GatherPerDBStats:      false,
		GatherPerColStats:     true,
		ColStatsDBs:           []string{"admin", "local", "config"},
		GatherTopStat:         false,
		mongos:                make(map[string]*Server),
	}
	err := input.gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host})
			// data.AddDefaultStats()
			data.AddDBStats()
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
		GatherClusterStats:    true,
		GatherPerDBStats:      false,
		GatherPerColStats:     false,
		ColStatsDBs:           []string{"admin", "local", "config"},
		GatherTopStat:         true,
		mongos:                make(map[string]*Server),
	}
	err := input.gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host})
			// data.AddDefaultStats()
			data.AddDBStats()
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

func TestTlsConnectCollect(t *testing.T) {
	input := &Input{
		Interval: datakit.Duration{Duration: 3 * time.Second},
		// Servers:               []string{"mongodb://127.0.0.1:27017"},
		Servers:               []string{"mongodb://10.200.7.21:27017"},
		GatherReplicaSetStats: true,
		GatherClusterStats:    true,
		GatherPerDBStats:      true,
		GatherPerColStats:     true,
		ColStatsDBs:           []string{""},
		GatherTopStat:         true,
		TLSClientConfig: net.TLSClientConfig{
			CaCerts:            []string{"/etc/ssl/certs/mongod.cert.pem"},
			Cert:               "/etc/ssl/certs/mongo.pem",
			CertKey:            "/etc/ssl/certs/mongo.key.pem",
			InsecureSkipVerify: true,
			ServerName:         "",
		},
		mongos: make(map[string]*Server),
	}
	err := input.gather()
	if err != nil {
		l.Panic(err.Error())
	}

	for _, srv := range input.mongos {
		if srv.lastResult != nil {
			data := NewMongodbData(NewStatLine(*srv.lastResult, *srv.lastResult, srv.URL.Host, true, 1), map[string]string{"hostname": srv.URL.Host})
			data.AddDefaultStats()
			data.AddDBStats()
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
} */
