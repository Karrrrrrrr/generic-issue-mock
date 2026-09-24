package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	openpgp "github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

// PGPPrivateKey PGP 私钥封装
type PGPPrivateKey = openpgp.EntityList

// PGPPublicKey PGP 公钥封装
type PGPPublicKey = openpgp.EntityList

type Pgp struct {
	ClientPublicKey      PGPPublicKey  `json:"client_public_key"`
	ServerPrivateKey     PGPPrivateKey `json:"server_private_key"`
	PrivateKeyPassphrase string        `json:"private_key_passphrase"`
}

type ValidateUpSignatureReq struct {
	Timestamp       string // timestamp
	SignatureHeader string // header中的Signature
	Payload         string // body
	SigningSecret   string // 密钥
}

func ValidateUpWebhookSignature(req *ValidateUpSignatureReq) error {
	valueToDigest := req.Payload + req.Timestamp
	mac := hmac.New(sha512.New, []byte(req.SigningSecret))
	mac.Write([]byte(valueToDigest))
	expect := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expect), []byte(req.SignatureHeader)) {
		return errors.New("signature mismatch")
	}
	return nil
}

func NewPGPConfig(clientPublicKey, serverPrivateKey, serverPrivateKeyPassphrase string) (*Pgp, error) {
	if clientPublicKey == "" || serverPrivateKey == "" {
		return nil, errors.New("uqpay pgp config error")
	}
	clientPub, err := loadPGPPublicKeyFromString(clientPublicKey)
	if err != nil {
		return nil, errors.New("uqpay clientPub error " + err.Error())
	}

	serverPriv, err := loadPGPPrivateKeyFromString(
		serverPrivateKey,
		serverPrivateKeyPassphrase,
	)
	if err != nil {
		return nil, errors.New("uqpay serverPriv error " + err.Error())
	}

	return &Pgp{
		ClientPublicKey:  clientPub,
		ServerPrivateKey: serverPriv,
	}, nil
}

// LoadPGPPublicKeyFromString 从字符串加载公钥
func loadPGPPublicKeyFromString(key string) (PGPPublicKey, error) {
	return openpgp.ReadArmoredKeyRing(strings.NewReader(key))
}

// LoadPGPPrivateKeyFromString 从字符串加载私钥
func loadPGPPrivateKeyFromString(key string, passphrase string) (PGPPrivateKey, error) {
	el, err := openpgp.ReadArmoredKeyRing(strings.NewReader(key))
	if err != nil {
		return nil, err
	}

	for _, e := range el {
		if e.PrivateKey != nil && e.PrivateKey.Encrypted {
			if err := e.PrivateKey.Decrypt([]byte(passphrase)); err != nil {
				return nil, err
			}
		}
		for _, sub := range e.Subkeys {
			if sub.PrivateKey != nil && sub.PrivateKey.Encrypted {
				if err := sub.PrivateKey.Decrypt([]byte(passphrase)); err != nil {
					return nil, err
				}
			}
		}
	}

	return el, nil
}

func PGPDecrypt(cipher []byte, priv PGPPrivateKey) ([]byte, error) {
	block, err := armor.Decode(bytes.NewReader(cipher))
	if err != nil {
		return nil, fmt.Errorf("pgp armor decode failed: %w", err)
	}

	md, err := openpgp.ReadMessage(
		block.Body,
		priv,
		nil,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("pgp read message failed: %w", err)
	}

	plaintext, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// PGPEncryptResponse 将响应对象序列化并用公钥加密
func PGPEncryptResponse(v any, pub PGPPublicKey) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return PGPEncrypt(raw, pub)
}

func PGPEncrypt(plain []byte, pub PGPPublicKey) ([]byte, error) {
	var buf bytes.Buffer

	armorWriter, err := armor.Encode(&buf, "PGP MESSAGE", nil)
	if err != nil {
		return nil, err
	}

	w, err := openpgp.Encrypt(armorWriter, pub, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(plain); err != nil {
		return nil, err
	}

	w.Close()
	armorWriter.Close()

	return buf.Bytes(), nil
}
