package http

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
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
	var a MsgGetInputConfig
	data ,_ := base64.StdEncoding.DecodeString("WyJhbnNpYmxlIiwiY3B1Il0=")

	//a := string(data)
	//t.Fatal(string(data),data)
	_ = json.Unmarshal(data,&a.Names)

}

func TestRename(t *testing.T) {
	path := "/Users/admin/Desktop/test/123.conf"
	//path := "/usr/local/cloudcare/dataflux/datakit/conf.d"
	a:= filepath.Dir(path)
	b,_ := ioutil.ReadFile(path)
	newName := fmt.Sprintf("%x.conf", md5.Sum(b))
	newPath := filepath.Join(a,newName)
	fmt.Println(newPath)
	err := os.Rename(path,newPath)
	fmt.Println(err)

}

func TestJson(t *testing.T) {
	a := "example"
	b,err := json.Marshal(a)
	fmt.Println(b,err)
}