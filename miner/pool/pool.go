package pool

import (
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/config"
	sys_log "github.com/hacash/blockmint/sys/log"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/x16rs"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type NewCreateBlockEvent struct {
	Block    block.Block
	Coinbase *transactions.Transaction_0_Coinbase
}

type MiningPool struct {

	// 状态
	StateData *PoolState

	// 储存
	StoreDB *Store

	// 数据通信
	NewCreateBlockCh    chan *NewCreateBlockEvent // 新挖掘的区块到来
	DiscoveryNewBlockCh chan block.Block          // 发现新的区块
	CalcSuccessBlockCh  chan x16rs.MiningSuccess
}

func NewMiningPool(logger sys_log.Logger) *MiningPool {

	fnp := config.Config.MiningPool.StatisticsDir
	if strings.Compare(fnp, "") == 0 {
		panic("config.Config.MiningPool.StatisticsDir must be set.")
	}

	sto := NewStore(fnp)
	sta := NewPoolState(config.Config.MiningPool.PayPassword, config.Config.MiningPool.Markword, config.Config.MiningPool.ClientMax)

	pool := &MiningPool{
		StoreDB:   sto,
		StateData: sta,

		NewCreateBlockCh:    make(chan *NewCreateBlockEvent, 8),
		DiscoveryNewBlockCh: make(chan block.Block, 8),
		CalcSuccessBlockCh:  make(chan x16rs.MiningSuccess, 8),
	}
	sta.setPool(pool)

	return pool

}

// 启动矿池
func (mp *MiningPool) Start() error {
	// 端口监听
	go mp.listenAndServeLoop()
	// 新区快到来监听
	go mp.newCreateBlockArriveLoop()
	// 定时清理tcp
	go mp.removeDeadConnLoop()
	// 定时发起打币
	go mp.transferRewardLoop()
	// 处理自动打币队列
	go mp.transfersLoop()
	// 处理自动打币队列
	go mp.createSendTransactionLoop()
	// 检查交易确认
	go mp.checkTransactionConfirmLoop()

	return nil
}

// 检查交易确认
func (mp *MiningPool) checkTransactionConfirmLoop() {
	for {
		time.Sleep(time.Second * 33)
		err := mp.checkTransactionConfirm()
		if err != nil {
			fmt.Println(err.Error()) // 创建错误
		}
	}
}

// 创建并发送交易
func (mp *MiningPool) createSendTransactionLoop() {
	for {
		time.Sleep(time.Second * 17)
		err := mp.createSendTransaction() // 创建交易并提交
		if err != nil {
			fmt.Println(err.Error()) // 创建错误
		}
	}
}

// 定时自动打币
func (mp *MiningPool) transfersLoop() {
	for {
		twk := <-mp.StateData.AutoTransferHandleQueue
		mp.transfer(twk) // 转账打币
	}
}

// 定时自动打币
func (mp *MiningPool) transferRewardLoop() {
	for {
		time.Sleep(time.Second * 37)   //(60*4+37))
		mp.StateData.transferRewards() // 自动打币
	}
}

// 定时清理过期未活跃的 conn
// 定时移除非活跃地址
func (mp *MiningPool) removeDeadConnLoop() {
	for {
		time.Sleep(time.Minute * 5)
		nnn := time.Now()
		mp.StateData.AllClients.Range(func(key interface{}, val interface{}) bool {
			client := val.(*Client)
			if nnn.Unix()-client.ActiveTimestamp.Unix() > 7*60 {
				mp.removeCloseClient(client)
			}
			return true
		})
	}
}

// 移除连接
func (mp *MiningPool) removeCloseClient(client *Client) {
	// fmt.Printf("removeCloseClient: %d\n", client.Id)
	mp.StateData.AllClients.Delete(client.Id) // 移除
	client.Conn.Close()                       // 关闭
	// 减少 worker 统计
	addr := client.RewordAddress
	if addr == nil || client.Worker == nil {
		return // 未注册
	}
	wk := client.Worker
	if wk != nil {
		mp.StateData.ClientCount -= 1 // 统计减一
		wk.ClientCount -= 1           // 统计减一
		if wk.ClientCount == 0 {
			// 保存 wk
			mp.StoreDB.SaveWorker(wk)
			// 从内存中去掉 worker
			mp.StateData.AllPowWorkers.Delete(string(*addr))
		}
	}
	// 复位
	client.RewordAddress = nil
	client.Worker = nil
}

// 启动端口监听
func (mp *MiningPool) listenAndServeLoop() {
	portstr := strconv.FormatUint(config.Config.MiningPool.Port, 10)
	mp.StateData.startListenAndServe(":" + portstr)
}

// 新区快到达
func (mp *MiningPool) newCreateBlockArriveLoop() {
	for {
		blk := <-mp.NewCreateBlockCh
		mp.StateData.createBlockArrive(blk)
	}
}

// 取出一个PowWorker
func (mp *MiningPool) getThePowWorker(addr *fields.Address) *PowWorker {
	// 检查内存
	wk := mp.StateData.getPowWorker(addr)
	if wk != nil {
		return wk
	}
	// 从磁盘加载
	wkd := mp.StoreDB.ReadWorker(addr)
	if wkd != nil {
		// 放入内存
		mp.StateData.putPowWorker(addr, wkd)
		return wkd
	}
	// 创建新 worker
	data := &AddressStatisticsStoreItem{
		0, 0, 0, 0, 0,
	}
	if mp.StateData.CurrentMiningBlock != nil {
		data.PrevTransferBlockHeight = uint32(mp.StateData.CurrentMiningBlock.Block.GetHeight())
	}
	wkn := &PowWorker{
		RewordAddress:           addr,
		RealtimePower:           big.NewInt(0),
		RealtimeWorkSubmitCount: 0,
		ClientCount:             0,
		StatisticsData:          data,
	}
	// 放入内存
	mp.StateData.putPowWorker(addr, wkn)
	return wkn

}
