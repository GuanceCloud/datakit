// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

import "encoding/json"

func (x *candidate) keepalive() (int, error) {
	body, err := x.puller.ElectionHeartbeat(x.namespace, x.id)
	if err != nil {
		log.Error(err)
		return electionIntervalDefault, err
	}

	e := electionResult{}
	if err := json.Unmarshal(body, &e); err != nil {
		log.Error(err)
		return electionIntervalDefault, err
	}

	log.Debugf("result body: %s", body)

	CurrentElected = e.Content.IncumbencyID

	switch e.Content.Status {
	case statusFail.String():
		electionStatusVec.Reset() // cleanup election status if election ok
		x.status = statusFail
		x.pausePlugins()

	case statusSuccess.String():
		log.Debugf("%s election keepalive ok", x.id)

	default:
		log.Warnf("unknown election status: %s", e.Content.Status)
	}
	return e.Content.Interval, nil
}
