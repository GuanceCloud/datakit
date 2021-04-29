package aliyuncms

import (
	"os"
	"testing"
)

func TestInput(t *testing.T) {

	ak := os.Getenv("ALIYUNYUN_AK")
	sk := os.Getenv("ALIYUNYUN_SK")

	ag := NewAgent("debug")
	ag.AccessKeyID = ak
	ag.AccessKeySecret = sk
	ag.RegionID = `cn-shanghai`

	ag.Project = []*Project{
		&Project{
			Namespace: `acs_rds_dashboard`,
		},
		&Project{
			Namespace: `acs_slb_dashboard`,
		},
	}

	ag.Run()
}
