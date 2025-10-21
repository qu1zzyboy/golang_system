package myCrypto

import (
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
)

const (
	signData   = "sign_data"
	signSecret = "secret_key"
)

var (
	SignHmacSha256Err = errorx.New(errCode.SIGN_HMAC_SHA256_ERROR, "HMAC SHA256签名错误")
)
