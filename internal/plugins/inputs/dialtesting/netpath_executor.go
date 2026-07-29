// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	netpathinput "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netpath"
)

const (
	netPathMetricName = "netpath_dial_testing"
	netPathTaskSource = "server"
)

type netPathExecutor struct {
	task               *dt.NetPathTask
	runProbe           func(context.Context, dt.NetPathProbeConfig, func(net.IP) error) netpathinput.DialProbeResult
	parentCtx          context.Context
	concurrency        chan struct{}
	validateResolvedIP func(net.IP) error

	// stopChans carries extra cancellation signals (e.g. the input stop
	// signal) that should also abort an in-flight probe. Nil channels are
	// dropped at construction time.
	stopChans []<-chan interface{}

	mu      sync.Mutex
	tags    map[string]string
	fields  map[string]interface{}
	reasons []string
	success bool
}

var _ dt.NetPathExecutor = (*netPathExecutor)(nil)

func newNetPathExecutor(task *dt.NetPathTask, stopChans ...<-chan interface{}) *netPathExecutor {
	e := &netPathExecutor{
		task:     task,
		runProbe: netpathinput.RunDialProbe,
		tags:     map[string]string{},
		fields:   map[string]interface{}{},
	}
	for _, ch := range stopChans {
		if ch != nil {
			e.stopChans = append(e.stopChans, ch)
		}
	}
	return e
}

func (e *netPathExecutor) withContext(ctx context.Context) *netPathExecutor {
	e.parentCtx = ctx
	return e
}

func (e *netPathExecutor) withConcurrency(concurrency chan struct{}) *netPathExecutor {
	e.concurrency = concurrency
	return e
}

func (e *netPathExecutor) withResolvedIPValidator(validate func(net.IP) error) *netPathExecutor {
	e.validateResolvedIP = validate
	return e
}

func (e *netPathExecutor) Run(ctx context.Context, cfg dt.NetPathProbeConfig) error {
	// The context handed in by NetPathTask.run is only canceled when the task
	// itself is stopped, so a probe keeps running until its own deadline when
	// DataKit exits or the input is terminated (the dialer's run loop blocks on
	// this call and never reaches its exit select). Derive a context that also
	// follows datakit.Exit and the extra stop signals so in-flight probes are
	// canceled promptly on shutdown.
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	watchCancel := func(done <-chan interface{}) {
		if done == nil {
			return
		}
		go func() {
			select {
			case <-done:
				cancel()
			case <-runCtx.Done():
			}
		}()
	}
	watchCancel(datakit.Exit.Wait())
	for _, ch := range e.stopChans {
		watchCancel(ch)
	}
	if e.parentCtx != nil {
		go func() {
			select {
			case <-e.parentCtx.Done():
				cancel()
			case <-runCtx.Done():
			}
		}()
	}

	if e.concurrency != nil {
		select {
		case e.concurrency <- struct{}{}:
			netPathProbesInFlight.Add(1)
			defer func() {
				netPathProbesInFlight.Add(-1)
				<-e.concurrency
			}()
		case <-runCtx.Done():
			return runCtx.Err()
		}
	}
	result := e.runProbe(runCtx, cfg, e.validateResolvedIP)
	reasons, success := e.evaluate(result)
	if result.Err != nil {
		reasons = append([]string{result.Err.Error()}, reasons...)
		success = false
	}
	e.setResult(result, reasons, success)
	return nil
}

func (e *netPathExecutor) Clear() {
	e.mu.Lock()
	e.tags = map[string]string{}
	e.fields = map[string]interface{}{}
	e.reasons = nil
	e.success = false
	e.mu.Unlock()
}

func (e *netPathExecutor) CheckResult() ([]string, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string(nil), e.reasons...), e.success
}

func (e *netPathExecutor) GetResults() (map[string]string, map[string]interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	tags := make(map[string]string, len(e.tags))
	for key, value := range e.tags {
		tags[key] = value
	}
	fields := make(map[string]interface{}, len(e.fields))
	for key, value := range e.fields {
		fields[key] = value
	}
	return tags, fields
}

func (e *netPathExecutor) SetError(message string) {
	now := time.Now()
	e.setResult(netpathinput.DialProbeResult{
		Tags:        map[string]string{},
		Fields:      map[string]interface{}{},
		ScheduledAt: now,
		StartedAt:   now,
		FinishedAt:  now,
	}, []string{message}, false)
}

func (e *netPathExecutor) evaluate(result netpathinput.DialProbeResult) ([]string, bool) {
	failures := []string{}
	passed := 0
	total := 0
	for _, success := range e.task.SuccessWhen {
		if success == nil {
			continue
		}
		for _, group := range netPathConditions(success) {
			for _, condition := range group.conditions {
				total++
				if err := checkNetPathCondition(group.field, condition, result); err != nil {
					failures = append(failures, err.Error())
				} else {
					passed++
				}
			}
		}
	}
	if total == 0 {
		return []string{"success_when is empty"}, false
	}
	if strings.EqualFold(e.task.SuccessWhenLogic, "or") {
		return failures, passed > 0
	}
	return failures, passed == total
}

// netPathFieldConditions groups the assertions of a single result field.
// evaluate walks the groups in the fixed declaration order below (kept
// identical to the cliutils NetPathSuccess.assertions order) so the
// fail_reason output is deterministic across runs.
type netPathFieldConditions struct {
	field      string
	conditions []*dt.NetPathCondition
}

func netPathConditions(success *dt.NetPathSuccess) []netPathFieldConditions {
	return []netPathFieldConditions{
		{field: "e2e_rtt_avg", conditions: success.E2ERTTAvg},
		{field: "e2e_rtt_min", conditions: success.E2ERTTMin},
		{field: "e2e_rtt_max", conditions: success.E2ERTTMax},
		{field: "e2e_rtt_variation_avg", conditions: success.E2ERTTVariationAvg},
		{field: "e2e_rtt_variation_max", conditions: success.E2ERTTVariationMax},
		{field: "e2e_probe_loss_percent", conditions: success.E2EProbeLossPercent},
		{field: "hop_count", conditions: success.HopCount},
		{field: "e2e_status", conditions: success.E2EStatus},
		{field: "traceroute_status", conditions: success.TracerouteStatus},
	}
}

func checkNetPathCondition(
	field string,
	condition *dt.NetPathCondition,
	result netpathinput.DialProbeResult,
) error {
	if condition == nil {
		return fmt.Errorf("%s assertion is null", field)
	}
	if field == "e2e_status" || field == "traceroute_status" {
		if !strings.EqualFold(strings.TrimSpace(condition.Op), "eq") {
			return fmt.Errorf("%s only supports eq", field)
		}
		actual, ok := result.Tags[field]
		if !ok || actual == "" {
			return fmt.Errorf("%s is missing", field)
		}
		target, err := netPathStringTarget(condition)
		if err != nil {
			return err
		}
		if actual != target {
			return fmt.Errorf("%s: expected %s, got %s", field, target, actual)
		}
		return nil
	}

	value, ok := result.Fields[field]
	if !ok {
		return fmt.Errorf("%s is missing", field)
	}
	actual, ok := numericValue(value)
	if !ok {
		return fmt.Errorf("%s is not numeric", field)
	}

	if strings.HasPrefix(field, "e2e_rtt") {
		target, err := netPathDurationTarget(condition)
		if err != nil {
			return fmt.Errorf("%s: %w", field, err)
		}
		if !compareNumber(actual, float64(target)/float64(time.Microsecond), condition.Op) {
			return fmt.Errorf("%s: %s does not satisfy %s %s",
				field, formatNetPathDuration(actual), condition.Op, target)
		}
		return nil
	}

	target, err := netPathNumberTarget(condition)
	if err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}
	if !compareNumber(actual, target, condition.Op) {
		return fmt.Errorf("%s: %v does not satisfy %s %v", field, actual, condition.Op, target)
	}
	return nil
}

func formatNetPathDuration(microseconds float64) string {
	nanoseconds := math.Round(microseconds * float64(time.Microsecond))
	return time.Duration(nanoseconds).String()
}

func netPathStringTarget(condition *dt.NetPathCondition) (string, error) {
	var target string
	if err := json.Unmarshal(condition.Target, &target); err != nil || strings.TrimSpace(target) == "" {
		return "", fmt.Errorf("target must be a non-empty string")
	}
	return target, nil
}

func netPathDurationTarget(condition *dt.NetPathCondition) (time.Duration, error) {
	target, err := netPathStringTarget(condition)
	if err != nil {
		return 0, err
	}
	duration, err := time.ParseDuration(target)
	if err != nil || duration < 0 {
		return 0, fmt.Errorf("target must be a duration")
	}
	return duration, nil
}

func netPathNumberTarget(condition *dt.NetPathCondition) (float64, error) {
	var target *float64
	if err := json.Unmarshal(condition.Target, &target); err != nil || target == nil {
		return 0, fmt.Errorf("target must be numeric")
	}
	return *target, nil
}

func numericValue(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case float32:
		return float64(typed), true
	case float64:
		return typed, true
	default:
		return 0, false
	}
}

func compareNumber(value, target float64, op string) bool {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "eq":
		return value == target
	case "lt":
		return value < target
	case "leq":
		return value <= target
	case "gt":
		return value > target
	case "geq":
		return value >= target
	default:
		return false
	}
}

func (e *netPathExecutor) setResult(
	result netpathinput.DialProbeResult,
	reasons []string,
	success bool,
) {
	task := e.task
	for i := range reasons {
		reasons[i] = redactNetPathSecureText(task, reasons[i])
	}
	port, err := strconv.ParseUint(strings.TrimSpace(task.Port), 10, 16)
	if err != nil {
		port = 0
	}
	name := strings.TrimSpace(task.Name)
	if name == "" {
		name = strings.TrimSpace(task.Host)
	}
	host := strings.TrimSpace(task.Host)
	portText := strings.TrimSpace(task.Port)
	protocol := strings.ToLower(strings.TrimSpace(task.Protocol))
	taskName := netpathinput.TaskDisplayName(name, host, uint16(port), protocol)
	tags := map[string]string{
		"name":              name,
		"task_name":         taskName,
		"task_source":       netPathTaskSource,
		"protocol":          protocol,
		"src_ip":            strings.TrimSpace(result.Tags["probe_source_ip"]),
		"src_port":          "*",
		"probe_source_ip":   strings.TrimSpace(result.Tags["probe_source_ip"]),
		"dest_host":         host,
		"dest_port":         portText,
		"dst_port":          portText,
		"source_name":       strings.TrimSpace(task.AdvanceOptions.SourceName),
		"target_name":       strings.TrimSpace(task.AdvanceOptions.TargetName),
		"traceroute_status": result.Tags["traceroute_status"],
		"e2e_status":        result.Tags["e2e_status"],
	}
	if tags["dst_port"] == "" {
		tags["dst_port"] = "*"
	}
	if net.ParseIP(host) == nil {
		tags["dst_domain"] = host
	}
	if destinationIP := netPathDestinationIP(result, host); destinationIP != "" {
		tags["dest_ip"] = destinationIP
		tags["dst_ip"] = destinationIP
	}
	if tags["target_name"] == "" {
		tags["target_name"] = host
	}
	tags["traceroute_protocol"] = result.Tags["traceroute_protocol"]
	if tags["traceroute_protocol"] == "" {
		tags["traceroute_protocol"] = tags["protocol"]
	}
	for key, value := range tags {
		if value == "" {
			delete(tags, key)
		}
	}
	if success {
		tags["status"] = "OK"
	} else {
		tags["status"] = "FAIL"
	}

	fields := make(map[string]interface{}, len(result.Fields)+12)
	for key, value := range result.Fields {
		fields[key] = value
	}
	fields["test_run_id"] = result.TestRunID
	fields["scheduled_at"] = result.ScheduledAt.UnixMicro()
	fields["started_at"] = result.StartedAt.UnixMicro()
	fields["duration"] = result.Duration.Microseconds()
	fields["task"] = e.sanitizedTask()
	fields["config_vars"] = e.safeConfigVars()
	if success {
		fields["success"] = int64(1)
		fields["fail_reason"] = ""
		fields["message"] = "netpath probe succeeded"
	} else {
		fields["success"] = int64(-1)
		fields["fail_reason"] = strings.Join(reasons, "; ")
		fields["message"] = fields["fail_reason"]
	}
	if result.FailType != "" {
		fields["traceroute_fail_type"] = result.FailType
	}

	e.mu.Lock()
	e.tags = tags
	e.fields = fields
	e.reasons = append([]string(nil), reasons...)
	e.success = success
	e.mu.Unlock()
}

func taskErrorForLog(task dt.ITask, err error) string {
	if err == nil {
		return ""
	}
	if netPathTask, ok := task.(*dt.NetPathTask); ok {
		return redactNetPathSecureText(netPathTask, err.Error())
	}
	return err.Error()
}

func redactNetPathSecureText(task *dt.NetPathTask, text string) string {
	if task == nil || task.Task == nil || text == "" {
		return text
	}
	variables := make([]*dt.ConfigVar, 0,
		len(task.ConfigVars)+len(task.ExtractedVars)+len(task.CustomVars))
	variables = append(variables, task.ConfigVars...)
	variables = append(variables, task.ExtractedVars...)
	variables = append(variables, task.CustomVars...)
	values := make([]string, 0, len(variables))
	for _, variable := range variables {
		if variable == nil || !variable.Secure || variable.Value == "" {
			continue
		}
		values = append(values, variable.Value)
	}
	sort.SliceStable(values, func(i, j int) bool {
		return len(values[i]) > len(values[j])
	})
	for _, value := range values {
		text = strings.ReplaceAll(text, value, "<redacted>")
	}
	return text
}

func netPathDestinationIP(result netpathinput.DialProbeResult, host string) string {
	if destinationIP := strings.TrimSpace(result.Tags["probe_dest_ip"]); destinationIP != "" {
		return destinationIP
	}
	if destinationIP, ok := result.Fields["e2e_dest_ip"].(string); ok {
		if destinationIP = strings.TrimSpace(destinationIP); destinationIP != "" {
			return destinationIP
		}
	}
	if net.ParseIP(host) != nil {
		return host
	}
	return ""
}

func (e *netPathExecutor) safeConfigVars() string {
	type safeVariable struct {
		Name   string `json:"name"`
		Value  string `json:"value,omitempty"`
		Secure bool   `json:"secure"`
	}

	vars := make([]safeVariable, 0, len(e.task.ConfigVars))
	for _, variable := range e.task.ConfigVars {
		if variable == nil {
			continue
		}
		item := safeVariable{Name: variable.Name, Secure: variable.Secure}
		if !variable.Secure {
			item.Value = variable.Value
		}
		vars = append(vars, item)
	}
	data, _ := json.Marshal(vars)
	return string(data)
}

func (e *netPathExecutor) sanitizedTask() string {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(e.task.String()), &raw); err != nil {
		return "{}"
	}
	delete(raw, "access_key")
	delete(raw, "post_url")
	if variables, ok := raw["config_vars"].([]interface{}); ok {
		for _, variable := range variables {
			item, ok := variable.(map[string]interface{})
			if !ok {
				continue
			}
			if secure, _ := item["secure"].(bool); secure {
				delete(item, "value")
			}
		}
	}
	data, _ := json.Marshal(raw)
	return string(data)
}
