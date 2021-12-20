package container

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type collector interface {
	Metric(context.Context, chan<- []*job)
	Object(context.Context, chan<- []*job)
	Logging(context.Context)
	Stop()
}

type job struct {
	measurement string
	tags        map[string]string
	fields      map[string]interface{}
	ts          time.Time
	cost        time.Duration
	category    string
}

func (j *job) merge(newJob *job) error {
	if newJob == nil {
		return nil
	}

	if j.measurement != newJob.measurement {
		return fmt.Errorf("two measurement is differect")
	}

	if j.category != newJob.category {
		return fmt.Errorf("two category is differect")
	}

	for k, v := range newJob.tags {
		j.tags[k] = v
	}

	for k, v := range newJob.fields {
		j.fields[k] = v
	}

	return nil
}

func (j *job) setMetric() {
	j.category = datakit.Metric
}

func (j *job) setObject() {
	j.category = datakit.Object
}

func (j *job) addTag(key string, value string) {
	if j.tags == nil {
		return
	}
	j.tags[key] = value
}

func (j *job) addField(key string, value interface{}) {
	if j.fields == nil {
		return
	}
	j.fields[key] = value
}

func (j *job) marshal() ([]byte, error) {
	temp := make(map[string]interface{}, len(j.tags)+len(j.fields))
	for k, v := range j.tags {
		temp[k] = v
	}
	for k, v := range j.fields {
		temp[k] = v
	}
	return json.Marshal(temp)
}
