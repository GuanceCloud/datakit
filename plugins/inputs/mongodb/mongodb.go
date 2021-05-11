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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gopkg.in/mgo.v2"
)

var (
	inputName    = "mongodb"
	sampleConfig = `
[[inputs.mongodb]]
	## gather interval
	interval = "10s"
	## An array of URLs of the form:
	##   "mongodb://" [user ":" pass "@"] host [ ":" port]
	## For example:
	##   mongodb://user:auth_key@10.10.3.30:27017,
	##   mongodb://10.10.3.33:18832,
	servers = ["mongodb://127.0.0.1:27017"]
	## When true, collect replica set stats
	gather_replica_set_stats = false
	## When true, collect cluster stats
	## Note that the query that counts jumbo chunks triggers a COLLSCAN, which may have an impact on performance.
	gather_cluster_stats = true
	## When true, collect per database stats
	gather_per_db_stats = true
	## When true, collect per collection stats
	gather_per_col_stats = true
	## List of db where collections stats are collected, If empty, all db are concerned
	col_stats_dbs = ["local"]
	## When true, collect top stats
	gather_top_stat = true
	## Optional TLS Config, enabled if true
	enable_tls = false
	[inputs.mongodb.tlsconf]
		# ca_certs = ["/etc/datakit/ca.pem"]
		# cert = "/etc/datakit/cert.pem"
		# cert_key = "/etc/datakit/key.pem"
		## Use TLS but skip chain & host verification
		# insecure_skip_verify = false
		# server_name = ""
`
	localhost = &url.URL{Host: "mongodb://127.0.0.1:27017"}
	l         = logger.SLogger(inputName)
)

type Input struct {
	Interval              datakit.Duration `toml:"interval"`
	Servers               []string         `toml:"servers"`
	GatherReplicaSetStats bool             `toml:"gather_replica_set_stats"`
	GatherClusterStats    bool             `toml:"gather_cluster_stats"`
	GatherPerDbStats      bool             `toml:"gather_per_db_stats"`
	GatherPerColStats     bool             `toml:"gather_per_col_stats"`
	ColStatsDbs           []string         `toml:"col_stats_dbs"`
	GatherTopStat         bool             `toml:"gather_top_stat"`
	EnableTls             bool             `toml:"enable_tls"`
	TlsConf               *TlsClientConfig `toml:"tlsconf"`
	mongos                map[string]*Server
}

func (m *Input) Catalog() string {
	return inputName
}

func (m *Input) SampleConfig() string {
	return sampleConfig
}

func (m *Input) SampleMeasurement() []inputs.Measurement {
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

	tick := time.NewTicker(m.Interval.Duration)
	for {
		select {
		case <-tick.C:
			// var lastErr error
			if err := m.Gather(); err != nil {
				// lastErr = err
				l.Errorf(err.Error())
				continue
			}
		case <-datakit.Exit.Wait():
			l.Info("mongodb input exits")

			return
		}
	}
}

func (m *Input) getMongoServer(url *url.URL) *Server {
	if _, ok := m.mongos[url.Host]; !ok {
		m.mongos[url.Host] = &Server{URL: url}
	}

	return m.mongos[url.Host]
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (m *Input) Gather() error {
	if len(m.Servers) == 0 {
		m.gatherServer(m.getMongoServer(localhost))

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
				l.Errorf("Error in plugin: %v", err)
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
		dialInfo.Direct = true
		dialInfo.Timeout = 5 * time.Second

		if m.EnableTls && m.TlsConf != nil {
			if tlsConfig, err := m.TlsConf.TlsConfig(); err != nil {
				return err
			} else {
				dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
					return tls.Dial("tcp", addr.String(), tlsConfig)
				}
			}
		}

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
			Interval:              datakit.Duration{Duration: 10 * time.Second},
			GatherReplicaSetStats: false,
			GatherClusterStats:    true,
			GatherPerDbStats:      true,
			GatherPerColStats:     true,
			ColStatsDbs:           []string{"local"},
			GatherTopStat:         true,
			mongos:                make(map[string]*Server),
		}
	})
}
