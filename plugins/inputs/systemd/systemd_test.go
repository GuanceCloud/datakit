// +build linux

package systemd

import "testing"

func TestMain(t *testing.T) {

	var s = Systemd{
		Interval: "3s",
		Tags: map[string]string{
			"AAA": "111",
			"BBB": "222",
		},
	}
	s.loadcfg()
	defer s.conn.Close()

	data, err := s.getMetrics()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}
