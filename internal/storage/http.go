// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package storage

func ConvertMapToMapEntries(src map[string][]string) (dst []*MapEntry) {
	if len(src) == 0 {
		return nil
	}

	dst = make([]*MapEntry, len(src))
	i := 0
	for k, v := range src {
		dst[i] = &MapEntry{
			Key:   k,
			Value: v,
		}
		i++
	}

	return
}

func ConvertMapEntriesToMap(src []*MapEntry) (dst map[string][]string) {
	if len(src) == 0 {
		return nil
	}

	dst = make(map[string][]string)
	for _, v := range src {
		dst[v.Key] = v.Value
	}

	return
}
