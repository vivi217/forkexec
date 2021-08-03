package message

import (
	"fmt"
	"testing"
)

func TestTransByteToHeader(t *testing.T) {
		b, err := IntToBytes(10)
		if err !=nil {
			t.Error(err)
			return
		}
		fmt.Println(len(b))
		a, err := BytesToInt(b)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("byte to int, [%d]\n", a)
		header := Header{1,12}
		m, err := TransHeaderToByte(header)
		if err != nil {
			t.Error(err)
			return
		}
		headerNew ,err := TransByteToHeader(m)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(headerNew)
}
