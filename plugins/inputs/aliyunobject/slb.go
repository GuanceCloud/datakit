package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	slbSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.slb]
    # ##(optional) ignore this object, default is false
    #disable = false

    # ##(optional) list of slb instanceid
    #instanceids = ['']

    # ##(optional) list of excluded slb instanceid
    #exclude_instanceids = ['']
`
)

type Slb struct {
	Disable            bool              `toml:"disable"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (s *Slb) disabled() bool {
	return s.Disable
}

func (s *Slb) run(ag *objectAgent) {
	var cli *slb.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = slb.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err == nil {
			break
		}
		moduleLogger.Errorf("%s", err)
		if ag.isTest() {
			ag.testError = err
			return
		}
		datakit.SleepContext(ag.ctx, time.Second*3)
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		pageNum := 1
		req := slb.CreateDescribeLoadBalancersRequest()
		req.PageNumber = requests.NewInteger(pageNum)
		req.PageSize = requests.NewInteger(100)

		if len(s.InstancesIDs) > 0 {
			data, err := json.Marshal(s.InstancesIDs)
			if err == nil {
				req.LoadBalancerId = string(data)
			}
		}

		for {
			resp, err := cli.DescribeLoadBalancers(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				s.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				if ag.isTest() {
					ag.testError = err
					return
				}
				break
			}

			if resp.TotalCount < resp.PageNumber*resp.PageSize {
				break
			}
			pageNum++
			req.PageNumber = requests.NewInteger(pageNum)
		}

		if ag.isTest() {
			break
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (s *Slb) handleResponse(resp *slb.DescribeLoadBalancersResponse, ag *objectAgent) {

	moduleLogger.Debugf("SLB TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	var objs []map[string]interface{}

	for _, inst := range resp.LoadBalancers.LoadBalancer {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s(%s)`, inst.LoadBalancerName, inst.LoadBalancerId), `aliyun_slb`, inst, inst.LoadBalancerId, s.ExcludeInstanceIDs, s.InstancesIDs); obj != nil {
			objs = append(objs, obj)
		} else {
			if err != nil {
				moduleLogger.Errorf("%s", err)
			}
		}
	}

	data, err := json.Marshal(&objs)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}

	if ag.isTest() {
		ag.testResult.Result = append(ag.testResult.Result, data...)
	} else if ag.isDebug() {
		fmt.Printf("%s\n", string(data))
	} else {
		io.NamedFeed(data, io.Object, inputName)
	}
}
