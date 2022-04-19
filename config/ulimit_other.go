//go:build !linux
// +build !linux

package config

func setUlimit(n uint64) error {
	return nil
}

func getUlimit() (soft, hard uint64, err error) {
	return 0, 0, nil
}
