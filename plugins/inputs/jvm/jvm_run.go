package jvm

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (j *JVM) gather() {
	duration, err := time.ParseDuration(j.Interval)
	if err != nil {
		l.Error(err)
		return
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := j.Collect(); err != nil {
				l.Error(err)
			} else {
				inputs.FeedMeasurement(j.MetricName, io.Metric, j.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: false})

				j.collectCache = j.collectCache[:] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			l.Infof("input %s exit", inputName)
			return
		}
	}
}

func (j *JVM) Collect() error {
	if j.client == nil {
		client, err := j.createClient(j.URLs)
		if err != nil {
			return err
		}
		j.client = client
	}

	repObjs, err := j.client.read()
	if err != nil {
		return err
	}

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["urls"] = j.URLs
	for _, obj := range repObjs {
		if obj.Status != 200 {
			l.Errorf("%s with err code %d", obj.Request, obj.Status)
			continue
		}

		key := genKey(obj.Request.Mbean, obj.Request.Attribute, obj.Request.Path)
		if v, ok := convertDict[key]; ok {
			fields[v] = obj.Value
		}
	}

	l.Error(fields)
	j.collectCache = append(j.collectCache, &JvmMeasurement{
		inputName,
		tags,
		fields,
		time.Now(),
	})

	return nil
}
