/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type CceJobStatus struct {
	// 作业的状态，有如下四种状态：  - JobPhaseInitializing JobPhase = \"Initializing\" - JobPhaseRunning JobPhase = \"Running\" - JobPhaseFailed JobPhase = \"Failed\" - JobPhaseSuccess JobPhase = \"Success\"
	Phase *string `json:"phase,omitempty"`
	// 作业变为当前状态的原因
	Reason *string `json:"reason,omitempty"`
}

func (o CceJobStatus) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CceJobStatus struct{}"
	}

	return strings.Join([]string{"CceJobStatus", string(data)}, " ")
}
