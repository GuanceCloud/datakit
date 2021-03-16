package file_collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	f, handler, err := r.FormFile("file")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if handler.Size > F.MaxUploadSize*1024*1024 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("file too large"))
		return
	}
	remotePath := F.getRemotePath(handler.Filename)
	defer f.Close()
	switch F.UploadType {
	case "oss":
		err = F.OssClient.OSSUPLoad(remotePath, f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		resp := map[string]string{
			"url": F.OssClient.GetOSSUrl(remotePath),
		}
		body, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(body)
	case "sftp":
		err = F.SftpClient.SFTPUPLoad(remotePath, f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
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
