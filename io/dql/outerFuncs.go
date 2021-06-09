package dql

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"

	kodom "gitlab.jiagouyun.com/cloudcare-tools/kodo/models"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
)

// outer delete func
// 删除查询的结果
func (qw *queryWorker) outerDeleteQuery(wsid string, q *ASTResult, explain bool) (*QueryResult, error) {

	deleteFunc, ok := q.AST.(*parser.DeleteFunc)
	if !ok || deleteFunc.StrDql == "" {
		return nil, fmt.Errorf("outer delete func error")
	}

	// (1) 删除部分数据
	if deleteFunc.DeleteIndex == false && deleteFunc.DeleteMeasurement == false {
		asts, err := Parse(deleteFunc.StrDql, qw.dqlparam)
		if err != nil {
			return nil, err
		}
		if len(asts) != 1 {
			return nil, fmt.Errorf("only single inner-query allowed within outer delete func")
		}
		return qw.deleteByQuery(wsid, asts[0])
	}

	// (2) 删除influxdb指定measurement
	if deleteFunc.DeleteMeasurement {
		dropStr := "drop measurement " + deleteFunc.StrDql
		return qw.deleteMeasurement(wsid, dropStr)
	}

	// (3) 删除ES指定索引
	if deleteFunc.DeleteIndex {
		return qw.deleteIndex(wsid, deleteFunc.StrDql)
	}

	return nil, nil

}

// 更新influxdb的查询语句 ，即 selete * from 变为 delete from
func getMerticDeleteDQL(ast interface{}) (string, error) {
	var (
		err       error
		deleteStr string
		selectStr = "SELECT * FROM"
	)
	switch v := ast.(type) {
	case *parser.DFQuery: // 删除其余条件
		v.GroupBy = nil
		v.Limit = nil
		v.SLimit = nil
		v.Targets = nil
		deleteStr, err = v.InfluxQL()
		if err != nil {
			return deleteStr, err
		}
		if len(deleteStr) > len(selectStr) && deleteStr[0:len(selectStr)] == selectStr {
			deleteStr = "delete from" + deleteStr[len(selectStr):]
		} else {
			return deleteStr, fmt.Errorf("unsupport dql, %s", deleteStr)
		}

	default:
		return deleteStr, fmt.Errorf("unsupport dql")
	}

	return deleteStr, nil
}

// 删除部分数据
func (qw *queryWorker) deleteByQuery(wsid string, astRes *ASTResult) (*QueryResult, error) {

	var (
		dqlStr string
		start  = time.Now()
	)

	switch astRes.Namespace {
	case "metric": // 删除influxdb数据
		dropStr, err := getMerticDeleteDQL(astRes.AST)
		if err != nil {
			return nil, err
		}
		return qw.deleteMeasurement(wsid, dropStr)
	default: // 删除es数据
		indexName := wsid + "_" + astRes.Namespace
		dqlStr = astRes.Q
		err := qw.esCli.DeleteByQuery(indexName, dqlStr)
		if err != nil {
			return nil, err
		}
	}

	elapsed := time.Since(start)
	return &QueryResult{
		Series:   nil,
		Cost:     fmt.Sprintf("%v", elapsed),
		RawQuery: "delete by query: " + dqlStr,
	}, nil
}

// 删除ES index
func (qw *queryWorker) deleteIndex(wsid, indexName string) (*QueryResult, error) {
	start := time.Now()
	dqlStr := "all"
	indexName = wsid + "_" + indexName
	err := qw.esCli.DeleteByQuery(indexName, dqlStr)

	if err != nil { // ES删除报错
		return nil, err
	}

	elapsed := time.Since(start)
	return &QueryResult{
		Series:   nil,
		Cost:     fmt.Sprintf("%v", elapsed),
		RawQuery: "delete by query: " + dqlStr,
	}, nil
}

// 删除influxdb measurement
func (qw *queryWorker) deleteMeasurement(wsid, dropStr string) (*QueryResult, error) {
	start := time.Now()

	// dqlStr := "drop measurement " + measurementName

	dbUUID, err := kodom.GetWsDBUID(wsid)
	if err != nil {
		return nil, err
	}
	ifdb, err := kodom.QueryInfluxInfoAdmin(dbUUID) // 需要admin权限

	if err != nil {
		return nil, err
	}

	cli, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{ // 每次删除都创建新的连接
		Addr:               ifdb.Host,
		Username:           ifdb.User,
		Password:           ifdb.Pwd,
		UserAgent:          "dql delete measurement",
		Timeout:            time.Duration(config.C.Influx.ReadTimeOut) * time.Second,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}

	defer cli.Close() // 删除结束，释放数据库连接

	influxq := influxdb.Query{
		Command:  dropStr,
		Database: ifdb.DB,
	}

	res, err := cli.Query(influxq)
	if err != nil {
		return nil, err
	} else if res.Error() != nil {
		return nil, res.Error()
	}

	elapsed := time.Since(start)

	return &QueryResult{
		Series:   nil,
		Cost:     fmt.Sprintf("%v", elapsed),
		RawQuery: "delete by query: " + dropStr,
	}, nil
}

// OuterFInfo 外层函数的基本属性
type OuterFInfo struct {
	fName string
	data  *QueryResult
	ses   *[]models.Row
}

// OuterHandler 外层函数的处理函数
type OuterHandler interface {
	doAction(*[]models.Row) (*[]models.Row, error)
	singleRowAction(models.Row) ([][]interface{}, error)
}

// DifferenceFInfo difference函数属性
type DifferenceFInfo struct {
	outerFInfo    *OuterFInfo
	allowNegative bool
}

// DerivativeFInfo derivative函数属性
type DerivativeFInfo struct {
	outerFInfo    *OuterFInfo
	allowNegative bool
}

// MovingAverageFInfo moving-average函数属性
type MovingAverageFInfo struct {
	outerFInfo *OuterFInfo
	size       int64
}

// LogFInfo log2，log10函数属性
type LogFInfo struct {
	outerFInfo *OuterFInfo
	base       float64
}

// CumSumFInfo cumsum函数属性
type CumSumFInfo struct {
	outerFInfo *OuterFInfo
}

// AbsFInfo abs函数属性
type AbsFInfo struct {
	outerFInfo *OuterFInfo
}

// MinFInfo min函数属性
type MinFInfo struct {
	outerFInfo *OuterFInfo
}

// MaxFInfo max函数属性
type MaxFInfo struct {
	outerFInfo *OuterFInfo
}

// AvgFInfo avg函数属性
type AvgFInfo struct {
	outerFInfo *OuterFInfo
}

// SumFInfo sum函数属性
type SumFInfo struct {
	outerFInfo *OuterFInfo
}

// FirstFInfo first函数属性
type FirstFInfo struct {
	outerFInfo *OuterFInfo
}

// LastFInfo last函数属性
type LastFInfo struct {
	outerFInfo *OuterFInfo
}

// CountFInfo count函数属性
type CountFInfo struct {
	outerFInfo     *OuterFInfo
	allowDuplicate bool //是否允许重复
}

// outer funcs, replace some func dql funcs
// 外层函数处理入口
func (qw *queryWorker) outerFuncQuery(wsid string, q *ASTResult, explain bool) (*QueryResult, error) {
	var (
		err   error
		start = time.Now()
	)

	outerFuncs, ok := q.AST.(*parser.OuterFuncs)
	if !ok {
		return nil, fmt.Errorf("outer func error")
	}

	if len(outerFuncs.Funcs) == 0 || len(outerFuncs.Funcs[0].FuncArgVals) < 1 {
		return nil, fmt.Errorf("outer func error")
	}

	str, ok := outerFuncs.Funcs[0].FuncArgVals[0].(string)
	if !ok {
		return nil, fmt.Errorf("outer func error")
	}
	innerQ := str

	asts, err := Parse(innerQ, qw.dqlparam)
	if err != nil {
		return nil, err
	}

	if len(asts) != 1 {
		return nil, fmt.Errorf("only single inner-query allowed within outer func")
	}

	data, err := qw.runQuery(wsid, asts[0], explain) // dql query
	if err != nil {
		return nil, err
	}

	rows, err := outerFuncsHandler(outerFuncs, data)
	if err != nil {
		return nil, err
	}
	elapsed := time.Since(start)

	if len(rows) == 0 {
		rows = nil
	}

	return &QueryResult{
		Series: rows,
		Cost:   fmt.Sprintf("%v", elapsed),
	}, nil
}

func singleFuncHandler(outerFunc *parser.OuterFunc, data *QueryResult) (OuterHandler, error) {
	var ohandler OuterHandler

	fInfo := OuterFInfo{
		fName: outerFunc.Func.Name,
		data:  data,
	}

	switch outerFunc.Func.Name {
	case "abs":
		nfInfo := AbsFInfo{outerFInfo: &fInfo}
		ohandler = &nfInfo

	case "avg":
		nfInfo := AvgFInfo{outerFInfo: &fInfo}
		ohandler = &nfInfo

	case "cumsum":
		nfInfo := CumSumFInfo{outerFInfo: &fInfo}
		ohandler = &nfInfo

	case "derivative":
		nfInfo := DerivativeFInfo{outerFInfo: &fInfo}
		nfInfo.allowNegative = true
		ohandler = &nfInfo

	case "difference":
		nfInfo := DifferenceFInfo{outerFInfo: &fInfo}
		nfInfo.allowNegative = true
		ohandler = &nfInfo

	case "first":
		nfInfo := FirstFInfo{outerFInfo: &fInfo}
		ohandler = &nfInfo

	case "last":
		nfInfo := LastFInfo{outerFInfo: &fInfo}
		ohandler = &nfInfo

	case "log10":
		nfInfo := LogFInfo{outerFInfo: &fInfo}
		nfInfo.base = float64(10)
		ohandler = &nfInfo

	case "log2":
		nfInfo := LogFInfo{outerFInfo: &fInfo}
		nfInfo.base = float64(2)
		ohandler = &nfInfo

	case "max":
		nfInfo := MaxFInfo{outerFInfo: &fInfo}
		ohandler = &nfInfo

	case "min":
		nfInfo := MinFInfo{outerFInfo: &fInfo}
		ohandler = &nfInfo

	case "moving_average":
		nfInfo := MovingAverageFInfo{outerFInfo: &fInfo}
		nfInfo.size = outerFunc.FuncArgVals[len(outerFunc.FuncArgVals)-1].(int64)
		ohandler = &nfInfo

	case "non_negative_derivative":
		nfInfo := DerivativeFInfo{outerFInfo: &fInfo}
		// nfInfo.allowNegative = false
		ohandler = &nfInfo

	case "non_negative_difference":
		nfInfo := DifferenceFInfo{outerFInfo: &fInfo}
		// nfInfo.allowNegative = false
		ohandler = &nfInfo

	case "sum":
		nfInfo := SumFInfo{outerFInfo: &fInfo}
		ohandler = &nfInfo

	case "count":
		nfInfo := CountFInfo{outerFInfo: &fInfo}
		nfInfo.allowDuplicate = true
		ohandler = &nfInfo

	case "count_distinct":
		nfInfo := CountFInfo{outerFInfo: &fInfo}
		// nfInfo.allowDuplicate = false
		ohandler = &nfInfo

	}
	return ohandler, nil
}

// 处理外层函数
func outerFuncsHandler(outerFuncs *parser.OuterFuncs, data *QueryResult) ([]models.Row, error) {
	var (
		ses *[]models.Row
		err error
	)
	handlers := []OuterHandler{}

	for _, outerFunc := range outerFuncs.Funcs {
		handler, err := singleFuncHandler(outerFunc, data)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, handler)
	}

	for _, handler := range handlers {
		ses, err = handler.doAction(ses)
		if err != nil {
			return nil, err
		}
	}

	return *ses, err
}

// interface{} to float64
// 获取float64类型的数值
func getNumberValue(v interface{}) (float64, error) {

	var (
		res float64
		err error
	)

	switch v.(type) {
	case nil: // 空值转化为0
		res = float64(0)
	case float64:
		res = v.(float64)
	case int64:
		res = float64(v.(int64))
	case json.Number:
		res, err = v.(json.Number).Float64()

	case string:
		err = fmt.Errorf("%v type is string, should be number type", v)
	default:
		err = fmt.Errorf("%v type is unknown, should be number type", v)
	}
	return res, err

}

// 删除nil值
func dropNil(rowValues [][]interface{}) [][]interface{} {
	res := [][]interface{}{}
	for _, v := range rowValues {
		existNil := false
		if v == nil {
			existNil = true
		} else {
			for _, iv := range v {
				if iv == nil {
					existNil = true
					break
				}
			}
		}
		if !existNil {
			res = append(res, v)
		}

	}
	return res
}

// 只删除非time字段值为nil的行
func dropNilExcludeTime(rowValues [][]interface{}) [][]interface{} {
	res := [][]interface{}{}
	for _, v := range rowValues {
		existNil := false
		for _, iv := range v[1:] {
			if iv == nil {
				existNil = true
				break
			}
		}
		if !existNil {
			res = append(res, v)
		}

	}
	return res
}

func genAction(preSes, oriSes *[]models.Row, d OuterHandler) (*[]models.Row, error) {
	var (
		ses = []models.Row{}
		err error
	)

	if preSes == nil {
		preSes = oriSes
	}

	if len(*preSes) == 0 {
		return &ses, err
	}

	for _, row := range *preSes {

		if len(row.Columns) != 2 {
			return nil, fmt.Errorf("inner query must return two columns, but has columns %v", row.Columns)
		}

		nValues := [][]interface{}{}

		// rowValues := dropNil(row.Values)

		if len(row.Values) > 0 {
			nValues, err = d.singleRowAction(row)
			if err != nil {
				return nil, err
			}
		}

		if len(nValues) == 0 {
			nValues = append(nValues, []interface{}{nil, nil})
		}
		nRow := models.Row{
			Name:    row.Name,
			Columns: row.Columns,
			Tags:    row.Tags,
			Values:  nValues,
			Partial: row.Partial,
		}
		ses = append(ses, nRow)

	}
	return &ses, err
}

func genTopAction(preSes, oriSes *[]models.Row, d OuterHandler) (*[]models.Row, error) {
	var (
		ses = []models.Row{}
		err error
	)

	if preSes == nil {
		preSes = oriSes
	}

	if len(*preSes) == 0 {
		return &ses, err
	}

	for _, row := range *preSes {

		if len(row.Columns) != 2 {
			return nil, fmt.Errorf("inner query must return two columns, but has columns %v", row.Columns)
		}

		nValues := [][]interface{}{}

		rowValues := dropNil(row.Values)

		if len(rowValues) > 0 {
			nValues, err = d.singleRowAction(row)
			if err != nil {
				return nil, err
			}
		}

		if len(nValues) == 0 {
			nValues = append(nValues, []interface{}{nil, nil})
		}
		nRow := models.Row{
			Name:    row.Name,
			Columns: row.Columns,
			Tags:    row.Tags,
			Values:  nValues,
			Partial: row.Partial,
		}
		ses = append(ses, nRow)

	}
	return &ses, err
}

// difference func
func (d *DifferenceFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// difference handle single row
func (d *DifferenceFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNil(row.Values)
	if len(rowValues) > 0 {
		init1, err := getNumberValue(rowValues[0][1])
		if err != nil {
			return nil, err
		}
		for _, v := range rowValues[1:] {
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			diff := v1 - init1
			init1 = v1
			if d.allowNegative == false && diff < 0 {
				continue
			}
			res = append(res, []interface{}{v[0], diff})
		}
	}
	return res, err
}

// derivative func
func (d *DerivativeFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// derivative handle single row
func (d *DerivativeFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)

	rowValues := dropNil(row.Values)

	if len(rowValues) > 0 {

		init0, err := getNumberValue(rowValues[0][0])
		if err != nil {
			return nil, err
		}
		init1, err := getNumberValue(rowValues[0][1])
		if err != nil {
			return nil, err
		}
		for _, v := range rowValues[1:] {
			v0, err := getNumberValue(v[0])
			if err != nil {
				return nil, err
			}
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			if v0 == init0 {
				return nil, fmt.Errorf("same timestamp")
			}
			diff := (v1 - init1) / ((v0 - init0) / 1000)
			init0, init1 = v0, v1
			if d.allowNegative == false && diff < 0 {
				continue
			}
			res = append(res, []interface{}{v[0], diff})
		}
	}
	return res, err
}

// moving average func
func (d *MovingAverageFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// moving average handle single row
func (d *MovingAverageFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)

	rowValues := dropNilExcludeTime(row.Values)

	if len(rowValues) > 0 && int64(len(rowValues)) >= d.size {
		i := d.size
		for i <= int64(len(rowValues)) {
			sum := float64(0)
			// get sum of windowSize points
			for _, rowVal := range rowValues[i-d.size : i] {
				v1, err := getNumberValue(rowVal[1])
				if err != nil {
					return nil, err
				}
				sum = sum + v1
			}
			res = append(res, []interface{}{rowValues[i-1][0], sum / float64(d.size)})
			i = i + 1
		}

	}
	return res, err
}

// math.log func
func (d *LogFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// math.log handle single row
func (d *LogFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)

	rowValues := dropNilExcludeTime(row.Values)

	if len(rowValues) > 0 {

		for _, v := range rowValues {
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			// 对数函数，x需要是正数
			if v1 > 0 {
				nV1 := math.Log(v1) / math.Log(d.base)
				res = append(res, []interface{}{v[0], nV1})
			}
		}
	}

	return res, err
}

// cumsum func
func (d *CumSumFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// cumsum handle single row
func (d *CumSumFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNilExcludeTime(row.Values)
	if len(rowValues) > 0 {
		sum := float64(0)
		for _, v := range rowValues {
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			sum = sum + v1
			res = append(res, []interface{}{v[0], sum})
		}
	}
	return res, err
}

// abs func
func (d *AbsFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// abs handle single row
func (d *AbsFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNilExcludeTime(row.Values)
	if len(rowValues) > 0 {
		for _, v := range rowValues {
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			res = append(res, []interface{}{v[0], math.Abs(v1)})
		}
	}
	return res, err
}

// min func
func (d *MinFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// min handle single row
func (d *MinFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNilExcludeTime(row.Values)
	if len(rowValues) > 0 {
		init0, err := getNumberValue(rowValues[0][0])
		if err != nil {
			return nil, err
		}
		init1, err := getNumberValue(rowValues[0][1])
		if err != nil {
			return nil, err
		}
		for _, v := range rowValues[1:] {
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			if init1 > v1 {
				v0, err := getNumberValue(v[0])
				if err != nil {
					return nil, err
				}
				init0 = v0
				init1 = v1
			}
		}
		res = append(res, []interface{}{init0, init1})
	}
	return res, err
}

// max func
func (d *MaxFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// max handle single row
func (d *MaxFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNilExcludeTime(row.Values) // 不需要time列

	if len(rowValues) > 0 {
		init0, err := getNumberValue(rowValues[0][0])
		if err != nil {
			return nil, err
		}
		init1, err := getNumberValue(rowValues[0][1])
		if err != nil {
			return nil, err
		}
		for _, v := range rowValues[1:] {
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			if init1 < v1 {
				v0, err := getNumberValue(v[0])
				if err != nil {
					return nil, err
				}
				init0 = v0
				init1 = v1
			}
		}
		res = append(res, []interface{}{init0, init1})
	}
	return res, err
}

// avg func
func (d *AvgFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// avg handle single row
func (d *AvgFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNilExcludeTime(row.Values)
	if len(rowValues) > 0 {
		sum := float64(0)
		lenValues := float64(len(rowValues))
		for _, v := range rowValues {
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			sum = sum + v1/lenValues
		}
		res = append(res, []interface{}{float64(0), sum})
	}
	return res, err
}

// sum func
func (d *SumFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// sum handle single row
func (d *SumFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNilExcludeTime(row.Values)
	if len(rowValues) > 0 {
		sum := float64(0)
		for _, v := range rowValues {
			v1, err := getNumberValue(v[1])
			if err != nil {
				return nil, err
			}
			sum = sum + v1
		}
		res = append(res, []interface{}{float64(0), sum})
	}
	return res, err
}

// first func
func (d *FirstFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// first handle single row
func (d *FirstFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNil(row.Values)

	if len(rowValues) > 0 {
		init0, err := getNumberValue(rowValues[0][0])
		if err != nil {
			return nil, err
		}
		init1, err := getNumberValue(rowValues[0][1])
		if err != nil {
			return nil, err
		}
		for _, v := range rowValues[1:] {
			v0, err := getNumberValue(v[0])
			if err != nil {
				return nil, err
			}
			if init0 > v0 { // 按照时间字段排序
				v1, err := getNumberValue(v[1])
				if err != nil {
					return nil, err
				}
				init0 = v0
				init1 = v1
			}
		}
		res = append(res, []interface{}{init0, init1})
	}
	return res, err
}

// last func
func (d *LastFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// last handle single row
func (d *LastFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNil(row.Values)

	if len(rowValues) > 0 {
		init0, err := getNumberValue(rowValues[0][0])
		if err != nil {
			return nil, err
		}
		init1, err := getNumberValue(rowValues[0][1])
		if err != nil {
			return nil, err
		}
		for _, v := range rowValues[1:] {
			v0, err := getNumberValue(v[0])
			if err != nil {
				return nil, err
			}
			if init0 < v0 { // 按照时间字段排序
				v1, err := getNumberValue(v[1])
				if err != nil {
					return nil, err
				}
				init0 = v0
				init1 = v1
			}
		}
		res = append(res, []interface{}{init0, init1})
	}
	return res, err
}

// count func
func (d *CountFInfo) doAction(preSes *[]models.Row) (*[]models.Row, error) {
	return genAction(preSes, &d.outerFInfo.data.Series, d)
}

// count handle single row
func (d *CountFInfo) singleRowAction(row models.Row) ([][]interface{}, error) {
	var (
		res = [][]interface{}{}
		err error
	)
	rowValues := dropNilExcludeTime(row.Values)
	count := 0
	if len(rowValues) > 0 {
		switch d.allowDuplicate {
		// 1) 如果不需要去重
		case true:
			count = len(rowValues)
		// 2) 如果需要去除重复
		default:
			vMap := map[string]int{}
			for _, v := range rowValues {
				strValue := fmt.Sprintf("%v", v[1])
				if _, ok := vMap[strValue]; !ok {
					vMap[strValue] = 0
				}
			}
			count = len(vMap)
		}
	}
	res = append(res, []interface{}{float64(0), count})
	return res, err
}
