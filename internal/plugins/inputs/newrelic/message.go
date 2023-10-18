// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package newrelic handle New Relic APM traces.
package newrelic

import (
	"context"
	"encoding/json"
	"net"
	"regexp"
	"time"
)

const (
	// app behavior.

	// DefaultConfigurableEventHarvestMs is the period for custom, error,
	// and transaction events if the connect response's
	// "event_harvest_config.report_period_ms" is missing or invalid.
	DefaultConfigurableEventHarvestMs = 60 * 1000
	// MaxPayloadSizeInBytes specifies the maximum payload size in bytes that
	// should be sent to any endpoint.
	MaxPayloadSizeInBytes = 1000 * 1000
	// MaxCustomEvents is the maximum number of Transaction Events that can be captured
	// per 60-second harvest cycle.
	MaxCustomEvents = 30 * 1000
	// MaxLogEvents is the maximum number of Log Events that can be captured per
	// 60-second harvest cycle.
	MaxLogEvents = 10 * 1000
	// MaxTxnEvents is the maximum number of Transaction Events that can be captured
	// per 60-second harvest cycle.
	MaxTxnEvents = 10 * 1000
	// MaxErrorEvents is the maximum number of Error Events that can be captured
	// per 60-second harvest cycle.
	MaxErrorEvents = 100
	// MaxSpanEvents is the maximum number of Spans Events that can be captured
	// per 60-second harvest cycle.
	MaxSpanEvents = 1000
	//
	DefaultReportPeriod = 60
)

type Connection struct {
	PID          int64           `json:"pid"`
	Language     string          `json:"language"`
	Host         string          `json:"host"`
	AppNames     []string        `json:"app_name"`
	AgentVersion string          `json:"agent_version"`
	Identifier   string          `json:"identifier"`
	Environments [][]interface{} `json:"environment"`
}

type securityPolicy struct {
	EnabledVal  bool `json:"enabled"`
	RequiredVal bool `json:"required"`
}

type SecurityPolicies struct {
	RecordSQL                 securityPolicy `json:"record_sql"`
	AttributesInclude         securityPolicy `json:"attributes_include"`
	AllowRawExceptionMessages securityPolicy `json:"allow_raw_exception_messages"`
	CustomEvents              securityPolicy `json:"custom_events"`
	CustomParameters          securityPolicy `json:"custom_parameters"`
}

type PreconnectReply struct {
	Collector        string           `json:"redirect_host"`
	SecurityPolicies SecurityPolicies `json:"security_policies"`
}

type reply struct {
	Reply *ConnectReply `json:"return_value"`
}

type segmentRule struct {
	Prefix   string   `json:"prefix"`
	Terms    []string `json:"terms"`
	TermsMap map[string]struct{}
}

// segmentRules is keyed by each segmentRule's Prefix field with any trailing
// slash removed.
type segmentRules map[string]*segmentRule

type metricRule struct {
	// 'Ignore' indicates if the entire transaction should be discarded if
	// there is a match.  This field is only used by "url_rules" and
	// "transaction_name_rules", not "metric_name_rules".
	Ignore              bool   `json:"ignore"`
	EachSegment         bool   `json:"each_segment"`
	ReplaceAll          bool   `json:"replace_all"`
	Terminate           bool   `json:"terminate_chain"`
	Order               int    `json:"eval_order"`
	OriginalReplacement string `json:"replacement"`
	RawExpr             string `json:"match_expression"`

	// Go's regexp backreferences use '${1}' instead of the Perlish '\1', so
	// we transform the replacement string into the Go syntax and store it
	// here.
	TransformedReplacement string
	re                     *regexp.Regexp // nolint:unused
}

// MetricRules is a collection of metric rules.
type MetricRules []*metricRule

// TrustedAccountSet is used for CAT.
type TrustedAccountSet map[int]struct{}

// IsTrusted reveals whether the account can be trusted.
func (t *TrustedAccountSet) IsTrusted(account int) bool {
	_, exists := (*t)[account]
	return exists
}

// UnmarshalJSON unmarshals the trusted set from the connect reply JSON.
func (t *TrustedAccountSet) UnmarshalJSON(data []byte) error {
	accounts := make([]int, 0)
	if err := json.Unmarshal(data, &accounts); err != nil {
		return err
	}

	*t = make(TrustedAccountSet)
	for _, account := range accounts {
		(*t)[account] = struct{}{}
	}

	return nil
}

type DialerFunc func(context.Context, string) (net.Conn, error)

// EventHarvestConfig contains fields relating to faster event harvest.
// This structure is used in the connect request (to send up defaults)
// and in the connect response (to get the server values).
//
// https://source.datanerd.us/agents/agent-specs/blob/master/Connect-LEGACY.md#event_harvest_config-hash
// https://source.datanerd.us/agents/agent-specs/blob/master/Connect-LEGACY.md#event-harvest-config
type EventHarvestConfig struct {
	ReportPeriodMs int `json:"report_period_ms,omitempty"`
	Limits         struct {
		TxnEvents    *uint `json:"analytic_event_data,omitempty"`
		CustomEvents *uint `json:"custom_event_data,omitempty"`
		LogEvents    *uint `json:"log_event_data,omitempty"`
		ErrorEvents  *uint `json:"error_event_data,omitempty"`
		SpanEvents   *uint `json:"span_event_data,omitempty"`
	} `json:"harvest_limits"`
}

// SpanEventHarvestConfig contains the Reporting period time and the given harvest limit.
type SpanEventHarvestConfig struct {
	ReportPeriod *uint `json:"report_period_ms"`
	HarvestLimit *uint `json:"harvest_limit"`
}

// ConnectReply contains all of the settings and state send down from the
// collector.  It should not be modified after creation.
type ConnectReply struct {
	RunID                 string            `json:"agent_run_id"`
	RequestHeadersMap     map[string]string `json:"request_headers_map"`
	MaxPayloadSizeInBytes int               `json:"max_payload_size_in_bytes"`
	EntityGUID            string            `json:"entity_guid"`

	// Transaction Name Modifiers
	SegmentTerms segmentRules `json:"transaction_segment_terms"`
	TxnNameRules MetricRules  `json:"transaction_name_rules"`
	URLRules     MetricRules  `json:"url_rules"`
	MetricRules  MetricRules  `json:"metric_name_rules"`

	// Cross Process
	EncodingKey     string            `json:"encoding_key"`
	CrossProcessID  string            `json:"cross_process_id"`
	TrustedAccounts TrustedAccountSet `json:"trusted_account_ids"`

	// Settings
	KeyTxnApdex            map[string]float64 `json:"web_transactions_apdex"`
	ApdexThresholdSeconds  float64            `json:"apdex_t"`
	CollectAnalyticsEvents bool               `json:"collect_analytics_events"`
	CollectCustomEvents    bool               `json:"collect_custom_events"`
	CollectTraces          bool               `json:"collect_traces"`
	CollectErrors          bool               `json:"collect_errors"`
	CollectErrorEvents     bool               `json:"collect_error_events"`
	CollectSpanEvents      bool               `json:"collect_span_events"`

	// RUM
	AgentLoader string `json:"js_agent_loader"`
	Beacon      string `json:"beacon"`
	BrowserKey  string `json:"browser_key"`
	AppID       string `json:"application_id"`
	ErrorBeacon string `json:"error_beacon"`
	JSAgentFile string `json:"js_agent_file"`

	// PreconnectReply fields are not in the connect reply, this embedding
	// is done to simplify code.
	PreconnectReply `json:"-"`

	Messages []struct {
		Message string `json:"message"`
		Level   string `json:"level"`
	} `json:"messages"`

	// TraceIDGenerator creates random IDs for distributed tracing.  It
	// exists here in the connect reply so it can be modified to create
	// deterministic identifiers in tests.
	TraceIDGenerator *TraceIDGenerator `json:"-"`
	// DistributedTraceTimestampGenerator allows tests to fix the outbound
	// DT header timestamp.
	DistributedTraceTimestampGenerator func() time.Time `json:"-"`
	// TraceObsDialer allows tests to connect to a local TraceObserver directly
	TraceObsDialer DialerFunc `json:"-"`

	// BetterCAT/Distributed Tracing
	AccountID                     string `json:"account_id"`
	TrustedAccountKey             string `json:"trusted_account_key"`
	PrimaryAppID                  string `json:"primary_application_id"`
	SamplingTarget                uint64 `json:"sampling_target"`
	SamplingTargetPeriodInSeconds int    `json:"sampling_target_period_in_seconds"`

	ServerSideConfig struct {
		TransactionTracerEnabled *bool `json:"transaction_tracer.enabled"`
		// TransactionTracerThreshold should contain either a number or
		// "apdex_f" if it is non-nil.
		TransactionTracerThreshold           interface{} `json:"transaction_tracer.transaction_threshold"`
		TransactionTracerStackTraceThreshold *float64    `json:"transaction_tracer.stack_trace_threshold"`
		ErrorCollectorEnabled                *bool       `json:"error_collector.enabled"`
		ErrorCollectorIgnoreStatusCodes      []int       `json:"error_collector.ignore_status_codes"`
		ErrorCollectorExpectStatusCodes      []int       `json:"error_collector.expected_status_codes"`
		CrossApplicationTracerEnabled        *bool       `json:"cross_application_tracer.enabled"`
	} `json:"agent_config"`

	// Faster Event Harvest
	EventData              EventHarvestConfig `json:"event_harvest_config"`
	SpanEventHarvestConfig `json:"span_event_harvest_config"`
	DataReportPeriod       int `json:"data_report_period"`
}

// ConnectReplyDefaults returns a newly allocated ConnectReply with the proper
// default settings.  A pointer to a global is not used to prevent consumers
// from changing the default settings.
func ConnectReplyDefaults() *ConnectReply {
	return &ConnectReply{
		RequestHeadersMap:     make(map[string]string),
		ApdexThresholdSeconds: 0.5,
		// CollectAnalyticsEvents:             true,
		// CollectCustomEvents:                true,
		CollectTraces: true,
		// CollectErrors:                      true,
		// CollectErrorEvents:                 true,
		CollectSpanEvents:                  true,
		MaxPayloadSizeInBytes:              MaxPayloadSizeInBytes,
		SamplingTarget:                     10,
		SamplingTargetPeriodInSeconds:      60,
		TraceIDGenerator:                   NewTraceIDGenerator(time.Now().UnixNano()),
		DistributedTraceTimestampGenerator: time.Now,
		// DataReportPeriod:                   60,
	}
}
