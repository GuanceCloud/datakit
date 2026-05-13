// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestBuildDatabaseObjectPoint(t *testing.T) {
	ipt := defaultInput()
	ipt.Election = false
	ipt.HostPort = "127.0.0.1:27017"
	ipt.Tags = map[string]string{"env": "testing"}
	ipt.Tagger = datakit.DefaultGlobalTagger()

	oldDefTags := defTags
	defer func() {
		defTags = oldDefTags
	}()
	defTags = ipt.Tags

	svr := &MongodbServer{
		host:             "127.0.0.1:27017",
		databaseInstance: "mongo.example.com:27017",
		ipt:              ipt,
	}
	status := &MongoStatus{
		ServerStatus: &ServerStatus{
			Host:    "mongo.example.com:27017",
			Version: "7.0.1",
			Process: "mongod",
			Uptime:  123,
			Connections: &ConnectionStats{
				Current:      10,
				Available:    90,
				TotalCreated: 100,
			},
			OpLatencies: &OpLatenciesStats{
				Reads: &LatencyStats{
					Latency: 3100,
					Ops:     5,
				},
			},
			StorageEngine: &StorageEngine{Name: "wiredTiger"},
		},
		DBStats: &DBStats{
			DBs: []DB{
				{
					Name: "app",
					DBStatsData: &DBStatsData{
						Collections: 2,
						Objects:     3,
						AvgObjSize:  4,
						DataSize:    5,
						StorageSize: 6,
						Indexes:     7,
						IndexSize:   8,
					},
				},
			},
		},
	}
	statLine := &StatLine{
		Insert:  1,
		Query:   2,
		Update:  3,
		Delete:  4,
		GetMore: 5,
		Command: 6,
	}
	settings := map[string]interface{}{
		"net": map[string]interface{}{
			"port": int32(27017),
		},
		"storage": map[string]interface{}{
			"engine": "wiredTiger",
		},
	}
	ptTS := int64(1710000000000000123)

	pt := svr.buildDatabaseObjectPoint(status, statLine, settings, ptTS)
	assert.NotNil(t, pt)
	assert.Equal(t, mongodbObjectMeasurementName, pt.Name())
	assert.Equal(t, ptTS, pt.Time().UnixNano())
	assert.Equal(t, "127.0.0.1:27017-mongo.example.com:27017", pt.GetTag("name"))
	assert.Equal(t, "127.0.0.1:27017", pt.GetTag("server"))
	assert.Equal(t, "mongo.example.com:27017", pt.GetTag("database_instance"))
	assert.Empty(t, pt.GetTag("host"))
	assert.Equal(t, "27017", pt.GetTag("port"))
	assert.Equal(t, "7.0.1", pt.GetTag("version"))
	assert.Equal(t, mongodbType, pt.GetTag("database_type"))
	assert.Equal(t, int64(123), pt.Get("uptime"))
	assert.Equal(t, float64(21), pt.Get("qps"))
	assert.Equal(t, float64(620), pt.Get("avg_query_time"))
	message, ok := pt.Get("message").(string)
	assert.True(t, ok)
	assert.Contains(t, message, `"setting"`)
	assert.Contains(t, message, `"storage"`)
	assert.Contains(t, message, `"engine":"wiredTiger"`)
	assert.Contains(t, message, `"net"`)
	assert.Contains(t, message, `"port":27017`)
	assert.NotContains(t, message, `"storage_engine"`)
	assert.NotContains(t, message, `"process"`)
	assert.NotContains(t, message, `"connections"`)
	assert.NotContains(t, message, `"databases"`)
	assert.NotContains(t, message, `"name":"app"`)
}

func TestBuildDatabaseObjectPointWithSRVServer(t *testing.T) {
	ipt := defaultInput()
	ipt.Election = false
	ipt.Tags = map[string]string{"env": "testing"}
	ipt.Tagger = datakit.DefaultGlobalTagger()

	oldDefTags := defTags
	defer func() {
		defTags = oldDefTags
	}()
	defTags = ipt.Tags

	svr := &MongodbServer{
		host:             "cluster0.example.com",
		databaseInstance: "mongo.example.com:27017",
		ipt:              ipt,
	}
	status := &MongoStatus{
		ServerStatus: &ServerStatus{
			Host:    "mongo.example.com:27017",
			Version: "7.0.1",
			Uptime:  123,
		},
	}

	pt := svr.buildDatabaseObjectPoint(status, nil, map[string]interface{}{}, 1710000000000000123)
	assert.NotNil(t, pt)
	assert.Equal(t, "cluster0.example.com-mongo.example.com:27017", pt.GetTag("name"))
	assert.Equal(t, "cluster0.example.com", pt.GetTag("server"))
	assert.Equal(t, "mongo.example.com:27017", pt.GetTag("database_instance"))
	assert.Equal(t, "cluster0.example.com", pt.GetTag("host"))
	assert.Empty(t, pt.GetTag("port"))
}
