package crypto

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// Signature 使用MD5withRSA对数据进行签名并Base64编码
func PhotonpaySignature(data []byte, privateKeyPEM string) (string, error) {
	// 1. 解析PEM格式的私钥
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block containing private key")
	}

	// 2. 解析PKCS8私钥
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// 3. 类型断言为RSA私钥
	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA type")
	}

	// 创建⼀个MD5哈希
	hash := md5.New()
	// 将数据写⼊哈希。
	_, err = hash.Write(data)
	if err != nil {
		return "", fmt.Errorf("error writing data to hash: %v", err)
	}

	// 5. 使用RSA对MD5哈希进行签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPriv, crypto.MD5, hash.Sum(nil))
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	// 6. 对签名结果进行Base64编码
	encodedSignature := base64.StdEncoding.EncodeToString(signature)

	return encodedSignature, nil
}

func VerifyPhotonpayWebhookSignature(data string, signatureBytes []byte, privateKeyPEM string) error {

	var publicKeyBytes []byte
	bloc, _ := pem.Decode([]byte(privateKeyPEM))
	if bloc == nil {
		// 如果不是PEM格式，直接base64解码
		decoded, err := base64.StdEncoding.DecodeString(privateKeyPEM)
		if err != nil {
			return errors.New("解析公钥失败: 无效的公钥格式")
		}
		publicKeyBytes = decoded
	} else {
		publicKeyBytes = bloc.Bytes
	}
	decodedSignature, err := base64.StdEncoding.DecodeString(string(signatureBytes))
	if err != nil {
		fmt.Println("Error decoding signature:", err)
		return err
	}
	// 解析PKCS#8格式私钥
	pub, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("解析PKCS8公钥失败: %v", err)
	}
	// 类型断言为RSA公钥
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("公钥不是RSA类型")
	}

	md5Hash := md5.Sum([]byte(data))

	// 测试验证签名
	err = rsa.VerifyPKCS1v15(rsaPub, crypto.MD5, md5Hash[:], decodedSignature)
	if err != nil {
		return fmt.Errorf("签名验证失败: %v", err)
	}
	return nil
}
