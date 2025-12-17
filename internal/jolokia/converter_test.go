// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jolokia

import (
	"reflect"
	"testing"
)

func TestMakeTagKey(t *testing.T) {
	tests := []struct {
		name     string
		tags     map[string]string
		expected string
	}{
		{
			name:     "empty tags",
			tags:     map[string]string{},
			expected: "",
		},
		{
			name:     "single tag",
			tags:     map[string]string{"key1": "value1"},
			expected: "key1\x00value1",
		},
		{
			name:     "multiple tags",
			tags:     map[string]string{"key1": "value1", "key2": "value2"},
			expected: "key1\x00value1\x00key2\x00value2",
		},
		{
			name:     "tags with special characters in value",
			tags:     map[string]string{"key1": "value|with|pipe", "key2": "value=with=equals"},
			expected: "key1\x00value|with|pipe\x00key2\x00value=with=equals",
		},
		{
			name:     "tags with null byte in value",
			tags:     map[string]string{"key1": "value\x00with\x00null"},
			expected: "key1\x00value\x00with\x00null",
		},
		{
			name:     "tags with unicode characters",
			tags:     map[string]string{"key1": "中文", "key2": "🚀"},
			expected: "key1\x00中文\x00key2\x00🚀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeTagKey(tt.tags)
			if result != tt.expected {
				t.Errorf("makeTagKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMakeTagKeyConsistency(t *testing.T) {
	// Test that same tag sets generate same keys
	tags1 := map[string]string{"key1": "value1", "key2": "value2"}
	tags2 := map[string]string{"key2": "value2", "key1": "value1"} // Different order

	key1 := makeTagKey(tags1)

	if key2 := makeTagKey(tags2); key1 != key2 {
		t.Errorf("Same tag sets with different key order should generate same key. key1=%q, key2=%q", key1, key2)
	}

	// Test that different tag sets generate different keys
	tags3 := map[string]string{"key1": "value1", "key2": "value3"}

	if key3 := makeTagKey(tags3); key1 == key3 {
		t.Errorf("Different tag sets should generate different keys. key1=%q, key3=%q", key1, key3)
	}
}

func TestCompactPoints(t *testing.T) {
	tests := []struct {
		name     string
		points   []jpoint
		expected []jpoint
	}{
		{
			name:     "empty points",
			points:   []jpoint{},
			expected: []jpoint{},
		},
		{
			name: "single point",
			points: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
			},
			expected: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
			},
		},
		{
			name: "no duplicate tags - should not compact",
			points: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
				{Tags: map[string]string{"tag2": "value2"}, Fields: map[string]interface{}{"field2": 2}},
			},
			expected: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
				{Tags: map[string]string{"tag2": "value2"}, Fields: map[string]interface{}{"field2": 2}},
			},
		},
		{
			name: "duplicate tags - should compact",
			points: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field2": 2}},
			},
			expected: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1, "field2": 2}},
			},
		},
		{
			name: "multiple duplicate tags - should compact all",
			points: []jpoint{
				{Tags: map[string]string{"tag1": "value1", "tag2": "value2"}, Fields: map[string]interface{}{"field1": 1}},
				{Tags: map[string]string{"tag1": "value1", "tag2": "value2"}, Fields: map[string]interface{}{"field2": 2}},
				{Tags: map[string]string{"tag1": "value1", "tag2": "value2"}, Fields: map[string]interface{}{"field3": 3}},
			},
			expected: []jpoint{
				{Tags: map[string]string{"tag1": "value1", "tag2": "value2"}, Fields: map[string]interface{}{"field1": 1, "field2": 2, "field3": 3}},
			},
		},
		{
			name: "mixed duplicate and unique tags",
			points: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field2": 2}},
				{Tags: map[string]string{"tag2": "value2"}, Fields: map[string]interface{}{"field3": 3}},
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field4": 4}},
			},
			expected: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1, "field2": 2, "field4": 4}},
				{Tags: map[string]string{"tag2": "value2"}, Fields: map[string]interface{}{"field3": 3}},
			},
		},
		{
			name: "overlapping field keys - later values should overwrite",
			points: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 2}},
			},
			expected: []jpoint{
				{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 2}},
			},
		},
		{
			name: "empty tags",
			points: []jpoint{
				{Tags: map[string]string{}, Fields: map[string]interface{}{"field1": 1}},
				{Tags: map[string]string{}, Fields: map[string]interface{}{"field2": 2}},
			},
			expected: []jpoint{
				{Tags: map[string]string{}, Fields: map[string]interface{}{"field1": 1, "field2": 2}},
			},
		},
		{
			name: "tags with special characters",
			points: []jpoint{
				{Tags: map[string]string{"tag1": "value|with|pipe"}, Fields: map[string]interface{}{"field1": 1}},
				{Tags: map[string]string{"tag1": "value|with|pipe"}, Fields: map[string]interface{}{"field2": 2}},
			},
			expected: []jpoint{
				{Tags: map[string]string{"tag1": "value|with|pipe"}, Fields: map[string]interface{}{"field1": 1, "field2": 2}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compactPoints(tt.points)

			// Compare results (order may differ, so we need to check by content)
			if len(result) != len(tt.expected) {
				t.Errorf("compactPoints() returned %d points, want %d", len(result), len(tt.expected))
				return
			}

			// Create a map to check if all expected points are present
			expectedMap := make(map[string]jpoint)
			for _, p := range tt.expected {
				key := makeTagKey(p.Tags)
				expectedMap[key] = p
			}

			resultMap := make(map[string]jpoint)
			for _, p := range result {
				key := makeTagKey(p.Tags)
				resultMap[key] = p
			}

			// Check each expected point exists in result
			for key, expectedPoint := range expectedMap {
				resultPoint, exists := resultMap[key]
				if !exists {
					t.Errorf("Expected point with tags %v not found in result", expectedPoint.Tags)
					continue
				}

				// Check tags match
				if !reflect.DeepEqual(resultPoint.Tags, expectedPoint.Tags) {
					t.Errorf("Tags mismatch for key %s: got %v, want %v", key, resultPoint.Tags, expectedPoint.Tags)
				}

				// Check fields match
				if !reflect.DeepEqual(resultPoint.Fields, expectedPoint.Fields) {
					t.Errorf("Fields mismatch for key %s: got %v, want %v", key, resultPoint.Fields, expectedPoint.Fields)
				}
			}
		})
	}
}

func TestCompactPointsOrderIndependence(t *testing.T) {
	// Test that order of input points doesn't affect the result
	points1 := []jpoint{
		{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
		{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field2": 2}},
		{Tags: map[string]string{"tag2": "value2"}, Fields: map[string]interface{}{"field3": 3}},
	}

	points2 := []jpoint{
		{Tags: map[string]string{"tag2": "value2"}, Fields: map[string]interface{}{"field3": 3}},
		{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field2": 2}},
		{Tags: map[string]string{"tag1": "value1"}, Fields: map[string]interface{}{"field1": 1}},
	}

	result1 := compactPoints(points1)
	result2 := compactPoints(points2)

	// Both should have 2 points (one for tag1, one for tag2)
	if len(result1) != 2 || len(result2) != 2 {
		t.Errorf("Both results should have 2 points. result1=%d, result2=%d", len(result1), len(result2))
	}

	// Create maps for comparison
	result1Map := make(map[string]jpoint)
	for _, p := range result1 {
		key := makeTagKey(p.Tags)
		result1Map[key] = p
	}

	result2Map := make(map[string]jpoint)
	for _, p := range result2 {
		key := makeTagKey(p.Tags)
		result2Map[key] = p
	}

	// Check that results are equivalent
	for key, p1 := range result1Map {
		p2, exists := result2Map[key]
		if !exists {
			t.Errorf("Key %s exists in result1 but not in result2", key)
			continue
		}

		if !reflect.DeepEqual(p1.Tags, p2.Tags) {
			t.Errorf("Tags mismatch for key %s: result1=%v, result2=%v", key, p1.Tags, p2.Tags)
		}

		if !reflect.DeepEqual(p1.Fields, p2.Fields) {
			t.Errorf("Fields mismatch for key %s: result1=%v, result2=%v", key, p1.Fields, p2.Fields)
		}
	}
}

func BenchmarkMakeTagKey(b *testing.B) {
	tags := map[string]string{
		"tag1": "value1",
		"tag2": "value2",
		"tag3": "value3",
		"tag4": "value4",
		"tag5": "value5",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = makeTagKey(tags)
	}
}

func BenchmarkCompactPoints(b *testing.B) {
	points := make([]jpoint, 1000)
	for i := 0; i < 1000; i++ {
		points[i] = jpoint{
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			},
			Fields: map[string]interface{}{
				"field1": i,
				"field2": i * 2,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = compactPoints(points)
	}
}

func BenchmarkCompactPointsWithDuplicates(b *testing.B) {
	points := make([]jpoint, 1000)
	for i := 0; i < 1000; i++ {
		// Create points with only 10 unique tag sets
		tagValue := i % 10
		points[i] = jpoint{
			Tags: map[string]string{
				"tag1": "value1",
				"tag2": string(rune('0' + tagValue)),
			},
			Fields: map[string]interface{}{
				"field1": i,
				"field2": i * 2,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = compactPoints(points)
	}
}
