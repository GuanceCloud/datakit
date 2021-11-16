package mongodb

import (
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/strarr"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type oplogEntry struct {
	Timestamp bson.MongoTimestamp `bson:"ts"`
}

type Server struct {
	URL        *url.URL
	Session    *mgo.Session
	lastResult *MongoStatus
}

func (s *Server) getDefaultTags() map[string]string {
	tags := make(map[string]string)
	tags["mongod_host"] = s.URL.Host
	for k, v := range defTags {
		tags[k] = v
	}

	return tags
}

func (s *Server) gatherServerStats() (*ServerStatus, error) {
	serverStatus := &ServerStatus{}
	err := s.Session.DB("admin").Run(bson.D{
		{
			Name:  "serverStatus",
			Value: 1,
		},
		{
			Name:  "recordStats",
			Value: 0,
		},
	}, serverStatus)
	if err != nil {
		return nil, err
	}

	return serverStatus, nil
}

func (s *Server) gatherReplSetStats() (*ReplSetStats, error) {
	ReplSetStats := &ReplSetStats{}
	err := s.Session.DB("admin").Run(bson.D{
		{
			Name:  "replSetGetStatus",
			Value: 1,
		},
	}, ReplSetStats)
	if err != nil {
		return nil, err
	}

	return ReplSetStats, nil
}

func (s *Server) getOplogReplLag(collection string) (*OplogStats, error) {
	query := bson.M{"ts": bson.M{"$exists": true}}

	var first oplogEntry
	err := s.Session.DB("local").C(collection).Find(query).Sort("$natural").Limit(1).One(&first)
	if err != nil {
		return nil, err
	}

	var last oplogEntry
	if err = s.Session.DB("local").C(collection).Find(query).Sort("-$natural").Limit(1).One(&last); err != nil {
		return nil, err
	}

	firstTime := time.Unix(int64(first.Timestamp>>32), 0)
	lastTime := time.Unix(int64(last.Timestamp>>32), 0)
	stats := &OplogStats{TimeDiff: int64(lastTime.Sub(firstTime).Seconds())}

	return stats, nil
}

// The "oplog.rs" collection is stored on all replica set members.
//
// The "oplog.$main" collection is created on the master node of a
// master-slave replicated deployment.  As of MongoDB 3.2, master-slave
// replication has been deprecated.
func (s *Server) gatherOplogStats() (*OplogStats, error) {
	stats, err := s.getOplogReplLag("oplog.rs")
	if err == nil {
		return stats, nil
	}

	return s.getOplogReplLag("oplog.$main")
}

func (s *Server) gatherClusterStats() (*ClusterStats, error) {
	chunkCount, err := s.Session.DB("config").C("chunks").Find(bson.M{"jumbo": true}).Count()
	if err != nil {
		return nil, err
	}

	return &ClusterStats{JumboChunksCount: int64(chunkCount)}, nil
}

func (s *Server) gatherShardConnPoolStats() (*ShardStats, error) {
	shardStats := &ShardStats{}
	err := s.Session.DB("admin").Run(bson.D{
		{
			Name: "shardConnPoolStats",
			// Name:  "connPoolStats",
			Value: 1,
		},
	}, &shardStats)
	if err != nil {
		return nil, err
	}

	return shardStats, nil
}

func (s *Server) gatherDBStats(name string) (*DB, error) {
	stats := &DBStatsData{}
	err := s.Session.DB(name).Run(bson.D{
		{
			Name:  "dbStats",
			Value: 1,
		},
	}, stats)
	if err != nil {
		return nil, err
	}

	return &DB{
		Name:        name,
		DBStatsData: stats,
	}, nil
}

func (s *Server) gatherCollectionStats(colStatsDBs []string) (*ColStats, error) {
	dbNames, err := s.Session.DatabaseNames()
	if err != nil {
		return nil, err
	}

	dbNames = strarr.Intersect(dbNames, colStatsDBs)
	results := &ColStats{}
	for _, dbName := range dbNames {
		var colls []string
		colls, err = s.Session.DB(dbName).CollectionNames()
		if err != nil {
			l.Errorf("Error getting collection names: %s", err)
			continue
		}

		for _, colName := range colls {
			colStatLine := &ColStatsData{}
			err = s.Session.DB(dbName).Run(bson.D{
				{
					Name:  "collStats",
					Value: colName,
				},
			}, colStatLine)
			if err != nil {
				l.Errorf("error getting col stats from %q: %w", colName, err)
				continue
			}

			collection := &Collection{
				Name:         colName,
				DBName:       dbName,
				ColStatsData: colStatLine,
			}
			results.Collections = append(results.Collections, *collection)
		}
	}

	return results, nil
}

func (s *Server) gatherTopStatData() (*TopStats, error) {
	topStats := &TopStats{}
	err := s.Session.DB("admin").Run(bson.D{
		{
			Name:  "top",
			Value: 1,
		},
	}, topStats)
	if err != nil {
		return nil, err
	}

	return topStats, nil
}

func (s *Server) gatherData(gatherReplicaSetStats bool,
	gatherClusterStats bool,
	gatherPerDBStats bool,
	gatherPerColStats bool,
	colStatsDBs []string,
	gatherTopStat bool) error {
	start := time.Now()

	s.Session.SetMode(mgo.Eventual, true)
	s.Session.SetSocketTimeout(0)

	serverStatus, err := s.gatherServerStats()
	if err != nil {
		l.Debugf("Error gathering server status")

		return err
	}

	// Get replica set status, an error indicates that the server is not a member of a replica set.
	var (
		ReplSetStats *ReplSetStats
		oplogStats   *OplogStats
	)
	if gatherReplicaSetStats {
		if ReplSetStats, err = s.gatherReplSetStats(); err != nil {
			l.Debugf("Unable to gather replica set status: %w", err)
		}
		// Gather the oplog if we are a member of a replica set. Non-replica set members do not have the oplog collections.
		if ReplSetStats != nil {
			if oplogStats, err = s.gatherOplogStats(); err != nil {
				l.Errorf("Unable to get oplog stats: %w", err)
			}
		}
	}

	var clusterStats *ClusterStats
	if gatherClusterStats {
		status, err := s.gatherClusterStats()
		if err != nil {
			l.Debugf("Unable to gather cluster status: %w", err)
		}
		clusterStats = status
	}

	shardStats, err := s.gatherShardConnPoolStats()
	if err != nil {
		l.Warnf("Unable to gather shard connection pool stats: %w", err)
	}

	dbStats := &DBStats{}
	if gatherPerDBStats {
		dbNames, err := s.Session.DatabaseNames()
		if err != nil {
			l.Debugf("Unable to get database names: %w", err)

			return err
		}

		for _, dbName := range dbNames {
			db, err := s.gatherDBStats(dbName)
			if err != nil {
				l.Debugf("Error getting db stats from %s: %w", dbName, err)
			}
			dbStats.DBs = append(dbStats.DBs, *db)
		}
	}

	var collectionStats *ColStats
	if gatherPerColStats {
		stats, err := s.gatherCollectionStats(colStatsDBs)
		if err != nil {
			l.Debugf("Unable to gather collection stats: %w", err)

			return err
		}
		collectionStats = stats
	}

	topStatData := &TopStats{}
	if gatherTopStat {
		topStats, err := s.gatherTopStatData()
		if err != nil {
			l.Debugf("Unable to gather top stat data: %w", err)

			return err
		}
		topStatData = topStats
	}

	result := &MongoStatus{
		ServerStatus: serverStatus,
		ReplSetStats: ReplSetStats,
		OplogStats:   oplogStats,
		ClusterStats: clusterStats,
		ShardStats:   shardStats,
		DBStats:      dbStats,
		ColStats:     collectionStats,
		TopStats:     topStatData,
	}

	result.SampleTime = time.Now()
	if s.lastResult != nil && result != nil {
		duration := result.SampleTime.Sub(s.lastResult.SampleTime)
		durationInSeconds := int64(duration.Seconds())
		if durationInSeconds == 0 {
			durationInSeconds = 1
		}

		data := NewMongodbData(NewStatLine(*s.lastResult, *result, s.URL.Host, true, durationInSeconds), s.getDefaultTags())
		data.AddDefaultStats()
		data.AddShardHostStats()
		data.AddDBStats()
		data.AddColStats()
		data.AddTopStats()
		data.append()
		data.flush(time.Since(start))
	}

	s.lastResult = result

	return nil
}
