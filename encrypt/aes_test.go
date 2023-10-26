package encrypt

import (
	"fmt"
	"testing"
)

var aesKey = "npWJKyIKFeJRowwvaFlMWzJvt2+CWYpH"

func TestAESCFB(t *testing.T) {
	data := "111 测试  aes 222"

	ed, err := AESCFBEncrypt([]byte(data), aesKey)
	if err != nil {
		t.Fatal(err)
	}

	dd, err := AESCFBDecrypt(ed, aesKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(dd))
	fmt.Println(string(dd) == data)
}

func TestAESCBC(t *testing.T) {
	data := "11csc 测试  aes 222"

	ed, err := AESCBCEncrypt([]byte(data), aesKey)
	if err != nil {
		t.Fatal(err)
	}

	dd, err := AESCBCDecrypt(ed, aesKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(dd))
	fmt.Println(string(dd) == data)
}
