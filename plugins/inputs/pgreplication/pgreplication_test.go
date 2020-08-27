package pgreplication

import (
	"testing"
)

func TestMain(t *testing.T) {
	testAssert = true

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
