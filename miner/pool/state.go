package pool

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/core/account"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner/difficulty"
	"github.com/hacash/x16rs"
	"math/big"
	"sync"
)

// 状态统计
type PoolState struct {
	pool *MiningPool

	Markword   []byte // 区块播报方 len:8
	FeeAccount *account.Account

	MaxClientCount uint64 // 连接地址数量上限
	ClientCount    uint64 // 实时统计的连接地址数量

	AllClients    sync.Map // []*Client // 	实时在线的全部客户端连接
	AllPowWorkers sync.Map // []*PowWorker // 实时在线的全部矿工

	CurrentMiningBlock         *NewCreateBlockEvent // 当前正在挖掘区块高度
	CurrentMiningBlockHashLoop int                  // 次数

	SuccessFindBlockHash          []byte          // 上一个成功找到的区块哈希
	SuccessFindBlockRewardAddress *fields.Address // 上一个成功找到的地址

	AutoincrementMiningCoinbaseStuffNum uint64 // 周期内自增，换周期清零的挖矿序号

	AutoTransferHandleQueue chan *PowWorker // 转账处理队列

}

func NewPoolState(feepassword, markword string, maxconn uint64) *PoolState {

	blkmsg := []byte("        ") // len:8
	copy(blkmsg, markword)       // []byte("pool.HCX")

	return &PoolState{
		FeeAccount:                          account.CreateAccountByPassword(feepassword),
		Markword:                            blkmsg,
		MaxClientCount:                      maxconn,
		ClientCount:                         0,
		AllClients:                          sync.Map{},
		AllPowWorkers:                       sync.Map{},
		CurrentMiningBlock:                  nil,
		SuccessFindBlockHash:                nil,
		AutoincrementMiningCoinbaseStuffNum: 0,
		AutoTransferHandleQueue:             make(chan *PowWorker, 100),
	}
}

func (ps *PoolState) setPool(pool *MiningPool) {
	ps.pool = pool
}

// 自动打币
func (ps *PoolState) transferRewards() {
	if ps.CurrentMiningBlock == nil {
		return
	}
	// 挑选出符合打币要求的地址
	trxWorkers := make([]*PowWorker, 0, 20)
	ps.AllPowWorkers.Range(func(key interface{}, val interface{}) bool {
		wk := val.(*PowWorker)
		rwds := wk.StatisticsData.DeservedRewards
		pthx := ps.CurrentMiningBlock.Block.GetHeight() - uint64(wk.StatisticsData.PrevTransferBlockHeight)
		if (rwds >= 2*10000*10000 && pthx > 10) || (rwds > 2*1000*10000 && pthx > 100) || (rwds > 2*100*10000 && pthx > 1000) {
			trxWorkers = append(trxWorkers, wk) // 2币10个块50分钟打， 0.2币100个块8.3小时打， 0.02币1000区块3.5天打
		}
		return true
	})
	// 开始打币
	go func() {
		for i := 0; i < len(trxWorkers); i++ {
			wk := trxWorkers[i]
			ps.AutoTransferHandleQueue <- wk // 加入打币队列
			//fmt.Printf("pool transfer address: %s, coin: ㄜ%.8f:240 \n", wk.RewordAddress.ToReadable(), float64(wk.StatisticsData.DeservedRewards)/100000000 )
		}
		//fmt.Println("---------")
	}()
}

// 新区块到达
func (ps *PoolState) createBlockArrive(blk *NewCreateBlockEvent) {
	// 检查并结算奖励
	if ps.SuccessFindBlockHash != nil && bytes.Compare(ps.SuccessFindBlockHash, blk.Block.GetPrevHash()) == 0 {
		coinnum := coin.BlockCoinBaseRewardNumber(blk.Block.GetHeight())
		ps.settlementAllWorkerRewards(uint32(coinnum)) // 结算奖励
	}
	// 记录上一区块的挖矿信息
	ps.CurrentMiningBlock = blk
	// 向所有客户端 发送挖掘数据，统计总算力
	ps.AllClients.Range(func(key interface{}, val interface{}) bool {
		client := val.(*Client)
		ps.sendMiningStuffData(client)
		return true
	})
}

// 结算奖励
func (ps *PoolState) settlementAllWorkerRewards(coinnum uint32) {
	coinRewards := uint64(coinnum) * 100000000
	poolfee := uint64(float64(coinRewards) * (config.Config.MiningPool.PayFeeRatio))
	totalRewards := coinRewards - poolfee // 应发奖励
	divRewards := totalRewards / 2        // 一半
	//fmt.Printf("divRewards: %d \n", divRewards)
	// 计算总算力值，除去挖出的地址
	var othertotalpower = big.NewInt(0)
	var otherworkers = make([]*PowWorker, 0, 40)
	var successworker *PowWorker = nil
	ps.AllPowWorkers.Range(func(key interface{}, val interface{}) bool {
		wk := val.(*PowWorker)
		if successworker == nil && bytes.Compare(*ps.SuccessFindBlockRewardAddress, *wk.RewordAddress) == 0 {
			// 出块者奖励
			wk.StatisticsData.FindBlocks += 1
			wk.StatisticsData.FindCoins += coinnum
			wk.StatisticsData.DeservedRewards += divRewards // 一半发给出块者
			if wk.StatisticsData.PrevTransferBlockHeight == 0 && ps.CurrentMiningBlock != nil {
				wk.StatisticsData.PrevTransferBlockHeight = uint32(ps.CurrentMiningBlock.Block.GetHeight()) // 打币时间记录
			}
			successworker = wk
			// 除去挖出的地址
		} else {
			// 剩余总算力统计
			othertotalpower = othertotalpower.Add(othertotalpower, wk.RealtimePower)
			otherworkers = append(otherworkers, wk)
		}
		return true
	})
	// 循环结算奖励
	length := len(otherworkers)
	//fmt.Printf("othertotalpower %d, otherworkers len %d\n", othertotalpower, length)
	if othertotalpower.Cmp(big.NewInt(0)) == 1 { // 避免除零错误
		for i := 0; i < length; i++ {
			cwk := otherworkers[i]
			rationum := new(big.Float).Quo(new(big.Float).SetInt(cwk.RealtimePower), new(big.Float).SetInt(othertotalpower))
			basemu, _ := rationum.Float64()
			rwdz := uint64(basemu * float64(divRewards))
			//fmt.Printf("%s  rwdz: %d \n", cwk.RewordAddress.ToReadable(), rwdz)
			cwk.StatisticsData.DeservedRewards += rwdz // 结算奖励
			// 清除数据统计，等待第二轮统计
			cwk.RealtimePower = big.NewInt(0)
			cwk.RealtimeWorkSubmitCount = 0
		}
	}
	// 保存 worker 进磁盘
	go func() {
		ps.pool.StoreDB.SaveWorker(successworker)
		for i := 0; i < length; i++ {
			ps.pool.StoreDB.SaveWorker(otherworkers[i])
		}
	}()
}

// 发送挖掘数据
func (ps *PoolState) sendMiningStuffData(client *Client) {
	if ps.CurrentMiningBlock != nil {
		// 记录上一区块算力信息
		client.MiningBlockStuffPrev = client.MiningBlockStuffCurrent
		// 发送新区块挖矿信息
		ps.AutoincrementMiningCoinbaseStuffNum += 1
		mlstuff := ps.updateSetCurrentMiningBlockCoinbaseMessage(ps.AutoincrementMiningCoinbaseStuffNum, nil)
		// 记录本区块
		client.MiningBlockStuffCurrent = mlstuff
		// 发送消息
		hei := ps.CurrentMiningBlock.Block.GetHeight()
		loopnum := hei/50000 + 1
		if loopnum > 16 {
			loopnum = 16
		}
		stuffmining := &x16rs.MiningPoolStuff{
			BlockHeight:   hei,
			MiningIndex:   ps.AutoincrementMiningCoinbaseStuffNum,
			Loopnum:       uint8(loopnum),
			TargetHash:    difficulty.Uint32ToHash(hei, ps.CurrentMiningBlock.Block.GetDifficulty()),
			BlockHeadMeta: mlstuff,
		}
		// 信息发送给客户端
		x16rs.MiningPoolWriteTcpMsgBytes(client.Conn, 1, stuffmining.Serialize())
	}
}

func (ps *PoolState) updateSetCurrentMiningBlockCoinbaseMessage(miningNum uint64, nonce []byte) []byte {
	// 重新计算区块 hash
	ps.fillCoinbaseMsg(ps.CurrentMiningBlock.Coinbase, miningNum)
	ps.CurrentMiningBlock.Block.SetMrklRoot(blocks.CalculateMrklRoot(ps.CurrentMiningBlock.Block.GetTransactions()))
	if nonce == nil {
		nonce = []byte{0, 0, 0, 0}
	}
	ps.CurrentMiningBlock.Block.SetNonce(binary.BigEndian.Uint32(nonce))
	mlstuff := blocks.CalculateBlockHashBaseStuff(ps.CurrentMiningBlock.Block)
	return mlstuff
}

//////////////////////////////////

func (ps *PoolState) fillCoinbaseMsg(coin *transactions.Transaction_0_Coinbase, miningNum uint64) {
	s1 := make([]byte, 8)
	binary.BigEndian.PutUint64(s1, miningNum)
	blkmsg := []byte("        ") // len=8
	copy(blkmsg, ps.Markword)    // []byte("pool.HCX")
	blkmsg = append(blkmsg, s1...)
	coin.Message = fields.TrimString16(blkmsg)
}

// 取出和保存 PowWorker
func (ps *PoolState) getPowWorker(addr *fields.Address) *PowWorker {
	wk, has := ps.AllPowWorkers.Load(string(*addr))
	if has {
		return wk.(*PowWorker)
	} else {
		return nil
	}
}
func (ps *PoolState) putPowWorker(addr *fields.Address, worker *PowWorker) {
	ps.AllPowWorkers.Store(string(*addr), worker)
}
