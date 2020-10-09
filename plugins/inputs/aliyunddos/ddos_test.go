package aliyunddos

import (
	"testing"
)

func TestHandle(t *testing.T) {
	t.Run("case-gatherGlobalStatuses", func(t *testing.T) {
		d := DDoS{}
		d.AccessKeyID = "LTAIqo2UBnC4q78J"
		d.AccessKeySecret = "t43b4XdKq9Bv50pzSy1yIYiIlwTtvd"
		d.Interval = "5s"

		d.Run()
		t.Log("ok")
	})
}
