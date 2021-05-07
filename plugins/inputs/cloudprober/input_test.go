package cloudprober

import (
	"testing"
	"github.com/prometheus/common/expfmt"

	"strings"
	"fmt"
)

func TestParse(t *testing.T) {

	type Metric struct {
		Tags map[string]string
		Fields map[string]interface{}
		t int64
	}

	var Measurement map[string][]*Metric

	s := `
#TYPE success counter
success{ptype="http",probe="google_homepage",dst="www.google.com"} 0 1620371363582
#TYPE timeouts counter
timeouts{ptype="http",probe="google_homepage",dst="www.google.com"} 157 1620371363582
#TYPE mallocs counter
mallocs{ptype="sysvars",probe="sysvars"} 360226 1620371375669
#TYPE version gauge
version{ptype="sysvars",probe="sysvars",val="v0.11.2"} 1 1620371375669
`
	var parse expfmt.TextParser
	m,err := parse.TextToMetricFamilies(strings.NewReader(s))
	if err != nil {
		l.Fatal(err)
	}
 	_ = Measurement
	for _,v := range m {

		fmt.Printf("%+#v\n",v.GetType())
		fmt.Printf("%s\n",v.GetMetric()[0])
	}

}
