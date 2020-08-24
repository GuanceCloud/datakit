package main

import (
	"log"
	"testing"
)

func TestUpdateLagacyConfig(_ *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	updateLagacyConfig("/usr/local/cloudcare/forethought/datakit")
}
