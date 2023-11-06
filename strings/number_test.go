package strings

import (
	"fmt"
	"testing"
)

func TestStringAllNumber(t *testing.T) {
	fmt.Println(StringAllNumber("sss111"))
	fmt.Println(StringAllNumber("0112"))
	fmt.Println(StringAllNumber("-0112"))
	fmt.Println(StringAllNumber("21343535"))
}
