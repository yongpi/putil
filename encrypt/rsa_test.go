package encrypt

import (
	"encoding/base64"
	"fmt"
	"testing"
)

var (
	publicKey = `-----BEGIN RSA PUBLIC KEY-----
MIGJAoGBAJ2AEc8rpigxWo68ie+RjKUAZWOd53iuHCJIAa+KA3QDQ7ioWXQAHxWQ
ajebTThG2tynP9vwjL5a604I23xweytulXOKpeRC+CU4SzE4e82iQ86JiVcbpaoq
txHi8EQ2Ai31mYYaGYA2tu0M6/TV8ZgvpAvdNPZBIEUWvIPcwcDDAgMBAAE=
-----END RSA PUBLIC KEY-----`

	privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCdgBHPK6YoMVqOvInvkYylAGVjned4rhwiSAGvigN0A0O4qFl0
AB8VkGo3m004Rtrcpz/b8Iy+WutOCNt8cHsrbpVziqXkQvglOEsxOHvNokPOiYlX
G6WqKrcR4vBENgIt9ZmGGhmANrbtDOv01fGYL6QL3TT2QSBFFryD3MHAwwIDAQAB
AoGAWT+mV+AjIql86F9cRn4S1blYus9SyGbZGG/3TJKHjGkBKhVzwzECbETOe74s
TtSP1vOLU0WHa6K3rhuEzIossLdAkmxYyTJtGgQQyXG+z/me2SbIZa8Q441XIo3k
cUxUfRofLdqDZeaoN5IJmIuB7DjGFLQPbzer/g4/CyTniUECQQDKOEsSEgTmg8ZO
y2ph5G5G4Hh7gPpbVKlcmqzFPvRyGM4jRc0QB7KrzBbQUlUssQmYc9fyZRq32+aE
aXlGT7bjAkEAx2MjC+UZu6LyRWOyR0ol0Jj+89H4X5IQzZiYEJvmPiWSQ2P8CA4Q
VtnH8ei1I+czJRf4H3n91IHMX/QjJjIUoQJAHbrG9qIljEpFRmJLgpbVy5/GtsmQ
hQreV1n6GomV4IxbCf6CFmA7WVyI4hmooghpE7u8PMu2cN9odYEYLkkb5wJAKokG
p/n29GV9o7nyBW1XBdotwZwQjWretM2R2zE2/BkNy9yfnqRJbg3FruDDC+a9rXMg
lq5yrQwHqoytlu9mIQJBAJfQHUspnxKT7BNA8y/h7CQqf1Ky83thc5NjXWQH4SWO
e4CMrZDC12iMH84/oXLzaX02cvLY4mm9dSgJ5l/RFok=
-----END RSA PRIVATE KEY-----`
)

func TestRSA(t *testing.T) {
	data := "999000"

	pbk := base64.StdEncoding.EncodeToString([]byte(publicKey))
	fmt.Println(pbk)

	dk, err := base64.StdEncoding.DecodeString(pbk)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(dk))

	ed, err := RSAEncrypt([]byte(data), publicKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(ed))

	ed, err = RSAEncrypt([]byte(data), publicKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(ed))

	dd, err := RSADecrypt(ed, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(dd))
	fmt.Println(string(dd) == data)
}

func TestRSADecrypt(t *testing.T) {
	ed := "FBx9Qcf69jpqQH21xWN7yKegUcS9r31KMiQWnY92fzvKXGjwyK9BVXFXhmxA10gGkxePqHW/eLf6nsI6QrM/On6xYp40XHqWMLp25Qd5Ulx6dyhARxoZLqzBUzYtUVnxFmd5V+ssqvrNug2zxRyy9StyJMX1yaTKIlNFz9ZHK98="

	dd, err := base64.StdEncoding.DecodeString(ed)
	if err != nil {
		t.Fatal(err)
	}

	data, err := RSADecrypt(dd, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))
}
