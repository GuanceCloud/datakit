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
	httpPath    = "/upload_file"
	F           = &Fc{}
	fileInfoMap = map[string]string{}
	mtx         = sync.RWMutex{}
	uploadChan  = make(chan UploadInfo)
	lines       []string
	fails       = make(chan UploadInfo)
)

func (_ *Fc) SampleConfig() string {
	return sampleConfig
}

func (_ *Fc) Catalog() string {
	return inputName
}

func (_ *Fc) RegHttpHandler() {
	httpd.RegHttpHandler("POST", httpPath, Handle)
}

func (_ *Fc) Test() (*inputs.TestResult, error) {
	testResult := &inputs.TestResult{}
	return testResult, nil
}

func (fc *Fc) initFsn() error {
	fc.ctx, fc.cancelFun = context.WithCancel(context.Background())

	watch, err := fsnotify.NewWatcher()
	if err != nil {

		return err
	}
	fc.watch = watch

	if fc.MaxUploadSize == 0 {
		fc.MaxUploadSize = 32
	}

	if fc.OssClient != nil {
		fc.UploadType = "oss"
		oc, err := fc.OssClient.GetOSSCli()
		if err != nil {
			return err
		}
		fc.OssClient.Cli = oc

	} else if fc.SftpClient != nil {
		fc.UploadType = "sftp"
		sc, err := fc.SftpClient.GetSFTPClient()
		if err != nil {
			return err
		}
		fc.SftpClient.Cli = sc

	}
	return nil

}

func (fc *Fc) Run() {
	l = logger.SLogger(inputName)
	l.Info("file_collector start")
	if !datakit.FileExist(fc.Path) {
		l.Errorf("[error] file:%s not exist", fc.Path)
		return
	}
	if fc.Path == datakit.DataDir {
		l.Errorf("[error] cannot set datakit data path")
		return
	}
	if err := fc.initFsn(); err != nil {
		l.Errorf("init file collector err :%s", err.Error())
		return
	}
	F = fc
	filepath.Walk(fc.Path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			err := fc.watch.Add(path)
			updateFileInfo(path)
			if err != nil {
				l.Errorf("[error] fsnotify add watch err:%s", err.Error())
				return err
			}
		}
		addFileInfo(path, "")
		return nil
	})
	go fc.handUpload()
	go fc.handFail()
	for {
		select {
		case ev := <-fc.watch.Events:
			notifyTime := time.Now()
			if ev.Op&fsnotify.Write == fsnotify.Write {
				fc.WriteLogByWrite(ev, notifyTime)
				continue
			}
			if ev.Op&fsnotify.Create == fsnotify.Create {
				fc.WriteLogByCreate(ev, notifyTime)
				continue
			}
			if ev.Op&fsnotify.Remove == fsnotify.Remove {
				fc.WriteLogByRemove(ev, notifyTime)
				continue
			}
			if ev.Op&fsnotify.Rename == fsnotify.Write {
				fc.WriteLogByRename(ev, notifyTime)
				continue
			}
		case <-datakit.Exit.Wait():
			l.Info("file_collector exit")
			fc.watch.Close()
			fc.cancelFun()
			return
		case err := <-fc.watch.Errors:
			l.Errorf("[error] fsnotify err:%s", err.Error())
			return
		}
	}
}

func (fc *Fc) handUpload() {
	tick := time.Tick(time.Second * 2)
	for {
		select {
		case <-fc.ctx.Done():
			return
		case u := <-uploadChan:
			if err := fc.LoadFile(u); err != nil {
				l.Error(err)
				fails <- u
				continue
			}
			fc.WriteLog(u.filename, u.Fields, u.CreateTime)
		case <-datakit.Exit.Wait():
			l.Info("file_collector upload exit")
			fc.watch.Close()
			return
		case <-tick:
			if len(lines) > 0 {
				io.NamedFeed([]byte(strings.Join(lines, "\n")), io.Logging, inputName)
				lines = []string{}
			}
		}
	}
}

func (fc *Fc) handFail() {
	for {
		select {
		case <-fc.ctx.Done():
			return
		case <-datakit.Exit.Wait():
			l.Info("file_collector handfail exit")
			return
		case u := <-fails:
			_, err := os.Stat(u.filename)
			if err != nil {
				continue
			}
			if err := fc.LoadFile(u); err != nil {
				continue
			}
			fc.WriteLog(u.filename, u.Fields, u.CreateTime)
		}
	}
}

func (fc *Fc) getRemotePath(name string) string {
	token := datakit.Cfg.MainCfg.DataWay.GetToken()
	hostName := datakit.Cfg.MainCfg.Hostname
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	if fc.UploadType == "sftp" {
		return filepath.Join(fc.SftpClient.UploadPath, token, hostName, name)
	}
	return filepath.Join(token, hostName, name)
}

func (fc *Fc) WriteLog(name string, fields map[string]interface{}, notifyTime time.Time) {
	tags := map[string]string{
		"source":      inputName,
		"path":        fc.Path,
		"filename":    name,
		"upload_type": fc.UploadType,
	}

	remotePath := fc.getRemotePath(name)
	switch fc.UploadType {
	case "oss":
		tags["endpoint"] = fc.OssClient.EndPoint
		tags["bucket"] = fc.OssClient.BucketName
		tags["remote_path"] = remotePath
		if _, err := fc.OssClient.ObjectExist(remotePath); err == nil {
			tags["url"] = fc.OssClient.GetOSSUrl(remotePath)
		}
	case "sftp":
		tags["user"] = fc.SftpClient.User
		tags["remote_path"] = remotePath
		tags["remote_host"] = fc.SftpClient.Host
	default:
	}
	line, err := io.MakeMetric(inputName, tags, fields, notifyTime)
	if err != nil {
		l.Errorf("[error] make metric err:%s", err.Error())
		return
	}
	lines = append(lines, string(line))
}

func (fc *Fc) LoadFile(u UploadInfo) error {
	if u.Size > fc.MaxUploadSize*1024*1024 {
		return nil
	}
	remotePath := fc.getRemotePath(u.filename)
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

	switch fc.UploadType {
	case "oss":
		if err := fc.OssClient.OSSUPLoad(remotePath, copyF); err != nil {
			return err

		}
	case "sftp":
		if err := fc.SftpClient.SFTPUPLoad(remotePath, copyF); err != nil {
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

func (fc *Fc) WriteLogByCreate(ev fsnotify.Event, notifyTime time.Time) {
	time.Sleep(time.Second)
	fi, err := os.Stat(ev.Name)
	if err != nil {
		return
	}
	if fi.IsDir() {
		fc.watch.Add(ev.Name)
		updateFileInfo(ev.Name)
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 内创建了新的文件%s", fc.Path, ev.Name),
		"size":    fi.Size(),
	}
	u := UploadInfo{
		filename:   ev.Name,
		Size:       fi.Size(),
		CreateTime: notifyTime,
		Fields:     fields,
	}
	if fc.UploadType != "" {
		uploadChan <- u
	} else {
		fc.WriteLog(ev.Name, fields, notifyTime)
		addFileInfo(ev.Name, "")
	}
}

func (fc *Fc) WriteLogByRemove(ev fsnotify.Event, notifyTime time.Time) {
	if _, ok := fileInfoMap[ev.Name]; !ok {
		return
	}
	dir := filepath.Dir(ev.Name)
	if !datakit.FileExist(dir) {
		_ = fc.watch.Remove(ev.Name)
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 中文件 %s 被删除了", fc.Path, ev.Name),
	}
	fc.WriteLog(ev.Name, fields, notifyTime)
	if !datakit.FileExist(ev.Name) {
		delete(fileInfoMap, ev.Name)
	}
}

func (fc *Fc) WriteLogByWrite(ev fsnotify.Event, notifyTime time.Time) {
	time.Sleep(time.Second)
	fi, err := os.Stat(ev.Name)
	if err != nil {
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 中文件 %s 被修改了", fc.Path, ev.Name),
		"size":    fi.Size(),
	}
	u := UploadInfo{
		filename:   ev.Name,
		Size:       fi.Size(),
		CreateTime: notifyTime,
		Fields:     fields,
	}
	if fc.UploadType != "" {
		uploadChan <- u
	} else {
		fc.WriteLog(ev.Name, fields, notifyTime)
		addFileInfo(ev.Name, "")
	}
}

func (fc *Fc) WriteLogByRename(ev fsnotify.Event, notifyTime time.Time) {
	if _, ok := fileInfoMap[ev.Name]; !ok {
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("文件夹 %s 中 %s 被重命名了", fc.Path, ev.Name),
	}
	delete(fileInfoMap, ev.Name)
	fc.WriteLog(ev.Name, fields, notifyTime)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Fc{}
	})
}
