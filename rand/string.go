package rand

import "math/rand"

const randomKey = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomString(length int) string {
	data := make([]byte, length)

	for i := 0; i < length; i++ {
		data[i] = randomKey[rand.Intn(len(randomKey))]
	}

	return string(data)
}
