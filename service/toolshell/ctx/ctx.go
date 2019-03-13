package ctx

import "github.com/hacash/blockmint/types/block"

type Context interface {
	NotLoadedYetAccountAddress(string) bool // 检测账户是否已经登录
	IsInvalidAccountAddress(string) bool // 检测是否为一个合法的账户名
	GetAllPrivateKeyBytes() map[string][]byte // 获取全部私钥，用于填充签名
	SetTxToRecord([]byte, block.Transaction) // 记录交易
	GetTxFromRecord([]byte) block.Transaction // 获取交易
	UseTimestamp() uint64 // 当前使用的时间戳
}


