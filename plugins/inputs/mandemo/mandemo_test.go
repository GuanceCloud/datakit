package mandemo

import (
	"testing"
)

func TestGetMan(t *testing.T) {
	md, err := GetMan()
	if err != nil {
		t.Error(err)
	}

	t.Log(md)
}
