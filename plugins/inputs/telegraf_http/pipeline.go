package telegraf_http

var pipelineCfg = map[string]string{
	"kubernetes_pod_container_cpu_mem_usage_percent": `
json_all()
expr(cpu_usage_core_nanoseconds / (cpu_usage_nanocores * 1000000000) * 100, cpu_usage)
`,

	"docker_container_cpu_usage_percent": `
json_all()
rename("cpu_usage", usage_percent)
`,

	"docker_container_mem_usage_percent": `
json_all()
rename("mem_usage_percent", usage_percent)
`,
}
