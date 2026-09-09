// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"encoding/json"
	"errors"
)

// decodeEmittedTestPayload keeps integration tests independent of the
// negotiated emitted-point codec. Production uses the same JSON/PPF1 split in
// decodeJITEmitted; this helper intentionally round-trips only ordinary UTF-8
// fixtures. Byte-preservation tests inspect DecodeFlatPoints directly.
func decodeEmittedTestPayload(payload []byte, target any) error {
	if len(payload) >= 4 && string(payload[:4]) == "PPF1" {
		points, err := DecodeFlatPoints(payload)
		if err != nil {
			return err
		}
		if len(points) != 1 {
			return errors.New("emitted PPF1 envelope must contain exactly one point")
		}
		payload, err = json.Marshal(points[0])
		if err != nil {
			return err
		}
	}
	return json.Unmarshal(payload, target)
}
