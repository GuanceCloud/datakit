package dql

import (
	"math"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/es"
)

const (
	maxDQLParseResult   = 5
	InfluxReadPrecision = "ms"
	DefaultRP           = "" // influxdb: empty RP means the deafult RP
)

var (
	qch       = make(chan *InnerQuery, 32)
	qchDebug  = make(chan *InnerQueryDebug, 4)
	qchBackup = make(chan *InnerQueryBackup, 4)
	l         = logger.DefaultSLogger("dql")
)

type QueryResult struct {
	Series      []models.Row  `json:"series"`
	Cost        string        `json:"cost"`
	RawQuery    string        `json:"raw_query,omitempty"`
	Totalhits   int64         `json:"total_hits,omitempty"`
	GroupByList []string      `json:"group_by,omitempty"`
	SearchAfter []interface{} `json:"search_after,omitempty"`
}

func InitLog() {
	l = logger.SLogger("dql")
}

type singleQuery struct {
	Query                string              `json:"query"`
	TimeRange            []int64             `json:"time_range"`
	Conditions           string              `json:"conditions"`
	MaxPoint             int64               `json:"max_point"`
	MaxDuration          string              `json:"max_duration"`
	OrderBy              []map[string]string `json:"orderby"`
	Limit                int64               `json:"limit"`
	Offset               int64               `json:"offset"`
	DisableSlimit        bool                `json:"disable_slimit"`
	DisableMultipleField bool                `json:"disable_multiple_field"`
	SearchAfter          []interface{}       `json:"search_after"`
	Highlight            bool                `json:"highlight"`
}

type InnerQuery struct {
	WorkspaceUUID string         `json:"workspace_uuid"`
	Token         string         `json:"token"`
	Queries       []*singleQuery `json:"queries"`
	EchoExplain   bool           `json:"echo_explain"`

	result chan interface{}
}

type InnerQueryDebug struct {
	WorkspaceUUID string `json:"workspace_uuid"`
	Namespace     string `json:"namespace"` // only accept "influxdb", "es"
	Query         string `json:"query"`

	result chan interface{}
}

// InnerQueryBackup query backup
type InnerQueryBackup struct {
	WorkspaceUUID string         `json:"workspace_uuid"`
	Queries       []*singleQuery `json:"queries"`
	EchoExplain   bool           `json:"echo_explain"`

	result chan interface{}
}

func Query(iq *InnerQuery) interface{} {
	iq.result = make(chan interface{})
	qch <- iq

	select {
	case res := <-iq.result: // blocking
		return res

		// FIXME: add timeout
	}
}

func QueryDebug(iq *InnerQueryDebug) interface{} {
	iq.result = make(chan interface{})
	qchDebug <- iq
	select {
	case res := <-iq.result: // blocking
		return res
		// FIXME: add timeout
	}
}

// QueryBackup backup query
func QueryBackup(iq *InnerQueryBackup) interface{} {
	iq.result = make(chan interface{})
	qchBackup <- iq

	select {
	case res := <-iq.result: // blocking
		return res

		// FIXME: add timeout
	}
}

func StartQueryWorkers() {
	if config.C.Global.Workers == 0 {
		return
	}

	for i := 0; i < config.C.Global.Workers; i++ {
		go func(idx int) {
			qw := queryWorker{
				influxClis: map[string]*influxQueryCli{},
				influxDBs:  map[string][]string{},
			}
			qw.run(idx)
		}(i)
	}

	const queryDebugWokers = 1

	for i := 0; i < queryDebugWokers; i++ {
		go func(idx int) {
			qw := queryWorker{
				influxClis: map[string]*influxQueryCli{},
				influxDBs:  map[string][]string{},
			}
			qw.runQueryDebug(idx)
		}(i)
	}

	// query backup
	for i := 0; i < 1; i++ {
		go func(idx int) {
			qw := queryWorker{
				influxClis: map[string]*influxQueryCli{},
				influxDBs:  map[string][]string{},
			}
			qw.runBackup(idx)
		}(i)
	}
}

// ESIndexInfo 索引的基本信息
type ESIndexInfo struct {
	IndexName string
	Start     int64
	End       int64
	IsWrite   string
}

// ESIndexFields 缓存索引的fields信息
type ESIndexFields struct {
	IndexName  string
	Dql        string
	Res        *QueryResult
	CreateTime time.Time
}

// ESWarmUpData 多个索引的信息
type ESWarmUpData struct {
	IndexInfos  map[string][]*ESIndexInfo
	IndexFields map[string][]*ESIndexFields
}

// WarmUpData 预热数据
var WarmUpData *ESWarmUpData

// WarmUpESCli escli
var WarmUpESCli *es.EsCli

// StartQueryWarmUp 预热数据
func StartQueryWarmUp() {
	var err error
	WarmUpESCli, err = es.InitEsCli(
		config.C.Es.Host,
		config.C.Es.User,
		config.C.Es.Passwd,
		config.C.Es.TimeOut,
	)

	WarmUpData = &ESWarmUpData{
		IndexInfos:  map[string][]*ESIndexInfo{},
		IndexFields: map[string][]*ESIndexFields{},
	}

	if err != nil {
		// l.Fatalf("[fatal] init Es failed : %s", err.Error())
		return
	}

	go func() {
		runWarmUpTask(true)

	}()
}

func runWarmUpTask(first bool) {
	var err error
	for {
		if first {
			err = WarmUpData.Update()
			if err == nil {
				first = false
			}
			time.Sleep(5 * time.Minute)
			continue
		}
		WarmUpData.Update()
		time.Sleep(1 * time.Hour)
	}

}

// Update 周期更新
func (w *ESWarmUpData) Update() error {
	aliasInfo, err := WarmUpESCli.GetWarmUpAliasInfo()
	if err != nil {
		l.Error(err)
		return err
	}
	for _, item := range aliasInfo {
		aIdxName := item[0]  // 索引别名
		indexName := item[1] // 具体索引名称
		isWrite := item[2]   // 是否正在写
		if _, ok := w.IndexInfos[aIdxName]; !ok {
			w.IndexInfos[aIdxName] = []*ESIndexInfo{}
		}
		isRepeat := false
		// 如果已经存在，且是非当前写入索引，不在重复请求
		for _, oItem := range w.IndexInfos[aIdxName] {
			if oItem.IndexName == indexName || oItem.IsWrite == "false" {
				isRepeat = true
				break
			}
		}
		if !isRepeat {
			start, end, err := WarmUpESCli.GetWarmUpTimeRange(indexName)
			if err != nil {
				l.Error(err)
				continue
			}
			if isWrite == "true" { // 如果正在写，需要修改结束时间
				end = math.MaxInt64
			}
			idxInfo := ESIndexInfo{
				IndexName: indexName,
				Start:     start,
				End:       end,
				IsWrite:   isWrite,
			}

			w.IndexInfos[aIdxName] = append(w.IndexInfos[aIdxName], &idxInfo)
			time.Sleep(time.Second * 5)
		}
	}
	return nil

}

// GetTimeRangeIndices 获取对应时间范围内的索引名称
func (w *ESWarmUpData) GetTimeRangeIndices(indexName string, start, end int64) []string {
	res := []string{}
	if w == nil || w.IndexInfos == nil {
		return res
	}
	if infos, ok := w.IndexInfos[indexName]; ok {
		for _, item := range infos {
			s := start >= item.Start && start <= item.End
			e := end >= item.Start && end <= item.End
			if s || e {
				res = append(res, item.IndexName)
			}
		}
	}
	return res

}

// GetShowFieldsRes get fields res
func (w *ESWarmUpData) GetShowFieldsRes(indexName string, dql string) *QueryResult {
	var res *QueryResult
	now := time.Now()
	if indexFields, ok := w.IndexFields[indexName]; ok {
		for _, item := range indexFields {
			isValidPeriod := now.Sub(item.CreateTime).Minutes() < 30
			if dql == item.Dql && isValidPeriod {
				res = item.Res
				break
			}
		}
	}
	return res
}

func (w *ESWarmUpData) UpdateShowFields(indexName string, dql string, res *QueryResult) {
	now := time.Now()
	indexFields, ok := w.IndexFields[indexName]
	nFields := ESIndexFields{
		IndexName:  indexName,
		Dql:        dql,
		Res:        res,
		CreateTime: now,
	}
	if !ok {

		w.IndexFields[indexName] = []*ESIndexFields{
			&nFields,
		}
	}
	for i, item := range indexFields {
		if dql == item.Dql {
			w.IndexFields[indexName][i] = &nFields
		}
	}
	return

}
