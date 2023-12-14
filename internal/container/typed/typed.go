// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package typed wrap point type
package typed

import (
	"encoding/json"
)

type PointKV struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func NewPointKV(name string) *PointKV {
	return &PointKV{
		name:   name,
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
	}
}

func (p *PointKV) Name() string { return p.name }

func (p *PointKV) Tags() map[string]string { return p.tags }

func (p *PointKV) Fields() map[string]interface{} { return p.fields }

func (p *PointKV) SetTag(key, value string) { p.tags[key] = value }

func (p *PointKV) SetTagIfNotEmpty(key, value string) {
	if value != "" {
		p.SetTag(key, value)
	}
}

func (p *PointKV) GetTag(key string) string { return p.tags[key] }

func (p *PointKV) SetTags(m map[string]string) {
	for k, v := range m {
		p.SetTag(k, v)
	}
}

func (p *PointKV) SetCustomerTags(m map[string]string, keys []string) {
	if len(keys) == 0 || len(m) == 0 {
		return
	}
	for _, key := range keys {
		if v, ok := m[key]; ok {
			p.SetTag(key, v)
		}
	}
}

func (p *PointKV) SetField(key string, value interface{}) { p.fields[key] = value }

func (p *PointKV) SetFields(m map[string]interface{}) {
	for k, v := range m {
		p.SetField(k, v)
	}
}

func (p *PointKV) GetField(key string) interface{} { return p.fields[key] }

func (p *PointKV) DeleteTag(key string) { delete(p.tags, key) }

func (p *PointKV) DeleteField(key string) { delete(p.fields, key) }

func (p *PointKV) String() string {
	if len(p.tags) == 0 && len(p.fields) == 0 {
		return ""
	}

	temp := make(map[string]interface{})

	for k, v := range p.tags {
		temp[k] = v
	}
	for k, v := range p.fields {
		temp[k] = v
	}

	b, err := json.Marshal(temp)
	if err != nil {
		return ""
	}

	return string(b)
}

func MapToJSON(m map[string]string) string {
	if len(m) == 0 {
		// If the map is empty, return an empty string instead of 'null'.
		return ""
	}

	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(b)
}

func TrimString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}
