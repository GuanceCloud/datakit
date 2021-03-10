package file_collector

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestFsnotify(t *testing.T) {
	dir := "/Users/admin/Desktop/aaa"

	os.MkdirAll(dir, 0777)
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		l.Fatal(err)
	}
	err = watch.Add(dir)
	if err != nil {
		l.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second * 1)
		//f,err := os.Create(filepath.Join(dir,"123.txt"))
		//if err != nil {
		//	l.Fatal(err)
		//}
		//f.Close()
		//time.Sleep(time.Second*1)
		//
		//ioutil.WriteFile(filepath.Join(dir,"123.txt"),[]byte("hahah"),0666)
		os.Rename(filepath.Join(dir, "222.txt"), filepath.Join(dir, "123.txt"))

	}()
	for {
		select {
		case ev := <-watch.Events:
			fmt.Println(ev.String())
		}
	}

}

func TestFsn_WriteLogByCreate(t *testing.T) {
	fsn := newfsn()
	go func() {
		time.Sleep(time.Second)
		f, _ := os.Create("123.txt")
		f.Close()
		time.Sleep(time.Second)
		os.Remove("123.txt")
	}()
	select {
	case ev := <-fsn.watch.Events:
		switch ev.Op {
		case fsnotify.Create:
			fsn.WriteLogByCreate(ev)

		}

	}
	time.Sleep(time.Second)
}

func newfsn() *Fsn {
	dir := "/Users/admin/Desktop/aaa"
	os.MkdirAll(dir, 0777)
	dw, _ := datakit.ParseDataway("http://10.100.64.140:9528?token=tkn_12595c1a660711ebb18e46cf65a67f12", "")
	datakit.Cfg.MainCfg.DataWay = dw
	region := "oss-cn-shanghai.aliyuncs.com"
	ak := "LTAI4G1E5j5QX5h1S4kT2qfg"
	sk := "aud5Bwb6tXExMoh5P1XEAinbZCH4kl"
	bucketName := "test20210223"
	cli, err := io.NewOSSClient(region, ak, sk, bucketName)

	//cli,err := io.NewSFTPClient("parallels","hjj940622","10.211.55.6","/home/parallels/Desktop/ccc",22)

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
		OssClient:     cli,
		watch:         watch,
		//SftpClient:cli,
	}
	return fsn

}

func TestFsn_Run(t *testing.T) {
	fsn := newfsn()
	fsn.Run()
}

func TestFileCopy(t *testing.T) {
	dir := filepath.Join(datakit.InstallDir, "log")
	f, err := os.Open(dir)
	if err != nil {
		l.Fatal(err)
	}
	defer f.Close()
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err = FileCopy(f, filepath.Join(datakit.DataDir, "log"))
		if err != nil {
			fmt.Printf("file copy err:%s", err.Error())
			return
		}
	}()

	wg.Wait()
}

func TestFsn_LoadFile(t *testing.T) {
	fsn := newfsn()
	ev := <-fsn.watch.Events
	f, err := os.Stat(ev.Name)
	if err != nil {
		l.Fatal(err)
	}
	fsn.LoadFile(f, ev)
}
