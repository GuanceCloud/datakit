package dql

import (
	"bytes"
	"context"
	"fmt"
	"time"

	imodels "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/es"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/models"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/rtpanic"
)

type queryWorker struct {
	idx        int
	influxClis map[string]*influxQueryCli // cache influx-instance-id -> influx-client
	influxDBs  map[string][]string        // cache workspace-id -> influx-dbname,influx-instance-id
	esCli      *es.EsCli
	dqlparam   *parser.ExtraParam
}

type DQLResults struct {
	Ses []*QueryResult
	Msg string
}

func (qw *queryWorker) getInfluxQueryCli(wsid string) (*influxQueryCli, error) {
	var influxID string

	if _, ok := qw.influxDBs[wsid]; !ok {

		ifdb, err := models.QueryWorkspaceInfoReadOnly(wsid) // query influx instance info of @wsid

		if err != nil {
			l.Errorf("load dbinfo(uuid: %s) failed: %s", wsid, err.Error())
			return nil, err
		}

		influxID = ifdb.InfluxInstanceUUID

		qw.influxDBs[wsid] = []string{ifdb.DB, influxID}

		if _, ok := qw.influxClis[influxID]; !ok { // instance not connected before
			cli, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{ // create new influx client
				Addr:               ifdb.Host,
				Username:           ifdb.User,
				Password:           ifdb.Pwd,
				UserAgent:          "dql-query",
				Timeout:            time.Duration(config.C.Influx.ReadTimeOut) * time.Second,
				InsecureSkipVerify: true,
			})

			if err != nil {
				l.Error(err)
				return nil, err
			}

			qw.influxClis[influxID] = &influxQueryCli{cli: cli}

			l.Debugf("[%d] new client %s:%+#v", qw.idx, influxID, qw.influxClis[influxID])
		}
	}

	return qw.influxClis[qw.influxDBs[wsid][1]], nil
}

func (qw *queryWorker) initEsCli() (err error) {
	qw.esCli, err = es.InitEsCli(config.C.Es.Host,
		config.C.Es.User,
		config.C.Es.Passwd,
		config.C.Es.TimeOut)
	return
}

func (qw *queryWorker) run(idx int) {

	l.Debugf("query worker %d started.", idx)

	if err := qw.initEsCli(); err != nil {
		l.Errorf("InitEs(): %s", err.Error())
		return
	}

	var f rtpanic.RecoverCallback

	var lastIQ *InnerQuery
	panicCnt := 0

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)

		if trace != nil || err != nil {
			panicCnt++
			l.Warnf("[%d] recoverd from panic. panic: %v\nstack trace:\n%s", panicCnt, err, string(bytes.TrimSpace(trace)))
			if lastIQ != nil {
				select {
				case lastIQ.result <- uhttp.Errorf(ErrQueryWorkerCrashed, "trace: %s", string(trace)):
				default:
				}
			}
		}

		for {
		start:
			select {
			case iq := <-qch:

				lastIQ = iq

				l.Debugf("query: %+#v", iq)

				var response []*QueryResult

				for _, query := range iq.Queries {
					pres, err := qw.parseQuery(query)
					if err != nil {
						iq.result <- uhttp.Errorf(ErrParseError, "parse error: %s", err.Error())
						goto start
					}

					results, err := qw.runQueries(iq, pres)

					if err != nil {
						iq.result <- uhttp.Errorf(ErrQueryError, "query error: %s", err.Error())
						goto start
					}

					response = append(response, results...)

				}

				iq.result <- response
			}
		}
	}

	f(nil, nil)
}

// ParseQuery parse query without qw
func ParseQuery(sq *singleQuery) (ASTResults, error) {
	dqlparam, err := newDQLParamWithQuery(sq)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	parseResults, err := Parse(sq.Query, dqlparam)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	if len(parseResults) > maxDQLParseResult {
		return nil, fmt.Errorf("do not query larger than %d", maxDQLParseResult)
	}

	return parseResults, nil

}

func (qw *queryWorker) parseQuery(sq *singleQuery) (ASTResults, error) {
	dqlparam, err := newDQLParamWithQuery(sq)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	qw.dqlparam = dqlparam

	parseResults, err := Parse(sq.Query, qw.dqlparam)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	if len(parseResults) > maxDQLParseResult {
		return nil, fmt.Errorf("do not query larger than %d", maxDQLParseResult)
	}

	return parseResults, nil
}

func (qw *queryWorker) runQueries(iq *InnerQuery, res ASTResults) ([]*QueryResult, error) {
	var datas []*QueryResult

	for _, ast := range res {
		data, err := qw.runQuery(iq.WorkspaceUUID, ast, iq.EchoExplain)
		if err != nil {
			return nil, err
		}

		if data != nil {
			datas = append(datas, data)
		}
	}

	return datas, nil
}

func (qw *queryWorker) runQuery(wsid string, ast *ASTResult, explain bool) (*QueryResult, error) {
	data, err := qw.runSingleQuery(wsid, ast, explain)
	if err != nil {
		return nil, err
	}

	if data == nil {
		l.Debugf("got no data on query: %s", ast.Q)
		return nil, nil
	}

	switch ast.Namespace {
	case NSFunc, NSFuncAbbr:
		// pass
	default:
		switch v := ast.AST.(type) {
		case *parser.DFQuery:
			data.GroupByList = v.GroupByList()
			/*
			 *  -> rewrite query results <-
			 */
			if err := RewriteResults(v, data.Series); err != nil {
				return nil, err
			}
		}
	}
	return data, nil
}

func (qw *queryWorker) runSingleQuery(wsid string, ast *ASTResult, explain bool) (*QueryResult, error) {
	l.Debugf("[%d] namespace: %s, query: %s", qw.idx, ast.Namespace, ast.Q)

	switch ast.Namespace {
	case NSMetric:
		qcli, err := qw.getInfluxQueryCli(wsid)
		if err != nil {
			return nil, err
		}
		if qcli == nil {
			l.Fatalf("[%d] qcli nil: %+#v", qw.idx, qw.influxClis)
		}

		qr, err := qw.query(ast.Q, qw.influxDBs[wsid][0], DefaultRP, InfluxReadPrecision, qcli)
		if err != nil {
			return nil, err
		}
		if explain {
			qr.RawQuery = ast.Q
		}

		return qr, nil

	case NSObject, NSLogging, NSEvent, NSTracing, NSRum, NSSecurity, NSBackupLogging:
		qrs, err := qw.esQuery(ast, wsid, explain)
		if err != nil {
			return nil, err
		}

		return qrs, nil

	case NSFunc, NSFuncAbbr:
		qr, err := qw.funcQuery(wsid, ast, qw.dqlparam, explain)
		if err != nil {
			return nil, err
		}
		return qr, nil

	case NSLambda:
		return qw.lambdaQuery(wsid, ast, explain)

	case NSOuterFunc: // outer func

		qr, err := qw.outerFuncQuery(wsid, ast, explain)
		if err != nil {
			return nil, err
		}
		return qr, nil

	case NSDeleteFunc: // outer delete func

		qr, err := qw.outerDeleteQuery(wsid, ast, explain)
		if err != nil {
			return nil, err
		}
		return qr, nil

	}

	return nil, fmt.Errorf("invalid namespace `%s'", ast.Namespace)
}

type influxQueryCli struct {
	cli influxdb.Client
}

func (qw *queryWorker) query(q, db, rp, precision string, c *influxQueryCli) (*QueryResult, error) {
	influxq := influxdb.Query{
		Command:         q,
		Database:        db,
		RetentionPolicy: rp,
		Precision:       precision,
		Parameters:      nil,
		Chunked:         false,
		ChunkSize:       0,
	}

	start := time.Now()

	res, err := c.cli.Query(influxq)
	if err != nil {
		l.Errorf("influxdb api error %s", err.Error())
		return nil, err
	}

	elapsed := time.Since(start)

	ses := []imodels.Row{}
	for _, r := range res.Results {
		ses = append(ses, r.Series...)
	}

	return &QueryResult{
		Series: ses,
		Cost:   fmt.Sprintf("%v", elapsed),
	}, nil
}

// ES: query & convert ES result -> Series
func (qw *queryWorker) esQuery(ast *ASTResult,
	wsID string,
	explain bool) (*QueryResult, error) {

	var qres *QueryResult // influx结构化后返回值

	indexName := ``
	switch ast.Namespace {
	case `object`, `O`:
		indexName = wsID + `_object`
	case `logging`, `L`:
		indexName = wsID + `_logging`
	case `keyevent`, `E`, `event`:
		indexName = wsID + `_keyevent`
	case `tracing`, `T`:
		indexName = wsID + `_tracing`
	case `rum`, `R`:
		indexName = wsID + `_rum`
	case `security`, `S`:
		indexName = wsID + `_security`
	default:
		l.Errorf(`No Support`)
		return nil, fmt.Errorf("no support namespace")
	}

	showAst, ok := ast.AST.(*parser.Show)
	if ok && showAst.Helper.ESTResPtr.ShowFields { // 满足show函数 且是 show fields
		res, err := qw.esCli.XPackSQL(indexName)
		if err != nil {
			err := qw.esCli.ErrorHandler(err) // 分析且格式化 es err
			return nil, err
		}
		timeField := showAst.Helper.ESTResPtr.TimeField
		// 查询是否在缓存中
		// showRes := WarmUpData.GetShowFieldsRes(indexName, ast.Q)
		// if showRes != nil {
		// 	return showRes, nil
		// }
		indexNames := []string{indexName}
		// indexNames := GetShowRealIndexNames(ast.AST, indexName)
		existRes, err := qw.esCli.MSearchFieldExists(indexNames, timeField, ast.Q, res)
		// existRes, err := qw.esCli.MSearchFieldExistsWithoutQuery(indexName, timeField, ast.Q, res)
		if err != nil {
			return nil, err
		}
		qres, err = esShowColumnsToInflux(existRes)
		// WarmUpData.UpdateShowFields(indexName, ast.Q, qres)
	} else {
		// indexNames := GetRealIndexNames(ast.AST, indexName)
		indexNames := []string{indexName}
		esRes, err := qw.esCli.Es.Search().
			Timeout(qw.esCli.TimeOut).
			Index(indexNames...).
			Source(ast.Q).
			Do(context.Background())

		if err != nil {
			nerr := qw.esCli.ErrorHandler(err) // 分析且格式化 es err
			l.Errorf(`dql:%s, indexName:%s, %s`, ast.Q, indexName, err.Error())
			return nil, nerr
		}

		if esRes.Error != nil {

			l.Error(`ES query result error: %s`, esRes.Error)
			return nil, fmt.Errorf(esRes.Error.Reason)
		}

		qres, err = es2influx(esRes, ast)
		if err != nil {
			return nil, err
		}
		// 添加search_after
		queryAST, ok := ast.AST.(*parser.DFQuery)
		if ok && queryAST.SearchAfter != nil {
			searchAfterRes, err := esSearchAfterRes(esRes, ast)
			if err != nil {
				return qres, err
			}
			qres.SearchAfter = searchAfterRes
		}
	}

	if explain {
		qres.RawQuery = ast.Q
	}

	return qres, nil
}

func (qw *queryWorker) runQueryDebug(idx int) {
	l.Debugf("query worker %d started.", idx)

	if err := qw.initEsCli(); err != nil {
		l.Errorf("InitEs(): %s", err.Error())
		return
	}

	var f rtpanic.RecoverCallback

	var lastIQ *InnerQuery
	panicCnt := 0

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)

		if trace != nil || err != nil {
			panicCnt++
			l.Warnf("[%d] recoverd from panic. panic: %v\nstack trace:\n%s", panicCnt, err, string(bytes.TrimSpace(trace)))
			if lastIQ != nil {
				select {
				case lastIQ.result <- uhttp.Errorf(ErrQueryWorkerCrashed, "trace: %s", string(trace)):
				default:
				}
			}
		}

		for {
			select {
			case iq := <-qchDebug:
				l.Debugf("query: %+#v", iq)

				wsID := iq.WorkspaceUUID

				switch iq.Namespace {
				case "metric":
					qcli, err := qw.getInfluxQueryCli(wsID)
					if err != nil {
						iq.result <- uhttp.Errorf(ErrQueryError, "query error: %s", err)
						continue
					}

					if qcli == nil {
						l.Fatalf("[%d] qcli nil: %+#v", qw.idx, qw.influxClis)
					}

					influxq := influxdb.Query{
						Command:         iq.Query,
						Database:        qw.influxDBs[wsID][0],
						RetentionPolicy: DefaultRP,
						Precision:       InfluxReadPrecision,
						Parameters:      nil,
						Chunked:         false,
						ChunkSize:       0,
					}

					response, err := qcli.cli.Query(influxq)
					if err != nil {
						iq.result <- uhttp.Errorf(ErrQueryError, "influxdb api error %s", err)
						continue
					}
					iq.result <- response

				default:
					indexName := ``
					switch iq.Namespace {
					case `object`:
						indexName = wsID + `_object`
					case `logging`:
						indexName = wsID + `_logging`
					case `keyevent`:
						indexName = wsID + `_keyevent`
					case `tracing`:
						indexName = wsID + `_tracing`
					case `rum`:
						indexName = wsID + `_rum`
					default:
						iq.result <- uhttp.Errorf(ErrQueryError, "invalid namespace: %s", iq.Namespace)
						continue
					}

					esRes, err := qw.esCli.Es.Search().
						Timeout(qw.esCli.TimeOut).
						Index(indexName).
						Source(iq.Query).
						Do(context.Background())

					if err != nil {
						err = qw.esCli.ErrorHandler(err) // 分析且格式化 es err
						iq.result <- uhttp.Errorf(ErrQueryError, "ES query result error: %s", err)
						continue
					}

					if esRes.Error != nil {
						iq.result <- uhttp.Errorf(ErrQueryError, "ES query result error: %s", esRes.Error)
						continue
					}

					iq.result <- esRes
				}
			}
		}
	}

	f(nil, nil)
}

// GetRealIndexNames 获取对应时间范围内的索引列表
func GetRealIndexNames(ast interface{}, indexName string) []string {
	res := []string{}
	rAst, ok := ast.(*parser.DFQuery)
	if !ok {
		res = append(res, indexName)
		return res
	}

	helper := rAst.Helper.ESTResPtr
	if helper.StartTime != 0 && helper.EndTime != 0 {
		indices := WarmUpData.GetTimeRangeIndices(indexName, helper.StartTime, helper.EndTime)
		if len(indices) > 0 {
			res = indices
		} else {
			res = append(res, indexName)
		}
	}
	if len(res) == 0 {
		res = append(res, indexName)
	}
	return res

}

// GetShowRealIndexNames 获取对应时间范围内的索引列表
func GetShowRealIndexNames(ast interface{}, indexName string) []string {
	res := []string{}
	rAst, ok := ast.(*parser.Show)
	if !ok {
		res = append(res, indexName)
		return res
	}

	helper := rAst.Helper.ESTResPtr
	if helper.StartTime != 0 && helper.EndTime != 0 {
		indices := WarmUpData.GetTimeRangeIndices(indexName, helper.StartTime, helper.EndTime)
		if len(indices) > 0 {
			res = indices
		} else {
			res = append(res, indexName)
		}
	}
	if len(res) == 0 {
		res = append(res, indexName)
	}
	return res

}

// runBackup run query backup
func (qw *queryWorker) runBackup(idx int) {

	l.Debugf("query backup worker %d started.", idx)

	if err := qw.initEsCli(); err != nil {
		l.Errorf("InitEs(): %s", err.Error())
		return
	}

	var f rtpanic.RecoverCallback

	var lastIQ *InnerQueryBackup
	panicCnt := 0

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)

		if trace != nil || err != nil {
			panicCnt++
			l.Warnf("[%d] recoverd from panic. panic: %v\nstack trace:\n%s", panicCnt, err, string(bytes.TrimSpace(trace)))
			if lastIQ != nil {
				select {
				case lastIQ.result <- uhttp.Errorf(ErrQueryWorkerCrashed, "trace: %s", string(trace)):
				default:
				}
			}
		}

		for {
		start:
			select {
			case iq := <-qchBackup:

				lastIQ = iq

				l.Debugf("query: %+#v", iq)

				var response []*QueryResult

				for _, query := range iq.Queries {
					pres, err := qw.parseQuery(query)
					if err != nil {
						iq.result <- uhttp.Errorf(ErrParseError, "parse error: %s", err.Error())
						goto start
					}

					results, err := qw.runBackupQueries(iq, pres)

					if err != nil {
						iq.result <- uhttp.Errorf(ErrQueryError, "query error: %s", err.Error())
						goto start
					}

					response = append(response, results...)

				}

				iq.result <- response
			}
		}
	}

	f(nil, nil)
}

func (qw *queryWorker) runBackupQueries(iq *InnerQueryBackup, res ASTResults) ([]*QueryResult, error) {
	var datas []*QueryResult

	for _, ast := range res {
		data, err := qw.runBackupQuery(iq.WorkspaceUUID, ast, iq.EchoExplain)
		if err != nil {
			return nil, err
		}

		if data != nil {
			datas = append(datas, data)
		}
	}

	return datas, nil
}

func (qw *queryWorker) runBackupQuery(wsid string, ast *ASTResult, explain bool) (*QueryResult, error) {
	data, err := qw.runBackupSingleQuery(wsid, ast, explain)
	if err != nil {
		return nil, err
	}

	if data == nil {
		l.Debugf("got no data on query: %s", ast.Q)
		return nil, nil
	}

	switch ast.Namespace {
	case NSFunc, NSFuncAbbr:
		// pass
	default:
		switch v := ast.AST.(type) {
		case *parser.DFQuery:
			data.GroupByList = v.GroupByList()
			/*
			 *  -> rewrite query results <-
			 */
			if err := RewriteResults(v, data.Series); err != nil {
				return nil, err
			}
		}
	}
	return data, nil
}

func (qw *queryWorker) runBackupSingleQuery(wsid string, ast *ASTResult, explain bool) (*QueryResult, error) {
	l.Debugf("[%d] namespace: %s, query: %s", qw.idx, ast.Namespace, ast.Q)

	switch ast.Namespace {

	case NSLogging:
		qrs, err := qw.esBackupQuery(ast, wsid, explain)
		if err != nil {
			return nil, err
		}

		return qrs, nil

	}

	return nil, fmt.Errorf("invalid namespace `%s'", ast.Namespace)
}

func (qw *queryWorker) esBackupQuery(ast *ASTResult,
	wsID string,
	explain bool) (*QueryResult, error) {

	var qres *QueryResult // influx结构化后返回值

	indexName := ``
	switch ast.Namespace {
	case `logging`, `L`:
		indexName = wsID + `_backup_log`
	default:
		l.Errorf(`No Support`)
		return nil, fmt.Errorf("no support namespace")
	}

	showAst, ok := ast.AST.(*parser.Show)
	if ok && showAst.Helper.ESTResPtr.ShowFields { // 满足show函数 且是 show fields
		res, err := qw.esCli.XPackSQL(indexName)
		if err != nil {
			err := qw.esCli.ErrorHandler(err) // 分析且格式化 es err
			return nil, err
		}
		timeField := showAst.Helper.ESTResPtr.TimeField
		indexNames := []string{indexName}
		existRes, err := qw.esCli.MSearchFieldExists(indexNames, timeField, ast.Q, res)

		if err != nil {
			return nil, err
		}
		qres, err = esShowColumnsToInflux(existRes)

	} else {
		indexNames := []string{indexName}
		esRes, err := qw.esCli.Es.Search().
			Timeout(qw.esCli.TimeOut).
			Index(indexNames...).
			Source(ast.Q).
			Do(context.Background())

		if err != nil {
			nerr := qw.esCli.ErrorHandler(err) // 分析且格式化 es err
			l.Errorf(`dql:%s, indexName:%s, %s`, ast.Q, indexName, err.Error())
			return nil, nerr
		}

		if esRes.Error != nil {

			l.Error(`ES query result error: %s`, esRes.Error)
			return nil, fmt.Errorf(esRes.Error.Reason)
		}

		qres, err = es2influx(esRes, ast)
		if err != nil {
			return nil, err
		}
	}

	if explain {
		qres.RawQuery = ast.Q
	}

	return qres, nil
}

// RunQuery direct query
func RunQuery(iq *InnerQuery, esCli *es.EsCli) (interface{}, error) {

	var response []*QueryResult

	qw := queryWorker{
		influxClis: map[string]*influxQueryCli{},
		influxDBs:  map[string][]string{},
		esCli:      esCli,
	}
	for _, query := range iq.Queries {
		pres, err := qw.parseQuery(query)
		if err != nil {
			return nil, err
		}

		results, err := qw.runQueries(iq, pres)

		if err != nil {
			return nil, err
		}
		response = append(response, results...)
	}

	return response, nil
}
