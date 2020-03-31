package mongodb

import (
	"reflect"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/vinllen/mgo/bson"
)

type PartialLog struct {
	Timestamp     bson.MongoTimestamp `bson:"ts"`
	Operation     string              `bson:"op"`
	Gid           string              `bson:"g"`
	Namespace     string              `bson:"ns"`
	Object        bson.D              `bson:"o"`
	Query         bson.M              `bson:"o2"`
	UniqueIndexes bson.M              `bson:"uk"`
	Lsid          interface{}         `bson:"lsid"`        // mark the session id, used in transaction
	FromMigrate   bool                `bson:"fromMigrate"` // move chunk

	/*
	 * Every field subsequent declared is NEVER persistent or
	 * transfer on network connection. They only be parsed from
	 * respective logic
	 */
	UniqueIndexesUpdates bson.M // generate by CollisionMatrix
	RawSize              int    // generate by Decorator
	SourceId             int    // generate by Validator
}

type mgodata struct {
	measurement string
	pointlist   map[string]string
	tags        map[string]string
	fields      map[string]interface{}
	time        time.Time
}

func newMgodata(sub *Subscribe) mgodata {

	m := mgodata{
		measurement: sub.Measurement,
		pointlist:   make(map[string]string, len(sub.Tags)+len(sub.Fields)),
		tags:        make(map[string]string),
		fields:      make(map[string]interface{}),
		time:        time.Now(),
	}

	for _, v := range sub.Tags {
		m.pointlist[v] = "tags"
	}

	for k, v := range sub.Fields {
		m.pointlist[k] = v
	}

	return m
}

func (md *mgodata) setTime(ts bson.MongoTimestamp) {
	md.time = time.Unix(int64(ts)>>32, 0)
}

func (md *mgodata) rematch(docelem bson.D, succkey string) {

	for _, elem := range docelem {
		completeKey := succkey + elem.Name

		if t, ok := md.pointlist[completeKey]; ok {
			switch t {
			case "tags":
				if vv, ok := elem.Value.(string); ok {
					md.tags[completeKey] = vv
				}
			case "int":
				if vv, ok := elem.Value.(float64); ok {
					md.fields[completeKey] = int64(vv)
				}
			case "float":
				if vv, ok := elem.Value.(float64); ok {
					md.fields[completeKey] = vv
				}
			case "bool":
				if vv, ok := elem.Value.(bool); ok {
					md.fields[completeKey] = vv
				}
			case "string":
				if vv, ok := elem.Value.(string); ok {
					md.fields[completeKey] = vv
				}
			default:
				// nil
			}
		}

		if reflect.DeepEqual(reflect.TypeOf(elem.Value), reflect.TypeOf(bson.D{})) {
			md.rematch(elem.Value.(bson.D), completeKey+"/")
		}
	}
}

func (md *mgodata) reset() {
	md.tags = make(map[string]string)
	md.fields = make(map[string]interface{})
	// FIXME:
	// md.time reset ?
}

func (md *mgodata) point() (*influxdb.Point, error) {
	return influxdb.NewPoint(
		md.measurement,
		md.tags,
		md.fields,
		md.time,
	)
}
