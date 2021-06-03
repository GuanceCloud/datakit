package mongodb

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/charset"
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
	tags["mongo_host"] = s.URL.Host
	for k, v := range defTags {
		tags[k] = v
	}

	return tags
}

func (s *Server) authLog(err error) {
	if strings.Contains(err.Error(), "not authorized") {
		l.Debug(err.Error())
	} else {
		l.Error(err.Error())
	}
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

func (s *Server) gatherDbStats(name string) (*Db, error) {
	stats := &DbStatsData{}
	err := s.Session.DB(name).Run(bson.D{
		{
			Name:  "dbStats",
			Value: 1,
		},
	}, stats)
	if err != nil {
		return nil, err
	}

	return &Db{
		Name:        name,
		DbStatsData: stats,
	}, nil
}

func (s *Server) gatherCollectionStats(colStatsDbs []string) (*ColStats, error) {
	dbNames, err := s.Session.DatabaseNames()
	if err != nil {
		return nil, err
	}

	dbNames = charset.Intersect(dbNames, colStatsDbs)
	results := &ColStats{}
	for _, dbName := range dbNames {
		var colls []string
		colls, err = s.Session.DB(dbName).CollectionNames()
		if err != nil {
			l.Errorf("Error getting collection names: %q", err.Error())
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
				s.authLog(fmt.Errorf("error getting col stats from %q: %v", colName, err))
				continue
			}

			collection := &Collection{
				Name:         colName,
				DbName:       dbName,
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

func (s *Server) gatherData(gatherReplicaSetStats bool, gatherClusterStats bool, gatherPerDbStats bool, gatherPerColStats bool, colStatsDbs []string, gatherTopStat bool) error {
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
			l.Debugf("Unable to gather replica set status: %q", err.Error())
		}
		// Gather the oplog if we are a member of a replica set. Non-replica set members do not have the oplog collections.
		if ReplSetStats != nil {
			if oplogStats, err = s.gatherOplogStats(); err != nil {
				s.authLog(fmt.Errorf("Unable to get oplog stats: %q", err.Error()))
			}
		}
	}

	var clusterStats *ClusterStats
	if gatherClusterStats {
		status, err := s.gatherClusterStats()
		if err != nil {
			l.Debugf("Unable to gather cluster status: %q", err.Error())
		}
		clusterStats = status
	}

	shardStats, err := s.gatherShardConnPoolStats()
	if err != nil {
		s.authLog(fmt.Errorf("Unable to gather shard connection pool stats: %q", err.Error()))
	}

	dbStats := &DbStats{}
	if gatherPerDbStats {
		dbNames, err := s.Session.DatabaseNames()
		if err != nil {
			l.Debugf("Unable to get database names: %q", err.Error())

			return err
		}

		for _, dbName := range dbNames {
			db, err := s.gatherDbStats(dbName)
			if err != nil {
				l.Debugf("Error getting db stats from %q: %q", dbName, err.Error())
			}
			dbStats.Dbs = append(dbStats.Dbs, *db)
		}
	}

	var collectionStats *ColStats
	if gatherPerColStats {
		stats, err := s.gatherCollectionStats(colStatsDbs)
		if err != nil {
			l.Debugf("Unable to gather collection stats: %q", err.Error())

			return err
		}
		collectionStats = stats
	}

	topStatData := &TopStats{}
	if gatherTopStat {
		topStats, err := s.gatherTopStatData()
		if err != nil {
			l.Debugf("Unable to gather top stat data: %q", err.Error())

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
		DbStats:      dbStats,
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

		data := NewMongodbData(NewStatLine(*s.lastResult, *result, s.URL.Host, true, durationInSeconds), s.getDefaultTags(), duration)
		data.AddDefaultStats()
		data.AddShardHostStats()
		data.AddDbStats()
		data.AddColStats()
		data.AddTopStats()
		data.append()
		data.flush()
	}

	s.lastResult = result

	return nil
}
