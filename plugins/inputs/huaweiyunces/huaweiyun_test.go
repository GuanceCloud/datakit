package huaweiyunces

import (
	"os"
	"testing"
)

func TestInput(t *testing.T) {

	ak := os.Getenv("HUAWEIYUN_AK")
	sk := os.Getenv("HUAWEIYUN_SK")

	ag := newAgent("debug")
	ag.AccessKeyID = ak
	ag.AccessKeySecret = sk
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
