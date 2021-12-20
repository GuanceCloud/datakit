package filecollector

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
)

var fileCollector = &FileCollector{
	maxSize:       32 * 1024 * 1024,
	MaxUploadSize: "32M",

	semStop: cliutils.NewSem(),
}

func Handle(c *gin.Context) {
	t := time.Now()
	fileHeader, err := c.FormFile("file")
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(http.ErrBadReq, err.Error()))
		return
	}
	if fileHeader.Size > fileCollector.maxSize {
		uhttp.HttpErr(c, uhttp.Error(http.ErrBadReq, "file too large"))
		return
	}
	remotePath := fileCollector.getRemotePath(fileHeader.Filename)
	f, err := fileHeader.Open()
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(http.ErrHTTPReadErr, err.Error()))
		return
	}
	defer f.Close() //nolint:errcheck
	switch fileCollector.UploadType {
	case "oss":
		err = fileCollector.OssClient.OSSUPLoad(remotePath, f)
		if err != nil {
			uhttp.HttpErr(c, uhttp.Error(http.ErrUploadFileErr, err.Error()))

			return
		}
		resp := map[string]string{
			"url": fileCollector.OssClient.GetOSSUrl(remotePath),
		}
		http.OK.HttpBody(c, resp)
	case "sftp":
		err = fileCollector.SftpClient.SFTPUPLoad(remotePath, f)
		if err != nil {
			uhttp.HttpErr(c, uhttp.Error(http.ErrUploadFileErr, err.Error()))
			return
		}
		http.OK.HttpBody(c, nil)
	default:
		uhttp.HttpErr(c, uhttp.Error(http.ErrBadReq, "check file_collector config"))
		return
	}
	fields := map[string]interface{}{
		"message": fmt.Sprintf("http通道上传了文件 %s", fileHeader.Filename),
		"size":    fileHeader.Size,
	}
	fileCollector.WriteLog(fileHeader.Filename, fields, t)
}
