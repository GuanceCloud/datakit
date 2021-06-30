package pgreplication

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	var r = Replication{
		Host:     "172.16.0.43",
		Port:     30432,
		User:     "testuser",
		Password: "testuser",
		Database: "test",
		Events:   []string{"INSERT"},
		// TagList:   []string{"name"},
		FieldList: []string{"name"},
		Tags:      map[string]string{"test": "DATAKIT"},
	}

	// var r = Replication{
	// 	Host:     "127.0.0.1",
	// 	Port:     5432,
	// 	User:     "repl",
	// 	Password: "abcd1234",
	// 	Database: "datakit_test_db",
	// 	Events:   []string{"INSERT"},
	// 	// TagList:   []string{"name"},
	// 	FieldList: []string{"name"},
	// 	Tags:      map[string]string{"test": "DATAKIT"},
	// }
	r.Run()

}

func TestRewriteCategory(t *testing.T) {
	testcase := []string{
		"metric",
		"logging",
		"invalid",
		"",
	}

	for _, tc := range testcase {
		m := &Replication{
			Category: tc,
		}
		t.Logf("source: %v\n", m.Category)

		m.rewriteCategory()
		t.Logf("valid:  %v\n\n", m.Category)
	}
}
