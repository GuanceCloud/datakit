package filefd

type Info struct {
	Allocated   int64   `json:"allocated"`
	Maximum     int64   `json:"maximum"`
	MaximumMega float64 `json:"maximum_mega"`
}

func GetFileFdInfo() (*Info, error) {
	info, err := collect()
	if err != nil {
		return nil, err
	}

	fileInfo := &Info{}
	if allocated, ok := info["allocated"]; ok {
		fileInfo.Allocated = allocated
	}

	//nolint:gomnd
	if maximum, ok := info["maximum"]; ok {
		fileInfo.Maximum = maximum
		fileInfo.MaximumMega = float64(fileInfo.Maximum/1000000) +
			float64(fileInfo.Maximum%1000000)/1000000.0
	}
	return fileInfo, nil
}
