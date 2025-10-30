// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"regexp"
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/point"
)

type CustomTags struct {
	lock       sync.RWMutex
	commonTags map[string]string
	regs       []*regexp.Regexp
}

func NewCustomTags(configTags []string, publicTags map[string]string) *CustomTags {
	c := &CustomTags{
		commonTags: publicTags,
		regs:       make([]*regexp.Regexp, 0),
	}

	for _, key := range configTags {
		if strings.HasPrefix(key, "reg:") {
			s := key[4:] // remove "reg:"
			log.Infof("add regexp common attribute: %s", s)
			re := regexp.MustCompile(s)
			if re != nil {
				c.regs = append(c.regs, re)
			}
		} else {
			c.commonTags[key] = strings.ReplaceAll(key, ".", "_")
		}
	}

	return c
}

func (c *CustomTags) AddTag(configTags map[string]string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, key := range configTags {
		if _, ok := c.commonTags[key]; !ok {
			c.commonTags[key] = strings.ReplaceAll(key, ".", "_")
		}
	}
}

func (c *CustomTags) OTELRegexKey(attrKey string, all bool) string {
	// otel 不需要加锁。
	// 固定 tags 匹配
	replaceKey, ok := c.commonTags[attrKey]
	if ok {
		return replaceKey
	}

	if all {
		return strings.ReplaceAll(attrKey, ".", "_")
	}

	// 正则匹配。
	for _, reg := range c.regs {
		if reg.MatchString(attrKey) {
			return strings.ReplaceAll(attrKey, ".", "_")
		}
	}
	return ""
}

func (c *CustomTags) DDTraceRegexKey(meta map[string]string) (map[string]string, point.KVs) {
	c.lock.Lock()
	defer c.lock.Unlock()
	var kvs point.KVs
	for k, v := range meta {
		if k == "runtime-id" {
			kvs = kvs.AddTag("runtime_id", v).AddTag("runtime-id", v)
			delete(meta, k)
			continue
		}

		if k == "db.type" { // db 类型。
			kvs = kvs.AddTag("db_host", meta["peer.hostname"])
		}

		replace, ok := c.commonTags[k]
		if !ok {
			// 尝试从正则中匹配
			for _, tagsReg := range c.regs {
				if tagsReg.MatchString(k) {
					replace = strings.ReplaceAll(k, ".", "_")
					break
				}
			}
		}
		if replace != "" {
			if len(v) > 1024 {
				kvs = kvs.Set(replace, v)
			} else {
				kvs = kvs.SetTag(replace, v)
			}
			// 从 message 中删除 key.
			delete(meta, k)
		}
	}

	return meta, kvs
}
