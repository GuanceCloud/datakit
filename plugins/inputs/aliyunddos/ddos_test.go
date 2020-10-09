package aliyunddos

import (
	"testing"
)

func TestHandle(t *testing.T) {
	t.Run("case-gatherGlobalStatuses", func(t *testing.T) {
		d := DDoS{}
		d.AccessKeyID = ""
		d.AccessKeySecret = ""
		d.Interval = "5s"

		d.Run()
		t.Log("ok")
	})
}
