package telegraf_http

import (
	influxm "github.com/influxdata/influxdb1-client/models"
)

var globalPointHandle = map[string]func(influxm.Point) ([]byte, error){

	// 添加新字段 cpu_usage，内容为 cpu_usage_core_nanoseconds 和 cpu_usage_nanocores 计算所得
	"kubernetes_pod_container": func(pt influxm.Point) ([]byte, error) {
		fields, err := pt.Fields()
		if err != nil {
			return nil, err
		}

		fields["cpu_usage"] = func() (usage float64) {
			// 两个值的类型都是 int64
			coreNanoseconds := fields["cpu_usage_nanocores"].(int64)
			if coreNanoseconds == 0 {
				return
			}
			nanocores := fields["cpu_usage_nanocores"].(int64)
			if nanocores == 0 {
				return
			}
			// cpu_usage_core_nanoseconds / (cpu_usage_nanocores * 1000000000) * 100
			return float64(coreNanoseconds) / float64((nanocores * 10000000))
		}()

		p, err := influxm.NewPoint(string(pt.Name()), pt.Tags(), fields, pt.Time())
		if err != nil {
			return nil, err
		}

		return []byte(p.String()), nil
	},

	// 添加新字段 cpu_usage，内容为 usage_percent
	"docker_container_cpu": func(pt influxm.Point) ([]byte, error) {
		fields, err := pt.Fields()
		if err != nil {
			return nil, err
		}

		if _, ok := fields["usage_percent"]; !ok {
			return []byte(pt.String()), nil
		}
		fields["cpu_usage"] = fields["usage_percent"]

		p, err := influxm.NewPoint(string(pt.Name()), pt.Tags(), fields, pt.Time())
		if err != nil {
			return nil, err
		}
		return []byte(p.String()), nil
	},

	// 添加新字段 mem_usage_percent，内容为 usage_percent
	"docker_container_mem": func(pt influxm.Point) ([]byte, error) {
		fields, err := pt.Fields()
		if err != nil {
			return nil, err
		}

		if _, ok := fields["usage_percent"]; !ok {
			return []byte(pt.String()), nil
		}
		fields["mem_usage_percent"] = fields["usage_percent"]

		p, err := influxm.NewPoint(string(pt.Name()), pt.Tags(), fields, pt.Time())
		if err != nil {
			return nil, err
		}
		return []byte(p.String()), nil
	},
}
