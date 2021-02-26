package fsnotify

import (
	"testing"
	"os"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"github.com/fsnotify/fsnotify"
	"time"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"path/filepath"
	"io/ioutil"
)

func TestFsnotify(t *testing.T) {
	dir := "/Users/admin/Desktop/aaa"

	os.MkdirAll(dir,0777)
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		l.Fatal(err)
	}
	err = watch.Add(dir)
	if err != nil {
		l.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second*2)
		ioutil.WriteFile(filepath.Join(dir,"123.txt"),[]byte("hahah"),0666)
	}()
	select {
	case ev := <- watch.Events:
		fmt.Println(ev.String())
	}
	os.RemoveAll(dir)
}













func TestFsn_WriteLogByCreate(t *testing.T) {
	fsn := newfsn()
	go func() {
		time.Sleep(time.Second)
		f,_:= os.Create("123.txt")
		f.Close()
		time.Sleep(time.Second)
		os.Remove("123.txt")
	}()
	select {
	case ev := <- fsn.watch.Events:
		switch ev.Op {
		case fsnotify.Create:
			fsn.WriteLogByCreate(ev)

		}

	}
	time.Sleep(time.Second)
}

func newfsn() *Fsn {
	dir := "/Users/admin/Desktop/aaa"
	os.MkdirAll(dir,0777)
	dw ,_:= datakit.ParseDataway("http://10.100.64.140:9528?token=tkn_12595c1a660711ebb18e46cf65a67f12","")
	datakit.Cfg.MainCfg.DataWay = dw
	region := "oss-cn-shanghai.aliyuncs.com"
	ak := "LTAI4G1E5j5QX5h1S4kT2qfg"
	sk := "aud5Bwb6tXExMoh5P1XEAinbZCH4kl"
	bucketName := "test20210223"
	cli,err := io.NewOSSClient(region,ak,sk,bucketName)

	if err != nil {
		l.Fatal(err)
	}
	watch, err := fsnotify.NewWatcher()
	watch.Add(dir)
	if err != nil {
		l.Fatal("[error] new watch err:%s", err.Error())

	}
	var fsn = &Fsn{
		Path:          dir,
		UploadType:    "oss",
		MaxUploadSize: 32,
		OssClient:    cli ,
		watch:watch,
	}
	return fsn

}

func TestFsn_Run(t *testing.T) {
	fsn := newfsn()
	fsn.Run()
}