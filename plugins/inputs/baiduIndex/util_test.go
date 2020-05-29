package baiduIndex

import (
	"fmt"
	"testing"
)

func TestDecrypt(t *testing.T) {
	key := "a.IzsGuLo2U58Z-57+86,.%1-49302"
	data := "8o8"
	res := decrypt(key, data)

	t.Log(res)
	fmt.Println(res)
}
