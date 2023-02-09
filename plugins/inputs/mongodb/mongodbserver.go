// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"context"
	"net"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/strarr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type oplogEntry struct {
	Timestamp int64 `bson:"ts"`
}

type MongodbServer struct {
	host       string
	cli        *mongo.Client
	lastResult *MongoStatus
	election   bool
}

func (svr *MongodbServer) getDefaultTags() map[string]string {
	tags := make(map[string]string)
	tags["mongod_host"] = svr.host
	setHostTagIfNotLoopback(tags, svr.host)
	for k, v := range defTags {
		tags[k] = v
	}

	return tags
}

func (svr *MongodbServer) gatherServerStats() (*ServerStatus, error) {
	rslt := svr.cli.Database("admin").RunCommand(context.TODO(), bson.M{"serverStatus": 1})
	if err := rslt.Err(); err != nil {
		return nil, err
	}

	svrStatus := &ServerStatus{}
	if err := rslt.Decode(svrStatus); err != nil {
		return nil, err
	}

	return svrStatus, nil
}

func (svr *MongodbServer) gatherReplSetStats() (*ReplSetStats, error) {
	rslt := svr.cli.Database("admin").RunCommand(context.TODO(), bson.M{"replSetGetStatus": 1})
	if err := rslt.Err(); err != nil {
		return nil, err
	}

	replSet := &ReplSetStats{}
	if err := rslt.Decode(replSet); err != nil {
		return nil, err
	}

	return replSet, nil
}

func (svr *MongodbServer) getOplogReplLag(col string) (*OplogStats, error) {
	query := bson.M{"ts": bson.M{"$exists": true}}
	rslt := svr.cli.Database("local").Collection(col).FindOne(context.TODO(), query, options.FindOne().SetSort(bson.M{"$natural": 1}))
	if err := rslt.Err(); err != nil {
		return nil, err
	}

	first := &oplogEntry{}
	if err := rslt.Decode(first); err != nil {
		return nil, err
	}

	rslt = svr.cli.Database("local").Collection(col).FindOne(context.TODO(), query, options.FindOne().SetSort(bson.M{"$natural": -1}))
	if err := rslt.Err(); err != nil {
		return nil, err
	}

	last := &oplogEntry{}
	if err := rslt.Decode(last); err != nil {
		return nil, err
	}

	firstTime := time.Unix(first.Timestamp>>32, 0)
	lastTime := time.Unix(last.Timestamp>>32, 0)
	stats := &OplogStats{TimeDiff: int64(lastTime.Sub(firstTime).Seconds())}

	return stats, nil
}

// The "oplog.rs" collection is stored on all replica set members.
// The "oplog.$main" collection is created on the master node of a
// master-slave replicated deployment.  As of MongoDB 3.2, master-slave
// replication has been deprecated.
func (svr *MongodbServer) gatherOplogStats() (*OplogStats, error) {
	stats, err := svr.getOplogReplLag("oplog.rs")
	if err == nil {
		return stats, nil
	}

	return svr.getOplogReplLag("oplog.$main")
}

func (svr *MongodbServer) gatherClusterStats() (*ClusterStats, error) {
	count, err := svr.cli.Database("config").Collection("chunks").CountDocuments(context.TODO(), bson.M{"jumbo": true})
	if err != nil {
		return nil, err
	}

	return &ClusterStats{JumboChunksCount: count}, nil
}

func (svr *MongodbServer) gatherShardConnPoolStats() (*ShardStats, error) {
	rslt := svr.cli.Database("admin").RunCommand(context.TODO(), bson.M{"connPoolStats": 1}) // connPoolStats
	if err := rslt.Err(); err != nil {
		return nil, err
	}

	shardStats := &ShardStats{}
	if err := rslt.Decode(shardStats); err != nil {
		return nil, err
	}

	return shardStats, nil
}

func (svr *MongodbServer) gatherDBStats(name string) (*DB, error) {
	rslt := svr.cli.Database(name).RunCommand(context.TODO(), bson.M{"dbStats": 1})
	if err := rslt.Err(); err != nil {
		return nil, err
	}

	stats := &DBStatsData{}
	if err := rslt.Decode(stats); err != nil {
		return nil, err
	}

	return &DB{
		Name:        name,
		DBStatsData: stats,
	}, nil
}

func (svr *MongodbServer) gatherCollectionStats(colStatsDBs []string) (*ColStats, error) {
	dbNames, err := svr.cli.ListDatabaseNames(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	dbNames = strarr.Intersect(dbNames, colStatsDBs)
	results := &ColStats{}
	for _, dbName := range dbNames {
		var colls []string
		colls, err = svr.cli.Database(dbName).ListCollectionNames(context.TODO(), bson.M{})
		if err != nil {
			log.Errorf("Error getting collection names: %s", err)
			continue
		}

		for _, colName := range colls {
			rslt := svr.cli.Database(dbName).RunCommand(context.TODO(), bson.M{"collStats": colName})
			if err := rslt.Err(); err != nil {
				log.Error(err.Error())
				continue
			}

			colStatLine := &ColStatsData{}
			if err := rslt.Decode(rslt); err != nil {
				log.Errorf("error getting col stats from %q: error: %s", colName, err.Error())
				continue
			}

			col := Collection{
				Name:         colName,
				DBName:       dbName,
				ColStatsData: colStatLine,
			}
			results.Collections = append(results.Collections, col)
		}
	}

	return results, nil
}

func (svr *MongodbServer) gatherTopStatData() (*TopStats, error) {
	rslt := svr.cli.Database("admin").RunCommand(context.TODO(), bson.M{"top": 1})
	if err := rslt.Err(); err != nil {
		return nil, err
	}

	topStats := &TopStats{}
	if err := rslt.Decode(topStats); err != nil {
		return nil, err
	}

	return topStats, nil
}

func (svr *MongodbServer) gatherData(gatherReplicaSetStats bool, gatherClusterStats bool, gatherPerDBStats bool, gatherPerColStats bool, colStatsDBs []string, gatherTopStat bool) error { // nolint:lll
	start := time.Now()
	serverStatus, err := svr.gatherServerStats()
	if err != nil {
		log.Debugf("gathering server failed: %s", err.Error())

		return err
	}

	// Get replica set status, an error indicates that the server is not a member of a replica set.
	var (
		ReplSetStats *ReplSetStats
		oplogStats   *OplogStats
	)
	if gatherReplicaSetStats {
		if ReplSetStats, err = svr.gatherReplSetStats(); err != nil {
			log.Debugf("Unable to gather replica set status: %w", err)
		}
		// Gather the oplog if we are a member of a replica set. Non-replica set members do not have the oplog collections.
		if ReplSetStats != nil {
			if oplogStats, err = svr.gatherOplogStats(); err != nil {
				log.Errorf("Unable to get oplog stats: %w", err)
			}
		}
	}

	var clusterStats *ClusterStats
	if gatherClusterStats {
		status, err := svr.gatherClusterStats()
		if err != nil {
			log.Debugf("Unable to gather cluster status: %w", err)
		}
		clusterStats = status
	}

	shardStats, err := svr.gatherShardConnPoolStats()
	if err != nil {
		log.Warnf("Unable to gather shard connection pool stats: %w", err)
	}

	dbStats := &DBStats{}
	if gatherPerDBStats {
		dbNames, err := svr.cli.ListDatabaseNames(context.TODO(), bson.M{})
		if err != nil {
			log.Debugf("Unable to get database names: %w", err)

			return err
		}

		for _, dbName := range dbNames {
			db, err := svr.gatherDBStats(dbName)
			if err != nil {
				log.Debugf("Error getting db stats from %s: %w", dbName, err)
			}
			dbStats.DBs = append(dbStats.DBs, *db)
		}
	}

	var colStats *ColStats
	if gatherPerColStats {
		stats, err := svr.gatherCollectionStats(colStatsDBs)
		if err != nil {
			log.Debugf("Unable to gather collection stats: %w", err)

			return err
		}
		colStats = stats
	}

	topStatData := &TopStats{}
	if gatherTopStat {
		topStats, err := svr.gatherTopStatData()
		if err != nil {
			log.Debugf("Unable to gather top stat data: %w", err)

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
		ColStats:     colStats,
		TopStats:     topStatData,
		SampleTime:   time.Now(),
	}
	log.Debugf("collected results: %v", *result)

	if svr.lastResult != nil {
		duration := result.SampleTime.Sub(svr.lastResult.SampleTime)
		durationInSeconds := int64(duration.Seconds())
		if durationInSeconds == 0 {
			durationInSeconds = 1
		}

		data := NewMongodbData(NewStatLine(svr.lastResult, result, svr.host, true, durationInSeconds), svr.getDefaultTags(), svr.election)
		data.AddDefaultStats()
		data.AddShardHostStats()
		data.AddDBStats()
		data.AddColStats()
		data.AddTopStats()
		data.append()
		data.flush(time.Since(start))
	}

	svr.lastResult = result

	return nil
}

func setHostTagIfNotLoopback(tags map[string]string, u string) {
	// input pattern:
	// localhost:27017/?authMechanism=SCRAM-SHA-256&authSource=admin
	// 127.0.0.1:27017,
	// 10.10.3.33:18832,
	uu, err := url.Parse("mongodb://" + u)
	if err != nil {
		log.Errorf("parse url: %v", err)
		return
	}
	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		log.Errorf("split host and port: %v", err)
		return
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		tags["host"] = host
	}
}
