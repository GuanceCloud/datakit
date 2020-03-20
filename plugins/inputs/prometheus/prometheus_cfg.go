package prometheus

const prometheusConfigSample = `
### You need to configure an [[targets]] for each exporter to be collected.
### host: prometheus exporter url
### interval: gather data every interval second, unit is second. The default value is 60.
### active: whether to gather data from exporter or not.
#[[targets]]
#	host="127.0.0.1:9090"
#	interval = 60
#	active = false

#[[targets]]
#	host="127.0.0.1:8080"
#	interval = 60
#	active = false
`
