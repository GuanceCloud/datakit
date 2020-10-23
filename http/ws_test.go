package http

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type MsgGetInputConfig struct {
	Names []string `json:"names"`
}

func TestB64(t *testing.T) {
	//var names []string
	//var a MsgGetInputConfig
	data, _ := base64.StdEncoding.DecodeString("W1tpbnB1dHMuY3B1XV0KCiAgIyMgV2hldGhlciB0byByZXBvcnQgcGVyLWNwdSBzdGF0cyBvciBub3QKICBwZXJjcHUgPSB0cnVlCiAgIyMgV2hldGhlciB0byByZXBvcnQgdG90YWwgc3lzdGVtIGNwdSBzdGF0cyBvciBub3QKICB0b3RhbGNwdSA9IHRydWUKICAjIyBJZiB0cnVlLCBjb2xsZWN0IHJhdyBDUFUgdGltZSBtZXRyaWNzLgogIGNvbGxlY3RfY3B1X3RpbWUgPSBmYWxzZQogICMjIElmIHRydWUsIGNvbXB1dGUgYW5kIHJlcG9ydCB0aGUgc3VtIG9mIGFsbCBub24taWRsZSBDUFUgc3RhdGVzLgogIHJlcG9ydF9hY3RpdmUgPSBmYWxzZQ==")
	fmt.Println(string(data))
	//a := string(data)
	//t.Fatal(string(data),data)
	//_ = json.Unmarshal(data, &a.Names)

}

func TestRename(t *testing.T) {
	path := "/Users/admin/Desktop/test/123.conf"
	//path := "/usr/local/cloudcare/dataflux/datakit/conf.d"
	a := filepath.Dir(path)
	b, _ := ioutil.ReadFile(path)
	newName := fmt.Sprintf("%x.conf", md5.Sum(b))
	newPath := filepath.Join(a, newName)
	fmt.Println(newPath)
	err := os.Rename(path, newPath)
	fmt.Println(err)

}



