package rum

const (
	configSample = `
[[inputs.rum]]
	# Tell datakit which http header contains the source ip, if empty use the client ip
	ip_header = 'X-Forwarded-For' # optional

	#pipeline = '' # optional
`

	pipelineSample = ``
)

var metricNames = map[string]bool{
	`rum_web_page_performance`:          true,
	`rum_web_resource_performance`:      true,
	`rum_app_startup`:                   true,
	`rum_app_system_performance`:        true,
	`rum_app_view`:                      true,
	`rum_app_freeze`:                    true,
	`rum_app_resource_performance`:      true,
	"rum_mini_app_startup":              true,
	"rum_mini_app_page_performance":     true,
	"rum_mini_app_resource_performance": true,
}

var esNames = map[string]bool{
	`js_error`: true,
	`page`:     true,
	`resource`: true,
	`view`:     true,
	`crash`:    true,
	`freeze`:   true,
}

func isMetricData(name string) bool {
	_, ok := metricNames[name]
	return ok
}

func isRUMData(name string) bool {
	_, ok := esNames[name]
	return ok
}
