package txpool

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/chain/state"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/sys/log"
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
	Length int    // 交易数量
	Size   uint64 // 交易总体积大小

	mulk sync.Mutex // 互斥锁

	// 事件订阅
	txFeed      event.Feed
	txFeedScope event.SubscriptionScope

	Log log.Logger
}

var (
	GlobalInstanceMemTxPoolMutex sync.Mutex
	GlobalInstanceMemTxPool      *MemTxPool = nil
)

func GetGlobalInstanceMemTxPool() service.TxPool {
	GlobalInstanceMemTxPoolMutex.Lock()
	defer GlobalInstanceMemTxPoolMutex.Unlock()
	if GlobalInstanceMemTxPool == nil {
		lg := config.GetGlobalInstanceLogger()
		GlobalInstanceMemTxPool = &MemTxPool{
			Length: 0,
			Size:   0,
			Log:    lg,
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

// 从交易池里查询一笔交易
func (this *MemTxPool) FindTxByHash(txhash []byte) (block.Transaction, bool) {
	hashave := this.pickUpTx(txhash, false)
	if hashave != nil && hashave.Tx != nil {
		return hashave.Tx, true
	} else {
		return nil, false
	}
}

// 检查交易是否已经存在
func (this *MemTxPool) CheckTxExist(block.Transaction) bool {
	panic("func not ok !")
	return false
}

// 取出一笔交易
func (this *MemTxPool) pickUpTx(hashnofee []byte, drop bool) *MemTxItem {
	if this.TxHead == nil {
		return nil
	}
	prev := this.TxHead
	next := this.TxHead
	for {
		if next == nil {
			break
		}
		if bytes.Compare(next.HashNoFee, hashnofee) == 0 {
			if drop {
				prev.next = next.next
				// 更新统计
				this.Length -= 1
				this.Size -= uint64(next.Tx.Size())
			}
			// 返回
			return next
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
	// 检查签名
	ok, e1 := tx.VerifyNeedSigns()
	if !ok || e1 != nil {
		return fmt.Errorf("Transaction Verify Signature error")
	}
	// 检查最低手续费
	if tx.FeePurity() < uint64(config.MinimalFeePurity) {
		return fmt.Errorf("The handling fee is too low for miners to accept.")
	}
	hashnofee := tx.HashNoFee()
	// 检查是否已经存在
	stoblk := store.GetGlobalInstanceBlocksDataStore()
	ext, e2 := stoblk.CheckTransactionExist(hashnofee)
	if e2 != nil {
		return fmt.Errorf("Transaction CheckTransactionExist error")
	}
	if ext {
		hashnofeestr := hex.EncodeToString(hashnofee)
		return fmt.Errorf("Transaction " + hashnofeestr + " already exist")
	}
	// 尝试执行，检查余额
	state_base := state.GetGlobalInstanceChainState()
	state_temp_tx := state.NewTempChainState(state_base)
	runerr := tx.ChangeChainState(state_temp_tx)
	state_temp_tx.Destroy() // 销毁临时执行栈
	if runerr != nil {
		return fmt.Errorf(runerr.Error()) // 执行失败
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
	// pick up
	hashave := this.pickUpTx(hashnofee, false)
	if hashave != nil {
		if txItem.FeePer <= hashave.FeePer { // 手续费不能比原有的低
			return fmt.Errorf("Tx FeePurity value equal or less than the exist")
		}
		this.pickUpTx(hashnofee, true) // 删除旧的交易
	}
	// size change
	this.Size += uint64(txsize)
	this.Length += 1
	// append
	if this.TxHead == nil {
		this.TxHead = txItem
		// 插入链表 广播添加事件
		this.feedTx(hashnofee, tx) // 广播
		return nil
	}
	txSeekPtr := this.TxHead
	for {
		if txSeekPtr.next == nil {
			txSeekPtr.next = txItem
			this.feedTx(hashnofee, tx) // 广播
			break
		}
		if txSeekPtr.next.FeePer < txItem.FeePer {
			txItem.next = txSeekPtr.next
			txSeekPtr.next = txItem
			// 插入链表 广播添加事件
			this.feedTx(hashnofee, tx) // 广播
			break
		}
		// 下一个
		txSeekPtr = txSeekPtr.next
	}

	return nil
}

// 广播交易
func (this *MemTxPool) feedTx(hashnofee []byte, tx block.Transaction) {
	this.Log.Info("mempool add tx", hex.EncodeToString(hashnofee), "send to subscribe feed")
	go this.txFeed.Send([]block.Transaction{tx})
}

// 弹出手续费最高的一笔交易
func (this *MemTxPool) PopTxByHighestFee() block.Transaction {
	this.Lock()
	defer this.Unlock()

	if this.TxHead == nil {
		return nil
	}
	head := this.TxHead
	this.TxHead = head.next
	// 更新统计
	this.Length -= 1
	this.Size -= uint64(head.Tx.Size())
	// 返回
	return head.Tx
}

// 订阅交易池加入新交易事件
func (this *MemTxPool) SubscribeNewTx(txCh chan<- []block.Transaction) event.Subscription {
	return this.txFeedScope.Track(this.txFeed.Subscribe(txCh))
}

func (this *MemTxPool) GetTxs() []block.Transaction {
	this.Lock()
	defer this.Unlock()
	var results = make([]block.Transaction, 0, this.Length)
	next := this.TxHead
	for {
		if next == nil {
			break
		}
		results = append(results, next.Tx)
		next = next.next
	}
	return results
}

// 过滤、清除交易
func (this *MemTxPool) RemoveTxs(txs []block.Transaction) {
	this.Lock()
	defer this.Unlock()

	for _, tx := range txs {
		this.pickUpTx(tx.HashNoFee(), true) // 取出并丢弃
	}

}
