package rand

import (
	"fmt"
	"testing"
)

func TestRandomString(t *testing.T) {
	fmt.Println(RandomString(10))
}

func BenchmarkRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Println(RandomString(8))
	}
}
