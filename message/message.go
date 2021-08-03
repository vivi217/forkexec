package message

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"syscall"
)

//MaxSize 防止消息太长，超过内存限制
const MaxSize = 102400
//Header 固定长度为8
type Header struct {
	DataLen int  //msg的长度
	MsgType int //msg类型,用来定义是什么类型的消息
}
//Msg 消息体
type Msg struct {
	Title   string
	Content string
	Size    int
}
//HeaderLen 获取header的长度, 但是这个获取的长度不是固定的，所以不能用
/*
func HeaderLen() (int, error) {
	header := Header{
		MsgType: 1,
		DataLen: 1024,
	}
	var h bytes.Buffer
	enc := gob.NewEncoder(&h)
	err := enc.Encode(header)
	if err != nil {
		return 0, err
	}
	return h.Len(), nil
}*/

//IntToBytes int转byte
func IntToBytes(n int) ([]byte, error) {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.BigEndian, x)
	if err != nil {
		return []byte{}, err
	}
	return bytesBuffer.Bytes(), nil
}

//BytesToInt 字节转换成整形
func BytesToInt(b []byte) (int, error) {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	err := binary.Read(bytesBuffer, binary.BigEndian, &x)
	if err != nil {
		return 0, err
	}
	return int(x), nil
}
//TransHeaderToByte 将header转成byte
func TransHeaderToByte(header Header)([]byte, error) {
	h := make([]byte, 8)
	b1, err := IntToBytes(header.MsgType)
	if err != nil {
		return h, err
	}
	b2, err := IntToBytes(header.DataLen)
	if err != nil {
		return h, err
	}
	copy(h,b1)
	copy(h[len(b1):], b2)
	return h, nil
}

// TransByteToHeader  将byte转成Header
func TransByteToHeader(h []byte)(Header, error) {
	var header Header
	n, err := BytesToInt(h[4:])
	if err != nil {
		return header, nil
	}
	header.DataLen = n
	n, err = BytesToInt(h[0:4])
	if err != nil {
		return header, nil
	}
	header.MsgType = n
	return header, nil
}
//SendMsg 发送消息
func SendMsg(fd int, msg Msg) error {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(msg)
	if err != nil {
		return err
	}
	header := Header{
		MsgType: 1,
		DataLen: b.Len(),
	}
	h,err := TransHeaderToByte(header)
	if err != nil {
		fmt.Println(err)
		return err
	}
	data := make([]byte, len(h) + b.Len())
	n := copy(data, h)
	n = copy(data[len(h):], b.Bytes())
	//fmt.Printf("package len[%d]\n", len(data))
	n, err = syscall.Write(fd, data)
	if err != nil {
		log.Fatalf("write error, %s", err)
		return err
	}
	log.Printf("send msg len[%d]", n)
	return nil
}
//ReadMsg 接收消息
func ReadMsg(fd int, msg *Msg) error {
	//先读消息头部,固定头部长度为8
	h := make([]byte, 8)
	n, err := syscall.Read(fd, h)
	if err != nil {
		log.Printf("read from fd[1] error, %s", err)
		return err
	}
	if n != 8 {
		log.Printf("read from size err, size[%d], %s",n, err)
		return err
	}
	header, err := TransByteToHeader(h)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("msg type[%d]\n", header.MsgType)
	//第二次读消息体,已经从消息头部拿到了消息的长度
	data := make([]byte, header.DataLen)
	n, err = syscall.Read(fd, data)
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
