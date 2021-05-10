package mongodb

import (
	"fmt"
	"net/url"
	"strings"
	"time"

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
	tags["hostname"] = s.URL.Host

	return tags
}

func (s *Server) authLog(err error) {
	if strings.Contains(err.Error(), "not authorized") {
		l.Debug(err.Error())
	} else {
		l.Error(err.Error())
	}
}

func (s *Server) gatherServerStatus() (*ServerStatus, error) {
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

func (s *Server) gatherReplSetStatus() (*ReplSetStatus, error) {
	replSetStatus := &ReplSetStatus{}
	err := s.Session.DB("admin").Run(bson.D{
		{
			Name:  "replSetGetStatus",
			Value: 1,
		},
	}, replSetStatus)
	if err != nil {
		return nil, err
	}

	return replSetStatus, nil
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

func (s *Server) gatherClusterStatus() (*ClusterStatus, error) {
	chunkCount, err := s.Session.DB("config").C("chunks").Find(bson.M{"jumbo": true}).Count()
	if err != nil {
		return nil, err
	}

	return &ClusterStatus{JumboChunksCount: int64(chunkCount)}, nil
}

func (s *Server) gatherShardConnPoolStats() (*ShardStats, error) {
	shardStats := &ShardStats{}
	err := s.Session.DB("admin").Run(bson.D{
		{
			Name:  "shardConnPoolStats",
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

	results := &ColStats{}
	for _, dbName := range dbNames {
		if stringInSlice(dbName, colStatsDbs) || len(colStatsDbs) == 0 {
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

func (s *Server) gatherData(gatherReplicaSetStatus bool, gatherClusterStatus bool, gatherPerDbStats bool, gatherPerColStats bool, colStatsDbs []string, gatherTopStat bool) error {
	s.Session.SetMode(mgo.Eventual, true)
	s.Session.SetSocketTimeout(0)

	serverStatus, err := s.gatherServerStatus()
	if err != nil {
		return err
	}

	// Get replica set status, an error indicates that the server is not a member of a replica set.
	var (
		replSetStatus *ReplSetStatus
		oplogStats    *OplogStats
	)
	if gatherReplicaSetStatus {
		if replSetStatus, err = s.gatherReplSetStatus(); err != nil {
			l.Debugf("Unable to gather replica set status: %q", err.Error())
		}
		// Gather the oplog if we are a member of a replica set. Non-replica set members do not have the oplog collections.
		if replSetStatus != nil {
			if oplogStats, err = s.gatherOplogStats(); err != nil {
				s.authLog(fmt.Errorf("Unable to get oplog stats: %q", err.Error()))
			}
		}
	}

	var clusterStatus *ClusterStatus
	if gatherClusterStatus {
		status, err := s.gatherClusterStatus()
		if err != nil {
			l.Debugf("Unable to gather cluster status: %q", err.Error())
		}
		clusterStatus = status
	}

	shardStats, err := s.gatherShardConnPoolStats()
	if err != nil {
		s.authLog(fmt.Errorf("Unable to gather shard connection pool stats: %q", err.Error()))
	}

	dbStats := &DbStats{}
	if gatherPerDbStats {
		names, err := s.Session.DatabaseNames()
		if err != nil {
			l.Debugf("Unable to get database names: %q", err.Error())

			return err
		}

		for _, name := range names {
			db, err := s.gatherDbStats(name)
			if err != nil {
				l.Debugf("Error getting db stats from %q: %q", name, err.Error())
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
		ServerStatus:  serverStatus,
		ReplSetStatus: replSetStatus,
		OplogStats:    oplogStats,
		ClusterStatus: clusterStatus,
		ShardStats:    shardStats,
		DbStats:       dbStats,
		ColStats:      collectionStats,
		TopStats:      topStatData,
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
		data.AddDbStats()
		data.AddColStats()
		data.AddShardHostStats()
		data.AddTopStats()
		data.append()
		data.flush()
	}

	s.lastResult = result

	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}
