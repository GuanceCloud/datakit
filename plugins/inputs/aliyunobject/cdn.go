package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	cdnSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.cdn]
	# ##(optional) ignore this object, default is false
	#disable = false

    # ##(optional) list of cdn DomainName
    #domainNames = []

    # ##(optional) list of excluded cdn exclude_domainNames
    #exclude_domainNames = []
`
)

type Cdn struct {
	Disable            bool              `toml:"disable"`
	Tags               map[string]string `toml:"tags,omitempty"`
	DomainNames        []string          `toml:"domainNames,omitempty"`
	ExcludeDomainNames []string          `toml:"exclude_domainNames,omitempty"`
}

func (e *Cdn) disabled() bool {
	return e.Disable
}

func (e *Cdn) run(ag *objectAgent) {
	var cli *cdn.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = cdn.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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
		req := cdn.CreateDescribeUserDomainsRequest()

		for {
			moduleLogger.Debugf("pageNume %v, pagesize %v", pageNum, pageSize)
			if len(e.DomainNames) > 0 {
				if pageNum <= len(e.DomainNames) {
					req.DomainName = e.DomainNames[pageNum-1]
				} else {
					break
				}
			} else {
				req.PageNumber = requests.NewInteger(pageNum)
				req.PageSize = requests.NewInteger(pageSize)
			}
			resp, err := cli.DescribeUserDomains(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				e.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				if len(e.DomainNames) > 0 {
					pageNum++
					continue
				}
				break
			}
			if len(e.DomainNames) <= 0 && resp.TotalCount < resp.PageNumber*resp.PageSize {
				break
			}

			pageNum++
			if len(e.DomainNames) <= 0 {
				req.PageNumber = requests.NewInteger(pageNum)
			}
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Cdn) handleResponse(resp *cdn.DescribeUserDomainsResponse, ag *objectAgent) {

	moduleLogger.Debugf("cdn TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	var objs []map[string]interface{}

	for _, inst := range resp.Domains.PageData {

		if obj, err := datakit.CloudObject2Json(inst.DomainName, `aliyun_cdn`, inst, inst.DomainName, e.ExcludeDomainNames, e.DomainNames); obj != nil {
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
