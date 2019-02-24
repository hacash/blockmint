package txpool

import (
	"fmt"
	"github.com/hacash/blockmint/types/block"
)

var (
	MemTxPoolMaxLimit = 500                          // 交易池的大小
	MemTxPoolMaxSize  = uint64(1024 * 1024 * 2 * 10) // 交易池的大小
)

type MemTxItem struct {
	FeePer    uint64 // 手续费比值， 用于排序 fee/txsize
	Size      uint32 //
	HashNoFee []byte // 哈希
	Tx        block.Transaction
	// 单链表
	next *MemTxItem
}

type MemTxPool struct {
	TxHead *MemTxItem
	Length int
	Size   uint64
}

func (this *MemTxPool) AddTx(tx block.Transaction) error {

	if this.Length > MemTxPoolMaxLimit {
		return fmt.Errorf("Mem Tx Pool Over Max Limit %d", MemTxPoolMaxLimit)
	}
	// 检查
	txsize := tx.Size()
	if uint64(txsize)+this.Size > MemTxPoolMaxSize {
		return fmt.Errorf("Mem Tx Pool Over Max Size %d", MemTxPoolMaxSize)
	}
	txfeepur := tx.FeePurity()
	txItem := &MemTxItem{
		FeePer:    txfeepur,
		Size:      txsize,
		HashNoFee: tx.HashNoFee(),
		Tx:        tx,
		next:      nil,
	}
	// append
	this.Size += uint64(txsize)
	this.Length += 1
	if this.TxHead == nil {
		this.TxHead = txItem
		return nil
	}
	txSeekPtr := this.TxHead
	for true {
		if txSeekPtr.next == nil {
			txSeekPtr.next = txItem
			break
		}
		if txSeekPtr.next.FeePer < txItem.FeePer {
			txItem.next = txSeekPtr.next
			txSeekPtr.next = txItem
			break // 插入链表
		}
		// 下一个
		txSeekPtr = txSeekPtr.next
	}

	return nil
}

/*
func (this *MemTxPool) RemoveTx(hashNoFee []byte) error {

	return nil
}
*/

// 弹出手续费最高的一笔交易
func (this *MemTxPool) PopTxByHighestFee() block.Transaction {
	if this.TxHead == nil {
		return nil
	}
	ret := this.TxHead
	this.TxHead = this.TxHead.next
	return ret.Tx
}
