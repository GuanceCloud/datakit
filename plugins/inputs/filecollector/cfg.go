package filecollector

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	sampleConfig = `
[[inputs.file_collector]]
   ### monitor file path ,required
   path = ""
   ## upload file max size example  100K 32M 2G
   # max_upload_size = "32M"
   # log status , alert/notice/info ,default info
   # status = "info"
   ## your oss  config
  [inputs.file_collector.oss]
      access_key_id = ""
      access_key_secret = ""
      bucket_name = ""
      endpoint = ""
      domain_name  = ""

  [inputs.file_collector.sftp]
   #   host = ""
   #   port = 22
   #   user = ""
   #   password = ""
   #   upload_path = ""

`
)

type FileCollector struct {
	Path          string `toml:"path"`
	UploadType    string `toml:"upload_type"`
	MaxUploadSize string `toml:"max_upload_size"`
	Status        string `toml:"status"`

	OssClient  *io.OSSClient  `toml:"oss,omitempty"`
	SftpClient *io.SFTPClient `toml:"sftp,omitempty"`

	watch *fsnotify.Watcher

	maxSize   int64
	ctx       context.Context
	cancelFun context.CancelFunc

	semStop          *cliutils.Sem // start stop signal
	semStopCompleted *cliutils.Sem // stop completed signal
}

type UploadInfo struct {
	filename   string
	Size       int64
	CreateTime time.Time
	Md5        string
	SuccessMd5 string
	Fields     map[string]interface{}
}
