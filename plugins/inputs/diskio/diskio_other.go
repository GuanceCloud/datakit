//go:build !linux
// +build !linux

package diskio

func (i *Input) diskInfo(devName string) (map[string]string, error) {
	return nil, nil
}
