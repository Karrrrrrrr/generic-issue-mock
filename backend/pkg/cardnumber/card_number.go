package cardnumber

import (
	"math/rand/v2"
	"strconv"
	"strings"

	"generic-mock/enums"
)

const length = 16

type GenerateRequest struct {
	Channel  enums.Channel
	Prefix   string
	Sequence int64
}

type GenerateResult struct {
	Bin    string
	Number string
}

func Generate(req GenerateRequest) (GenerateResult, bool) {
	if req.Sequence <= 0 {
		return GenerateResult{}, false
	}

	prefixes := strings.Split(req.Prefix, ",")
	if len(prefixes) > 1 && req.Channel != enums.Channel_PingPong {
		return GenerateResult{}, false
	}

	suffix := strconv.FormatInt(req.Sequence, 10)
	for index, candidate := range prefixes {
		prefix := strings.TrimSpace(candidate)
		if prefix == "" || len(prefix)+len(suffix) > length {
			return GenerateResult{}, false
		}
		for _, digit := range prefix {
			if digit < '0' || digit > '9' {
				return GenerateResult{}, false
			}
		}
		prefixes[index] = prefix
	}

	prefix := prefixes[rand.IntN(len(prefixes))]
	return GenerateResult{
		Bin:    prefix,
		Number: prefix + strings.Repeat("0", length-len(prefix)-len(suffix)) + suffix,
	}, true
}
