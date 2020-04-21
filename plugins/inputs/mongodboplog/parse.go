package mongodboplog

import (
	"reflect"
	"strconv"
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

		if reflect.DeepEqual(reflect.TypeOf(elem.Value), reflect.TypeOf(bson.D{})) {
			md.rematch(elem.Value.(bson.D), completeKey+"/")
		}

		if array, ok := elem.Value.([]interface{}); ok {
			for k, v := range array {
				arraykey := completeKey + "[" + strconv.Itoa(k) + "]"

				if reflect.DeepEqual(reflect.TypeOf(v), reflect.TypeOf(bson.D{})) {
					md.rematch(v.(bson.D), arraykey+"/")
				}
				md.typeAssert(arraykey, v)
			}
		}

		md.typeAssert(completeKey, elem.Value)
	}
}

func (md *mgodata) typeAssert(completeKey string, value interface{}) {
	if typee, ok := md.pointlist[completeKey]; ok {
		switch typee {
		case "tags":
			if v, ok := value.(string); ok {
				md.tags[completeKey] = v
			}
		case "int":
			if v, ok := value.(float64); ok {
				md.fields[completeKey] = int64(v)
			}
		case "float":
			if v, ok := value.(float64); ok {
				md.fields[completeKey] = v
			}
		case "bool":
			if v, ok := value.(bool); ok {
				md.fields[completeKey] = v
			}
		case "string":
			if v, ok := value.(string); ok {
				md.fields[completeKey] = v
			}
		default:
			// nil
		}
	}

}

func (md *mgodata) reset() {
	md.tags = make(map[string]string)
	md.fields = make(map[string]interface{})
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
