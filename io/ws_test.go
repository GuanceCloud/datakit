package io

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

type MsgGetInputConfig struct {
	Names []string `json:"names"`
}

func TestB64(t *testing.T) {
    //var names []string
	var a MsgGetInputConfig
	data ,_ := base64.StdEncoding.DecodeString("WyJhbnNpYmxlIiwiY3B1Il0=")

	//a := string(data)
	//t.Fatal(string(data),data)
	_ = json.Unmarshal(data,&a.Names)

}