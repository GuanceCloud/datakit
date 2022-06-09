package k8sinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	assert.Equal(t, true, MatchLabel(map[string]string{"a": "b"},
		map[string]string{"a": "b"}))

	assert.Equal(t, true, MatchLabel(map[string]string{}, nil))

	assert.Equal(t, true, MatchLabel(map[string]string{"c": "x"},
		map[string]string{"c": "x", "s": "x"}))

	assert.Equal(t, true, MatchLabel(nil, map[string]string{"a": "b"}))
	assert.Equal(t, true, MatchLabel(map[string]string{},
		map[string]string{"a": "b"}))

	assert.Equal(t, false, MatchLabel(map[string]string{"a": "b"},
		map[string]string{"a": "x"}))

	assert.Equal(t, false, MatchLabel(map[string]string{"a": "b"},
		map[string]string{"c": "x"}))

	assert.Equal(t, false, MatchLabel(map[string]string{"a": "b", "s": "x"},
		map[string]string{"c": "x"}))

	assert.Equal(t, false, MatchLabel(map[string]string{"a": "b", "s": "x"},
		map[string]string{"c": "x", "v": "x"}))
}
