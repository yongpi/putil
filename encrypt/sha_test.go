package encrypt

import (
	"fmt"
	"testing"
)

func TestSHA256(t *testing.T) {
	password := "ssssdsdsfw22"

	a := SHA256([]byte(password))
	b := SHA256([]byte(password))

	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(a == b)

	fmt.Println(SHA256([]byte(password)))
	fmt.Println(SHA256([]byte(password)))
}
