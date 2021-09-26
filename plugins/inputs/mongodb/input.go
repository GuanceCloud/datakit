package mongodb

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gopkg.in/mgo.v2"
)

var _ inputs.ElectionInput = (*Input)(nil)

var (
	defInterval      = datakit.Duration{Duration: 10 * time.Second}
	defMongoUrl      = "mongodb://127.0.0.1:27017"
	defTlsCaCert     = "/etc/ssl/certs/mongod.cert.pem"
	defTlsCert       = "/etc/ssl/certs/mongo.cert.pem"
	defTlsCertKey    = "/etc/ssl/certs/mongo.key.pem"
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
    # insecure_skip_verify = true
    # server_name = ""

  ## Mongod log
  # [inputs.mongodb.log]
  # #Log file path check your mongodb config path usually under '/var/log/mongodb/mongod.log'.
  # files = ["` + defMongodLogPath + `"]
  # #Grok pipeline script file.
  # pipeline = "` + defPipeline + `"

  ## Customer tags, if set will be seen with every metric.
  [inputs.mongodb.tags]
    # "key1" = "value1"
    # "key2" = "value2"
		# ...
`
	pipelineConfig = `
  json(_, t, "tmp")
  json(tmp, ` + "`" + "$date" + "`" + `, "time")
  json(_, s, "status")
  json(_, c, "component")
  json(_, msg, "msg")
  json(_, ctx, "context")
  drop_key(tmp)
  default_time(time)
`
	l = logger.DefaultSLogger(inputName)
)

type mongodblog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

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
	TlsConf               *dknet.TLSClientConfig `toml:"tlsconf"`
	EnableMongodLog       bool                   `toml:"enable_mongod_log"`
	Log                   *mongodblog            `toml:"log"`
	Tags                  map[string]string      `toml:"tags"`

	mongos  map[string]*Server
	tail    *tailer.Tailer
	pause   bool
	pauseCh chan bool
}

func (*Input) Catalog() string { return inputName }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) PipelineConfig() map[string]string { return map[string]string{"mongod": pipelineConfig} }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mongodbMeasurement{},
		&mongodbDbMeasurement{},
		&mongodbColMeasurement{},
		&mongodbShardMeasurement{},
		&mongodbTopMeasurement{},
	}
}

func (m *Input) RunPipeline() {
	if !m.EnableMongodLog || m.Log == nil || len(m.Log.Files) == 0 {
		return
	}

	if m.Log.Pipeline == "" {
		m.Log.Pipeline = "mongod.p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		GlobalTags:        m.Tags,
		IgnoreStatus:      m.Log.IgnoreStatus,
		CharacterEncoding: m.Log.CharacterEncoding,
		MultilineMatch:    m.Log.MultilineMatch,
	}

	pl := filepath.Join(datakit.PipelineDir, m.Log.Pipeline)
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	var err error
	m.tail, err = tailer.NewTailer(m.Log.Files, opt)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	go m.tail.Start()
}

func (m *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("mongodb input started")

	defTags = m.Tags

	tick := time.NewTicker(m.Interval.Duration)
	for {
		select {
		case <-datakit.Exit.Wait():
			if m.tail != nil {
				m.tail.Close()
				l.Info("mongodb log exits")
			}
			l.Info("mongodb input exits")
			return

		case <-tick.C:
			if m.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			if err := m.gather(); err != nil {
				l.Error(err.Error())
				io.FeedLastError(inputName, err.Error())
			}

		case m.pause = <-m.pauseCh:
			// nil
		}
	}
}

func (m *Input) getMongoServer(url *url.URL) *Server {
	if _, ok := m.mongos[url.Host]; !ok {
		m.mongos[url.Host] = &Server{URL: url}
	}

	return m.mongos[url.Host]
}

// Reads stats from all configured servers.
// Returns one of the errors encountered while gather stats (if any).
func (m *Input) gather() error {
	if len(m.Servers) == 0 {
		return m.gatherServer(m.getMongoServer(&url.URL{Host: defMongoUrl}))
	}

	var wg sync.WaitGroup
	for i, serv := range m.Servers {
		if !strings.HasPrefix(serv, "mongodb://") {
			serv = "mongodb://" + serv
			l.Warnf("using %q as connection URL; please update your configuration to use an URL", serv)
			m.Servers[i] = serv
		}

		u, err := url.Parse(serv)
		if err != nil {
			l.Errorf("unable to parse address %q: %s", serv, err.Error())
			continue
		}
		if u.Host == "" {
			l.Errorf("unable to parse address %q", serv)
			continue
		}

		wg.Add(1)
		go func(srv *Server) {
			defer wg.Done()

			if err := m.gatherServer(srv); err != nil {
				l.Errorf("error in plugin: %s,%v", srv.URL.String(), err)
			}
		}(m.getMongoServer(u))
	}
	wg.Wait()

	return nil
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
			if tlsConfig, err := m.TlsConf.TLSConfig(); err != nil {
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

func (m *Input) Pause() error {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	select {
	case m.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (m *Input) Resume() error {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	select {
	case m.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
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
			TlsConf: &dknet.TLSClientConfig{
				CaCerts:            []string{defTlsCaCert},
				Cert:               defTlsCert,
				CertKey:            defTlsCertKey,
				InsecureSkipVerify: true,
			},
			EnableMongodLog: false,
			Log:             &mongodblog{Files: []string{defMongodLogPath}, Pipeline: defPipeline},
			mongos:          make(map[string]*Server),
			pauseCh:         make(chan bool, 1),
		}
	})
}
