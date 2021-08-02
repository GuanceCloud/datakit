package filefd

type FileFdInfo struct {
	Allocated   int64   `json:"allocated"`
	Maximum     int64   `json:"maximum"`
	MaximumMega float64 `json:"maximum_mega"`
}

func GetFileFdInfo() *FileFdInfo {
	info, err := FileFdCollect()
	fileInfo := &FileFdInfo{}

	if err != nil {
		// l.Warnf("fail to get filefd stats, %s", err)
	} else {
		if allocated, ok := info["allocated"]; ok {
			fileInfo.Allocated = allocated
		}

		if maximum, ok := info["maximum"]; ok {
			fileInfo.Maximum = maximum
			fileInfo.MaximumMega = float64(fileInfo.Maximum/1000000) + float64(fileInfo.Maximum%1000000)/1000000.0
		}
	}

	return fileInfo
}
