package hostobject

import (
	"testing"
)

func TestInput(t *testing.T) {

	ag := newInput("debug")
	ag.Tags = map[string]string{
		"k1": "v1",
	}
	ag.Run()
}
