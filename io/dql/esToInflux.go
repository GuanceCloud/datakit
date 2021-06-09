package dql

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	"github.com/olivere/elastic/v7"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/utils"
)

type Tag struct {
	Key   string
	Value string
}

type Value struct {
	Key   string
	Value interface{}
}

type STags []*Tag
type SValues [][]*Value

type DataSingleDoc struct {
	Tags   []*Tag
	Values []*Value
}

// es结果转为influx结构
func es2influx(res *elastic.SearchResult, astRes *ASTResult) (*QueryResult, error) {
	switch astRes.AST.(type) {
	case *parser.DFQuery:
		// (1) 查询结果结构化(近似influxdb)
		return astRes.queryToInflux(res)

	case *parser.Show:
		// (2) show结果结构化(近似influxdb)
		return astRes.showToInflux(res)
	}
	return nil, nil
}

// es search_after结果
func esSearchAfterRes(res *elastic.SearchResult, astRes *ASTResult) ([]interface{}, error) {
	if res.Hits != nil && res.Hits.Hits != nil {
		hits := res.Hits.Hits
		lenHits := len(hits)
		if lenHits > 0 {
			lastSearchHit := hits[lenHits-1]
			if lastSearchHit.Sort != nil {
				return lastSearchHit.Sort, nil
			}
		}
	}
	return []interface{}{}, nil
}

// query res to influx
func (astRes *ASTResult) queryToInflux(res *elastic.SearchResult) (*QueryResult, error) {
	var (
		ses       []models.Row
		err       error
		totalHits int64
	)
	ses, err = astRes.doTransResult(res)
	if err != nil {
		return nil, err
	}

	dfQuery, ok := (astRes.AST).(*parser.DFQuery)
	if !ok {
		return nil, fmt.Errorf("es translate error")
	}

	// 保证ses每个元素都有value
	ses, err = validSes(ses)
	if err != nil {
		return nil, err
	}

	// 判断ses是否为空
	if len(ses) > 0 {
		if dfQuery.Helper != nil && dfQuery.Helper.ESTResPtr != nil {
			updateSesFuncs := []func(*parser.ESTRes, []models.Row) []models.Row{
				sortSes,       // ES 结果排序
				aliasSes,      // ES 别名替换
				multiRowSes,   // ES first，last多个函数，多行合并处理
				aggsPagingSes, // ES 聚合分页
			}
			for _, f := range updateSesFuncs {
				ses = f(dfQuery.Helper.ESTResPtr, ses)
			}
		}
		totalHits = res.TotalHits()
	} else {
		ses = nil
	}

	return &QueryResult{
		Series:    ses,
		Cost:      fmt.Sprintf(`%v`, time.Millisecond*time.Duration(res.TookInMillis)),
		Totalhits: totalHits,
	}, nil
}

//  show to influx
func (astRes *ASTResult) showToInflux(res *elastic.SearchResult) (*QueryResult, error) {
	var (
		ses       []models.Row
		err       error
		totalHits int64
	)
	// show查询结果结构化
	ses, err = doTransShowResult(res)
	if err != nil {
		// l.Errorf(`doTransResult %d:%s %s`, i, astRes.AST.Q, err.Error())
		return nil, err
	}
	if len(ses) == 0 {
		ses = nil
	}
	return &QueryResult{
		Series:    ses,
		Cost:      fmt.Sprintf(`%v`, time.Millisecond*time.Duration(res.TookInMillis)),
		Totalhits: totalHits,
	}, nil
}

// 返回ses验证
func validSes(ses []models.Row) ([]models.Row, error) {
	nSes := []models.Row{}
	for _, sPtr := range ses {
		if len(sPtr.Columns) > 0 && len(sPtr.Values) > 0 {
			nSes = append(nSes, sPtr)
		}
	}
	return nSes, nil
}

//sortSes ES 字段排序
func sortSes(esResPtr *parser.ESTRes, ses []models.Row) []models.Row {
	sortFields := esResPtr.SortFields
	lenFields := len(sortFields)

	if lenFields > 0 {
		for i, sPtr := range ses {
			// 返回值的顺序
			sortIndex := []int{}
			for range sortFields {
				sortIndex = append(sortIndex, -1)
			}

			// 返回顺序转为查询顺序
			for fi, fv := range sortFields {
				fExist := false
				for ci, cv := range sPtr.Columns {
					if fv == cv {
						sortIndex[fi] = ci
					}
				}
				if fExist == false {
					//error
				}
			}
			// 变更返回的columns
			ses[i].Columns = sortFields

			// 对value结果转换顺序
			for vi, vv := range sPtr.Values {

				sValues := []interface{}{}
				for range sortFields {
					sValues = append(sValues, nil)
				}
				for fi, fv := range sortIndex {
					if fv > -1 {
						sValues[fi] = vv[fv]
					} else {
						sValues[fi] = nil
					}
				}
				ses[i].Values[vi] = sValues
			}

		}
	}
	return ses
}

//aliasSes ES字段别名变更
func aliasSes(esResPtr *parser.ESTRes, ses []models.Row) []models.Row {
	for j, sPtr := range ses {
		ses[j].Name = esResPtr.ClassNames // 添加分类名称
		for ci, c := range sPtr.Columns {
			if ac, ok := esResPtr.Alias[c]; ok {
				ses[j].Columns[ci] = ac
			}
		}
	}
	return ses
}

// 多个first，last函数，多行返回结果合并为一行
func multiRowSes(esResPtr *parser.ESTRes, ses []models.Row) []models.Row {
	if esResPtr.FLFuncCount > 1 {
		for i, item := range ses {
			nRow := mergeRowValues(item)
			ses[i] = nRow
		}
	}
	return ses
}

// 具体合并values
func mergeRowValues(row models.Row) models.Row {
	nValue := make([]interface{}, len(row.Columns))
	for _, v := range row.Values {
		for j, fieldValue := range v {
			if fieldValue != nil && nValue[j] == nil {
				nValue[j] = fieldValue // 取后面的非nil值
			}
		}
	}
	row.Values = [][]interface{}{nValue}
	return row
}

// aggsPagingSes 聚合分页
func aggsPagingSes(esResPtr *parser.ESTRes, ses []models.Row) []models.Row {
	size := esResPtr.AggsFromSize
	rSize := len(ses)
	if size > 0 {
		if size < rSize {
			ses = ses[esResPtr.AggsFromSize:]
		} else {
			ses = nil
		}
	}
	return ses

}

// show查询结果结构化
func doTransShowResult(sr *elastic.SearchResult) ([]models.Row, error) {
	var res = []models.Row{}
	var values = [][]interface{}{}
	var err error

	if sr.Aggregations != nil {
		if aggs1, ok := sr.Aggregations["aggs1"]; ok {
			var outer map[string]interface{}
			err = json.Unmarshal(aggs1, &outer)
			if err != nil {
				return res, err
			}
			if bucketsItem, ok := outer["buckets"]; ok {
				buckets, _ := bucketsItem.([]interface{})
				for _, bucketItem := range buckets {
					bucket, _ := bucketItem.(map[string]interface{})
					values = append(values, []interface{}{bucket["key"]})
				}
			}
		}

	}
	var seriesItem = models.Row{
		Name:    "measurements",
		Columns: []string{"name"},
		Values:  values,
	}
	res = append(res, seriesItem)
	return res, nil
}

// es 结果集转化为 influxdb 时间线格式
func (astRes *ASTResult) doTransResult(sr *elastic.SearchResult) ([]models.Row, error) {
	// series := []*Series{}
	series := []models.Row{}

	// hits  basic query result
	if sr.Hits != nil && len(sr.Hits.Hits) > 0 {
		serie, err := astRes.hitSourceTransform(sr)
		if err != nil {
			l.Errorf(`dql:%+#v, %s`, sr, err.Error())
			return nil, err
		}

		series = append(series, *serie)
	}

	// aggs results
	aggses, err := aggsTrans(sr)
	if err != nil {

		l.Errorf(`aggsTrans: %s`, err.Error())
		return nil, err
	}

	if len(aggses) > 0 {
		series = append(series, aggses...)
	}

	return series, nil
}

// time 字段 第一列处理
func doColumnToTimeFirst(columns []string) []string {
	cols := []string{}

	exitTime := false
	for _, col := range columns {
		switch col {
		case utils.Times, utils.EsKeepTimeStampM:
			exitTime = true
		default:
			cols = append(cols, col)
			sort.Strings(cols)
		}

	}

	res := []string{}
	if exitTime {
		res = append(res, utils.Times)
	}

	res = append(res, cols...)
	return res

}

// __source column 采集
// TODO  不同层级的field key 相同名字 ？ 带路径 ?
func walkSourceCols(expr map[string]interface{}, columns []string) ([]string, error) {

	for k, v := range expr {
		switch v.(type) {
		case map[string]interface{}:

			cols, err := walkSourceCols(v.(map[string]interface{}), columns)
			if err != nil {
				return nil, err
			}

			for _, col := range cols {
				if !utils.ContainsValue(col, columns) {
					columns = append(columns, col)
				}
			}

		default:
			if !utils.ContainsValue(k, columns) {
				columns = append(columns, k)
			}

		}
	}

	return columns, nil

}

// __source , values采集
func walkSourceVals(expr map[string]interface{}, tg map[string]interface{}) (map[string]interface{}, error) {

	for k, v := range expr {
		switch v.(type) {
		case map[string]interface{}:

			vals, err := walkSourceVals(v.(map[string]interface{}), tg)
			if err != nil {
				l.Errorf(`%s`, err.Error())
				return nil, err
			}

			for valsk, valsv := range vals {
				tg[valsk] = valsv
			}

		default:
			tg[k] = v
		}
	}

	return tg, nil
}

// basic query 示例, 结果集 _source 提取解析
//"hits" : [
// {
//     "_index" : "wksp_e2db2d02837a11eaba9a8671df186910_rum-000001",
//     "_type" : "_doc",
//     "_id" : "L-m5PHYB7hvW6zoN8nio",
//     "_score" : 1.0,
//     "_source" : {
//       "__meta" : {
//         "__esCreateTime" : 1607336587943
//       },
// 	  	 "__tags" :
// 	  	 {
// 		  ....
// 	  	 }
// 	  }
// }]
func (astRes *ASTResult) hitSourceTransform(source *elastic.SearchResult) (*models.Row, error) {
	series := &models.Row{}
	columns := []string{}

	if source.Hits.TotalHits.Value <= 0 {
		l.Warnf(`No Data`)
		return nil, nil
	}

	for _, h := range source.Hits.Hits {
		source := map[string]interface{}{}
		err := json.Unmarshal(h.Source, &source)
		if err != nil {
			l.Errorf(`%s`, err.Error())
			return nil, err
		}

		// first, walk for columns
		columns, err = walkSourceCols(source, columns)

		if err != nil {
			return nil, err
		}

	}

	// 添加高亮字段
	dfQuery, ok := (astRes.AST).(*parser.DFQuery)
	if !ok {
		return nil, fmt.Errorf("es translate error")
	}
	eHighlights := []string{}

	if dfQuery.Helper != nil && dfQuery.Helper.ESTResPtr != nil {
		if dfQuery.Helper.ESTResPtr.HighlightFields != nil && len(dfQuery.Helper.ESTResPtr.HighlightFields) > 0 {
			eHighlights = dfQuery.Helper.ESTResPtr.HighlightFields
		}
	}
	isHighlight := len(eHighlights) > 0

	// field values list
	for _, h := range source.Hits.Hits {
		source := map[string]interface{}{}
		err := json.Unmarshal(h.Source, &source)
		if err != nil {
			l.Errorf(`%s`, err.Error())
			return nil, err
		}

		target := map[string]interface{}{}
		walkSourceVals(source, target)
		values := []interface{}{}

		// 添加对于highlight的解析
		if isHighlight {
			highlights := h.Highlight
			for _, col := range columns {
				// highlight fields
				hv, ok := highlights[col]
				if ok {
					if len(hv) > 0 {
						values = append(values, hv[0])
					} else {
						values = append(values, target[col])
					}
				} else {
					values = append(values, target[col])
				}
			}
		} else {
			for _, col := range columns {
				values = append(values, target[col])
			}
		}

		if len(values) > 0 {
			series.Values = append(series.Values, values)
		}
	}

	series.Columns = columns

	return series, nil
}

// agg query, 结果集解析
func aggsTrans(source *elastic.SearchResult) ([]models.Row, error) {
	sds := []*DataSingleDoc{}
	var err error

	sNoAg := DataSingleDoc{
		Values: []*Value{},
	}

	for k, v := range source.Aggregations {

		var aggs interface{}
		err = json.Unmarshal(v, &aggs)
		if err != nil {

			l.Errorf(`%s`, err.Error())
			return nil, err
		}

		switch aggs.(type) {
		case map[string]interface{}:
			sd := DataSingleDoc{
				Tags:   []*Tag{},
				Values: []*Value{},
			}

			ss, err := doAggsTransform(aggs.(map[string]interface{}), k, &sd)
			if err != nil {
				l.Errorf(`%s`, err.Error())
				return nil, err
			}

			// 非桶聚合，metric /pipeline aggs 处理
			if !HasBuckets(aggs.(map[string]interface{})) {

				sNoAg.Values = append(sNoAg.Values, sd.Values...)

			} else { // buckets  aggs 处理
				sds = append(sds, ss...)
			}

		default:
			// TODO NOT SUPPORT
		}

	}

	if len(sNoAg.Values) > 0 {
		sds = append(sds, &sNoAg)
	}

	series, err := DocTransForSeries(sds)
	if err != nil {

		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	return series, nil
}

// tags 计算md5值
func doTagsMd5(source []*Tag) (string, error) {
	tkeys := []string{}
	for _, t := range source {
		tkeys = append(tkeys, t.Key)
	}

	target := []*Tag{}
	sort.Strings(tkeys)
	for _, tkey := range tkeys {
		for _, t := range source {
			if t.Key == tkey {
				target = append(target, t)
				break
			}
		}
	}

	data, err := json.Marshal(target)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return "", err
	}

	digest := md5.New()
	digest.Write(data)
	md5Cur := fmt.Sprintf("%x", digest.Sum(nil))

	return md5Cur, nil
}

// 中间结构转时间线，按时间线 tags 分组，组合时间线
func DocTransForSeries(source []*DataSingleDoc) ([]models.Row, error) {
	ses := []models.Row{}
	target := map[string]SValues{}
	mtags := map[string]STags{}

	for _, dsd := range source {
		md5Key, err := doTagsMd5(dsd.Tags)
		if err != nil {

			l.Errorf(`%s`, err.Error())
			return nil, err
		}

		if _, ok := mtags[md5Key]; !ok {
			mtags[md5Key] = dsd.Tags
		}

		_, ok := target[md5Key]
		if !ok {
			target[md5Key] = SValues{}
		}

		//相同的tags，则应该属于同一时间线
		target[md5Key] = append(target[md5Key], dsd.Values)
	}

	// column 计算
	for tgk, tgs := range mtags {

		columns := []string{}
		vals := [][]interface{}{}

		for _, colv := range target[tgk] {
			for _, col := range colv {

				if !utils.ContainsValue(col.Key, columns) {
					columns = append(columns, col.Key)
				}
			}
		}

		//values 填充
		for _, colv := range target[tgk] {

			values := []interface{}{}
			for _, col := range columns {

				exist := false
				for _, val := range colv {
					if val.Key == col {
						values = append(values, val.Value)
						exist = true
						break
					}
				}

				if !exist {
					values = append(values, nil)
				}

			}

			if len(values) > 0 {
				vals = append(vals, values)
			}
		}

		tags := map[string]string{}
		for _, t := range tgs {
			tags[t.Key] = t.Value
		}

		ses = append(ses, models.Row{
			Tags:    tags,
			Columns: columns,
			Values:  vals,
		})

	}

	return ses, nil
}

//判断是否是桶聚合的返回结果
func HasBuckets(source map[string]interface{}) bool {
	for k := range source {
		switch k {
		case `buckets`, `hits`:
			return true
		default:
		}
	}
	return false
}

// 聚合结果转换，主要区分为桶聚合和非桶聚合
// metric aggs/pipeline aggs:
// 		"aggs": {
// 			"agg_alias": {
// 				"value": xxx,
// 	or			"values": {
// 					"key1": xxx,
// 					"key2": xxx,
// 					....
// 				}
// or  			"max": xxx,
// 				"min": xxx,
// 				...
// bucket aggs
// or 			"buckets": {....}
// 				}
// 			}
//
func doAggsTransform(source map[string]interface{}, key string, sd *DataSingleDoc) ([]*DataSingleDoc, error) {
	sds := []*DataSingleDoc{}

	for k, v := range source {
		switch k {
		case `value`, `doc_count`: // 单字段聚合时

			newV := &Value{
				Key:   key,
				Value: v,
			}
			//TODO

			sd.Values, _ = doAppendVals(newV, sd.Values, false)

			//sds = append(sds, &sd)

		case `values`: //单字段 percents 聚合
			//
			for vk, vv := range v.(map[string]interface{}) {

				newV := &Value{
					Key:   key + `_` + vk,
					Value: vv,
				}

				//TODO
				sd.Values, _ = doAppendVals(newV, sd.Values, false)
			}

			//sds = append(sds, &sd)
		case `doc_count_error_upper_bound`, `sum_other_doc_count`: // 固定的不使用字段
		case `from`, `to`: //TODO
		case `hits`: //top bottom

			hits, err := doHitTransf(v.(map[string]interface{}))
			if err != nil {
				l.Errorf(`%s`, err.Error())
				return nil, err
			}

			for _, vl := range hits {

				newSd := &DataSingleDoc{
					Tags:   sd.Tags,
					Values: sd.Values,
				}

				for vlk, vlv := range vl {

					newSd.Values = append(newSd.Values, &Value{
						Key:   vlk,
						Value: vlv,
					})

				}

				sds = append(sds, newSd)

			}

		case `buckets`: //桶聚合
			sdcs, err := doBucketsTransform(v.([]interface{}), key, sd)
			if err != nil {

				l.Errorf(`%s`, err.Error())
				return nil, err
			}

			sds = append(sds, sdcs...)

		case `after_key`:
			//TODO  afterKey  map[string]interface{}

		default:

			// 非桶聚合 普通情况
			switch v.(type) {
			case map[string]interface{}:
				return doAggsTransform(v.(map[string]interface{}), key+`_`+k, sd)

			default:

				newV := &Value{
					Key:   key + `_` + k,
					Value: v,
				}
				//TODO
				sd.Values, _ = doAppendVals(newV, sd.Values, false)
			}

			//sds = append(sds, &sd)

		}
	}

	return sds, nil
}

//是否存在子聚合结果集
func HasSubAgger(source map[string]interface{}) bool {
	for k := range source {
		switch k {
		case `key`, `doc_count`:
		default:
			return true
		}
	}

	return false
}

// 相同列名时，values处理，发现同名，加后缀_1,存在，则变为_2,... 依次类推
func doAppendVals(newv *Value, source []*Value, sameKeyIgnore bool) ([]*Value, error) {

	if newv == nil || newv.Key == `` || newv.Value == `` || newv.Value == nil {
		return source, nil
	}

	target := []*Value{}
	keys := []string{}
	for _, v := range source {
		keys = append(keys, v.Key)
	}

	var err error
	target = append(target, source...)

	if !utils.ContainsValue(newv.Key, keys) {
		target = append(target, newv)
		return target, nil
	}

	if sameKeyIgnore {
		return source, nil
	}

	index := 0

	ms := strings.Split(newv.Key, `_`)

	if len(ms) == 2 {
		index, err = strconv.Atoi(ms[len(ms)-1])
		if err != nil {
			return nil, err
		}
	}

	index++

	keyNew := fmt.Sprintf(`%s_%d`, ms[0], index)
	newv.Key = keyNew

	return doAppendVals(newv, source, false)

}

// 桶聚合结果集
// {
// 	"key" : "zipkin",
// 	"doc_count" : 10543,
// 	"__serviceName" : {   // 子聚合
// 	  "doc_count_error_upper_bound" : 0,
// 	  "sum_other_doc_count" : 0,
//    .....
//   }
// }
func doBucketsTransform(source []interface{}, key string, sd *DataSingleDoc) ([]*DataSingleDoc, error) {
	res := []*DataSingleDoc{}

	for _, b := range source {
		singleDoc := DataSingleDoc{
			Tags:   []*Tag{},
			Values: []*Value{},
		}

		vals := []*Value{}
		tags := []*Tag{}

		// key 一般都是作为tag ,先处理
		for k, v := range b.(map[string]interface{}) {

			vals = append(vals, sd.Values...)

			tags = append(tags, sd.Tags...)

			switch k {
			case `key`:

				tgs, vl := doKeyTransf(v, key)

				var err error
				vals, err = doAppendVals(vl, vals, false)
				if err != nil {
					l.Errorf(`%s`, err.Error())
				}

				if HasSubAgger(b.(map[string]interface{})) {
					tags = append(tags, tgs...)
				} else {
					for _, t := range tgs {
						vl := &Value{
							Key:   t.Key,
							Value: t.Value,
						}

						vals, err = doAppendVals(vl, vals, false)
						if err != nil {
							l.Errorf(`%s`, err.Error())
						}

					}
				}

				break
			}
		}

		eTag := false //非桶聚合的标志

		rs := []*DataSingleDoc{}

		singleDoc.Tags = tags
		singleDoc.Values = vals

		//value 再依次处理
		for k, v := range b.(map[string]interface{}) {

			switch k {
			case `key`, `key_as_string`:
			case `doc_count`: // 有子聚合时，忽略不计
				if !HasSubAgger(b.(map[string]interface{})) {

					res = append(res, &DataSingleDoc{
						Tags:   tags,
						Values: vals,
					})
				}

			default:

				if HasSubAgger(b.(map[string]interface{})) {

					rsd, err := doAggsTransform(v.(map[string]interface{}), k, &singleDoc)
					if err != nil {

						l.Errorf(`%s`, err.Error())
						return nil, err
					}

					if len(rsd) > 0 {
						rs = append(rs, rsd...)
					} else {
						eTag = true
					}

				}

			}
		}

		if eTag { // 非桶聚合情况下的，结果集组values,水平扩展；桶聚合，则垂直扩展
			if len(rs) == 0 { //
				res = append(res, &singleDoc)
			} else {
				for _, r := range rs {
					for _, s := range singleDoc.Values {
						r.Values, _ = doAppendVals(s, r.Values, true)
					}
				}

				res = append(res, rs...)
			}
		} else {
			res = append(res, rs...)
		}
	}

	return res, nil
}

// top/botton hit 处理
// "metric-top_hits-4-0" : {
// 	"hits" : {
// 	  "total" : {
// 		"value" : 10543,
// 		"relation" : "eq"
// 	  },
// 	  "max_score" : 2.2356224,
// 	  "hits" : [
// 		{
// 		  "_index" : "wksp_e2db2d02837a11eaba9a8671df186910_logging-000007",
// 		  "_type" : "_doc",
// 		  "_id" : "pONa_XQBn_tsqfhKHzJd",
// 		  "_score" : 2.2356224,
// 		  "_source" : {
// 			"__timestampUs" : 1601978320247848
// 		  }
// 		}
// 	  ]
// 	}
// }

func doHitTransf(source map[string]interface{}) ([]map[string]interface{}, error) {

	sr, err := json.Marshal(source)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	var hts elastic.SearchHits

	err = json.Unmarshal(sr, &hts)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	vals := []map[string]interface{}{}
	for _, ht := range hts.Hits {
		source := map[string]interface{}{}
		err := json.Unmarshal(ht.Source, &source)
		if err != nil {
			l.Errorf(`%s`, err.Error())
			return nil, err
		}

		target := map[string]interface{}{}
		walkSourceVals(source, target)

		vals = append(vals, target)
	}

	return vals, nil

}

// key 处理，通常作为tag, 目前业务需求数值型是作为field value
func doKeyTransf(source interface{}, key string) (tags []*Tag, val *Value) {

	switch source.(type) {

	case map[string]interface{}:
		for k, v := range source.(map[string]interface{}) {
			tags = append(tags, &Tag{
				Key:   k,
				Value: v.(string),
			})
			//tags[k] = v.(string)
		}

	case string:
		tags = append(tags, &Tag{
			Key:   key,
			Value: source.(string),
		})

		//tags[key] = source.(string)

	default:

		val = &Value{
			Key:   key,
			Value: source,
		}

	}

	return tags, val
}

// walk 递归调用
func doTransForm(source map[string]interface{}, target map[string]interface{}) {
	for k, v := range source {
		switch v.(type) {
		case map[string]interface{}:
		default:
			target[k] = v
		}
	}
}

// walk 示例
func doSourceTrans(source map[string]interface{}) map[string]interface{} {
	target := map[string]interface{}{}
	doTransForm(source, target)
	for _, v := range source {
		switch v.(type) {
		case map[string]interface{}:
			ts := walkSource(v.(map[string]interface{}), func(s map[string]interface{}, t map[string]interface{}) {
				doTransForm(s, target)
			})

			for tsk, tsv := range ts {
				target[tsk] = tsv
			}
		default:
		}
	}
	return target
}

type walkCol func(map[string]interface{}, map[string]interface{})

func walkSource(expr map[string]interface{}, f walkCol) map[string]interface{} {

	target := map[string]interface{}{}
	for _, v := range expr {
		switch v.(type) {
		case map[string]interface{}:
			expr = walkSource(v.(map[string]interface{}), f)

		default:
			f(expr, target)
		}
	}

	return expr
}

// esShowColumnsToInflux
func esShowColumnsToInflux(res [][]interface{}) (*QueryResult, error) {

	// (1) 查询结果结构化(近似influxdb)
	ses := []models.Row{}
	if len(res) > 0 {
		var row models.Row
		row.Name = "fields"
		row.Columns = []string{"fieldKey", "fieldType"}
		row.Values = res
		ses = append(ses, row)
	}
	return &QueryResult{
		Series: ses,
		// Cost:   fmt.Sprintf(`%v`, time.Millisecond*time.Duration(res.TookInMillis)),
	}, nil

}
