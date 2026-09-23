// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"compress/gzip"
	"container/list"
	"fmt"
	"io"
	"sync"

	pprofile "github.com/google/pprof/profile"
	"google.golang.org/protobuf/encoding/protowire"
)

const (
	maxKubernetesProfileDecodedSize = 8 * MiB
	maxKubernetesProfileFields      = 100000
	defaultKubernetesDeltaCacheMB   = 32
	profileDeltaEntryOverhead       = 256
)

type profileDeltaKey struct {
	owner *GoProfiler
	kind  string
}

type profileDeltaEntry struct {
	key  profileDeltaKey
	data []byte
}

type profileDeltaCache struct {
	mu      sync.Mutex
	limit   int64
	used    int64
	entries map[profileDeltaKey]*list.Element
	order   *list.List
}

func newProfileDeltaCache(limit int64) *profileDeltaCache {
	return &profileDeltaCache{limit: limit, entries: make(map[profileDeltaKey]*list.Element), order: list.New()}
}

func (cache *profileDeltaCache) remove(element *list.Element) []byte {
	entry := element.Value.(profileDeltaEntry)
	delete(cache.entries, entry.key)
	cache.order.Remove(element)
	cache.used -= int64(len(entry.data)) + profileDeltaEntryOverhead
	return entry.data
}

func (cache *profileDeltaCache) exchange(owner *GoProfiler, kind string, data []byte) []byte {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if owner.deltaClosed {
		return nil
	}
	key := profileDeltaKey{owner: owner, kind: kind}
	var previous []byte
	if element := cache.entries[key]; element != nil {
		previous = cache.remove(element)
	}
	size := int64(len(data)) + profileDeltaEntryOverhead
	if size > cache.limit {
		observeKubernetesProfileSkipped("delta_cache_limit")
		return nil
	}
	for cache.used+size > cache.limit {
		cache.remove(cache.order.Front())
		observeKubernetesProfileSkipped("delta_cache_evicted")
	}
	cache.entries[key] = cache.order.PushBack(profileDeltaEntry{key: key, data: append([]byte(nil), data...)})
	cache.used += size
	return previous
}

func (cache *profileDeltaCache) release(owner *GoProfiler) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	owner.deltaClosed = true
	for kind := range profileConfigMap {
		if element := cache.entries[profileDeltaKey{owner: owner, kind: kind}]; element != nil {
			cache.remove(element)
		}
	}
}

func parseBoundedProfile(data []byte, bodyLimit int64) (*pprofile.Profile, []byte, error) {
	limit := int64(maxKubernetesProfileDecodedSize)
	if bodyLimit < limit {
		limit = bodyLimit
	}
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid gzip profile")
		}
		decoded, readErr := io.ReadAll(io.LimitReader(reader, limit+1))
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			return nil, nil, fmt.Errorf("cannot decompress profile")
		}
		data = decoded
	}
	if int64(len(data)) > limit {
		return nil, nil, fmt.Errorf("decompressed profile exceeds %d bytes", limit)
	}
	budget := maxKubernetesProfileFields
	if err := checkProfileWireBudget(data, "profile", &budget); err != nil {
		return nil, nil, err
	}
	parsed, err := pprofile.ParseUncompressed(data)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid protobuf profile")
	}
	if err := parsed.CheckValid(); err != nil || len(parsed.SampleType) == 0 {
		return nil, nil, fmt.Errorf("invalid profile structure")
	}
	if !profileExpansionWithinLimit(parsed, limit) {
		return nil, nil, fmt.Errorf("profile exceeds expanded sample budget")
	}
	return parsed, data, nil
}

func profileExpansionWithinLimit(parsed *pprofile.Profile, limit int64) bool {
	for _, sample := range parsed.Sample {
		limit -= int64(len(sample.Value)) * 8
		for key, values := range sample.Label {
			for _, value := range values {
				limit -= int64(len(key) + len(value))
			}
		}
		for key, values := range sample.NumLabel {
			limit -= int64(len(key)+8) * int64(len(values))
		}
		for key, values := range sample.NumUnit {
			for _, value := range values {
				limit -= int64(len(key) + len(value))
			}
		}
		for _, location := range sample.Location {
			limit -= 8
			if mapping := location.Mapping; mapping != nil {
				limit -= int64(len(mapping.File) + len(mapping.BuildID))
			}
			for _, line := range location.Line {
				limit -= 8
				if function := line.Function; function != nil {
					limit -= int64(len(function.Name) + len(function.SystemName) + len(function.Filename))
				}
				if limit < 0 {
					return false
				}
			}
			if limit < 0 {
				return false
			}
		}
		if limit < 0 {
			return false
		}
	}
	return true
}

func checkProfileWireBudget(data []byte, message string, budget *int) error {
	for len(data) > 0 {
		field, wireType, tagSize := protowire.ConsumeTag(data)
		if tagSize < 0 {
			return fmt.Errorf("invalid profile field")
		}
		data = data[tagSize:]
		if wireType == protowire.StartGroupType || wireType == protowire.EndGroupType {
			return fmt.Errorf("invalid profile wire type")
		}
		fieldSize := protowire.ConsumeFieldValue(field, wireType, data)
		if fieldSize < 0 {
			return fmt.Errorf("invalid profile field value")
		}
		*budget -= 1
		if wireType == protowire.BytesType {
			payload, consumed := protowire.ConsumeBytes(data)
			if consumed < 0 {
				return fmt.Errorf("invalid profile field length")
			}
			nested := ""
			switch {
			case message == "profile" && field == 2:
				nested = "sample"
			case message == "profile" && field == 4:
				nested = "location"
			case message == "profile" && (field == 1 || field == 3 || field == 5 || field == 11),
				message == "sample" && field == 3, message == "location" && field == 4:
				nested = "leaf"
			case message == "sample" && (field == 1 || field == 2), message == "profile" && field == 13:
				*budget -= len(payload)
			}
			if nested != "" {
				if err := checkProfileWireBudget(payload, nested, budget); err != nil {
					return err
				}
			}
		}
		if *budget < 0 {
			return fmt.Errorf("profile exceeds parsing budget")
		}
		data = data[fieldSize:]
	}
	return nil
}
