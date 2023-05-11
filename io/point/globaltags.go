// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

var (
	globalHostTags     = map[string]string{}
	globalElectionTags = map[string]string{}
)

////////////////////////////////////////////////////////////////////////////////

func SetGlobalHostTags(k, v string) {
	globalHostTags[k] = v
}

func SetGlobalHostTagsByMap(in map[string]string) {
	for k, v := range in {
		globalHostTags[k] = v
	}
}

////////////////////////////////////////////////////////////////////////////////

func SetGlobalElectionTags(k, v string) {
	globalElectionTags[k] = v
}

func SetGlobalElectionTagsByMap(in map[string]string) {
	for k, v := range in {
		globalElectionTags[k] = v
	}
}

////////////////////////////////////////////////////////////////////////////////

func GlobalHostTags() map[string]string {
	return copyMapString(globalHostTags)
}

func GlobalElectionTags() map[string]string {
	return copyMapString(globalElectionTags)
}

////////////////////////////////////////////////////////////////////////////////

func ClearGlobalTags() {
	globalHostTags = map[string]string{}
	globalElectionTags = map[string]string{}
}

////////////////////////////////////////////////////////////////////////////////

func copyMapString(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
