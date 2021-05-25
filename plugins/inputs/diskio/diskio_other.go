// +build !linux

package diskio

type diskInfoCache struct{}

func (i *Input) diskInfo(devName string) (map[string]string, error) {
	return nil, nil
}
