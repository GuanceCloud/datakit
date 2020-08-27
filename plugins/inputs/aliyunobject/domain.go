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
#[inputs.aliyunobject.domain]

# ## @param - custom tags - [list of Domain instanceid] - optional
#instanceids = []

# ## @param - custom tags - [list of excluded Domain instanceid] - optional
#exclude_instanceids = []

# ## @param - custom tags for Domain object - [list of key:value element] - optional
#[inputs.aliyunobject.domain.tags]
# key1 = 'val1'
`
)

type Domain struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstanceIDs        []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
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
			moduleLogger.Infof("pageNume %v, pagesize %v", pageNum, pageSize)

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

	var objs []*map[string]interface{}

	for _, d := range resp.Data.Domain {
		//moduleLogger.Debugf("dbinstanceInfo %+#v", db)

		inc := false
		for _, isid := range ag.Domain.InstanceIDs {
			if isid == d.InstanceId {
				inc = true
				break
			}
		}

		if len(ag.Domain.InstanceIDs) > 0 && !inc {
			continue
		}

		exclude := false
		for _, isId := range ag.Domain.ExcludeInstanceIDs {
			if d.InstanceId == isId {
				exclude = true
				break
			}
		}

		if exclude {
			continue
		}

		tags := map[string]interface{}{
			"__class":                  "aliyun_domain",
			"__provider":               "aliyun",
			"InstanceId":               d.InstanceId,
			"DomainStatus":             d.DomainStatus,
			"ZhRegistrantOrganization": d.ZhRegistrantOrganization,
			"Email":                    d.Email,
			"DomainType":               d.DomainType,
			"DomainName":               d.DomainName,
			"ProductId":                d.ProductId,
			"ExpirationDateStatus":     d.ExpirationDateStatus,
			"RegistrantType":           d.RegistrantType,
			"Premium":                  d.Premium,
			"DomainAuditStatus":        d.DomainAuditStatus,
			"DomainGroupName":          d.DomainGroupName,
			"RegistrantOrganization":   d.RegistrantOrganization,
			"DomainGroupId":            d.DomainGroupId,
		}

		//add Domain object custom tags
		for k, v := range dm.Tags {
			tags[k] = v
		}

		//add global tags
		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

		obj := &map[string]interface{}{
			"__name":                 fmt.Sprintf(`Domain_%s`, d.InstanceId),
			"__tags":                 tags,
			"RegistrationDate":       d.RegistrationDate,
			"RegistrationDateLong":   d.RegistrationDateLong,
			"Remark":                 d.Remark,
			"ExpirationDateLong":     d.ExpirationDateLong,
			"ExpirationDate":         d.ExpirationDate,
			"ExpirationCurrDateDiff": d.ExpirationCurrDateDiff,
			"DnsList":                d.DnsList,
		}

		objs = append(objs, obj)
	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}
