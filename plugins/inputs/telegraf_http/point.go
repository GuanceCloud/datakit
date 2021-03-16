package telegraf_http

import (
	influxm "github.com/influxdata/influxdb1-client/models"
)

var globalPointHandle = map[string]func(influxm.Point) ([]byte, error){

	// 添加新字段 cpu_usage，内容为 cpu_usage_core_nanoseconds 和 cpu_usage_nanocores 计算所得
	// 添加新字段 mem_usage_percent，内容为 memory_usage_bytes 和 memory_available_bytes 计算所得
	"kubernetes_node": func(pt influxm.Point) ([]byte, error) {
		p, err := addK8sCPUUsage(pt)
		if err != nil {
			return nil, err
		}
		p, err = addK8sMemUsagePercent(p)
		if err != nil {
			return nil, err
		}
		return []byte(p.String()), nil
	},

	// 添加新字段 cpu_usage，内容为 cpu_usage_core_nanoseconds 和 cpu_usage_nanocores 计算所得
	"kubernetes_pod_container": func(pt influxm.Point) ([]byte, error) {
		return addK8sCPUUsageHandle(pt)
	},

	// 添加新字段 cpu_usage，内容为 usage_percent
	"docker_container_cpu": func(pt influxm.Point) ([]byte, error) {
		return copyFieldHandle(pt, "usage_percent", "cpu_usage")
	},

	// 添加新字段 mem_usage_percent，内容为 usage_percent
	"docker_container_mem": func(pt influxm.Point) ([]byte, error) {
		return copyFieldHandle(pt, "usage_percent", "mem_usage_percent")
	},
}

func addK8sCPUUsageHandle(pt influxm.Point) ([]byte, error) {
	p, err := addK8sCPUUsage(pt)
	if err != nil {
		return nil, err
	}
	return []byte(p.String()), nil
}

func addK8sCPUUsage(pt influxm.Point) (influxm.Point, error) {
	fields, err := pt.Fields()
	if err != nil {
		return nil, err
	}

	// 如果没有找到所需字段，原样返回
	// 两个值的类型都是 int64

	var coreNanoseconds int64
	if v, ok := fields["cpu_usage_core_nanoseconds"]; !ok {
		return pt, nil
	} else {
		coreNanoseconds = v.(int64)
	}

	var usageNanocores int64
	if v, ok := fields["cpu_usage_nanocores"]; !ok {
		return pt, nil
	} else {
		usageNanocores = v.(int64)
	}

	// source link: https://github.com/kubernetes/heapster/issues/650#issuecomment-147795824
	// cpu_usage_core_nanoseconds / (cpu_usage_nanocores * 1000000000) * 100
	fields["cpu_usage"] = float64(coreNanoseconds) / float64(usageNanocores*1000000000) * 100

	return influxm.NewPoint(string(pt.Name()), pt.Tags(), fields, pt.Time())
}

func addK8sMemUsagePercent(pt influxm.Point) (influxm.Point, error) {
	fields, err := pt.Fields()
	if err != nil {
		return nil, err
	}

	// 如果没有找到所需字段，原样返回
	// 两个值的类型都是 int64

	var usageBytes int64
	if v, ok := fields["memory_usage_bytes"]; !ok {
		return pt, nil
	} else {
		usageBytes = v.(int64)
	}

	var availableBytes int64
	if v, ok := fields["memory_available_bytes"]; !ok {
		return pt, nil
	} else {
		usageBytes = v.(int64)
	}

	// mem_usage_percent = memory_usage_bytes / (memory_usage_bytes + memory_available_bytes)
	fields["mem_usage_percent"] = float64(usageBytes) / float64(usageBytes+availableBytes)

	return influxm.NewPoint(string(pt.Name()), pt.Tags(), fields, pt.Time())
}

func copyFieldHandle(pt influxm.Point, oldName, newName string) ([]byte, error) {
	p, err := copyField(pt, oldName, newName)
	if err != nil {
		return nil, err
	}
	return []byte(p.String()), nil
}

func copyField(pt influxm.Point, oldName, newName string) (influxm.Point, error) {
	fields, err := pt.Fields()
	if err != nil {
		return nil, err
	}

	// 没有找到源字段，原样返回
	if _, ok := fields[oldName]; !ok {
		return pt, nil
	}

	fields[newName] = fields[oldName]

	return influxm.NewPoint(string(pt.Name()), pt.Tags(), fields, pt.Time())
}
