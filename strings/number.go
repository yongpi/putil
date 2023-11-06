package strings

import "unicode"

func StringAllNumber(data string) bool {
	for _, item := range []rune(data) {
		if !unicode.IsNumber(item) {
			return false
		}
	}

	return true
}
