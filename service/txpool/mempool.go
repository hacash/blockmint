package txpool

import (
	"bytes"
	"fmt"
	"github.com/hacash/blockmint/types/block"
	"sync"
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

	mulk sync.Mutex // 互斥锁
}

var GlobalInstanceMemTxPool *MemTxPool = nil

func GetGlobalInstanceMemTxPool() *MemTxPool {
	if GlobalInstanceMemTxPool == nil {
		GlobalInstanceMemTxPool = &MemTxPool{
			Length: 0,
			Size:   0,
		}
	}
	return GlobalInstanceMemTxPool
}

func (this *MemTxPool) Lock() {
	//fmt.Println("MemTxPool Lock +++++++++++")
	this.mulk.Lock()
}
func (this *MemTxPool) Unlock() {
	//fmt.Println("MemTxPool Unlock ---------")
	this.mulk.Unlock()
}

// 检查交易是否已经存在
func (this *MemTxPool) CheckTxExist(block.Transaction) bool {
	panic("func not ok !")
	return false
}

func (this *MemTxPool) pickUpTrs(hashnofee []byte) *MemTxItem {
	if this.TxHead == nil {
		return nil
	}
	prev := this.TxHead
	next := this.TxHead
	for true {
		if next == nil {
			break
		}
		if bytes.Compare(next.HashNoFee, hashnofee) == 0 {
			prev.next = next.next
			return next // 返回
		}
		prev = next
		next = next.next
	}
	return nil
}

func (this *MemTxPool) AddTx(tx block.Transaction) error {
	this.Lock()
	defer this.Unlock()

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
	hashave := this.pickUpTrs(tx.HashNoFee())
	if hashave != nil {
		if txItem.FeePer <= hashave.FeePer { // 手续费不能比原有的低
			return fmt.Errorf("Tx FeePurity value equal or less than the exist")
		}
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
	this.Lock()
	defer this.Unlock()

	if this.TxHead == nil {
		return nil
	}
	ret := this.TxHead
	this.TxHead = this.TxHead.next
	return ret.Tx
}
