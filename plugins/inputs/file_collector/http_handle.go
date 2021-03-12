package file_collector

import (
	"fmt"
	"net/http"
	"time"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	f, handler, err := r.FormFile("file")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	remotePath := fc.getRemotePath(handler.Filename)
	defer f.Close()
	switch fc.UploadType {
	case "oss":
		go func() {
			err = fc.OssClient.OSSUPLoad(remotePath, f)
			if err != nil {
				l.Errorf("http oss upload err:%s", err.Error())
				return
			}
		}()

	case "sftp":
		go func() {
			err = fc.SftpClient.SFTPUPLoad(remotePath, f)
			if err != nil {
				l.Errorf("http sftp upload err:%s", err.Error())
				return
			}
		}()
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("http通道 上传了文件 %s", handler.Filename),
		"size":    handler.Size,
	}
	fc.WriteLog(handler.Filename, fields, time.Now())
	w.WriteHeader(http.StatusOK)
}
