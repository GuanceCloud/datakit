package huaweiyunobject

import (
	"testing"
)

func TestInput(t *testing.T) {

	ag := newAgent("debug")
	ag.AccessKeyID = ``
	ag.AccessKeySecret = ``

	ag.Run()
}
