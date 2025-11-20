// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	"github.com/GuanceCloud/pipeline-go/ptinput"
)

type ScriptGRPCRequestResponse struct {
	Body string `json:"body"`
}

func (h *ScriptGRPCRequestResponse) String() (string, error) {
	bytes, err := json.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("response marshal failed: %w", err)
	}
	return string(bytes), nil
}

type ScriptGRPCMessage struct {
	Response *ScriptGRPCRequestResponse `json:"response"`
	Vars     *Vars                      `json:"vars"`
}

// postScriptDoGRPC run pipeline script for gRPC response and return result.
//
// bodyBytes is the JSON body of the gRPC response.
func postScriptDoGRPC(script string, bodyBytes []byte) (*ScriptResult, error) {
	if script == "" || bodyBytes == nil {
		return &ScriptResult{}, nil
	}

	response := &ScriptGRPCRequestResponse{
		Body: string(bodyBytes),
	}

	result, err := runPipelineGRPC(script, response, nil)
	if err != nil {
		return nil, fmt.Errorf("run pipeline failed: %w", err)
	}
	return result, nil
}

func runPipelineGRPC(script string, response *ScriptGRPCRequestResponse, vars *Vars) (*ScriptResult, error) {
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

	message := &ScriptGRPCMessage{
		Response: response,
		Vars:     vars,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("message marshal failed: %w", err)
	}
	messageString := string(messageBytes)

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
