package txpool

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/service"
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

	// 事件订阅
	txFeed      event.Feed
	txFeedScope event.SubscriptionScope
}

var (
	GlobalInstanceMemTxPoolMutex sync.Mutex
	GlobalInstanceMemTxPool      *MemTxPool = nil
)

func GetGlobalInstanceMemTxPool() service.TxPool {
	GlobalInstanceMemTxPoolMutex.Lock()
	defer GlobalInstanceMemTxPoolMutex.Unlock()
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

func (this *MemTxPool) checkTx(tx block.Transaction) error {
	if this.Length > MemTxPoolMaxLimit {
		return fmt.Errorf("Mem Tx Pool Over Max Limit %d", MemTxPoolMaxLimit)
	}

	ok, e1 := tx.VerifyNeedSigns()
	if !ok || e1 != nil {
		return fmt.Errorf("Transaction Verify Signature error")
	}

	hashnofee := tx.HashNoFee()

	stoblk := store.GetGlobalInstanceBlocksDataStore()
	ext, e2 := stoblk.CheckTransactionExist(hashnofee)
	if e2 != nil {
		return fmt.Errorf("Transaction CheckTransactionExist error")
	}
	if ext {
		hashnofeestr := hex.EncodeToString(hashnofee)
		return fmt.Errorf("Transaction " + hashnofeestr + " already exist")
	}
	// 最低手续费
	if tx.FeePurity() < uint64(config.MinimalFeePurity) {
		return fmt.Errorf("The handling fee is too low for miners to accept.")
	}

	// pass check !
	return nil
}

// entrance 是否为
func (this *MemTxPool) AddTx(tx block.Transaction) error {

	if e := this.checkTx(tx); e != nil {
		return e
	}

	this.Lock()
	defer this.Unlock()

	// 检查
	txsize := tx.Size()
	if uint64(txsize)+this.Size > MemTxPoolMaxSize {
		return fmt.Errorf("Mem Tx Pool Over Max Size %d", MemTxPoolMaxSize)
	}

	hashnofee := tx.HashNoFee()

	txfeepur := tx.FeePurity()
	txItem := &MemTxItem{
		FeePer:    txfeepur,
		Size:      txsize,
		HashNoFee: hashnofee,
		Tx:        tx,
		next:      nil,
	}
	hashave := this.pickUpTrs(hashnofee)
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
		// 插入链表 广播添加事件
		go this.txFeed.Send([]block.Transaction{tx})
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
			// 插入链表 广播添加事件
			go this.txFeed.Send([]block.Transaction{tx})
			break
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

// 订阅交易池加入新交易事件
func (this *MemTxPool) SubscribeNewTx(txCh chan<- []block.Transaction) event.Subscription {
	return this.txFeedScope.Track(this.txFeed.Subscribe(txCh))
}
