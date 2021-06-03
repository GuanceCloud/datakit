package sqlserver

import "testing"
import (
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
)

func TestCon(t *testing.T) {
	n := Input{
		Host:     "10.100.64.133:1433",
		User:     "_",
		Password: "_",
	}
	if err := n.initDb(); err != nil {
		l.Error(err.Error())
		return
	}

	n.getMetric()
	for _, v := range collectCache {
		fmt.Println(v.String())
	}
}
