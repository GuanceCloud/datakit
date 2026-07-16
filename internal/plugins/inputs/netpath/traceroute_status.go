// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

func setTracerouteResultStatus(tags map[string]string, fields map[string]interface{}, payload traceroutePayload,
	reached bool, runErr error,
) {
	tags["traceroute_status"] = "failed"
	switch {
	case reached:
		tags["traceroute_status"] = "reached"
	case traceroutePayloadHasReachableHop(payload):
		tags["traceroute_status"] = "partial"
	}

	if runErr != nil {
		fields["traceroute_fail_reason"] = runErr.Error()
	} else if tags["traceroute_status"] == "failed" {
		fields["traceroute_fail_reason"] = "traceroute returned no reachable hops"
	}
}

func traceroutePayloadHasReachableHop(payload traceroutePayload) bool {
	for _, run := range payload.Runs {
		for _, hop := range run.Hops {
			if hop.Reachable {
				return true
			}
		}
	}
	return false
}
