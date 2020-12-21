package configtemplate

import (
	"log"
	"testing"
)

func TestCfgTemp(t *testing.T) {
	ct := NewCfgTemplate(".")
	err := ct.InstallConfigs("test.tar.gz", []byte(`base64://eyJBbGlZdW5fQUsiOiJ0aGlzIGlzIGFrIn0=`))
	if err != nil {
		log.Fatalf("%s", err)
	}
}
