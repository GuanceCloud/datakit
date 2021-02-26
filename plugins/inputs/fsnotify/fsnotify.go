package fsnotify

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"path/filepath"
	"os"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"github.com/fsnotify/fsnotify"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"time"
	"fmt"
	"strings"
)

var (
	inputName  = `fsnotiry`
	l          = logger.DefaultSLogger("fsnotify")
	modifyFile = make(map[string]time.Time)
)

func (_ *Fsn) SampleConfig() string {
	return sampleConfig
}

func (_ *Fsn) Catalog() string {
	return inputName
}

func (_ *Fsn) Test() (*inputs.TestResult, error) {
	testResult := &inputs.TestResult{}
	return testResult, nil
}

func (fsn *Fsn) initFsn() {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		l.Errorf("[error] new watch err:%s", err.Error())
		return
	}
	fsn.watch = watch

	switch fsn.UploadType {
	case "oss":
		oc, err := fsn.OssClient.GetOSSCli()
		if err != nil {
			l.Errorf("[error] oss new client err:%s", err.Error())
			return
		}
		fsn.OssClient.Cli = oc
	case "ft-oss":
	case "sftp":
		sc, err := fsn.SftpClient.GetSFTPClient()
		if err != nil {
			l.Errorf("[error] sftp new client err:%s", err.Error())
			return
		}
		fsn.SftpClient.Cli = sc
	}

}

func (fsn *Fsn) Run() {
	l = logger.SLogger(inputName)
	if !datakit.FileExist(fsn.Path) {
		l.Errorf("[error] file:%s not exist", fsn.Path)
		return
	}
	fsn.initFsn()

	filepath.Walk(fsn.Path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			err := fsn.watch.Add(path)
			if err != nil {
				l.Errorf("[error] fsnotify add watch err:%s", err.Error())
				return err
			}
		}
		return nil
	})

	for {
		select {
		case ev := <-fsn.watch.Events:
			switch ev.Op {
			case fsnotify.Create:
				fsn.WriteLogByCreate(ev)
			case fsnotify.Chmod:
				fsn.WriteLogByChmod(ev)
			case fsnotify.Remove:
				fsn.WriteLogByRemove(ev)
			case fsnotify.Rename:
				fsn.WriteLogByRename(ev)
			case fsnotify.Write:
				fsn.WriteLogByWrite(ev)
			}
		case <-datakit.Exit.Wait():
			l.Info("fsnotify exit")
			fsn.watch.Close()
			return
		case err := <-fsn.watch.Errors:
			l.Errorf("[error] fsnotify err:%s", err.Error())
			return
		}

	}

}

func (fsn *Fsn) getRemotePath(name string) string {
	token := datakit.Cfg.MainCfg.DataWay.GetToken()
	hostName := datakit.Cfg.MainCfg.Hostname
	name = strings.ReplaceAll(name,"/","-")
	name = strings.ReplaceAll(name,"\\","-")
	return filepath.Join(token, hostName, name)
}

func (fsn *Fsn) WriteLog(ev fsnotify.Event, fields map[string]interface{}) {
	tags := map[string]string{
		"source":      inputName,
		"path":        fsn.Path,
		"filename":    ev.Name,
		"upload_type": fsn.UploadType,
	}

	remotePath := fsn.getRemotePath(ev.Name)
	switch fsn.UploadType {
	case "oss", "ft-oss":
		tags["endpoint"] = fsn.OssClient.EndPoint
		tags["bucket"] = fsn.OssClient.BucketName
		tags["remote_path"] = remotePath
		tags["url"] = fmt.Sprintf("https://%s.%s/%s", fsn.OssClient.BucketName, fsn.OssClient.EndPoint, remotePath)
	case "sftp":
		tags["user"] = fsn.SftpClient.User
		tags["remote_path"] = remotePath
		tags["remote_host"] = fsn.SftpClient.Host
	default:
	}
	fmt.Println(fields["message"])
	io.NamedFeedEx(inputName, io.Logging, inputName, tags, fields, time.Now())
}

func (fsn *Fsn) LoadFile(fi os.FileInfo, ev fsnotify.Event) {
	remotePath := fsn.getRemotePath(ev.Name)
	if fi.Size() <= fsn.MaxUploadSize*1024*1024 {
		f, err := os.Open(ev.Name)
		if err != nil {
			l.Errorf("[error] fsnotify openfile err:%s", err.Error())
			return
		}
		defer f.Close()
		switch fsn.UploadType {
		case "oss", "ft-oss":
			if err := fsn.OssClient.OSSUPLoad(remotePath, f); err != nil {
				l.Errorf("[error] fsnotify ossupload err:%s", err.Error())
				return
			}
		case "sftp":
			if err := fsn.SftpClient.SFTPUPLoad(remotePath, f); err != nil {
				l.Errorf("[error] fsnotify sftpupload err:%s", err.Error())
				return
			}
		}
	}
}

func (fsn *Fsn) WriteLogByCreate(ev fsnotify.Event) {
	fi, err := os.Stat(ev.Name)
	if err != nil {
		return
	}
	if fi.IsDir() {
		fsn.watch.Add(ev.Name)
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 内创建了新的文件%s", fsn.Path, ev.Name),
		"size":    fi.Size(),
	}
	fsn.WriteLog(ev, fields)
	go fsn.LoadFile(fi, ev)
}

func (fsn *Fsn) WriteLogByRemove(ev fsnotify.Event) {
	dir := filepath.Dir(ev.Name)
	if !datakit.FileExist(dir) {
		_ = fsn.watch.Remove(ev.Name)
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 内删除了文件%s", fsn.Path, ev.Name),
	}
	fsn.WriteLog(ev, fields)
}

func (fsn *Fsn) WriteLogByChmod(ev fsnotify.Event) {
	fi, err := os.Stat(ev.Name)
	if err != nil {
		l.Errorf("[error] fsnotify os.stat err:%s", err.Error())
		return
	}
	if fi.IsDir() {
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 内文件 %s 变更了权限", fsn.Path, ev.Name),
		"size":    fi.Size(),
	}
	fsn.WriteLog(ev, fields)
}

func (fsn *Fsn) WriteLogByWrite(ev fsnotify.Event) {
	fi, err := os.Stat(ev.Name)
	if err != nil {
		l.Errorf("[error] fsnotify os.stat err:%s", err.Error())
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 中文件 %s 被修改了", fsn.Path, ev.Name),
		"size":    fi.Size(),
	}
	if t, ok := modifyFile[ev.Name]; ok {
		if time.Since(t) < time.Minute*5 {
			return
		}
	}
	modifyFile[ev.Name] = time.Now()
	fsn.WriteLog(ev, fields)
	go fsn.LoadFile(fi, ev)
}

func (fsn *Fsn) WriteLogByRename(ev fsnotify.Event) {
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 中 %s 被重命名了", fsn.Path, ev.Name),
	}
	if t, ok := modifyFile[ev.Name]; ok {
		if time.Since(t) < time.Minute*5 {
			return
		}
	}
	modifyFile[ev.Name] = time.Now()
	fsn.WriteLog(ev, fields)
}


func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Fsn{}
	})
}
