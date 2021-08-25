package msgpack

import (
	"bytes"
	"log"
	"testing"
)

type Texture struct {
	Name string `codec:"name-name"`
	Id   int64  `codec:"id+id"`
}

func TestMsgPack(t *testing.T) {
	src := &Texture{Name: "tnt", Id: 5678987654}
	buf, err := Marshal(src)
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println(string(buf))

	if buf, err = Marshal(src); err != nil {
		log.Fatalln(err.Error())
	}
	log.Println(string(buf))

	newtexture := &Texture{}
	if err = Unmarshal(bytes.NewBuffer(buf), newtexture); err != nil {
		log.Fatalln(err.Error())
	}
	log.Println(*newtexture)
}
