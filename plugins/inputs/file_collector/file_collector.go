package file_collector

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	goIO "io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	inputName   = `file_collector`
	l           = logger.DefaultSLogger("file_collector")
	fc          = &Fsn{}
	fileInfoMap = map[string]string{}
	mtx         = sync.RWMutex{}
	uploadChan  = make(chan UploadInfo)
	lines       []string
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

func (fsn *Fsn) initFsn() error {
	fsn.ctx, fsn.cancelFun = context.WithCancel(context.Background())

	watch, err := fsnotify.NewWatcher()
	if err != nil {

		return err
	}
	fsn.watch = watch

	if fsn.MaxUploadSize == 0 {
		fsn.MaxUploadSize = 32
	}

	if fsn.OssClient != nil {
		fsn.UploadType = "oss"
		oc, err := fsn.OssClient.GetOSSCli()
		if err != nil {
			return err
		}
		fsn.OssClient.Cli = oc

	} else if fsn.SftpClient != nil {
		fsn.UploadType = "sftp"
		sc, err := fsn.SftpClient.GetSFTPClient()
		if err != nil {
			return err
		}
		fsn.SftpClient.Cli = sc

	}
	return nil

}

func (fsn *Fsn) Run() {
	l = logger.SLogger(inputName)

	if !datakit.FileExist(fsn.Path) {
		l.Errorf("[error] file:%s not exist", fsn.Path)
		return
	}
	if fsn.Path == datakit.DataDir {
		l.Errorf("[error] cannot set datakit data path")
		return
	}
	if err := fsn.initFsn(); err != nil {
		l.Errorf("init file collector err :%s", err.Error())
		return
	}
	fc = fsn
	httpd.RegHttpHandler("POST", "/"+inputName, Handle)

	filepath.Walk(fsn.Path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			err := fsn.watch.Add(path)
			updateFileInfo(f.Name())
			if err != nil {
				l.Errorf("[error] fsnotify add watch err:%s", err.Error())
				return err
			}
		}
		return nil
	})
	go func() {

		tick := time.Tick(time.Second * 5)
		for {
			select {
			case <-fsn.ctx.Done():
				return
			case u := <-uploadChan:
				if err := fsn.LoadFile(u); err != nil {
					l.Error(err)
					continue
				}
				fsn.WriteLog(u.filename, u.Fields, u.CreateTime)
			case <-datakit.Exit.Wait():
				l.Info("fsnotify upload exit")
				fsn.watch.Close()
				return
			case <-tick:
				io.NamedFeed([]byte(strings.Join(lines, "\n")), io.Logging, inputName)
				lines = []string{}
			}
		}
	}()

	for {
		select {
		case ev := <-fsn.watch.Events:
			notifyTime := time.Now()
			if ev.Op&fsnotify.Write == fsnotify.Write {
				fsn.WriteLogByWrite(ev, notifyTime)
				continue
			}
			if ev.Op&fsnotify.Create == fsnotify.Create {
				fsn.WriteLogByCreate(ev, notifyTime)
				continue
			}
			if ev.Op&fsnotify.Remove == fsnotify.Remove {
				fsn.WriteLogByRemove(ev, notifyTime)
				continue
			}
			if ev.Op&fsnotify.Rename == fsnotify.Write {
				fsn.WriteLogByRename(ev, notifyTime)
				continue
			}
		case <-datakit.Exit.Wait():
			l.Info("fsnotify exit")
			fsn.watch.Close()
			fsn.cancelFun()
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
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	if fsn.UploadType == "sftp" {
		return filepath.Join(fsn.SftpClient.UploadPath, token, hostName, name)
	}
	return filepath.Join(token, hostName, name)
}

func (fsn *Fsn) WriteLog(name string, fields map[string]interface{}, notifyTime time.Time) {
	tags := map[string]string{
		"source":      inputName,
		"path":        fsn.Path,
		"filename":    name,
		"upload_type": fsn.UploadType,
	}

	remotePath := fsn.getRemotePath(name)
	switch fsn.UploadType {
	case "oss":
		tags["endpoint"] = fsn.OssClient.EndPoint
		tags["bucket"] = fsn.OssClient.BucketName
		tags["remote_path"] = remotePath
		tags["url"] = fsn.OssClient.GetOSSUrl(remotePath)
	case "sftp":
		tags["user"] = fsn.SftpClient.User
		tags["remote_path"] = remotePath
		tags["remote_host"] = fsn.SftpClient.Host
	default:
	}
	line, err := io.MakeMetric(inputName, tags, fields, notifyTime)
	if err != nil {
		l.Errorf("[error] make metric err:%s", err.Error())
		return
	}
	lines = append(lines, string(line))
}

func (fsn *Fsn) LoadFile(u UploadInfo) error {
	if u.Size > fsn.MaxUploadSize*1024*1024 {
		return nil
	}
	remotePath := fsn.getRemotePath(u.filename)
	f, err := os.Open(u.filename)
	if err != nil {
		return err
	}
	MD5, err := getFileMd5(f)
	if err != nil {
		return err
	}
	//防止同一个文件重复上传
	if v, ok := fileInfoMap[u.filename]; ok {
		if v == MD5 {
			return nil
		}
	}

	tmpPath := filepath.Join(datakit.DataDir, remotePath)
	if err := FileCopy(f, tmpPath); err != nil {
		return err
	}
	f.Close()
	copyF, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer copyF.Close()
	defer os.Remove(tmpPath)

	switch fsn.UploadType {
	case "oss":
		if err := fsn.OssClient.OSSUPLoad(remotePath, copyF); err != nil {
			return err

		}
	case "sftp":
		if err := fsn.SftpClient.SFTPUPLoad(remotePath, copyF); err != nil {
			return err
		}
	}
	u.SuccessMd5 = MD5
	addFileInfo(u.filename, MD5)
	return nil
}

func FileCopy(f *os.File, tmpPath string) error {
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return err
	}
	err := os.MkdirAll(filepath.Dir(tmpPath), 0777)
	if err != nil {
		return err
	}
	tmpf, err := os.Create(tmpPath)
	defer tmpf.Close()
	if err != nil {
		return err
	}

	if _, err := goIO.Copy(tmpf, f); err != nil {
		return fmt.Errorf("copy err :%s", err.Error())
	}

	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

func (fsn *Fsn) WriteLogByCreate(ev fsnotify.Event, notifyTime time.Time) {

	fi, err := os.Stat(ev.Name)
	if err != nil {
		return
	}
	if fi.IsDir() {
		fsn.watch.Add(ev.Name)
		updateFileInfo(ev.Name)
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 内创建了新的文件%s", fsn.Path, ev.Name),
		"size":    fi.Size(),
	}
	u := UploadInfo{
		filename:   ev.Name,
		Size:       fi.Size(),
		CreateTime: notifyTime,
		Fields:     fields,
	}
	if fsn.UploadType != "" {
		uploadChan <- u
	} else {
		fsn.WriteLog(ev.Name, fields, notifyTime)
		addFileInfo(ev.Name, "")
	}
}

func (fsn *Fsn) WriteLogByRemove(ev fsnotify.Event, notifyTime time.Time) {
	if _, ok := fileInfoMap[ev.Name]; !ok {
		return
	}
	dir := filepath.Dir(ev.Name)
	if !datakit.FileExist(dir) {
		_ = fsn.watch.Remove(ev.Name)
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 中文件 %s 被删除了", fsn.Path, ev.Name),
	}
	delete(fileInfoMap, ev.Name)
	fsn.WriteLog(ev.Name, fields, notifyTime)
}

func (fsn *Fsn) WriteLogByWrite(ev fsnotify.Event, notifyTime time.Time) {
	if _, ok := fileInfoMap[ev.Name]; !ok {
		return
	}
	time.Sleep(time.Second)

	fi, err := os.Stat(ev.Name)
	if err != nil {
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 中文件 %s 被修改了", fsn.Path, ev.Name),
		"size":    fi.Size(),
	}
	u := UploadInfo{
		filename:   ev.Name,
		Size:       fi.Size(),
		CreateTime: notifyTime,
		Fields:     fields,
	}
	if fsn.UploadType != "" {
		uploadChan <- u
	} else {
		fsn.WriteLog(ev.Name, fields, notifyTime)
		addFileInfo(ev.Name, "")
	}
}

func (fsn *Fsn) WriteLogByRename(ev fsnotify.Event, notifyTime time.Time) {
	if _, ok := fileInfoMap[ev.Name]; !ok {
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 中 %s 被重命名了", fsn.Path, ev.Name),
	}
	delete(fileInfoMap, ev.Name)
	fsn.WriteLog(ev.Name, fields, notifyTime)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Fsn{}
	})
}
