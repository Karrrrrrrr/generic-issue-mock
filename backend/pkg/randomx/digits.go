package randomx

import "crypto/rand"

func Digits(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	for index := range bytes {
		bytes[index] = '0' + bytes[index]%10
	}

	return string(bytes)
}
