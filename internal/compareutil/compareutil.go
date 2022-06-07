// Package compareutil contains compare utils
package compareutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func CompareListDisordered(listA interface{}, listB interface{}) bool {
	return assert.ElementsMatch(&testing.T{}, listA, listB)
}
