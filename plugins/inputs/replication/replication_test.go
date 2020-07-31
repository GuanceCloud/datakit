package replication

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func __init() {
	logger.SetGlobalRootLogger("", logger.DEBUG, logger.OPT_DEFAULT)
	l = logger.SLogger(inputName)
	testAssert = true
}

func TestMain(t *testing.T) {
	__init()

	var r = Replication{
		Host:      "127.0.0.1",
		Port:      5432,
		User:      "repl",
		Password:  "abcd1234",
		Database:  "datakit_test_db",
		Events:    []string{"insert"},
		TagList:   []string{"name"},
		FieldList: []string{"age"},
	}

	r.Run()

}
