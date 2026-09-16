package cardnumber

import (
	"strconv"
	"strings"
)

const length = 16

func Generate(prefix string, sequence int64) (string, bool) {
	if prefix == "" || len(prefix) >= length || sequence <= 0 {
		return "", false
	}

	suffix := strconv.FormatInt(sequence, 10)
	remainingLength := length - len(prefix)
	if len(suffix) > remainingLength {
		return "", false
	}

	return prefix + strings.Repeat("0", remainingLength-len(suffix)) + suffix, true
}
