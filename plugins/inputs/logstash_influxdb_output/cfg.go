package logstash_influxdb_output

import "sync"

const (
	configSample = `
#[[inputs.logstash_influxdb_output]]
#pipeline=''
`

	pipelineSample = ``
)

type logstashInfluxdbOutput struct {
	Pipeline string `toml:"pipeline"`

	pipelinePool *sync.Pool
}
