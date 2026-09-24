package crypto

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/wenzhenxi/gorsa"
)

func PayfulEncryptRSA(originData []byte, privateKey string) (encryptData string, err error) {
	dataContent, err := gorsa.PriKeyEncrypt(string(base64.StdEncoding.EncodeToString(originData)), privateKey)
	if err != nil {
		return "", err
	}

	originByte, _ := base64.StdEncoding.DecodeString(dataContent)
	return hex.EncodeToString(originByte), nil
}

func PayfulDecodeRSA(encryptData string, publicKey string) (originData string, err error) {
	if encryptData == "" {
		return "", nil
	}

	encryptByte, err := hex.DecodeString(encryptData)
	if err != nil {
		return
	}

	originHex, err := gorsa.PublicDecrypt(base64.StdEncoding.EncodeToString(encryptByte), publicKey)
	if err != nil {
		return
	}

	originByte, err := hex.DecodeString(originHex)
	if err != nil {
		return
	}

	originDataByte, err := base64.StdEncoding.DecodeString(string(originByte))
	if err != nil {
		return
	}

	originData = string(originDataByte)
	return
}
