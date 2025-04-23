// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	"github.com/GuanceCloud/pipeline-go/ptinput"
)

const MaxErrorMessageSize = 1024

type ScriptHTTPRequestResponse struct {
	Header     http.Header `json:"header"`
	Body       string      `json:"body"`
	StatusCode int         `json:"status_code"`
}

func (h *ScriptHTTPRequestResponse) String() (string, error) {
	if bytes, err := json.Marshal(h); err != nil {
		return "", fmt.Errorf("response marshal failed: %w", err)
	} else {
		return string(bytes), nil
	}
}

type ScriptHTTPResult struct {
	IsFailed     bool   `json:"is_failed"`
	ErrorMessage string `json:"error_message"`
}

type Vars map[string]interface{}

type ScriptResult struct {
	Result ScriptHTTPResult `json:"result"`
	Vars   Vars             `json:"vars"`
}

type ScriptHTTPMessage struct {
	Response *ScriptHTTPRequestResponse `json:"response"`
	Vars     *Vars                      `json:"vars"`
}

func (m *ScriptHTTPMessage) String() (string, error) {
	if bytes, err := json.Marshal(m); err != nil {
		return "", fmt.Errorf("response marshal failed: %w", err)
	} else {
		return string(bytes), nil
	}
}

// postScriptDo run pipeline script and return result.
//
// bodyBytes is the body of the response and resp is the response from server.
func postScriptDo(script string, bodyBytes []byte, resp *http.Response) (*ScriptResult, error) {
	if script == "" || resp == nil {
		return &ScriptResult{}, nil
	}

	response := &ScriptHTTPRequestResponse{
		Header:     resp.Header,
		Body:       string(bodyBytes),
		StatusCode: resp.StatusCode,
	}

	if result, err := runPipeline(script, response, nil); err != nil {
		return nil, fmt.Errorf("run pipeline failed: %w", err)
	} else {
		return result, nil
	}
}

func runPipeline(script string, response *ScriptHTTPRequestResponse, vars *Vars) (*ScriptResult, error) {
	scriptName := "script"

	script = fmt.Sprintf(`
	content = load_json(_)
	response = content["response"]
	vars = content["vars"]
	result = {} 

	%s	

	add_key(result, result)
	add_key(vars, vars)
	`, script)

	pls, errs := platypus.NewScripts(
		map[string]string{scriptName: script},
		lang.WithCat(point.Logging),
	)

	defer func() {
		for _, pl := range pls {
			pl.Cleanup()
		}
	}()

	for k, v := range errs {
		return nil, fmt.Errorf("new scripts failed: %s, %w", k, v)
	}

	pl, ok := pls[scriptName]
	if !ok {
		return nil, fmt.Errorf("script %s not found", scriptName)
	}

	if vars == nil {
		vars = &Vars{}
	}

	message := &ScriptHTTPMessage{
		Response: response,
		Vars:     vars,
	}

	messageString, err := message.String()
	if err != nil {
		return nil, fmt.Errorf("message marshal failed: %w", err)
	}

	fileds := map[string]interface{}{
		"message": messageString,
	}

	pt := ptinput.NewPlPoint(point.Logging, "test", nil, fileds, time.Now())

	if err := pl.Run(pt, nil, nil); err != nil {
		return nil, fmt.Errorf("run failed: %w", err)
	}

	resultFields := pt.Fields()

	result := ScriptHTTPResult{}

	if val, ok := resultFields["result"]; !ok {
		return nil, fmt.Errorf("result not found")
	} else if err := json.Unmarshal([]byte(getFiledString(val)), &result); err != nil {
		return nil, fmt.Errorf("unmarshal result failed: %w", err)
	}

	if val, ok := resultFields["vars"]; !ok {
		return nil, fmt.Errorf("vars not found")
	} else if err := json.Unmarshal([]byte(getFiledString(val)), &vars); err != nil {
		return nil, fmt.Errorf("unmarshal vars failed: %w", err)
	}

	// limit error message length
	if len(result.ErrorMessage) > MaxErrorMessageSize {
		result.ErrorMessage = result.ErrorMessage[:MaxErrorMessageSize] + "..."
	}

	return &ScriptResult{
		Result: result,
		Vars:   *vars,
	}, nil
}

func getFiledString(filed any) string {
	switch v := filed.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", filed)
	}
}
