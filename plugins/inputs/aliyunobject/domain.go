package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/domain"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	domainSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.domain]
    # ##(optional) ignore this object, default is false
    #disable = false

    # ##(optional) list of Domain instanceid
    #instanceids = []

    # ##(optional) list of excluded Domain instanceid
    #exclude_instanceids = []
`
)

type Domain struct {
	Disable            bool              `toml:"disable"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstanceIDs        []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (dm *Domain) disabled() bool {
	return dm.Disable
}

func (dm *Domain) run(ag *objectAgent) {
	var cli *domain.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = domain.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err == nil {
			break
		}
		moduleLogger.Errorf("%s", err)
		datakit.SleepContext(ag.ctx, time.Second*3)
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		pageNum := 1
		pageSize := 100
		req := domain.CreateQueryDomainListRequest()

		for {
			moduleLogger.Debugf("pageNume %v, pagesize %v", pageNum, pageSize)

			req.PageNum = requests.NewInteger(pageNum)
			req.PageSize = requests.NewInteger(pageSize)

			resp, err := cli.QueryDomainList(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				dm.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)

				break
			}

			if !resp.NextPage {
				break
			}

			pageNum++

			req.PageNum = requests.NewInteger(pageNum)

		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (dm *Domain) handleResponse(resp *domain.QueryDomainListResponse, ag *objectAgent) {

	moduleLogger.Debugf("TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalItemNum, resp.PageSize, resp.CurrentPageNum)

	var objs []map[string]interface{}

	for _, d := range resp.Data.Domain {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`Domain_%s`, d.InstanceId), `aliyun_domain`, d, d.InstanceId, dm.ExcludeInstanceIDs, dm.InstanceIDs); obj != nil {
			objs = append(objs, obj)
		} else {
			if err != nil {
				moduleLogger.Errorf("%s", err)
			}
		}
	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}
	if ag.isDebug() {
		fmt.Printf("%s\n", string(data))
	} else {
		io.NamedFeed(data, io.Object, inputName)
	}
}
