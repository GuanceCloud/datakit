package fsnotify

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"github.com/fsnotify/fsnotify"
)

const (
	sampleConfig = `
[[inputs.fsnotify]]
   ### monitor file path
   path = "" 
   ### upload file type, in "oss","ft-oss","sftp","" ; oss mean upload to your oss ;
   ### ft-oss mean upload to dataflux oss ; sftp mean use sftp ; "" or comment out mean not upload
   # upload_type = "" 
   ## upload file max size ,unit MB
   # max_upload_size = 32 
   ## your oss  config
   # [inputs.fsnotify.oss]
   #   access_key_id = "" 
   #   access_key_secret = ""
   #   bucket_name = ""
   #   endpoint = "" 
   # [inputs.fsnotify.sftp]
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
	MaxUploadSize int64 `toml:"max_upload_size"`

	OssClient  *io.OSSClient  `toml:"oss,omitempty"`
	SftpClient *io.SFTPClient `toml:"sftp,omitempty"`

	watch *fsnotify.Watcher
}
