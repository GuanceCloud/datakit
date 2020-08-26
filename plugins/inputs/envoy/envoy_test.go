package envoy

import (
	"testing"
)

func TestMain(t *testing.T) {
	testAssert = true

	var envoyer = Envoy{
		Host:     "127.0.0.1",
		Port:     9901,
		Interval: "10s",
		TLSOpen:  false,
	}

	envoyer.Run()

}
