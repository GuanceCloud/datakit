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
	if handler.Size > fileCollector.maxSize {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("file too large"))
		return
	}
	remotePath := fileCollector.getRemotePath(handler.Filename)
	defer f.Close()
	switch fileCollector.UploadType {
	case "oss":
		err = fileCollector.OssClient.OSSUPLoad(remotePath, f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		resp := map[string]string{
			"url": fileCollector.OssClient.GetOSSUrl(remotePath),
		}
		body, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(body)
	case "sftp":
		err = fileCollector.SftpClient.SFTPUPLoad(remotePath, f)
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
		"message": fmt.Sprintf("http通道上传了文件 %s", handler.Filename),
		"size":    handler.Size,
	}
	fileCollector.WriteLog(handler.Filename, fields, time.Now())
	w.WriteHeader(http.StatusOK)
}
