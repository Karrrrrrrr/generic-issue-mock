package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

func CheckSumsubPayloadDigest(alg, payloadDigest string, body []byte, secret string) bool {
	hashFunc, ok := sumsubPayloadDigestHash(alg)
	if !ok {
		return false
	}

	mac := hmac.New(hashFunc, []byte(secret))
	mac.Write(body)
	calculatedDigest := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(calculatedDigest), []byte(payloadDigest))
}

func sumsubPayloadDigestHash(alg string) (func() hash.Hash, bool) {
	switch alg {
	case "HMAC_SHA256_HEX":
		return sha256.New, true
	case "HMAC_SHA512_HEX":
		return sha512.New, true
	default:
		return nil, false
	}
}
