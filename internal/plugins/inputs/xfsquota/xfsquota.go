// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package xfsquota

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type QuotaInfo struct {
	Type       string
	ProjectID  string
	UsedBlocks int64
	SoftLimit  int64
	HardLimit  int64
	WarnGrace  string
}

func getXFSQuota(binaryPath string, filesystemPath string) (string, error) {
	cmd := exec.Command(binaryPath, "-x", "-c", "report -ap", filesystemPath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command execution failed: %w, error output: %s", err, stderr.String())
	}
	return out.String(), nil
}

// var quotaInfoRegex = regexp.MustCompile(`^(User|Group)\s+([\w#]+)\s+(\d+)\s+(\d+)\s+(\d+)\s+([\d-]+)\s+([\[\]\w-]+)`).
var quotaInfoRegex = regexp.MustCompile(`(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\S+\s*\[.*\])`)

func parseQuotaOutput(output string) ([]QuotaInfo, error) {
	var results []QuotaInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}

		matches := quotaInfoRegex.FindStringSubmatch(line)
		if len(matches) < 6 {
			continue
		}

		used, _ := strconv.ParseInt(matches[2], 10, 64)
		soft, _ := strconv.ParseInt(matches[3], 10, 64)
		hard, _ := strconv.ParseInt(matches[4], 10, 64)

		results = append(results, QuotaInfo{
			ProjectID:  matches[1],
			UsedBlocks: used,
			SoftLimit:  soft,
			HardLimit:  hard,
			// WarnGrace:  matches[5],
			// GracePeriod: matches[6] + " " + matches[7],
		})
	}

	return results, nil
}
