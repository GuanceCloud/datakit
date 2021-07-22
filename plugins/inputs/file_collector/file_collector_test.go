package file_collector

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestFsnotify(t *testing.T) {
	dir, _ := os.Getwd()

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

		f, err := os.Create(filepath.Join(dir, "123.txt"))
		if err != nil {
			l.Fatal(err)
		}
		f.Close()

	}()
	select {
	case ev := <-watch.Events:
		fmt.Println(ev.String())
	}

	_ = os.Remove(filepath.Join(dir, "123.txt"))

}

// func TestFsn_WriteLogByCreate(t *testing.T) {
// 	fsn := newfc()
// 	go func() {
// 		time.Sleep(time.Second)
// 		f, _ := os.Create("123.txt")
// 		f.Close()
// 		time.Sleep(time.Second)
// 		os.Remove("123.txt")
// 	}()
// 	select {
// 	case ev := <-fsn.watch.Events:
// 		switch ev.Op {
// 		case fsnotify.Create:
// 			fsn.WriteLogByCreate(ev, time.Now())

// 		}

// 	}
// 	time.Sleep(time.Second)
// }

// func newfc() *FileCollector {
// 	dir, _ := os.Getwd()
// 	dw, _ := datakit.ParseDataway("http://10.100.64.140:9528?token=tkn_12595c1a660711ebb18e46cf65a67f12", "")
// 	datakit.Cfg.DataWay = dw
// 	region := ""
// 	ak := ""
// 	sk := ""
// 	bucketName := ""
// 	cli, err := io.NewOSSClient(region, ak, sk, bucketName)
// 	if err != nil {
// 		l.Fatal(err)
// 	}
// 	watch, err := fsnotify.NewWatcher()
// 	watch.Add(dir)
// 	if err != nil {
// 		l.Fatal("[error] new watch err:%s", err.Error())

// 	}
// 	var fsn = &FileCollector{
// 		Path:          dir,
// 		UploadType:    "oss",
// 		MaxUploadSize: 32 * 1024 * 1024,
// 		OssClient:     cli,
// 		watch:         watch,
// 		//SftpClient:cli,
// 	}
// 	return fsn

// }

// func TestFsn_Run(t *testing.T) {
// 	fsn := newfc()
// 	fsn.Run()
// }

// func TestFileCopy(t *testing.T) {
// 	dir := filepath.Join(datakit.InstallDir, "123.txt")
// 	f, err := os.Open(dir)
// 	if err != nil {
// 		l.Fatal(err)
// 	}

// 	wg := sync.WaitGroup{}
// 	wg.Add(1)

// 	go func() {
// 		defer wg.Done()
// 		err = FileCopy(f, filepath.Join(datakit.DataDir, "123.txt"))
// 		if err != nil {
// 			fmt.Printf("file copy err:%s", err.Error())
// 			return
// 		}
// 		f.Close()
// 	}()

// 	wg.Wait()
// }

// func TestFsn_LoadFile(t *testing.T) {
// 	fsn := newfc()
// 	ev := <-fsn.watch.Events
// 	f, err := os.Stat(ev.Name)
// 	if err != nil {
// 		l.Fatal(err)
// 	}
// 	var u = UploadInfo{
// 		filename:   ev.Name,
// 		Size:       f.Size(),
// 		CreateTime: time.Now(),
// 	}
// 	fsn.LoadFile(u)
// }
