package service

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/hacash/blockmint/types/block"
)

type TxPool interface {
	// 检查交易是否已经存在
	CheckTxExist(block.Transaction) bool
	// 添加交易
	AddTx(block.Transaction) error
	// 从交易池里查询一笔交易
	FindTxByHash([]byte) (block.Transaction, bool)
	// 获取全部交易
	GetTxs() []block.Transaction
	// 过滤、清除交易
	RemoveTxs([]block.Transaction)
	// 获取手续费最高的一笔交易
	PopTxByHighestFee() block.Transaction
	// 订阅交易池加入新交易事件
	SubscribeNewTx(chan<- []block.Transaction) event.Subscription
}
