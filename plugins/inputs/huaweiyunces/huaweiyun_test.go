package huaweiyunces

import (
	"testing"
)

func TestInput(t *testing.T) {

	ag := newAgent("debug")
	ag.AccessKeyID = "CGG2MSXKM57HDF0B089B"
	ag.AccessKeySecret = "7ff1iStWknEqAkW872BMwO6B3GHdEsV2VkWFQWy9"
	ag.ApiFrequency = 20

	ag.IncludeMetrics = []string{
		`SYS.ECS`,
		`SYS.OBS`,
		`AGT.ECS`,
		`SYS.VPC`,
		`SYS.EVS`,
		`SYS.ELB`,
		`SYS.RDS`,
	}

	ag.Run()
}
