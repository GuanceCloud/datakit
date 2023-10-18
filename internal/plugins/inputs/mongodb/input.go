// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mongodb collects MongoDB metrics.
package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var _ inputs.ElectionInput = &Input{}

var (
	catalogName  = "db"
	inputName    = "mongodb"
	sampleConfig = `
[[inputs.mongodb]]
  ## Gathering interval
  interval = "10s"

  ## A list of Mongodb servers URL
  ## Note: must escape special characters in password before connect to Mongodb server, otherwise parse will failed.
  ## Form: "mongodb://[user ":" pass "@"] host [ ":" port]"
  ## Some examples:
  ## mongodb://user:pswd@localhost:27017/?authMechanism=SCRAM-SHA-256&authSource=admin
  ## mongodb://user:pswd@127.0.0.1:27017,
  ## mongodb://10.10.3.33:18832,
  servers = ["mongodb://127.0.0.1:27017"]

  ## When true, collect replica set stats
  gather_replica_set_stats = false

  ## When true, collect cluster stats
  ## Note that the query that counts jumbo chunks triggers a COLLSCAN, which may have an impact on performance.
  gather_cluster_stats = false

  ## When true, collect per database stats
  gather_per_db_stats = true

  ## When true, collect per collection stats
  gather_per_col_stats = true

  ## List of db where collections stats are collected, If empty, all dbs are concerned.
  col_stats_dbs = []

  ## When true, collect top command stats.
  gather_top_stat = true

  ## Set true to enable election
  election = true

  ## TLS connection config
  # ca_certs = ["/etc/ssl/certs/mongod.cert.pem"]
  # cert = "/etc/ssl/certs/mongo.cert.pem"
  # cert_key = "/etc/ssl/certs/mongo.key.pem"
  # insecure_skip_verify = true
  # server_name = ""

  ## Mongodb log files and Grok Pipeline files configuration
  # [inputs.mongodb.log]
    # files = ["/var/log/mongodb/mongod.log"]
    # pipeline = "mongod.p"

  ## Customer tags, if set will be seen with every metric.
  # [inputs.mongodb.tags]
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
)

var (
	log     = logger.DefaultSLogger(inputName)
	defTags map[string]string
)

type mongodblog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type Input struct {
	TLSConf               *dknet.TLSClientConfig `toml:"tlsconf"` // deprecated
	Interval              datakit.Duration       `toml:"interval"`
	Servers               []string               `toml:"servers"`
	GatherReplicaSetStats bool                   `toml:"gather_replica_set_stats"`
	GatherClusterStats    bool                   `toml:"gather_cluster_stats"`
	GatherPerDBStats      bool                   `toml:"gather_per_db_stats"`
	GatherPerColStats     bool                   `toml:"gather_per_col_stats"`
	ColStatsDBs           []string               `toml:"col_stats_dbs"`
	GatherTopStat         bool                   `toml:"gather_top_stat"`
	Election              bool                   `toml:"election"`
	*dknet.TLSClientConfig
	MgoDBLog *mongodblog       `toml:"log"`
	Tags     map[string]string `toml:"tags"`
	mgoSvrs  []*MongodbServer
	tail     *tailer.Tailer
	pause    bool
	pauseCh  chan bool
	semStop  *cliutils.Sem // start stop signal
	feeder   dkio.Feeder
	Tagger   dkpt.GlobalTagger
}

func (*Input) Catalog() string { return catalogName }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mongodbMeasurement{},
		&mongodbDBMeasurement{},
		&mongodbColMeasurement{},
		&mongodbShardMeasurement{},
		&mongodbTopMeasurement{},
	}
}

func (*Input) SampleConfig() string { return sampleConfig }

//nolint:lll
func (*Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"MongoDB log": `{"t":{"$date":"2021-06-03T09:12:19.977+00:00"},"s":"I", "c":"STORAGE", "id":22430, "ctx":"WTCheckpointThread","msg":"WiredTiger message","attr":{"message":"[1622711539:977142][1:0x7f1b9f159700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 653, snapshot max: 653 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)"}}`,
		},
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{inputName: pipelineConfig}
}

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.MgoDBLog != nil {
					return ipt.MgoDBLog.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (ipt *Input) RunPipeline() {
	if ipt.MgoDBLog == nil || len(ipt.MgoDBLog.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          ipt.MgoDBLog.Pipeline,
		GlobalTags:        ipt.Tags,
		IgnoreStatus:      ipt.MgoDBLog.IgnoreStatus,
		CharacterEncoding: ipt.MgoDBLog.CharacterEncoding,
		MultilinePatterns: []string{ipt.MgoDBLog.MultilineMatch},
		Done:              ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.MgoDBLog.Files, opt)
	if err != nil {
		log.Errorf("NewTailer: %s", err)

		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_mongodb"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)

	ipt.pauseCh = make(chan bool, inputs.ElectionPauseChannelLength)
	ipt.semStop = cliutils.NewSem()
	defTags = ipt.Tags

	for _, v := range ipt.Servers {
		mgocli, err := ipt.createMgoClient(v)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		var (
			host string
			li   = strings.LastIndexByte(v, '@')
		)
		if li > 0 {
			host = v[li+1:]
		} else {
			host = strings.TrimPrefix(v, "mongodb://")
		}
		ipt.mgoSvrs = append(ipt.mgoSvrs, &MongodbServer{
			host: host,
			cli:  mgocli,
			ipt:  ipt,
		})
	}
	if len(ipt.mgoSvrs) == 0 {
		log.Errorf("connect to all Mongodb servers failed")

		return
	}

	log.Infof("%s input started", inputName)

	tick := time.NewTicker(ipt.Interval.Duration)
	for {
		if ipt.pause {
			log.Debugf("not leader, skipped")
		} else if err := ipt.gather(); err != nil {
			log.Error(err.Error())
			ipt.feeder.FeedLastError(err.Error(), dkio.WithLastErrorInput(inputName))
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			log.Info("mongodb input exit")

			return
		case <-ipt.semStop.Wait():
			ipt.exit()
			log.Info("mongodb input return")

			return
		case <-tick.C:
		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) createMgoClient(url string) (*mongo.Client, error) {
	cliOpts := options.Client()
	cliOpts.ApplyURI(url)

	var tlsConfig *dknet.TLSClientConfig
	if ipt.TLSConf != nil {
		tlsConfig = ipt.TLSConf
	} else if ipt.TLSClientConfig != nil {
		tlsConfig = ipt.TLSClientConfig
	}
	if tlsConfig != nil {
		if tlscfg, err := tlsConfig.TLSConfig(); err != nil {
			log.Errorf("connect to mongodb with TLS failed: %s backe to INSECURE connection", err.Error())
		} else {
			cliOpts.SetTLSConfig(tlscfg)
		}
	}
	cliOpts.SetConnectTimeout(time.Second * 10)
	mgocli, err := mongo.Connect(context.TODO(), cliOpts)
	if err != nil {
		return nil, err
	}
	if err = mgocli.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	return mgocli, nil
}

// Reads stats from all configured servers.
// Returns one of the errors encountered while gather stats (if any).
func (ipt *Input) gather() error {
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_mongodb"})
	for _, svr := range ipt.mgoSvrs {
		func(svr *MongodbServer) {
			g.Go(func(ctx context.Context) error {
				return svr.gatherData(ipt.GatherReplicaSetStats, ipt.GatherClusterStats, ipt.GatherPerDBStats,
					ipt.GatherPerColStats, ipt.ColStatsDBs, ipt.GatherTopStat)
			})
		}(svr)
	}

	return g.Wait()
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil

	case <-datakit.Exit.Wait():
		log.Info("pause mongodb interrupted by global exit.")
		return nil

	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		log.Info("mongodb log exits")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:  dkio.DefaultFeeder(),
		semStop: cliutils.NewSem(),
		Tagger:  dkpt.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
