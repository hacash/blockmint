package rpc

import (
	"encoding/hex"
	"github.com/hacash/blockmint/core/account"
)

func newAccountByPassword(params map[string]string) map[string]string {

	result := make(map[string]string)
	passstr, ok1 := params["password"]
	if !ok1 {
		result["err"] = "password must"
		return result
	}
	// 创建账户
	acc := account.CreateAccountByPassword(passstr)

	result["address"] = string(acc.AddressReadable)
	result["private_key"] = hex.EncodeToString(acc.PrivateKey)

	return result
}
