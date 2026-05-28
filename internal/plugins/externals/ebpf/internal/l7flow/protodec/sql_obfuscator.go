//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
)

var sqlObfuscatorPool = sync.Pool{
	New: func() any {
		return obfuscate.NewObfuscator(nil)
	},
}

func obfuscateSQLBytes(clean []byte) []byte {
	obf, _ := sqlObfuscatorPool.Get().(*obfuscate.Obfuscator)
	if obf == nil {
		obf = obfuscate.NewObfuscator(nil)
	}
	defer sqlObfuscatorPool.Put(obf)

	if output, err := obf.Obfuscate("sql", string(clean)); err == nil && output != nil {
		o := []byte(output.Query)
		validLen := utf8ValidLength(o)
		return o[:validLen]
	}
	return []byte(string(bytes.Runes(clean)))
}
