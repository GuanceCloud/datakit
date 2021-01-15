package hostobject

import (
	"encoding/json"
	"log"
	"testing"
)

func TestMessageObj(t *testing.T) {
	msg := getHostObjectMessage()
	msgdata, _ := json.Marshal(msg)
	log.Printf("%s", string(msgdata))
}
