package lua

import (
	"crypto/md5"
	"fmt"
	"io"
	"sync"
)

// lscript global connection pool to databases
var connPool sync.Map

func connKey(s string, arr ...string) string {

	h := md5.New()
	io.WriteString(h, s)

	for _, a := range arr {
		io.WriteString(h, a)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
