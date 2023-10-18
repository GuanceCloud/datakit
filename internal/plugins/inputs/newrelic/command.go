// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package newrelic handle New Relic APM traces.
package newrelic

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

const (
	cmdPreconnect       = "preconnect"
	cmdConnect          = "connect"
	cmdAgentSettings    = "agent_settings"
	cmdGetAgentCommands = "get_agent_commands"
	cmdMetrics          = "metric_data"
	cmdCustomEvents     = "custom_event_data"
	cmdLogEvents        = "log_event_data"
	cmdTxnEvents        = "analytic_event_data"
	cmdErrorEvents      = "error_event_data"
	cmdErrorData        = "error_data"
	cmdTxnTraces        = "transaction_sample_data"
	cmdSlowSQLs         = "sql_trace_data"
	cmdSpanEvents       = "span_event_data"
	cmdShutdown         = "shutdown"
)

var compound map[string]CommandHandler = map[string]CommandHandler{
	cmdPreconnect:       &preConnect{},
	cmdConnect:          &connect{mapper: &appIDMapper{m: make(map[string]*appWithID)}},
	cmdAgentSettings:    &agentSettings{},
	cmdGetAgentCommands: &getAgentCommands{},
	cmdMetrics:          &metricData{},
	cmdCustomEvents:     &customEventData{},
	cmdLogEvents:        &logEventData{},
	cmdTxnEvents:        &analyticEventData{},
	cmdErrorEvents:      &errorEventData{},
	cmdErrorData:        &errorData{},
	cmdTxnTraces:        &transactionSampleData{},
	cmdSlowSQLs:         &sqlTraceData{},
	cmdSpanEvents:       &spanEventData{},
	cmdShutdown:         &shutdown{},
}

type CommandHandler interface {
	ProcessMethod(resp http.ResponseWriter, q *query)
}

type preConnect struct{}

func (x *preConnect) ProcessMethod(resp http.ResponseWriter, q *query) {
	preconn := PreconnectReply{
		Collector: "",
		SecurityPolicies: SecurityPolicies{
			RecordSQL:                 securityPolicy{EnabledVal: true, RequiredVal: true},
			AttributesInclude:         securityPolicy{EnabledVal: true, RequiredVal: true},
			AllowRawExceptionMessages: securityPolicy{EnabledVal: true, RequiredVal: true},
			CustomEvents:              securityPolicy{EnabledVal: true, RequiredVal: true},
			CustomParameters:          securityPolicy{EnabledVal: true, RequiredVal: true},
		},
	}

	stdReply(q.acceptEncoding, q.contentType, resp, preconn, nil)
}

type connect struct {
	mapper *appIDMapper
}

func (x *connect) ProcessMethod(resp http.ResponseWriter, q *query) {
	connArg := []Connection{}
	err := json.Unmarshal(q.body.Bytes(), &connArg)
	if err != nil {
		log.Debug(err.Error())
		writeEmptyJSON(resp)

		return
	}

	var rid string
	if len(connArg) != 0 && len(connArg[0].AppNames) != 0 {
		rid = x.mapper.update(connArg[0].Host, connArg[0].AgentVersion, connArg[0].Identifier, connArg[0].AppNames[0])
	} else {
		buf := make([]byte, 30)
		_, _ = rand.Read(buf)
		rid = fmt.Sprintf("incomplete-conn-%s", base64.StdEncoding.EncodeToString(buf))
	}

	defReply := ConnectReplyDefaults()
	defReply.RunID = rid
	defReply.HarvestLimit = uintptr(1000)
	defReply.EventData.ReportPeriodMs = DefaultConfigurableEventHarvestMs
	defReply.EventData.Limits.CustomEvents = uintptr(MaxCustomEvents)
	defReply.EventData.Limits.ErrorEvents = uintptr(MaxErrorEvents)
	defReply.EventData.Limits.LogEvents = uintptr(MaxLogEvents)
	defReply.EventData.Limits.SpanEvents = uintptr(MaxSpanEvents)
	defReply.EventData.Limits.TxnEvents = uintptr(MaxTxnEvents)

	defReply.SpanEventHarvestConfig.HarvestLimit = uintptr(MaxSpanEvents)
	defReply.SpanEventHarvestConfig.ReportPeriod = uintptr(DefaultReportPeriod)

	defReply.DataReportPeriod = DefaultReportPeriod

	conn := &reply{Reply: defReply}

	stdReply(q.acceptEncoding, q.contentType, resp, conn, nil)
}

func (x *connect) findAppNameBy(id string) (string, bool) {
	return x.mapper.find(id)
}

type agentSettings struct{}

func (x *agentSettings) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type getAgentCommands struct{}

func (x *getAgentCommands) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type metricData struct{}

func (x *metricData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type customEventData struct{}

func (x *customEventData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type logEventData struct{}

func (x *logEventData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type analyticEventData struct{}

func (x *analyticEventData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type errorEventData struct{}

func (x *errorEventData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type errorData struct{}

func (x *errorData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type transactionSampleData struct{}

func (x *transactionSampleData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)

	trans := &transaction{}
	err := json.Unmarshal(q.body.Bytes(), trans)
	if err != nil {
		log.Debug(err.Error())
	}

	trace := transformToDkTrace(trans)
	if len(trace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{trace})
	}

	writeEmptyJSON(resp)
}

type sqlTraceData struct{}

func (x *sqlTraceData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type spanEventData struct{}

func (x *spanEventData) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}

type shutdown struct{}

func (x *shutdown) ProcessMethod(resp http.ResponseWriter, q *query) {
	printQueryInfo(q)
	writeEmptyJSON(resp)
}
