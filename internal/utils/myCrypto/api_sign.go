package myCrypto

import (
	"crypto"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
)

const (
	KeyTypeHmac    = "HMAC"
	KeyTypeRsa     = "RSA"
	KeyTypeEd25519 = "ED25519"
)

func SignFunc(keyType string) (func([]byte, []byte) (string, error), error) {
	switch keyType {
	case KeyTypeHmac:
		return HmacSha256, nil
	case KeyTypeRsa:
		return Rsa, nil
	case KeyTypeEd25519:
		return Ed25519, nil
	default:
		return nil, fmt.Errorf("unsupported keyType=%s", keyType)
	}
}

func HmacSha256(secretKey, data []byte) (string, error) {
	mac := hmac.New(sha256.New, secretKey)
	_, err := mac.Write(data)
	if err != nil {
		return "", SignHmacSha256Err.WithCause(err).WithMetadata(map[string]string{
			signData:   string(data),
			signSecret: string(secretKey),
		})
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func HmacSha256Fast(secretKey, data []byte, dst *[]byte) error {
	mac := hmac.New(sha256.New, secretKey)
	if _, err := mac.Write(data); err != nil {
		return SignHmacSha256Err.WithCause(err).WithMetadata(map[string]string{
			signData:   string(data),
			signSecret: string(secretKey),
		})
	}
	*dst = (*dst)[:64]             // 重置长度
	hex.Encode(*dst, mac.Sum(nil)) // 写入 hex 编码结果(64字节)
	return nil
}

func Sha256(secretKey string, data string) (*string, error) {
	mac := hmac.New(sha256.New, []byte(secretKey))
	_, err := mac.Write([]byte(data))
	if err != nil {
		return nil, err
	}
	encodeData := hex.EncodeToString(mac.Sum(nil))
	return &encodeData, nil
}

func Rsa(secretKey, data []byte) (string, error) {
	block, _ := pem.Decode(secretKey)
	if block == nil {
		return "", errors.New("rsa pem.Decode failed, invalid pem format secretKey")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("rsa ParsePKCS8PrivateKey failed, error=%v", err.Error())
	}
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("rsa convert PrivateKey failed")
	}
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	return encodedSignature, nil
}

func Ed25519(secretKey, data []byte) (string, error) {
	block, _ := pem.Decode(secretKey)
	if block == nil {
		return "", fmt.Errorf("ed25519 pem.Decode failed, invalid pem format secretKey")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("ed25519 call ParsePKCS8PrivateKey failed, error=%v", err.Error())
	}
	ed25519PrivateKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return "", fmt.Errorf("ed25519 convert PrivateKey failed")
	}
	pk := ed25519.PrivateKey(ed25519PrivateKey)
	signature := ed25519.Sign(pk, data)
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	return encodedSignature, nil
}
