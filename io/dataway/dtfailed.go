// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import sync "sync"

// send failed used to remember the fail count of dialtesting, see issue:
//  https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/54

var (
	dtFailed    = map[string]int{}
	dtfailedMtx sync.Mutex
)

func updateDTFailInfo(url string, sendOK bool) {
	dtfailedMtx.Lock()
	defer dtfailedMtx.Unlock()

	if _, ok := dtFailed[url]; ok {
		if sendOK {
			dtFailed[url] = 0 // reset
		} else {
			dtFailed[url]++
		}
	} else if !sendOK {
		dtFailed[url] = 1
	}
}

// GetDTFailInfo return the failed count on request the url.
func GetDTFailInfo(url string) int {
	dtfailedMtx.Lock()
	defer dtfailedMtx.Unlock()

	return dtFailed[url]
}
