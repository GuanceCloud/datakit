package huaweicloud

import (
	"log"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	ak        = `Z82IZCNXFPYCESM1RO0I`
	sk        = `mxvSedXBYgBwIFO5hZVRs86B7FDYi5174MEQFzd1`
	endPoint  = `rds.cn-east-3.myhuaweicloud.com`
	projectId = `09e22a05420025c82f7dc01f0efb10f2`
	l         = logger.DefaultSLogger("huaweicloud")
)

func TestListRDS(t *testing.T) {

	cli := NewHWClient(ak, sk, endPoint, projectId, l)
	opts := map[string]string{
		`datastore_type`: `MySQL`,
	}
	res, err := cli.RdsList(opts)
	if err != nil {
		log.Printf("[error] %v", err)
		return
	}

	log.Printf("%+#v", res)

}
