package huaweiyunces

import (
	"fmt"
	"strings"
)

type (
	Dimension struct {
		Name  string `toml:"name" json:"name"`
		Value string `toml:"value" json:"value"`
	}

	Namespace struct {
		Name          string      `toml:"name"`
		MetricSetName string      `toml:"metric_set_name"`
		MetricNames   []string    `toml:"metric_names"`
		Property      []*Property `toml:"property"`

		properties map[string]*Property
	}
)

func (r *metricsRequest) copy() *metricsRequest {
	var m metricsRequest
	m = *r
	return &m
}

func (p *Namespace) checkProperties() error {
	p.properties = make(map[string]*Property)

	for _, prop := range p.Property {

		if prop.Dimensions == nil && p.Name != "SYS.ECS" {
			err := fmt.Errorf("invalid property for %s, dimensions cannot be empty", p.Name)
			moduleLogger.Errorf("%s", err)
			return err
		}

		if prop.Name == "" {
			p.properties["*"] = prop
		} else {
			parts := strings.Split(prop.Name, ",")
			if len(parts) > 1 {
				for _, sn := range parts {
					sn = strings.TrimSpace(sn)
					np := &Property{}
					np.Dimensions = prop.Dimensions
					np.Interval = prop.Interval
					np.Period = prop.Period
					np.Name = sn
					np.Tags = prop.Tags
					np.Filter = prop.Filter
					p.properties[sn] = np
				}
			} else if len(parts) == 1 {
				p.properties[prop.Name] = prop
			}
		}

	}

	return nil
}

func (p *Namespace) applyProperty(req *metricsRequest, instanceIDs []string) (adds []*metricsRequest) {

	ecs := (p.Name == "SYS.ECS")

	metricName := req.metricname

	var property *Property
	property = p.properties[metricName]
	if property == nil {
		property = p.properties["*"]
	}

	if property == nil {
		if ecs {
			property = &Property{
				Name: metricName,
			}
		} else {
			return nil
		}
	}

	if property.Period > 0 {
		req.period = property.Period
	}
	if property.Filter != "" {
		req.filter = property.Filter
	}
	if property.Interval.Duration != 0 {
		req.interval = property.Interval.Duration
	}
	req.tags = property.Tags

	if len(property.Dimensions) == 0 && ecs {
		//如果是ecs且没有设置dimension，则默认拿所有instance
		for idx, inst := range instanceIDs {
			if idx == 0 {
				req.dimsOld = []*Dimension{
					&Dimension{
						Name:  "instance_id",
						Value: inst,
					},
				}
			} else {
				newreq := req.copy()
				newreq.dimsOld = []*Dimension{
					&Dimension{
						Name:  "instance_id",
						Value: inst,
					},
				}
				adds = append(adds, newreq)
			}
		}

	} else {
		dims := []*Dimension{}
		for k, v := range property.Dimensions {
			dim := &Dimension{
				Name:  k,
				Value: v,
			}
			dims = append(dims, dim)
		}
		req.dimsOld = dims
	}

	return
}

func (p *Namespace) genMetricReq(metric string) *metricsRequest {

	req := &metricsRequest{
		namespace:  p.Name,
		metricname: metric,
		filter:     "average",
		period:     300,
	}

	return req
}
