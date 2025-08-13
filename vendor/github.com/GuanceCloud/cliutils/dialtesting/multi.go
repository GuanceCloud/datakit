// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"encoding/json"
	"fmt"
	"text/template"
	"time"
)

var (
	_ TaskChild = (*MultiTask)(nil)
	_ ITask     = (*MultiTask)(nil)
)

type MultiStepRetry struct {
	Retry    int `json:"retry"`    // retry times
	Interval int `json:"interval"` // ms
}

type MultiExtractedVar struct {
	Name   string `json:"name"`
	Field  string `json:"field"`
	Secure bool   `json:"secure"`
	Value  string `json:"value,omitempty"`
}

type MultiStep struct {
	Type          string              `json:"type"` // http or wait
	Name          string              `json:"name"` // name
	AllowFailure  bool                `json:"allow_failure"`
	Retry         *MultiStepRetry     `json:"retry"`
	TaskString    string              `json:"task,omitempty"`
	Value         int                 `json:"value,omitempty"` // wait seconds for wait task
	ExtractedVars []MultiExtractedVar `json:"extracted_vars,omitempty"`

	result           map[string]interface{}
	postScriptResult *ScriptResult
}

type MultiTask struct {
	*Task
	Steps []*MultiStep `json:"steps"`

	postScriptResult *ScriptResult
	duration         time.Duration
	extractedVars    []MultiExtractedVar
	lastStep         int
	lastHTTPStep     int
}

func (t *MultiTask) clear() {
	t.duration = 0
	t.extractedVars = nil
	t.lastStep = -1
	t.lastHTTPStep = -1

	for _, step := range t.Steps {
		step.result = nil
	}
}

func (t *MultiTask) stop() {}

func (t *MultiTask) class() string {
	return ClassMulti
}

func (t *MultiTask) metricName() string {
	return `multi_dial_testing`
}

func (t *MultiTask) getResults() (tags map[string]string, fields map[string]interface{}) {
	fields = map[string]interface{}{
		"success":       -1,
		"response_time": int64(t.duration) / 1000,
		"last_step":     t.lastStep,
	}

	tags = map[string]string{
		"status": "FAIL",
		"name":   t.Name,
	}
	for k, v := range t.Tags {
		tags[k] = v
	}

	if t.lastHTTPStep > -1 {
		step := t.Steps[t.lastHTTPStep]
		if step.result != nil {
			if step.result["status"] == "OK" {
				tags["status"] = "OK"
				fields["success"] = 1
			} else if step.result["fail_reason"] != nil {
				fields["fail_reason"] = step.result["fail_reason"]
			}
			fields["message"] = step.result["message"]
		}
	}

	steps := []map[string]interface{}{}

	for _, s := range t.Steps {
		// extraced vars
		extractedVars := []MultiExtractedVar{}
		for _, v := range s.ExtractedVars {
			ev := MultiExtractedVar{
				Name:   v.Name,
				Field:  v.Field,
				Secure: v.Secure,
			}

			if !v.Secure {
				ev.Value = v.Value
			}

			extractedVars = append(extractedVars, ev)
		}
		result := map[string]interface{}{}
		if s.result != nil {
			for k, v := range s.result {
				result[k] = v
			}
		}
		result["extracted_vars"] = extractedVars
		steps = append(steps, result)
	}

	bytes, _ := json.Marshal(steps)
	fields["steps"] = string(bytes)

	return tags, fields
}

func (t *MultiTask) runHTTPStep(step *MultiStep) (map[string]interface{}, error) {
	var err error
	var task ITask
	runCount := 0
	maxCount := 1
	interval := time.Millisecond

	result := map[string]interface{}{}
	if step == nil {
		return nil, fmt.Errorf("step should not be nil")
	}

	if step.Retry != nil {
		if step.Retry.Retry > 0 {
			maxCount = step.Retry.Retry + 1
		}
		interval = time.Duration(step.Retry.Interval) * time.Millisecond
	}

	for runCount < maxCount {
		httpTask := &HTTPTask{}
		task, err = NewTask(step.TaskString, httpTask)
		if err != nil {
			return nil, fmt.Errorf("new task failed: %w", err)
		}
		task.SetOption(t.GetOption())
		for _, v := range t.extractedVars {
			task.AddExtractedVar(&ConfigVar{
				Name:   v.Name,
				Secure: v.Secure,
				Value:  v.Value,
			})
		}
		// inject custom vars
		task.SetCustomVars(t.ConfigVars)
		err = task.RenderTemplateAndInit(nil)
		if err != nil {
			err = fmt.Errorf("init http step task failed: %w", err)
		} else {
			err = task.Run()
			if err != nil {
				err = fmt.Errorf("run http step task failed: %w", err)
			}
			tags, fields := task.GetResults()
			for k, v := range tags {
				result[k] = v
			}
			for k, v := range fields {
				result[k] = v
			}
		}

		if httpTask.postScriptResult != nil { // set extracted vars
			step.postScriptResult = httpTask.postScriptResult
			for i, v := range step.ExtractedVars {
				value, ok := httpTask.postScriptResult.Vars[v.Field]
				varValue := ""
				if value != nil {
					varValue = fmt.Sprintf("%v", value)
				}
				if ok && !v.Secure && value != nil {
					step.ExtractedVars[i].Value = varValue
				}

				// set extracted vars, which can be used in next step
				t.extractedVars = append(t.extractedVars,
					MultiExtractedVar{
						Name:   v.Name,
						Value:  varValue,
						Secure: v.Secure,
						Field:  v.Field,
					})
			}
		}
		task.Stop()

		runCount++
		if runCount < maxCount {
			time.Sleep(interval)
		}
	}

	if err == nil {
		if len(result) > 0 && result["status"] != "OK" {
			err = fmt.Errorf("run HTTP step task failed: %s", result["fail_reason"])
		}
	}

	return result, err
}

func (t *MultiTask) run() error {
	if len(t.Steps) == 0 {
		return fmt.Errorf("no steps found")
	}
	now := time.Now()
	lastStep := -1     // last step which is run
	lastHTTPStep := -1 // last step which is not wait

	isLastStepFailed := false
	for i, step := range t.Steps {
		step.result = map[string]interface{}{
			"type":          step.Type,
			"allow_failure": step.AllowFailure,
			"task":          step.TaskString,
		}

		if !isLastStepFailed {
			step.result["task_start_time"] = time.Now().UnixMilli()
			lastStep = i
		}

		// run step
		switch step.Type {
		case "http":
			if isLastStepFailed {
				httpTask := &HTTPTask{}
				_, err := NewTask(step.TaskString, httpTask)
				if err == nil {
					step.result["name"] = httpTask.Name
					step.result["method"] = httpTask.Method
					step.result["url"] = httpTask.URL
				}
			} else {
				if i > lastHTTPStep {
					lastHTTPStep = i
				}

				result, err := t.runHTTPStep(step)

				for k, v := range result {
					step.result[k] = v
				}
				if err != nil {
					step.result["fail_reason"] = fmt.Sprintf("run task failed: %s", err.Error())
					if !step.AllowFailure {
						isLastStepFailed = true
					}
				}
				// set post script result
				if i == len(t.Steps)-1 {
					t.postScriptResult = step.postScriptResult
				}
			}

		case "wait":
			step.result["value"] = step.Value
			if !isLastStepFailed {
				time.Sleep(time.Duration(step.Value) * time.Second)
			}
		default:
			return fmt.Errorf("step type should be wait or http")
		}
	}

	t.duration = time.Since(now)
	t.lastStep = lastStep
	t.lastHTTPStep = lastHTTPStep

	return nil
}

func (t *MultiTask) check() error {
	for _, step := range t.Steps {
		if step.Retry != nil {
			if step.Retry.Retry < 0 || step.Retry.Retry > 5 {
				return fmt.Errorf("retry should be in 0 ~ 5")
			}

			if step.Retry.Interval < 0 || step.Retry.Interval > 5000 {
				return fmt.Errorf("retry interval should be in 0 ~ 5000")
			}
		}
		switch step.Type {
		case "wait":
			if step.Value <= 0 || step.Value > 180 {
				return fmt.Errorf("wait step value should be in 1 ~ 180")
			}

		case "http":
			if step.TaskString == "" {
				return fmt.Errorf("http step task should not be empty")
			}

			task, err := NewTask(step.TaskString, &HTTPTask{})
			if err != nil {
				return fmt.Errorf("new task failed: %w", err)
			}

			if err := task.CheckTask(); err != nil {
				return fmt.Errorf("check task failed: %w", err)
			}

			for _, v := range step.ExtractedVars {
				if !isValidVariableName(v.Name) {
					return fmt.Errorf("invalid variable name %s", v.Name)
				}
			}

		default:
			return fmt.Errorf("step type should be wait or http")
		}
	}

	return nil
}

func (t *MultiTask) checkResult() (reasons []string, succFlag bool) {
	return nil, false
}

func (t *MultiTask) init() error {
	return nil
}

// TODO.
func (t *MultiTask) getHostName() ([]string, error) {
	hostNames := []string{}
	for _, step := range t.Steps {
		if step.Type == "http" {
			ct := &HTTPTask{}
			task, err := NewTask(step.TaskString, ct)
			if err != nil {
				return nil, fmt.Errorf("new task failed: %w", err)
			}

			if v, err := task.GetHostName(); err != nil {
				return nil, fmt.Errorf("get host name failed: %w", err)
			} else {
				hostNames = append(hostNames, v...)
			}
		}
	}

	return hostNames, nil
}

// TODO.
func (t *MultiTask) getVariableValue(variable Variable) (string, error) {
	return "", fmt.Errorf("not support")
}

func (t *MultiTask) renderTemplate(fm template.FuncMap) error {
	return nil
}

func (t *MultiTask) getRawTask(taskString string) (string, error) {
	task := MultiTask{}

	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal multi task failed: %w", err)
	}

	task.Task = nil

	bytes, _ := json.Marshal(task)
	return string(bytes), nil
}

func (t *MultiTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
}
