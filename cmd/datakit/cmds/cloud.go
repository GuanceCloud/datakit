package cmds

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/hostobject"
)

func ShowCloudInfo(cloud string) (map[string]interface{}, error) {

	hostobject.SetLog()

	x := hostobject.Input{}
	return x.SyncCloudInfo(cloud)
}
