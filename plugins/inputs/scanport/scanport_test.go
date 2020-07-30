package scanport

import (
	"testing"
)

func TestHandle(t *testing.T) {
	t.Run("case-tracerouter", func(t *testing.T) {
		scan := Scanport{}
		scan.Ips = []string{"127.0.0.1"}
		scan.PortStart = 3000
		scan.PortEnd = 60000
		scan.Protocol = []string{"tcp"}
		scan.handle()
	})
}
