package logstash_influxdb_output

const (
	configSample = `
#[[inputs.logstash_influxdb_output]]
#pipeline=''
`

	pipelineSample = ``
)

type logstashInfluxdbOutput struct {
	Pipeline string `toml:"pipeline"`
}
