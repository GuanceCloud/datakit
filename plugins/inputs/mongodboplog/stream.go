package mongodboplog

import (
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/vinllen/mgo"
	"github.com/vinllen/mgo/bson"
)

type stream struct {
	mgo *Mongodboplog
	//
	sub *Subscribe
	// mongodb namespace is 'database.collection'
	namespace string
	// stream iterator
	iter *mgo.Iter
	// start time
	receivedTime time.Time
	// mate data
	mdata mgodata

	points []*influxdb.Point
}

func newStream(sub *Subscribe, mongo *Mongodboplog) *stream {
	return &stream{
		mgo:          mongo,
		sub:          sub,
		namespace:    sub.Database + "." + sub.Collection,
		receivedTime: time.Now(),
		mdata:        newMgodata(sub),
	}
}

func (s *stream) start(wg *sync.WaitGroup) error {
	s.points = []*influxdb.Point{}
	defer wg.Done()

	session, err := mgo.Dial(s.sub.MongodbURL)
	if err != nil {
		return err
	}

	session.SetPoolLimit(2)
	session.SetSocketTimeout(0)
	session.SetMode(mgo.Primary, true)

	var query = bson.M{}

	// bson.MongoTimestamp int64
	// |----------32---------|-----------32-----------|
	// |   timestamp second  |          count         |
	query["ts"] = bson.M{"$gt": bson.MongoTimestamp(s.receivedTime.Unix() << 32)}

	s.iter = session.DB("local").C("oplog.rs").Find(query).LogReplay().Tail(-1)
	s.runloop()
	return nil
}

func (s *stream) runloop() {
	var log *bson.Raw

	for {
		log = new(bson.Raw)
		if s.iter.Next(log) {
			p := new(PartialLog)
			bson.Unmarshal(log.Data, p)

			if p.Namespace != s.namespace {
				continue
			}
			switch p.Operation {
			case "i", "u", "d":
				// fmt.Printf("\n%#v\n\n", p)
				s.mdata.setTime(p.Timestamp)
				s.mdata.rematch(p.Object, "/")

				if p, err := s.mdata.point(); err == nil {
					s.points = append(s.points, p)
				}
				s.flush()
				s.mdata.reset()
			}
		}
	}
}

func (s *stream) flush() (err error) {
	// fmt.Println(s.points)
	err = s.mgo.ProcessPts(s.points)
	s.points = nil
	return err
}
