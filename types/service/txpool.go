package service

import "github.com/hacash/blockmint/types/block"

type TxPool interface {
	AddTx(block.Transaction) error
	RemoveTxs([]block.Transaction) error
	// 获取手续费最高的一笔交易
	GetTxByHighestFee() (block.Transaction, error)
}
