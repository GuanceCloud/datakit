package kubernetes

import (
	"github.com/gobuffalo/packr/v2"
)

var (
	kubeBox = packr.New("kubernetes", "./templetes")
)

func Get(name string) (string, error) {
	return kubeBox.FindString(name + ".tpl")
}
