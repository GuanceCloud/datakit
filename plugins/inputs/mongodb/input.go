package mongodb

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gopkg.in/mgo.v2"
)

var (
	defInterval      = datakit.Duration{Duration: 10 * time.Second}
	defMongoUrl      = "mongodb://127.0.0.1:27017"
	defTlsCaCert     = datakit.InstallDir + "/conf.d/mongodb/ssh/ca.pem"
	defTlsCert       = datakit.InstallDir + "/conf.d/mongodb/ssh/cert.pem"
	defTlsCertKey    = datakit.InstallDir + "/conf.d/mongodb/ssh/key.pem"
	defMongodLogPath = "/var/log/mongodb/mongod.log"
	defPipeline      = "mongod.p"
	defTags          map[string]string
)

var (
	inputName    = "mongodb"
	sampleConfig = `
[[inputs.mongodb]]
  ## Gathering interval
  # interval = "` + defInterval.UnitString(time.Second) + `"

  ## An array of URLs of the form:
  ##   "mongodb://" [user ":" pass "@"] host [ ":" port]
  ## For example:
  ##   mongodb://user:auth_key@10.10.3.30:27017,
  ##   mongodb://10.10.3.33:18832,
  # servers = ["` + defMongoUrl + `"]

  ## When true, collect replica set stats
  # gather_replica_set_stats = false

  ## When true, collect cluster stats
  ## Note that the query that counts jumbo chunks triggers a COLLSCAN, which may have an impact on performance.
  # gather_cluster_stats = false

  ## When true, collect per database stats
  # gather_per_db_stats = true

  ## When true, collect per collection stats
  # gather_per_col_stats = true

  ## List of db where collections stats are collected, If empty, all dbs are concerned.
  # col_stats_dbs = []

  ## When true, collect top command stats.
  # gather_top_stat = true

  ## Optional TLS Config, enabled if true.
  # enable_tls = false

	## Optional local Mongod log input config, enabled if true.
	# enable_mongod_log = false

  ## TLS connection config
  [inputs.mongodb.tlsconf]
    # ca_certs = ["` + defTlsCaCert + `"]
    # cert = "` + defTlsCert + `"
    # cert_key = "` + defTlsCertKey + `"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false
    # server_name = ""

	## MongoD log
	[inputs.mongodb.log]
		## Log file path check your mongodb config path usually under '/var/log/mongodb/mongod.log'.
		# files = ["` + defMongodLogPath + `"]
		## Grok pipeline script file.
		# pipeline = "` + defPipeline + `"

  ## Customer tags, if set will be seen with every metric.
  [inputs.mongodb.tags]
    # "key1" = "value1"
    # "key2" = "value2"
`
	piplelineConfig = `
	json_all()
	rename("component", c)
	rename("severity", s)
	rename("context", ctx)
	rename("date", ` + "`t.$date`)" + `
	drop_key(id)
`
	l = logger.SLogger(inputName)
)

type Input struct {
	Interval              datakit.Duration       `toml:"interval"`
	Servers               []string               `toml:"servers"`
	GatherReplicaSetStats bool                   `toml:"gather_replica_set_stats"`
	GatherClusterStats    bool                   `toml:"gather_cluster_stats"`
	GatherPerDbStats      bool                   `toml:"gather_per_db_stats"`
	GatherPerColStats     bool                   `toml:"gather_per_col_stats"`
	ColStatsDbs           []string               `toml:"col_stats_dbs"`
	GatherTopStat         bool                   `toml:"gather_top_stat"`
	EnableTls             bool                   `toml:"enable_tls"`
	TlsConf               *dknet.TlsClientConfig `toml:"tlsconf"`
	EnableMongodLog       bool                   `toml:"enable_mongod_log"`
	Log                   *inputs.TailerOption   `toml:"log"`
	Tags                  map[string]string      `toml:"tags"`
	mongos                map[string]*Server
	tailer                *inputs.Tailer
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{"mongod": piplelineConfig}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mongodbMeasurement{},
		&mongodbDbMeasurement{},
		&mongodbColMeasurement{},
		&mongodbShardMeasurement{},
		&mongodbTopMeasurement{},
	}
}

func (m *Input) Run() {
	l.Info("mongodb input started")

	if m.EnableMongodLog && m.Log != nil && len(m.Log.Files) != 0 {
		l.Info("mongod_log input started")

		go func() {
			inputs.JoinPipelinePath(m.Log, m.Log.Pipeline)
			m.Log.Source = "mongod_log"
			m.Log.Tags = make(map[string]string)
			for k, v := range m.Tags {
				m.Log.Tags[k] = v
			}

			var err error
			if m.tailer, err = inputs.NewTailer(m.Log); err != nil {
				l.Errorf("init tailf err:%s", err.Error())
			} else {
				m.tailer.Run()
			}
		}()
	}

	defTags = m.Tags

	tick := time.NewTicker(m.Interval.Duration)
	for {
		select {
		case <-tick.C:
			if err := m.gather(); err != nil {
				l.Error(err.Error())
				io.FeedLastError(inputName, err.Error())
				continue
			}
		case <-datakit.Exit.Wait():
			l.Info("mongodb input exits")

			return
		}
	}
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (m *Input) gather() error {
	if len(m.Servers) == 0 {
		m.gatherServer(m.getMongoServer(&url.URL{Host: defMongoUrl}))

		return nil
	}

	var wg sync.WaitGroup
	for i, serv := range m.Servers {
		if !strings.HasPrefix(serv, "mongodb://") {
			serv = "mongodb://" + serv
			l.Warnf("Using %q as connection URL; please update your configuration to use an URL", serv)
			m.Servers[i] = serv
		}

		u, err := url.Parse(serv)
		if err != nil {
			l.Errorf("Unable to parse address %q: %s", serv, err.Error())
			continue
		}
		if u.Host == "" {
			l.Errorf("Unable to parse address %q", serv)
			continue
		}

		wg.Add(1)
		go func(srv *Server) {
			defer wg.Done()

			if err := m.gatherServer(srv); err != nil {
				l.Errorf("Error in plugin: %s,%v", srv.URL.String(), err)
			}
		}(m.getMongoServer(u))
	}
	wg.Wait()

	return nil
}

func (m *Input) getMongoServer(url *url.URL) *Server {
	if _, ok := m.mongos[url.Host]; !ok {
		m.mongos[url.Host] = &Server{URL: url}
	}

	return m.mongos[url.Host]
}

func (m *Input) gatherServer(server *Server) error {
	if server.Session == nil {
		var dialAddrs []string
		if server.URL.User != nil {
			dialAddrs = []string{server.URL.String()}
		} else {
			dialAddrs = []string{server.URL.Host}
		}

		dialInfo, err := mgo.ParseURL(dialAddrs[0])
		if err != nil {
			return fmt.Errorf("unable to parse URL %q: %s", dialAddrs[0], err.Error())
		}

		if m.EnableTls && m.TlsConf != nil {
			if tlsConfig, err := m.TlsConf.TlsConfig(); err != nil {
				return err
			} else {
				dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
					return tls.Dial("tcp", addr.String(), tlsConfig)
				}
			}
		}

		dialInfo.Direct = true
		dialInfo.Timeout = 5 * time.Second

		sess, err := mgo.DialWithInfo(dialInfo)
		if err != nil {
			return fmt.Errorf("unable to connect to MongoDB: %s", err.Error())
		}
		server.Session = sess
	}

	return server.gatherData(m.GatherReplicaSetStats, m.GatherClusterStats, m.GatherPerDbStats, m.GatherPerColStats, m.ColStatsDbs, m.GatherTopStat)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval:              defInterval,
			Servers:               []string{defMongoUrl},
			GatherReplicaSetStats: false,
			GatherClusterStats:    false,
			GatherPerDbStats:      true,
			GatherPerColStats:     true,
			ColStatsDbs:           []string{},
			GatherTopStat:         true,
			EnableTls:             false,
			EnableMongodLog:       false,
			Log:                   &inputs.TailerOption{Files: []string{defMongodLogPath}, Pipeline: defPipeline},
			mongos:                make(map[string]*Server),
		}
	})
}
