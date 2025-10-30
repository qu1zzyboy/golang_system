package myCrypto

import (
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
)

const (
	signData   = "sign_data"
	signSecret = "secret_key"
)

var (
	SignHmacSha256Err = errorx.New(errCode.SIGN_HMAC_SHA256_ERROR, "HMAC SHA256签名错误")
)
