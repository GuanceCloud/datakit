package main

import (
	"log"
	"testing"
)

func TestUpdateLagacyConfig(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	updateLagacyConfig("/usr/local/cloudcare/forethought/datakit")
}
