package file_collector

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"time"
)

const (
	sampleConfig = `
[[inputs.file_collector]]
   ### monitor file path
   path = ""
   ## upload file max size ,unit MB
   # max_upload_size = 32 
   ## your oss  config
   # [inputs.file_collector.oss]
   #   access_key_id = "" 
   #   access_key_secret = ""
   #   bucket_name = ""
   #   endpoint = "" 
   # [inputs.file_collector.sftp]
   #   host = ""
   #   port = 22
   #   user = ""
   #   password = ""
   #   upload_path = ""

`
)

type Fsn struct {
	Path          string `toml:"path"`
	UploadType    string `toml:"upload_type"`
	MaxUploadSize int64  `toml:"max_upload_size"`

	OssClient  *io.OSSClient  `toml:"oss,omitempty"`
	SftpClient *io.SFTPClient `toml:"sftp,omitempty"`

	watch *fsnotify.Watcher

	ctx       context.Context
	cancelFun context.CancelFunc
}

type UploadInfo struct {
	filename   string
	Size       int64
	CreateTime time.Time
	Md5        string
	SuccessMd5 string
	Fields     map[string]interface{}
}
