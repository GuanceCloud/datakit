package jvm

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (i *Input) gather() {
	duration, err := time.ParseDuration(i.Interval)
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
			if err := i.Collect(); err != nil {
				l.Error(err)
			} else {
				inputs.FeedMeasurement(i.MetricName, io.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: false})

				i.collectCache = i.collectCache[:] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			l.Infof("input %s exit", inputName)
			return
		}
	}
}

func (i *Input) Collect() error {
	if i.client == nil {
		client, err := i.createClient(i.URLs)
		if err != nil {
			return err
		}
		i.client = client
	}

	repObjs, err := i.client.read()
	if err != nil {
		return err
	}

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["urls"] = i.URLs
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
	i.collectCache = append(i.collectCache, &JvmMeasurement{
		inputName,
		tags,
		fields,
		time.Now(),
	})

	return nil
}
