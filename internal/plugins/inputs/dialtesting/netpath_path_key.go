// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func makeNetPathDialtestingPathKey(nodeID, taskID string) string {
	nodeID = strings.TrimSpace(nodeID)
	taskID = strings.TrimSpace(taskID)
	if nodeID == "" || taskID == "" {
		return ""
	}

	sum := sha256.Sum256([]byte("v1\x00" + nodeID + "\x00" + taskID))
	return "np-v1-" + hex.EncodeToString(sum[:16])
}
