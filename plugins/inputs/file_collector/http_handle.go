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
	remotePath := F.getRemotePath(handler.Filename)
	defer f.Close()
	switch F.UploadType {
	case "oss":
		go func() {
			err = F.OssClient.OSSUPLoad(remotePath, f)
			if err != nil {
				l.Errorf("http oss upload err:%s", err.Error())
				return
			}
		}()

	case "sftp":
		go func() {
			err = F.SftpClient.SFTPUPLoad(remotePath, f)
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
		"message": fmt.Sprintf("http 通道 上传了文件 %s", handler.Filename),
		"size":    handler.Size,
	}
	F.WriteLog(handler.Filename, fields, time.Now())
	w.WriteHeader(http.StatusOK)
}
