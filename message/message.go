package message

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"syscall"
)

type Msg struct {
	Title   string
	Content string
	Size    int
}

func SendMsg(fd int, msg Msg) error {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(msg)
	if err != nil {
		return err
	}
	n, err := syscall.Write(fd, b.Bytes())
	if err != nil {
		log.Fatalf("write error, %s", err)
		return err
	}
	log.Printf("send msg len[%d]", n)
	return nil
}

func ReadMsg(fd int, msg *Msg) error {
	data := make([]byte, 1024)
	n, err := syscall.Read(fd, data)
	if err != nil {
		log.Printf("read from fd[1] error, %s", err)
		return err
	}
	var b bytes.Buffer
	_, err = b.Write(data)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(&b)
	err = dec.Decode(msg)
	log.Printf("child readmsg content:")
	log.Println(msg)
	if n < 1 {
		log.Println("read msg len is 0")
		return errors.New("read msg len is 0")
	}
	return nil
}
