package strarr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		set      []string
		elem     string
		expected bool
	}{
		{
			name:     "empty set",
			set:      []string{},
			elem:     "test",
			expected: false,
		},
		{
			name:     "nil set",
			set:      nil,
			elem:     "test",
			expected: false,
		},
		{
			name:     "element exists",
			set:      []string{"a", "b", "c"},
			elem:     "b",
			expected: true,
		},
		{
			name:     "element does not exist",
			set:      []string{"a", "b", "c"},
			elem:     "d",
			expected: false,
		},
		{
			name:     "empty element",
			set:      []string{"a", "b", "c"},
			elem:     "",
			expected: false,
		},
		{
			name:     "empty element exists in set",
			set:      []string{"a", "", "c"},
			elem:     "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.set, tt.elem)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiffer(t *testing.T) {
	tests := []struct {
		name     string
		source   []string
		compare  []string
		expected []string
	}{
		{
			name:     "empty source",
			source:   []string{},
			compare:  []string{"a", "b"},
			expected: []string{},
		},
		{
			name:     "empty compare",
			source:   []string{"a", "b"},
			compare:  []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "nil source",
			source:   nil,
			compare:  []string{"a", "b"},
			expected: []string{},
		},
		{
			name:     "nil compare",
			source:   []string{"a", "b"},
			compare:  nil,
			expected: []string{"a", "b"},
		},
		{
			name:     "no difference",
			source:   []string{"a", "b"},
			compare:  []string{"a", "b"},
			expected: []string{},
		},
		{
			name:     "some elements differ",
			source:   []string{"a", "b", "c", "d"},
			compare:  []string{"b", "d"},
			expected: []string{"a", "c"},
		},
		{
			name:     "completely different",
			source:   []string{"a", "b"},
			compare:  []string{"c", "d"},
			expected: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Differ(tt.source, tt.compare)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestIntersect(t *testing.T) {
	tests := []struct {
		name     string
		set1     []string
		set2     []string
		expected []string
	}{
		{
			name:     "empty set1",
			set1:     []string{},
			set2:     []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "empty set2",
			set1:     []string{"a", "b"},
			set2:     []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "nil set1",
			set1:     nil,
			set2:     []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "nil set2",
			set1:     []string{"a", "b"},
			set2:     nil,
			expected: []string{"a", "b"},
		},
		{
			name:     "no intersection",
			set1:     []string{"a", "b"},
			set2:     []string{"c", "d"},
			expected: []string{},
		},
		{
			name:     "some intersection",
			set1:     []string{"a", "b", "c"},
			set2:     []string{"b", "c", "d"},
			expected: []string{"b", "c"},
		},
		{
			name:     "complete intersection",
			set1:     []string{"a", "b"},
			set2:     []string{"a", "b"},
			expected: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersect(tt.set1, tt.set2)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}
