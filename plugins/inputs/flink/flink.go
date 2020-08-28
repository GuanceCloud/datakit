package flink

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "flink"

	sampleCfg = `
[[inputs.flink]]
    # require
    db = "flink"

    # [inputs.flink.tags]
    # tags1 = "value1"
`
)

var (
	l      = logger.DefaultSLogger(inputName)
	dbList = Flinks{m: make(map[string]map[string]string), mut: &sync.RWMutex{}}
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Flink{}
	})
}

type Flink struct {
	DB   string            `toml:"db"`
	Tags map[string]string `toml:"tags"`
}

func (*Flink) SampleConfig() string {
	return sampleCfg
}

func (*Flink) Catalog() string {
	return inputName
}

func (f *Flink) Run() {
	l = logger.SLogger(inputName)
	l.Infof("flink input started...")

	dbList.Store(f.DB, f.Tags)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	db := r.URL.Query().Get("db")
	if _, ok := dbList.Load(db); !ok {
		l.Errorf("not open db %s", db)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Errorf("failed to read body, err: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := extract(db, r.URL.Query().Get("precision"), body); err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		l.Errorf("failed to handle, %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}
}

func extract(db, prec string, body []byte) error {
	pts, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), prec)
	if err != nil {
		return err
	}

	if len(pts) == 0 {
		return nil
	}

	var fields = make(map[string]interface{}, len(pts))
	for _, pt := range pts {
		ptFields, _ := pt.Fields()
		fields[string(pt.Name())] = ptFields["value"]
	}

	tags, _ := dbList.Load(db)
	data, err := io.MakeMetric(db, tags, fields, pts[0].Time())
	if err != nil {
		return err
	}

	if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
		return err
	}

	return nil
}

type Flinks struct {
	// map[db]map[tags1]value1
	m   map[string]map[string]string
	mut *sync.RWMutex
}

func (f *Flinks) Store(key string, value map[string]string) {
	f.mut.Lock()
	defer f.mut.Unlock()

	f.m[key] = value
}

func (f *Flinks) Load(key string) (map[string]string, bool) {
	f.mut.Lock()
	defer f.mut.Unlock()

	v, ok := f.m[key]
	return v, ok
}
