// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

type DWOption func(*Dataway)

// WithGlobalTags add new global tags for the dataway.
func WithGlobalTags(maps ...map[string]string) DWOption {
	return func(dw *Dataway) {
		if dw.globalTags == nil {
			dw.globalTags = map[string]string{}
		}

		for _, tags := range maps {
			for k, v := range tags {
				dw.globalTags[k] = v
			}
		}

		l.Infof("dataway set globals: %+#v", dw.globalTags)
	}
}

// WithURLs add new dataway URLs for the dataway.
func WithURLs(urls ...string) DWOption {
	return func(dw *Dataway) {
		dw.URLs = append(dw.URLs, urls...)
	}
}

// WithWALWorkers set WAL flush workers.
func WithWALWorkers(n int) DWOption {
	return func(dw *Dataway) {
		if n > 0 {
			dw.WAL.Workers = n
		}
	}
}

// TODO: add more options for dataway instance.
