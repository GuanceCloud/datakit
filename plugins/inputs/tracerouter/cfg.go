// +build !windows

package tracerouter

const sampleConfig = `
	# trace domain
	# addr = "www.dataflux.cn"
`

type TraceRouter struct {
	Addr string
}
