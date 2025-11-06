// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dialtesting defined dialtesting tasks and task implements.
package dialtesting

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/GuanceCloud/cliutils"
	log "github.com/GuanceCloud/cliutils/logger"
	"github.com/robfig/cron/v3"
)

const (
	StatusStop = "stop"

	ClassHTTP      = "HTTP"
	ClassTCP       = "TCP"
	ClassWebsocket = "WEBSOCKET"
	ClassICMP      = "ICMP"
	ClassDNS       = "DNS"
	ClassHeadless  = "BROWSER"
	ClassOther     = "OTHER"
	ClassWait      = "WAIT"
	ClassMulti     = "MULTI"

	ScheduleTypeCron      = "crontab"
	ScheduleTypeFrequency = "frequency"
)

var logger = log.DefaultSLogger("icmp")

var (
	setupLock        sync.Mutex // setup global variable
	MaxMsgSize       = 100 * 1024
	ICMPConcurrentCh chan struct{}
)

type ConfigVar struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name"`
	Value   string `json:"value,omitempty"`
	Example string `json:"example,omitempty"`
	Secure  bool   `json:"secure"`
}

var TypeVariableGlobal = "global"

type Variable struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	TaskID      string `json:"task_id,omitempty"`
	TaskVarName string `json:"task_var_name,omitempty"`
	Value       string `json:"value,omitempty"`
	Secure      bool   `json:"secure,omitempty"`
	PostScript  string `json:"post_script,omitempty"`

	UpdatedAt       int64  `json:"updated_at,omitempty"`       // update time
	ValueUpdatedAt  int64  `json:"value_updated_at,omitempty"` // update value time
	FailCount       int    `json:"fail_count,omitempty"`       // update fail count
	DeletedAt       int64  `json:"deleted_at,omitempty"`
	OwnerExternalID string `json:"owner_external_id,omitempty"`
	CreatedAt       int64  `json:"-"`
}
type TaskChild interface {
	ITask
	run() error
	init() error
	checkResult() ([]string, bool)
	getResults() (map[string]string, map[string]interface{})
	stop()
	check() error
	clear()
	getVariableValue(variable Variable) (string, error)
	class() string
	metricName() string
	getHostName() ([]string, error)
	getRawTask(string) (string, error)
	renderTemplate(fm template.FuncMap) error
	initTask()
	setReqError(string)
}

func getHostName(host string) (string, error) {
	reqURL, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("parse host error: %w", err)
	}

	return reqURL.Hostname(), nil
}

var nameRe = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*$`)

func isValidVariableName(name string) bool {
	return nameRe.MatchString(name)
}

type ITask interface {
	ID() string
	Status() string
	Run() error
	CheckResult() ([]string, bool)
	Class() string
	GetResults() (map[string]string, map[string]interface{})
	PostURLStr() string
	MetricName() string
	Stop()
	RegionName() string
	AccessKey() string
	Check() error
	CheckTask() error
	UpdateTimeUs() int64
	GetFrequency() string
	GetOwnerExternalID() string
	GetExternalID() string
	SetOwnerExternalID(string)
	GetLineData() string
	GetHostName() ([]string, error)
	GetWorkspaceLanguage() string
	GetDFLabel() string
	GetScheduleType() string
	GetCrontab() string

	SetOption(map[string]string)
	GetOption() map[string]string
	SetRegionID(string)
	SetAk(string)
	SetStatus(string)
	SetUpdateTime(int64)
	SetChild(TaskChild)
	SetTaskJSONString(string)
	GetTaskJSONString() string
	SetDisabled(uint8)

	GetVariableValue(Variable) (string, error)
	GetGlobalVars() []string
	RenderTemplateAndInit(globalVariables map[string]Variable) error
	AddExtractedVar(*ConfigVar)
	SetCustomVars([]*ConfigVar)
	GetPostScriptVars() Vars
	GetIsTemplate() bool
	SetIsTemplate(bool)
	SetBeforeRun(func(*Task) error)

	String() string
}

type Task struct {
	ExternalID        string `json:"external_id"`
	Name              string `json:"name"`
	AK                string `json:"access_key"`
	PostURL           string `json:"post_url"`
	CurStatus         string `json:"status"`
	Frequency         string `json:"frequency"`
	Region            string `json:"region"`
	OwnerExternalID   string `json:"owner_external_id"`
	WorkspaceLanguage string `json:"workspace_language,omitempty"`
	TagsInfo          string `json:"tags_info,omitempty"` // deprecated
	DFLabel           string `json:"df_label,omitempty"`
	ScheduleType      string `json:"schedule_type,omitempty"` // "frequency" or "crontab"
	Crontab           string `json:"crontab,omitempty"`       // crontab expression like "0 0 * * *"

	Tags          map[string]string `json:"tags,omitempty"`
	Labels        []string          `json:"labels,omitempty"`
	UpdateTime    int64             `json:"update_time,omitempty"`
	ConfigVars    []*ConfigVar      `json:"config_vars,omitempty"`
	ExtractedVars []*ConfigVar
	CustomVars    []*ConfigVar

	taskJSONString string
	rawTask        string
	child          TaskChild

	Disabled uint8 `json:"disabled"`

	isTemplate bool
	globalVars map[string]Variable
	option     map[string]string
	fm         template.FuncMap
	beforeRun  func(*Task) error
}

type TaskConfig struct {
	MaxMsgSize        int `json:"max_msg_size,omitempty"`
	MaxICMPConcurrent int `json:"max_icmp_concurrent,omitempty"`
	Logger            *log.Logger
}

func Setup(c *TaskConfig) {
	setupLock.Lock()
	defer setupLock.Unlock()
	if c.MaxMsgSize > 0 {
		MaxMsgSize = c.MaxMsgSize
	}

	if c.MaxICMPConcurrent > 0 {
		ICMPConcurrentCh = make(chan struct{}, c.MaxICMPConcurrent)
	}

	if c.Logger != nil {
		logger = c.Logger
	} else {
		logger = log.SLogger("dialtesting")
	}
}

func CreateTaskChild(taskType string) (TaskChild, error) {
	var ct TaskChild
	switch taskType {
	case "http", "https", ClassHTTP:
		ct = &HTTPTask{}

	case "multi", ClassMulti:
		ct = &MultiTask{}

	case "headless", "browser", ClassHeadless:
		return nil, errors.New("headless task deprecated")

	case "tcp", ClassTCP:
		ct = &TCPTask{}

	case "websocket", ClassWebsocket:
		ct = &WebsocketTask{}

	case "icmp", ClassICMP:
		ct = &ICMPTask{}

	default:
		return nil, fmt.Errorf("unknown task type %s", taskType)
	}

	return ct, nil
}

func NewTask(taskString string, task TaskChild) (ITask, error) {
	if task == nil {
		return nil, errors.New("invalid task")
	}

	if taskString != "" {
		if err := json.Unmarshal([]byte(taskString), &task); err != nil {
			return nil, fmt.Errorf("json.Unmarshal failed: %w, task json: %s", err, taskString)
		}
	} else {
		bytes, _ := json.Marshal(task)
		taskString = string(bytes)
	}

	task.initTask()

	if t, ok := task.(ITask); !ok {
		return nil, errors.New("invalid task, not ITask")
	} else {
		t.SetTaskJSONString(taskString)
		t.SetChild(task)
		t.SetIsTemplate(hasTemplateTags(taskString))
		return t, nil
	}
}

func (t *Task) String() string {
	b, _ := json.Marshal(t.child)
	return string(b)
}

func (t *Task) NewRawTask(child TaskChild) error {
	if t.taskJSONString == "" {
		return errors.New("task json string is empty")
	}

	if err := json.Unmarshal([]byte(t.taskJSONString), &child); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w, task json: %s", err, t.taskJSONString)
	}

	return nil
}

func (t *Task) SetChild(child TaskChild) {
	t.child = child
}

func (t *Task) SetOption(opt map[string]string) {
	if opt != nil {
		t.option = opt
	}
}

func (t *Task) GetOption() map[string]string {
	if t.option == nil {
		t.option = make(map[string]string)
	}
	return t.option
}

func (t *Task) UpdateTimeUs() int64 {
	return t.UpdateTime
}

func (t *Task) Clear() {
	t.child.clear()
	t.ExtractedVars = t.ExtractedVars[:0]
}

func (t *Task) ID() string {
	if t.ExternalID == `` {
		return cliutils.XID("dtst_")
	}
	return fmt.Sprintf("%s_%s", t.AK, t.ExternalID)
}

func (t *Task) GetOwnerExternalID() string {
	return t.OwnerExternalID
}

func (t *Task) GetExternalID() string {
	return t.ExternalID
}

func (t *Task) SetOwnerExternalID(exid string) {
	t.OwnerExternalID = exid
}

func (t *Task) SetRegionID(regionID string) {
	t.Region = regionID
}

func (t *Task) SetAk(ak string) {
	t.AK = ak
}

func (t *Task) SetDisabled(disabled uint8) {
	t.Disabled = disabled
}

func (t *Task) SetStatus(status string) {
	t.CurStatus = status
}

func (t *Task) SetUpdateTime(ts int64) {
	t.UpdateTime = ts
}

func (t *Task) Stop() {
	t.child.stop()
}

func (t *Task) Status() string {
	return t.CurStatus
}

func (t *Task) Class() string {
	return t.child.class()
}

func (t *Task) MetricName() string {
	return t.child.metricName()
}

func (t *Task) PostURLStr() string {
	return t.PostURL
}

func (t *Task) GetFrequency() string {
	return t.Frequency
}

func (t *Task) GetLineData() string {
	return ""
}

func (t *Task) GetIsTemplate() bool {
	return t.isTemplate
}

func (t *Task) SetIsTemplate(isTemplate bool) {
	t.isTemplate = isTemplate
}

func (t *Task) GetResults() (tags map[string]string, fields map[string]interface{}) {
	tags, fields = t.child.getResults()

	// add config_vars
	vars := []ConfigVar{}
	for _, v := range t.ConfigVars {
		variable := ConfigVar{
			Name:   v.Name,
			Secure: v.Secure,
		}

		if !v.Secure {
			variable.Value = v.Value
		}

		vars = append(vars, variable)
	}

	// config_vars
	bytes, _ := json.Marshal(vars)
	fields[`config_vars`] = string(bytes)

	// task
	fields["task"] = t.getRawTask()

	return tags, fields
}

func (t *Task) getRawTask() string {
	if t.rawTask == "" {
		if v, err := t.child.getRawTask(t.taskJSONString); err != nil {
			return ""
		} else {
			t.rawTask = v
		}
	}

	return t.rawTask
}

func (t *Task) RegionName() string {
	return t.Region
}

func (t *Task) AccessKey() string {
	return t.AK
}

func (t *Task) CheckTask() error {
	for _, v := range t.ConfigVars {
		if !isValidVariableName(v.Name) {
			return fmt.Errorf("invalid variable name %s", v.Name)
		}
	}

	for _, v := range t.ExtractedVars {
		if !isValidVariableName(v.Name) {
			return fmt.Errorf("invalid variable name %s", v.Name)
		}
	}

	if err := t.child.check(); err != nil {
		return err
	}

	return t.init()
}

func (t *Task) Check() error {
	// TODO: check task validity
	if t.ExternalID == "" {
		return fmt.Errorf("external ID missing")
	}

	if t.ScheduleType == "" {
		t.ScheduleType = ScheduleTypeFrequency
	}

	if t.ScheduleType == ScheduleTypeCron {
		if t.Crontab == "" {
			return fmt.Errorf("crontab missing")
		}
		_, err := cron.ParseStandard(t.Crontab)
		if err != nil {
			return fmt.Errorf("invalid crontab: %w", err)
		}
	} else {
		_, err := time.ParseDuration(t.Frequency)
		if err != nil {
			return fmt.Errorf("invalid frequency: %w", err)
		}
	}

	return t.CheckTask()
}

func (t *Task) Run() error {
	t.Clear()
	if t.fm != nil {
		if err := t.child.renderTemplate(t.fm); err != nil {
			return fmt.Errorf("render template failed: %w", err)
		}
	}
	// before run
	if t.beforeRun != nil && t.Class() != ClassMulti {
		if err := t.beforeRun(t); err != nil {
			t.child.setReqError(err.Error())
			return nil
		}
	}

	return t.child.run()
}

func (t *Task) SetBeforeRun(beforeRun func(*Task) error) {
	t.beforeRun = beforeRun
}

func (t *Task) init() error {
	if strings.EqualFold(t.CurStatus, StatusStop) {
		return nil
	}

	if t.option == nil {
		t.option = map[string]string{}
	}

	return t.child.init()
}

func (t *Task) Init() error {
	return t.init()
}

func (t *Task) GetHostName() ([]string, error) {
	return t.child.getHostName()
}

func (t *Task) GetWorkspaceLanguage() string {
	if t.WorkspaceLanguage == "en" {
		return "en"
	}
	return "zh"
}

func (t *Task) GetDFLabel() string {
	if t.DFLabel != "" {
		return t.DFLabel
	}
	return t.TagsInfo
}

func (t *Task) SetTaskJSONString(s string) {
	t.taskJSONString = s
}

func (t *Task) GetTaskJSONString() string {
	return t.taskJSONString
}

func (t *Task) GetGlobalVars() []string {
	vars := []string{}
	for _, v := range t.ConfigVars {
		if v.Type == TypeVariableGlobal {
			vars = append(vars, v.ID)
		}
	}
	return vars
}

func getDefaultFunc() template.FuncMap {
	return template.FuncMap{
		"timestamp": func(unit string) int64 {
			utcTime := time.Now().UTC()
			switch unit {
			case "s":
				return utcTime.Unix()
			case "ms":
				return utcTime.UnixMilli()
			case "us":
				return utcTime.UnixMicro()
			case "ns":
				return utcTime.UnixNano()
			default:
				return 0
			}
		},

		"date": func(format string) string {
			switch format {
			case "rfc3339":
				format = time.RFC3339
			case "iso8601":
				format = "2006-01-02T15:04:05Z"
			}

			return time.Now().Format(format)
		},

		"urlencode": url.QueryEscape,
	}
}

// RenderTemplateAndInit render template and init task.
func (t *Task) RenderTemplateAndInit(globalVariables map[string]Variable) error {
	if globalVariables == nil {
		globalVariables = make(map[string]Variable)
	}

	t.globalVars = globalVariables

	fm := getDefaultFunc()

	allVars := append(t.ConfigVars, t.ExtractedVars...) // nolint:gocritic
	allVars = append(allVars, t.CustomVars...)

	for _, v := range allVars {
		value := v.Value
		if v.Type == TypeVariableGlobal && v.ID != "" { // global variables
			if gv, ok := globalVariables[v.ID]; ok {
				value = gv.Value
				v.Secure = gv.Secure
			}
		}

		fm[v.Name] = func() string {
			return value
		}

		v.Value = value
	}

	t.fm = fm

	// multi task does not need to render template
	// render template only for its child task
	if t.Class() == ClassMulti {
		return nil
	}

	if err := t.child.renderTemplate(fm); err != nil {
		return fmt.Errorf("render template error: %w", err)
	}

	return t.init()
}

func (t *Task) AddExtractedVar(v *ConfigVar) {
	if v == nil {
		return
	}
	if t.ExtractedVars == nil {
		t.ExtractedVars = []*ConfigVar{}
	}

	t.ExtractedVars = append(t.ExtractedVars, v)
}

func (t *Task) SetCustomVars(vars []*ConfigVar) {
	t.CustomVars = vars
}

func (t *Task) GetVariableValue(variable Variable) (string, error) {
	return t.child.getVariableValue(variable)
}

func (t *Task) CheckResult() ([]string, bool) {
	return t.child.checkResult()
}

func (t *Task) GetPostScriptVars() Vars {
	if ct, ok := t.child.(*HTTPTask); ok {
		if ct.postScriptResult != nil {
			return ct.postScriptResult.Vars
		}
		return nil
	}

	if ct, ok := t.child.(*MultiTask); ok {
		if ct.postScriptResult != nil {
			return ct.postScriptResult.Vars
		}
		return nil
	}

	return nil
}

var variableRe = regexp.MustCompile(`function "([^']*)" not defined`)

func getTemplateError(err error) error {
	if err == nil {
		return err
	}
	msg := err.Error()
	matches := variableRe.FindStringSubmatch(msg)

	if len(matches) > 1 {
		return fmt.Errorf("variable '%s' not defined", matches[1])
	}

	return err
}

func (t *Task) GetParsedString(text string, fm template.FuncMap) (string, error) {
	if text == "" {
		return "", nil
	}

	tmpl, err := template.New("text").Funcs(fm).Option("missingkey=zero").Parse(text)
	if err != nil {
		return "", fmt.Errorf("parse template error: %w", getTemplateError(err))
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("execute template error: %w", err)
	}

	return buf.String(), nil
}

func hasTemplateTags(s string) bool {
	re := regexp.MustCompile(`{{.*}}`)
	return re.MatchString(s)
}

func (t *Task) renderSuccessOption(v, dest *SuccessOption, fm template.FuncMap) error {
	if text, err := t.GetParsedString(v.Is, fm); err != nil {
		return fmt.Errorf("render body is failed: %w", err)
	} else {
		dest.Is = text
	}

	if text, err := t.GetParsedString(v.IsNot, fm); err != nil {
		return fmt.Errorf("render body is not failed: %w", err)
	} else {
		dest.IsNot = text
	}

	if text, err := t.GetParsedString(v.Contains, fm); err != nil {
		return fmt.Errorf("render body contains failed: %w", err)
	} else {
		dest.Contains = text
	}

	if text, err := t.GetParsedString(v.NotContains, fm); err != nil {
		return fmt.Errorf("render body not contains failed: %w", err)
	} else {
		dest.NotContains = text
	}

	if text, err := t.GetParsedString(v.MatchRegex, fm); err != nil {
		return fmt.Errorf("render body match regex failed: %w", err)
	} else {
		dest.MatchRegex = text
	}

	if text, err := t.GetParsedString(v.NotMatchRegex, fm); err != nil {
		return fmt.Errorf("render body not match regex failed: %w", err)
	} else {
		dest.NotMatchRegex = text
	}

	return nil
}

func (t *Task) GetScheduleType() string {
	if t.ScheduleType == "" {
		return "frequency" // default to frequency for backward compatibility
	}
	return t.ScheduleType
}

func (t *Task) GetCrontab() string {
	return t.Crontab
}
