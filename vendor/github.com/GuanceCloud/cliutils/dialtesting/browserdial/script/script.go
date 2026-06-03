package script

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	errorsx "github.com/GuanceCloud/cliutils/dialtesting/browserdial/errors"
	"gopkg.in/yaml.v3"
)

type Script struct {
	Name              string            `json:"name" yaml:"name"`
	Target            string            `json:"target" yaml:"target"`
	PostURL           string            `json:"post_url" yaml:"post_url"`
	TimeoutMS         int               `json:"timeout_ms" yaml:"timeout_ms"`
	Headers           map[string]string `json:"headers" yaml:"headers"`
	Cookies           []Cookie          `json:"cookies" yaml:"cookies"`
	IgnoreHTTPSErrors bool              `json:"ignore_https_errors" yaml:"ignore_https_errors"`
	ProxyURL          string            `json:"proxy_url" yaml:"proxy_url"`
	Tags              map[string]string `json:"tags" yaml:"tags"`
	Metadata          map[string]any    `json:"metadata" yaml:"metadata"`
	Auth              Auth              `json:"auth" yaml:"auth"`
	ConfigVars        []ConfigVar       `json:"config_vars" yaml:"config_vars"`
	Steps             []Step            `json:"steps" yaml:"steps"`
}

type Auth struct {
	Mode  string `json:"mode" yaml:"mode"`
	Steps []Step `json:"steps" yaml:"steps"`
}

type ConfigVar struct {
	Name   string `json:"name" yaml:"name"`
	Value  string `json:"value" yaml:"value"`
	Secure bool   `json:"secure" yaml:"secure"`
}

type Cookie struct {
	Name      string `json:"name" yaml:"name"`
	Value     string `json:"value" yaml:"value"`
	ValueFrom string `json:"value_from" yaml:"value_from"`
	Domain    string `json:"domain" yaml:"domain"`
	Path      string `json:"path" yaml:"path"`
	Secure    bool   `json:"secure" yaml:"secure"`
	HTTPOnly  bool   `json:"http_only" yaml:"http_only"`
	SameSite  string `json:"same_site" yaml:"same_site"`
}

type Step struct {
	Name      string `json:"name" yaml:"name"`
	Action    string `json:"action" yaml:"action"`
	URL       string `json:"url" yaml:"url"`
	Selector  string `json:"selector" yaml:"selector"`
	Value     string `json:"value" yaml:"value"`
	ValueFrom string `json:"value_from" yaml:"value_from"`
	Text      string `json:"text" yaml:"text"`
	Contains  string `json:"contains" yaml:"contains"`
	Equals    string `json:"equals" yaml:"equals"`
	TimeoutMS int    `json:"timeout_ms" yaml:"timeout_ms"`
	Sensitive bool   `json:"sensitive" yaml:"sensitive"`
}

var supportedActions = map[string]struct{}{
	"goto":              {},
	"wait_for_selector": {},
	"click":             {},
	"fill":              {},
	"assert_title":      {},
	"assert_url":        {},
	"assert_text":       {},
	"eval":              {},
}

func Load(path string) (Script, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Script{}, errorsx.ScriptLoadError{Message: err.Error()}
	}

	var out Script
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(content, &out); err != nil {
			return Script{}, errorsx.ScriptLoadError{Message: fmt.Sprintf("parse %s: %v", path, err)}
		}
	default:
		if err := yaml.Unmarshal(content, &out); err != nil {
			return Script{}, errorsx.ScriptLoadError{Message: fmt.Sprintf("parse %s: %v", path, err)}
		}
	}

	if err := out.Validate(); err != nil {
		return Script{}, errorsx.ScriptLoadError{Message: err.Error()}
	}
	return out, nil
}

func (s Script) Validate() error {
	if err := s.validateConfigVars(); err != nil {
		return err
	}
	if err := s.validateHeadersAndCookies(); err != nil {
		return err
	}
	if err := s.validateAuth(); err != nil {
		return err
	}
	if len(s.Steps) == 0 {
		return fmt.Errorf("script must define at least one step")
	}
	for index, step := range s.Steps {
		if err := s.validateStep("step", index, step); err != nil {
			return err
		}
	}
	return nil
}

func (s Script) validateHeadersAndCookies() error {
	for key := range s.Headers {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("headers must not contain an empty key")
		}
	}
	for index, cookie := range s.Cookies {
		if strings.TrimSpace(cookie.Name) == "" {
			return fmt.Errorf("cookies %d name is required", index+1)
		}
		if strings.TrimSpace(cookie.ValueFrom) != "" && cookie.Value != "" {
			return fmt.Errorf("cookies %q must not define both value and value_from", cookie.Name)
		}
		switch strings.ToLower(strings.TrimSpace(cookie.SameSite)) {
		case "", "lax", "strict", "none":
		default:
			return fmt.Errorf("cookies %q same_site must be Lax, Strict, or None", cookie.Name)
		}
	}
	return nil
}

func (s Script) validateConfigVars() error {
	seen := map[string]struct{}{}
	for index, variable := range s.ConfigVars {
		name := strings.TrimSpace(variable.Name)
		if name == "" {
			return fmt.Errorf("config_vars %d name is required", index+1)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("config_vars %q is duplicated", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func (s Script) validateAuth() error {
	mode := strings.TrimSpace(strings.ToLower(s.Auth.Mode))
	if mode == "" {
		mode = "none"
	}
	switch mode {
	case "none":
		if len(s.Auth.Steps) > 0 {
			return fmt.Errorf("auth.steps requires auth.mode form")
		}
	case "form":
		if len(s.Auth.Steps) == 0 {
			return fmt.Errorf("auth.steps must define at least one step when auth.mode is form")
		}
	default:
		return fmt.Errorf("auth.mode %q is not supported", s.Auth.Mode)
	}
	for index, step := range s.Auth.Steps {
		if err := s.validateStep("auth step", index, step); err != nil {
			return err
		}
	}
	return nil
}

func (s Script) validateStep(label string, index int, step Step) error {
	action := strings.TrimSpace(step.Action)
	if action == "" {
		return fmt.Errorf("%s %d must define action", label, index+1)
	}
	if _, ok := supportedActions[action]; !ok {
		return fmt.Errorf("%s %d action %q is not supported", label, index+1, action)
	}
	switch action {
	case "goto":
		if step.URL == "" && s.Target == "" {
			return fmt.Errorf("%s %d goto requires url or top-level target", label, index+1)
		}
	case "wait_for_selector", "click", "fill":
		if step.Selector == "" {
			return fmt.Errorf("%s %d %s requires selector", label, index+1, action)
		}
		if action == "fill" && step.Sensitive {
			if strings.TrimSpace(step.ValueFrom) == "" {
				return fmt.Errorf("%s %d sensitive fill requires value_from", label, index+1)
			}
			if step.Value != "" {
				return fmt.Errorf("%s %d sensitive fill must not define value", label, index+1)
			}
		}
	case "assert_text":
		if step.Selector == "" {
			return fmt.Errorf("%s %d assert_text requires selector", label, index+1)
		}
		if expectedText(step) == "" {
			return fmt.Errorf("%s %d assert_text requires contains, equals, or text", label, index+1)
		}
	case "assert_title", "assert_url":
		if expectedText(step) == "" {
			return fmt.Errorf("%s %d %s requires contains, equals, or text", label, index+1, action)
		}
	case "eval":
		if step.Value == "" && step.Text == "" {
			return fmt.Errorf("%s %d eval requires value or text", label, index+1)
		}
	}
	return nil
}

func (s Step) StepName(index int) string {
	if strings.TrimSpace(s.Name) != "" {
		return s.Name
	}
	return fmt.Sprintf("%02d %s", index, s.Action)
}

func Expected(step Step) (mode string, expected string) {
	if step.Equals != "" {
		return "equals", step.Equals
	}
	if step.Contains != "" {
		return "contains", step.Contains
	}
	return "contains", step.Text
}

func expectedText(step Step) string {
	_, expected := Expected(step)
	return expected
}
