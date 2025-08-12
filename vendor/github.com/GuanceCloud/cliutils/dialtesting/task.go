// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dialtesting defined dialtesting tasks and task implements.
package dialtesting

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/GuanceCloud/cliutils"
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

	MaxMsgSize = 100 * 1024
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
	beforeFirstRender()
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
	initTask()
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

	String() string
}

type Task struct {
	ExternalID        string            `json:"external_id"`
	Name              string            `json:"name"`
	AK                string            `json:"access_key"`
	PostURL           string            `json:"post_url"`
	CurStatus         string            `json:"status"`
	Disabled          uint8             `json:"disabled"`
	Frequency         string            `json:"frequency"`
	Region            string            `json:"region"`
	OwnerExternalID   string            `json:"owner_external_id"`
	Tags              map[string]string `json:"tags,omitempty"`
	Labels            []string          `json:"labels,omitempty"`
	WorkspaceLanguage string            `json:"workspace_language,omitempty"`
	TagsInfo          string            `json:"tags_info,omitempty"` // deprecated
	DFLabel           string            `json:"df_label,omitempty"`
	UpdateTime        int64             `json:"update_time,omitempty"`
	ConfigVars        []*ConfigVar      `json:"config_vars,omitempty"`
	ExtractedVars     []*ConfigVar
	CustomVars        []*ConfigVar

	taskJSONString       string
	parsedTaskJSONString string
	child                TaskChild

	rawTask    string
	inited     bool
	globalVars map[string]Variable
	option     map[string]string
}

func CreateTaskChild(taskType string) (TaskChild, error) {
	var ct TaskChild
	switch taskType {
	case "http", "https", ClassHTTP:
		ct = &HTTPTask{}

	case "multi", ClassMulti:
		ct = &MultiTask{}

	case "headless", "browser", ClassHeadless:
		return nil, fmt.Errorf("headless task deprecated")

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
		return nil, fmt.Errorf("invalid task")
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
		return nil, fmt.Errorf("invalid task, not ITask")
	} else {
		t.SetTaskJSONString(taskString)
		t.SetChild(task)
		return t, nil
	}
}

func (t *Task) String() string {
	b, _ := json.Marshal(t.child)
	return string(b)
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

	_, err := time.ParseDuration(t.Frequency)
	if err != nil {
		return err
	}

	return t.CheckTask()
}

func (t *Task) Run() error {
	t.Clear()
	return t.child.run()
}

func (t *Task) init() error {
	defer func() {
		t.inited = true
	}()

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

// RenderTemplateAndInit render template and init task.
func (t *Task) RenderTemplateAndInit(globalVariables map[string]Variable) error {
	// first render
	if !t.inited {
		t.child.beforeFirstRender()
	}

	if globalVariables == nil {
		globalVariables = make(map[string]Variable)
	}

	t.globalVars = globalVariables

	fm := template.FuncMap{}

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

	// multi task does not need to render template
	// render template only for its child task
	if t.Class() == ClassMulti {
		return nil
	}

	tmpl, err := template.New("task").Funcs(fm).Option("missingkey=zero").Parse(t.taskJSONString)
	if err != nil {
		return fmt.Errorf("parse template error: %w", getTemplateError(err))
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		return fmt.Errorf("execute template error: %w", err)
	}

	parsedString := buf.String()

	// no need to re-parse
	if parsedString != t.parsedTaskJSONString {
		t.parsedTaskJSONString = parsedString
		if err := json.Unmarshal([]byte(parsedString), t.child); err != nil {
			return fmt.Errorf("unmarshal parsed template error: %w", err)
		}
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
