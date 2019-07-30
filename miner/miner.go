package miner

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/store"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/chain/state"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner/difficulty"
	"github.com/hacash/blockmint/miner/pool"
	"github.com/hacash/blockmint/protocol/block1def"
	"github.com/hacash/blockmint/service/txpool"
	"github.com/hacash/blockmint/sys/log"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/service"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	insertBlocksChSize = 255

	miningSleepNanosecond = uint64(0) // 矿工休眠时间

)

type HacashMiner struct {
	PowMiningWorkHashPower *big.Int // 哈希率，算力值
	PowMiningWorkTime      uint64   // 哈希率，算力值

	State *MinerState

	TxPool service.TxPool

	// 矿工状态标识
	miningStatusCh chan bool

	// 正在插入区块
	insertBlock sync.Mutex

	// 当前能否添加区块
	insertBlocksCh chan *DiscoveryNewBlockEvent

	// 插入区块进度 事件订阅
	insertBlockFeed      event.Feed
	insertBlockFeedScope event.SubscriptionScope

	// 成功挖掘新区快 事件订阅
	discoveryNewBlockFeed      event.Feed
	discoveryNewBlockFeedScope event.SubscriptionScope

	// 当前正在计算的区块
	CurrentPenddingBlock block.Block

	Log log.Logger
}

type DiscoveryNewBlockEvent struct {
	Success  bool // 成功写入
	Already  bool // 已经存在
	Block    block.Block
	Bodys    []byte
	insertCh chan DiscoveryNewBlockEvent
}

var (
	globalInstanceHacashMinerMutex sync.Mutex
	globalInstanceHacashMiner      *HacashMiner = nil
)

func GetGlobalInstanceHacashMiner() *HacashMiner {
	globalInstanceHacashMinerMutex.Lock()
	defer globalInstanceHacashMinerMutex.Unlock()
	if globalInstanceHacashMiner == nil {
		lg := config.GetGlobalInstanceLogger()
		globalInstanceHacashMiner = NewHacashMiner(lg)
	}
	return globalInstanceHacashMiner
}

func NewHacashMiner(logger log.Logger) *HacashMiner {
	// 检查配置
	sleepnano, e1 := strconv.ParseUint(config.Config.Miner.Stepsleepnano, 10, 0)
	if e1 != nil {
		panic("config.Config.Miner.Stepsleepnano not be " + config.Config.Miner.Stepsleepnano)
	}
	if sleepnano > 0 {
		logger.Note("miner step calculation sleep", sleepnano, "nanosecond")
	}
	miningSleepNanosecond = sleepnano
	// 创建
	miner := &HacashMiner{}
	miner.Log = logger
	miner.State = NewMinerState(logger)
	miner.State.FetchLoad()
	miner.TxPool = txpool.GetGlobalInstanceMemTxPool()
	miner.miningStatusCh = make(chan bool, 200)
	miner.insertBlocksCh = make(chan *DiscoveryNewBlockEvent, insertBlocksChSize)
	miner.CurrentPenddingBlock = nil
	miner.PowMiningWorkHashPower = big.NewInt(0) // 算力
	miner.PowMiningWorkTime = 0

	return miner
}

// 获取实时算力值，哈希率
func (this *HacashMiner) GetHashPower() {

}

/////////////  start interface  /////////////

// 获取当前基于的钻石区块hash
func (this *HacashMiner) GetPrevDiamondHash() (uint32, []byte) {
	return this.State.GetPrevDiamondBlockHash()
}

// 设置钻石区块hash
func (this *HacashMiner) SetPrevDiamondHash(diamond_number uint32, blkhash []byte) {
	this.State.SetPrevDiamondBlockHash(diamond_number, blkhash)
}

////////////  end interface  //////////////

// 开始挖矿
func (this *HacashMiner) Start() {
	go this.miningLoop()
	go this.insertBlockLoop()
}

// 开始挖矿
func (this *HacashMiner) StartMining() {
	//this.Log.Noise("hacash miner will start mining by call func StartMining()")
	this.miningStatusCh <- true
	/*
		if 0 == len(this.startingCh) {
			// 如果是停止状态
			this.Log.Info("start mining")
			this.startingCh <- true
			// 切换到
		}
	*/
}

// 开始挖矿
func (this *HacashMiner) StopMining() {
	//this.Log.Noise("hacash miner will stop mining by call func StopMining()")
	this.miningStatusCh <- false
	/*
		if 1 == atomic.LoadUint32(&this.miningStatus) {
			this.Log.Info("stop mining")
			this.stopingCh <- true
			//this.Log.Noise("stop mining ok!!!")
		}
	*/
}

// 挖矿循环
func (this *HacashMiner) miningLoop() {
	for {
		this.Log.Info("mining loop wait to start")
		select {
		case stat := <-this.miningStatusCh:
			if stat == false {
				continue // 停止状态
			}
			this.Log.Info("do mining start")
			err := this.doMining()
			if err != nil {
				this.Log.Info("mining process out for", err)
			} else {
				this.Log.Info("mining process out")
				// 继续挖掘下一个区块
				this.StartMining()
			}
		}
	}
}

func gpuMinerHttpGet(url string, statusCh chan bool) string {

	var bodyretCh = make(chan string, 0)

	go func() {
		resp, err := http.Get(url)
		if err != nil {
			bodyretCh <- ""
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			bodyretCh <- ""
			return
		}
		bodyretCh <- string(body)
	}()

	select {
	case stat := <-statusCh:
		if stat == false {
			return ""
		}
	case body := <-bodyretCh:
		return body
	}
	return ""
	//return string(body), nil
}

// 执行挖矿
func (this *HacashMiner) doMining() error {
	isMiningPoolWork := len(config.Config.MiningPool.StatisticsDir) > 0
	//return fmt.Errorf("not do")
	// 创建区块
	var cbrwaddr *fields.Address = nil
	if isMiningPoolWork {
		plm := GetGlobalInstanceMiningPool()
		cbrwaddr = &plm.StateData.FeeAccount.Address // 矿池地址
	}
	newBlock, _, coinbase, _, e := this.CreateNewBlock(cbrwaddr)
	this.CurrentPenddingBlock = newBlock
	if e != nil {
		this.Log.Warning("create new block for mining error", e)
		return e
	}

	var gpumineraddr = config.Config.GpuMiner.Address
	var usegpuminer = strings.Compare(gpumineraddr, "") != 0

	var rewardAddr *fields.Address
	var rewardAddrReadble string
	var targetFinishHash []byte

	//fmt.Println(config.Config.MiningPool.StatisticsDir)
	// 是否为矿池
	if isMiningPoolWork {
		// 发送新区块给矿池
		plm := GetGlobalInstanceMiningPool()
		ncm := &pool.NewCreateBlockEvent{
			newBlock,
			coinbase,
		}
		// 清空队列
		for {
			select {
			case <-plm.CalcSuccessBlockCh:
			default:
				goto CLEAN_CalcSuccessBlockCh
			}
		}
	CLEAN_CalcSuccessBlockCh:
		// 通知挖新区块
		plm.NewCreateBlockCh <- ncm
		// 等待矿池的返回
		select {
		case stat := <-this.miningStatusCh: // 新区快到来
			if stat == false {
				this.reputAllTxsFromBlock(newBlock)                        // 重新放入所有交易到交易池
				return fmt.Errorf("mining break by set sign stoping chan") // 停止挖矿
			}
		// 矿池挖出
		case <-plm.CalcSuccessBlockCh:
			this.Log.Info("find a valid nonce for block", "height", newBlock.GetHeight())
			targetFinishHash = newBlock.Hash()
			goto MINING_SUCCESS
		}

		// 是否为多线程挖矿
	} else if config.Config.Miner.Supervene > 0 {
		// 多线程并发挖矿
		_, rewardAddrReadble = this.setMinerForCoinbase(coinbase, false)
		newBlock = this.calculateNextBlock(newBlock, coinbase)
		if newBlock == nil {
			return fmt.Errorf("mining break by set sign stoping chan on supervene") // 停止挖矿
		}
		targetFinishHash = newBlock.Hash()

	} else {

		// 普通挖矿 或者 GPU 挖掘计算
		targetHash := difficulty.Uint32ToHash(newBlock.GetHeight(), newBlock.GetDifficulty())
		targetDifficulty := difficulty.HashToBig(newBlock.GetHeight(), targetHash)

	RESTART_TO_MINING:

		rewardAddr, rewardAddrReadble = this.setMinerForCoinbase(coinbase, true) // coinbase
		newBlock.SetMrklRoot(blocks.CalculateMrklRoot(newBlock.GetTransactions())) // update mrkl root
		this.Log.News("set new coinbase address", rewardAddrReadble, "height", newBlock.GetHeight(), "do mining...")

		if usegpuminer {
			// 是否为GPU挖矿
			this.Log.News("config.Config.GpuMiner.Address: " + gpumineraddr)
			// http 访问请求
			stuff := blocks.CalculateBlockHashBaseStuff(newBlock)
			url := fmt.Sprintf("http://%s/dominer?height=%d&targethash=%s&blockstuff=%s&coinaddr=%s&coinmsg=%s",
				gpumineraddr,
				newBlock.GetHeight(),
				hex.EncodeToString(targetHash),
				hex.EncodeToString(stuff),
				hex.EncodeToString(*rewardAddr),
				hex.EncodeToString([]byte(coinbase.Message)),
			)
			this.Log.Info(url)
			retstr := gpuMinerHttpGet(url, this.miningStatusCh)
			this.Log.Info(retstr)
			if strings.Compare(retstr, "") == 0 {
				this.reputAllTxsFromBlock(newBlock)   // 重新放入所有交易到交易池
				return fmt.Errorf("gpu mining break") // 停止挖矿
			}
			retstrs := strings.Split(retstr, ".")
			if len(retstrs) >= 2 {
				if strings.Compare(retstrs[1], "retry") == 0 {
					// 重试
					goto RESTART_TO_MINING
				} else if strings.Compare(retstrs[1], "success") == 0 {
					// 挖矿成功
					nonce, _ := hex.DecodeString(retstrs[2])
					noncenum := binary.BigEndian.Uint32(nonce)
					newBlock.SetNonce(noncenum)
					targetFinishHash = newBlock.HashFresh()
					// OK !!!!!!!!!!!!!!!
					goto MINING_SUCCESS
				} else {
					// 挖矿退出
					this.reputAllTxsFromBlock(newBlock)                            // 重新放入所有交易到交易池
					return fmt.Errorf("gpu mining break by set sign stoping chan") // 停止挖矿
				}
			} else {
				this.reputAllTxsFromBlock(newBlock) // 重新放入所有交易到交易池
				return fmt.Errorf("gpu mining error return: %s", retstr)
			}

		} else {
			// 普通单核CPU挖矿
			//this.Log.Noise("mining by simply CPU")
			for i := uint32(0); i < 4294967295; i++ {
				//this.Log.Noise(i)
				select {
				case stat := <-this.miningStatusCh:
					if stat == false {
						this.reputAllTxsFromBlock(newBlock) // 重新放入所有交易到交易池
						// this.Log.Debug("mining break and stop mining -…………………………………………………………………………")
						return fmt.Errorf("mining break by set sign stoping chan") // 停止挖矿
					}
				default:

				}
				if miningSleepNanosecond > 0 {
					time.Sleep(time.Duration(miningSleepNanosecond) * time.Nanosecond)
				}
				newBlock.SetNonce(i)
				targetFinishHash = newBlock.HashFresh()
				// curdiff := difficulty.BigToCompact(difficulty.HashToBig_v1(&targetHash))
				curdiff := difficulty.HashToBig(newBlock.GetHeight(), targetFinishHash)
				//fmt.Println(curdiff, targetDifficulty)
				if curdiff.Cmp(targetDifficulty) == -1 {
					this.Log.Info("find a valid nonce for block", "height", newBlock.GetHeight())
					// OK !!!!!!!!!!!!!!!
					goto MINING_SUCCESS
				}
			}
			goto RESTART_TO_MINING // 下一轮次
		}
	}

MINING_SUCCESS:

	// 挖矿成功！！！
	this.CurrentPenddingBlock = nil

	// 插入并等待结果
	insert := this.InsertBlockWait(newBlock, nil)
	if insert.Success {
		targethashhex := hex.EncodeToString(targetFinishHash)
		// 广播新区快信息
		this.Log.Info("mining success one block", "hash", targethashhex)
		go this.discoveryNewBlockFeed.Send(DiscoveryNewBlockEvent{
			Block: newBlock,
			Bodys: insert.Bodys,
		})
		// 打印相关信息
		str_time := time.Unix(int64(newBlock.GetTimestamp()), 0).Format("01/02 15:04:05")
		this.Log.Note(fmt.Sprintf("⬤  %s, bh: %d, tx: %d, hx: %s, px: %s, df: %d, cm: %s, tt: %s",
			coinbase.Reward.ToFinString(),
			int(newBlock.GetHeight()),
			len(newBlock.GetTransactions())-1,
			targethashhex,
			hex.EncodeToString(newBlock.GetPrevHash()[0:10])+"...",
			newBlock.GetDifficulty(),
			rewardAddrReadble,
			str_time,
		))
	} else {
		length := this.reputAllTxsFromBlock(newBlock)
		this.Log.Warning("mining finish block", "height", newBlock.GetHeight(), "hash", hex.EncodeToString(newBlock.Hash()), "insert chain fail, reappend all txs", length, "and clear block")
	}
	return nil
}

// 将交易全部放入交易池
func (this *HacashMiner) reputAllTxsFromBlock(newBlock block.Block) int {
	blktxs := newBlock.GetTransactions()
	length := len(blktxs) - 1
	for i := length; i > 0; i-- { // drop coinbase
		this.TxPool.AddTx(blktxs[i]) // 倒序重新放入交易池
	}
	return length
}

// 插入区块
func (this *HacashMiner) InsertBlock(blk block.Block, bodys []byte, insertCh chan DiscoveryNewBlockEvent) {
	this.insertBlocksCh <- &DiscoveryNewBlockEvent{
		false,
		false,
		blk,
		bodys,
		insertCh,
	}
}

// 插入区块
func (this *HacashMiner) InsertBlocks(blks []block.Block) {
	go func() {

		for _, blk := range blks {
			this.insertBlocksCh <- &DiscoveryNewBlockEvent{
				false,
				false,
				blk,
				nil,
				nil,
			}
		}

	}()
}

// 插入区块
func (this *HacashMiner) insertBlockLoop() {
	for {
		//tk := time.NewTimer(time.Second * 9)
		select {
		case blk := <-this.insertBlocksCh:
			this.Log.Info("insert loop get one block", "height", blk.Block.GetHeight())
			// tk.Stop()
			this.StopMining() // 停止挖矿
			err := this.doInsertBlock(blk)
			if err != nil {
				this.Log.Error("do insert block loop", "height", blk.Block.GetHeight(), "error", err)
			} else {
				this.Log.Info("insert block ok", "height", blk.Block.GetHeight())
			}
			//case <-tk.C:
			// this.Log.Noise("no block to insert")
			// this.StartMining() // 几秒后没有区块插入则自动开始挖矿
		}
	}
}

// 插入区块
func (this *HacashMiner) doInsertBlock(blk *DiscoveryNewBlockEvent) error {
	if blk.Block == nil && (blk.Bodys == nil || len(blk.Bodys) == 0) {
		return fmt.Errorf("data is empty")
	}
	if blk.Block == nil {
		b, _, e := blocks.ParseBlock(blk.Bodys, 0)
		if e != nil {
			return e
		}
		blk.Block = b
	}
	block := blk.Block
	successInsert := false
	alreadyInsert := false
	var blockBytes []byte = nil
	defer func() {
		// fmt.Println("go this.insertBlockFeed.Send(DiscoveryNewBlockEvent{  successInsert == ", successInsert, " alreadyInsert == ", alreadyInsert)
		// 插入处理事件通知
		event := DiscoveryNewBlockEvent{
			successInsert,
			alreadyInsert,
			block,
			blockBytes,
			nil,
		}
		// 写入回调
		if blk.insertCh != nil {
			blk.insertCh <- event
		}
		// 发送消息
		go this.insertBlockFeed.Send(event)
	}()

	// 判断重复

	// 判断区块重复，已经收到这个区块，则立即返回正确
	if bytes.Compare(this.State.CurrentBlockHash(), block.Hash()) == 0 {
		// fmt.Println("bytes.Compare(this.State.CurrentBlockHash(), block.Hash()) == 0,  successInsert = true,  alreadyInsert = true")
		successInsert = true
		alreadyInsert = true // 已经存在
		if blk.Bodys == nil || len(blk.Bodys) == 0 {
			blockBytes, _ = block.Serialize()
		} else {
			blockBytes = blk.Bodys
		}
		return nil
	}

	// 判断高度
	var fail_height = this.State.CurrentHeight()+1 != block.GetHeight()
	var fail_prevhash = bytes.Compare(this.State.CurrentBlockHash(), block.GetPrevHash()) != 0
	if fail_height || fail_prevhash {
		// 错误
		var typestr = "prev hash"
		if fail_height {
			typestr += "height"
		}
		return fmt.Errorf("not accepted block with wrong %s, height=%d, hash=%s, target_prev_hash=%s, base_height=%d, base_hash=%s, base_prev_hash=%s",
			typestr,
			block.GetHeight(),
			hex.EncodeToString(block.Hash()),
			hex.EncodeToString(block.GetPrevHash()),
			this.State.CurrentHeight(),
			hex.EncodeToString(this.State.CurrentBlockHash()),
			hex.EncodeToString(this.State.prevBlockHead.GetPrevHash()),
		)
	}
	// 判断时间
	prevblocktime := this.State.GetBlockHead().GetTimestamp()
	blktime := block.GetTimestamp()
	if blktime <= prevblocktime {
		str_time := time.Unix(int64(blktime), 0).Format("01/02 15:04:05")
		str_time_prev := time.Unix(int64(prevblocktime), 0).Format("01/02 15:04:05")
		return fmt.Errorf("block %d timestamp %s cannot be accept, prev blocktime is %s", block.GetHeight(), str_time, str_time_prev)
	}
	if blktime > uint64(time.Now().Unix()) {
		str_time := time.Unix(int64(blktime), 0).Format("01/02 15:04:05")
		str_time_system := time.Now().Format("01/02 15:04:05")
		return fmt.Errorf("block %d timestamp %s cannot be accept, cannot more than current system time %s", block.GetHeight(), str_time, str_time_system)
	}
	// 检查难度值
	blkhash := block.HashFresh()
	hxdift := difficulty.HashToUint32(block.GetHeight(), blkhash)
	tardgbig, tardift := this.State.TargetDifficultyCompact(block.GetHeight(), nil)
	if tardgbig.Cmp(difficulty.HashToBig(block.GetHeight(), blkhash)) == -1 {
		return fmt.Errorf("difficulty not satisfy, height %d, accept %d, but got %d", block.GetHeight(), tardift, hxdift)
	}
	// 判断默克尔root
	mrklRoot := blocks.CalculateMrklRoot(block.GetTransactions())
	if bytes.Compare(mrklRoot, block.GetMrklRoot()) != 0 {
		return fmt.Errorf("block %d mrkl root hash wrong, accept %s, but got %s",
			block.GetHeight(),
			hex.EncodeToString(mrklRoot),
			hex.EncodeToString(block.GetMrklRoot()))
	}
	// 检查交易数量
	if uint32(len(block.GetTransactions())) != block.GetTransactionCount() {
		return fmt.Errorf("block %d transaction count wrong, accept %d, but got %d",
			block.GetHeight(),
			len(block.GetTransactions()),
			block.GetTransactionCount())
	}
	// 检查coinbase
	txs := block.GetTransactions()
	if len(txs) < 1 {
		return fmt.Errorf("block %d need coinbase transaction", block.GetHeight())
	}
	coinbase, ok := txs[0].(*transactions.Transaction_0_Coinbase)
	if !ok {
		return fmt.Errorf("block %d need coinbase transaction", block.GetHeight())
	}
	targetreward := coin.BlockCoinBaseReward(block.GetHeight())
	if !targetreward.Equal(coinbase.GetReward()) {
		return fmt.Errorf("block %d coinbase reward need %s but get %s", block.GetHeight(), targetreward.ToFinString(), coinbase.GetReward().ToFinString())
	}
	// 验证全部交易签名
	sigok, e1 := block.VerifyNeedSigns()
	if e1 != nil {
		return e1
	}
	if !sigok {
		return fmt.Errorf("block signature verify faild")
	}
	// 判断交易是否已经存在
	blockdb := store.GetGlobalInstanceBlocksDataStore()
	if len(txs) < 1 {
		return fmt.Errorf("block is empty")
	}
	for i := 1; i < len(txs); i++ {
		txhashnofee := txs[i].HashNoFee()
		if ok, e := blockdb.CheckTransactionExist(txhashnofee); ok || e != nil {
			return fmt.Errorf("tx %s is exist", hex.EncodeToString(txhashnofee))
		}
	}
	// 验证交易
	newBlockChainState := state.NewTempChainState(nil)
	newBlockChainState.SetBlock(block) // 设置当前处理的区块
	newBlockChainState.SetMiner(this)  // 矿工状态
	blksterr := block.ChangeChainState(newBlockChainState)
	if blksterr != nil {
		return blksterr
	}
	// 存储区块数据
	var sverr error
	blockBytes, sverr = blockdb.Save(block)
	if sverr != nil {
		return sverr
	}
	// 保存用户状态
	chainstate := state.GetGlobalInstanceChainState()
	chainstate.TraversalCopy(newBlockChainState)
	// 从交易池内清除已经出块的交易
	this.TxPool.RemoveTxs(block.GetTransactions())
	// 更新矿工状态
	this.State.SetNewBlock(block)
	diamondNumber, diamondHash := newBlockChainState.GetPrevDiamondHash()
	if diamondNumber > 0 && diamondHash != nil {
		this.SetPrevDiamondHash(diamondNumber, diamondHash)
	}

	// 成功状态
	successInsert = true

	return nil
}

// 插入区块进度事件
func (this *HacashMiner) SubscribeInsertBlock(insertCh chan<- DiscoveryNewBlockEvent) event.Subscription {
	return this.insertBlockFeedScope.Track(this.insertBlockFeed.Subscribe(insertCh))
}

// 订阅挖掘出新区快事件
func (this *HacashMiner) SubscribeDiscoveryNewBlock(discoveryCh chan<- DiscoveryNewBlockEvent) event.Subscription {
	return this.discoveryNewBlockFeedScope.Track(this.discoveryNewBlockFeed.Subscribe(discoveryCh))
}

// 插入区块，等待插入状态返回
func (this *HacashMiner) InsertBlockWait(blk block.Block, bodys []byte) DiscoveryNewBlockEvent {
	// 写入区块
	insertCh := make(chan DiscoveryNewBlockEvent, 1)
	this.Log.Debug("insert block to chain state with wait", "height", blk.GetHeight())
	this.InsertBlock(blk, bodys, insertCh)
	return <-insertCh
}

// 创建区块
func (this *HacashMiner) CreateNewBlock(minerAddress *fields.Address) (block.Block, *state.ChainState, *transactions.Transaction_0_Coinbase, *fields.Amount, error) {
	nextblock := blocks.NewEmptyBlock_v1(this.State.prevBlockHead)

	//////////////// test ////////////////
	//timeset := coin.GetGenesisBlock().GetTimestamp() + nextblock.GetHeight()*300 - 150 + uint64(rand.Int63n(300))
	//nextblock.CreateTimestamp = fields.VarInt5(timeset)
	//if timeset > uint64(time.Now().Unix()) {
	//	panic("okokokok")
	//}
	//////////////// test ////////////////

	hei, _, dfct, print := this.State.NextHeightTargetDifficultyCompact()
	if print != nil && *print != "" {
		this.Log.Note(*print)
	}
	nextblock.Height = fields.VarInt5(hei)
	nextblock.Difficulty = fields.VarInt4(dfct)
	coinbase := this.createCoinbaseTx(nextblock)
	if minerAddress != nil {
		coinbase.Address = *minerAddress // 自定义矿工地址
	}
	nextblock.TransactionCount = 1
	nextblock.Transactions = append(nextblock.Transactions, coinbase)
	// 获取交易并验证
	tempBlockState := state.NewTempChainState(nil)
	tempBlockState.SetBlock(nextblock) // 设置当前处理的区块
	tempBlockState.SetMiner(this)
	// 添加交易
	stoblk := store.GetGlobalInstanceBlocksDataStore()
	blockSize := uint32(block1def.ByteSizeBlockBeforeTransaction)
	blockTotalFee := fields.NewEmptyAmount()
	cacheNextTrs := make([]block.Transaction, 0)
	for {
		trs := this.TxPool.PopTxByHighestFee()
		if trs == nil {
			break // nothing
		}
		hashnofee := trs.HashNoFee() // 交易是否已经存在
		ext, e2 := stoblk.CheckTransactionExist(hashnofee)
		if e2 != nil || ext {
			continue // drop tx
		}
		blockSize += trs.Size() // 区块大小上限2MB
		if int64(blockSize) > int64(1024)*1024*2 {
			this.TxPool.AddTx(trs)
			break // over block size
		}
		hxstate := state.NewTempChainState(tempBlockState)
		hxstate.SetBlock(nextblock) // 设置当前处理的区块
		hxstate.SetMiner(this)
		errun := trs.ChangeChainState(hxstate)
		if errun != nil {
			// fmt.Println(errun)
			if strings.HasPrefix(errun.Error(), "{BACKTOPOOL}") {
				// put back to pool // 保留，下一个区块处理
				cacheNextTrs = append(cacheNextTrs, trs)
			}
			// 出现其他错误则直接丢弃交易
			hxstate.Destroy()
			continue // error , give up tx
		}
		// ok copy state
		tempBlockState.TraversalCopy(hxstate)
		hxstate.Destroy()
		nextblock.Transactions = append(nextblock.Transactions, trs)
		nextblock.TransactionCount += 1
		// 手续费
		fee := fields.ParseAmount(trs.GetFee(), 0)
		blockTotalFee, _ = blockTotalFee.Add(fee)
		// put back to pool // 成功的交易，保留，供其他连接的节点下载同步
		cacheNextTrs = append(cacheNextTrs, trs)
	}
	// 将交易扔回交易池，下一个区块处理或者供其他连接的节点下载同步
	// 出块成功时将移除已经出块的交易
	for _, trs := range cacheNextTrs {
		this.TxPool.AddTx(trs)
	}

	this.Log.Info("create new block", "height", nextblock.Height, "transaction", nextblock.TransactionCount-1)

	return nextblock, tempBlockState, coinbase, blockTotalFee, nil
}

// 创建coinbase交易
func (this *HacashMiner) createCoinbaseTx(block block.Block) *transactions.Transaction_0_Coinbase {
	// coinbase
	coinbase := transactions.NewTransaction_0_Coinbase()
	coinbase.Reward = *(coin.BlockCoinBaseReward(uint64(block.GetHeight())))
	this.setMinerForCoinbase(coinbase, false)
	return coinbase
}

// 设置coinbase交易
func (this *HacashMiner) setMinerForCoinbase(coinbase *transactions.Transaction_0_Coinbase, randmsgtail bool) (*fields.Address, string) {
	addrreadble := config.GetRandomMinerRewardAddress()
	addr, e := fields.CheckReadableAddress(addrreadble)
	if e != nil {
		panic("Miner Reward Address `" + addrreadble + "` Error !")
	}
	coinbase.Address = *addr

	// 末尾随机数
	if randmsgtail {
		markwork := []byte(config.Config.Miner.Markword)
		if len(markwork) > 11 {
			panic("config.Config.Miner.Markword length too long over 11")
		}
		minermsg := make([]byte, 16)
		binary.BigEndian.PutUint32(minermsg[12:], rand.Uint32())
		copy(minermsg, markwork)
		coinbase.Message = fields.TrimString16(minermsg)
	}

	return addr, addrreadble
}

// 倒退区块
func (this *HacashMiner) BackTheWorldToHeight(target_height uint64) ([]block.Block, error) {

	current_height := this.State.CurrentHeight()
	if target_height >= current_height {
		// do nothing
		return []block.Block{}, nil
	}

	// 被回退的区块
	var backblks = make([]block.Block, current_height-target_height-1)

	db := store.GetGlobalInstanceBlocksDataStore()
	state := state.GetGlobalInstanceChainState()
	state.SetMiner(this)
	for {
		_, blkbts, err := db.GetBlockBytesByHeight(current_height, true, true, 0)
		if err != nil {
			return backblks, err
		}
		blkobj, _, e2 := blocks.ParseBlock(blkbts, 0)
		if e2 != nil {
			return backblks, e2
		}
		backblks = append(backblks, blkobj)
		this.Log.Note("delete height", current_height, "hash", hex.EncodeToString(blkobj.Hash()), "prev_hash", hex.EncodeToString(blkobj.GetPrevHash()[0:16])+"...")
		// 回退状态
		blkobj.RecoverChainState(state)
		// 删除数据
		db.DeleteBlockForceUnsafe(blkobj)
		// 是否完成
		current_height--
		if current_height <= target_height {
			break
		}
	}
	// 修改矿工状态
	_, blkhdbts, e0 := db.GetBlockBytesByHeight(current_height, true, true, 0)
	if e0 != nil {
		return backblks, e0
	}
	//fmt.Println("head bytes ", hex.EncodeToString(blkhdbts))
	blkhead, _, e2 := blocks.ParseBlock(blkhdbts, 0)
	if e2 != nil {
		return backblks, e2
	}
	prev288blkhei := current_height - (current_height % config.ChangeDifficultyBlockNumber)
	if prev288blkhei == 0 {
		rootblk := coin.GetGenesisBlock()
		this.State.prev288BlockTimestamp = rootblk.GetTimestamp() // 起始时间戳
	} else {
		_, blkhdbts_prev288, e1 := db.GetBlockBytesByHeight(prev288blkhei, true, false, 0)
		if e1 != nil {
			return backblks, e1
		}
		blkhead_prev288, _, e3 := blocks.ParseBlockHead(blkhdbts_prev288, 0)
		if e3 != nil {
			return backblks, e3
		}
		this.State.prev288BlockTimestamp = blkhead_prev288.GetTimestamp() // 时间戳
		//fmt.Println("修改矿工状态 height", blkhead.GetHeight(), "hash", hex.EncodeToString(blkhead.Hash()))
	}
	// 修改矿工状态
	this.State.SetNewBlock(blkhead)
	diamondNumber, diamondHash := state.GetPrevDiamondHash()
	if diamondNumber >= 0 && diamondHash != nil {
		this.SetPrevDiamondHash(diamondNumber, diamondHash)
		state.SetPrevDiamondHash(0, nil) // 状态复位
	}
	// ok
	return backblks, nil
}

/////////////////////////////////////////////////////////////////////////////////
