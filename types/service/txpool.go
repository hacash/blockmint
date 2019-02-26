package service

import "github.com/hacash/blockmint/types/block"

type TxPool interface {
	// 检查交易是否已经存在
	CheckTxExist(block.Transaction) bool
	// 添加交易
	AddTx(block.Transaction) error
	// RemoveTxs([]block.Transaction) error
	// 获取手续费最高的一笔交易
	PopTxByHighestFee() block.Transaction
}
