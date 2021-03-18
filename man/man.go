package man

import (
	"github.com/gobuffalo/packr/v2"
)

var (
	ManualBox = packr.New("manulas", "./manuals")
)

func Get(name string) (string, error) {
	return ManualBox.FindString(name + ".md")
}
