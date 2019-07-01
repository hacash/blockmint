package miner

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/core/coin"
	"github.com/hacash/blockmint/miner/difficulty"
	"github.com/hacash/blockmint/sys/file"
	sys_log "github.com/hacash/blockmint/sys/log"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/x16rs"
	"log"
	"math/big"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

/**
 * 矿池程序
 */

// 结算数据储存
type PoolPeriodStateData struct {
	flushFileName string

	dataIndex uint32 // 数据标号

	SuccessRewards uint32 // 成功的区块奖励HCX数量
	// addressMaxNumber uint32 // 使用的地址后缀编号上限
	NodeNumber            uint64   // 参与的地址数量
	TotalMiningPower      *big.Int // 总算力大小
	MiningPowerStatistics sync.Map // map[string]*big.Int 算力分别统计

}

// 解析数据
func (p *PoolPeriodStateData) Parse(stuff []byte, seek uint64) error {
	p.dataIndex = binary.BigEndian.Uint32(stuff[seek : seek+4])
	seek += 4
	p.SuccessRewards = binary.BigEndian.Uint32(stuff[seek : seek+4])
	seek += 4
	p.NodeNumber = binary.BigEndian.Uint64(stuff[seek : seek+8])
	seek += 8
	p.TotalMiningPower = (new(big.Int)).SetBytes(stuff[seek : seek+128])
	seek += 128
	for i := 0; i < int(p.NodeNumber); i++ {
		addr := fields.Address(stuff[seek : seek+21])
		addrstr := (addr).ToReadable()
		seek += 21
		value := (new(big.Int)).SetBytes(stuff[seek : seek+32])
		p.MiningPowerStatistics.Store(addrstr, value)
		seek += 32
	}
	return nil
}

// 保存数据
func (p *PoolPeriodStateData) Serialize() []byte {
	buf := bytes.NewBuffer([]byte{})
	s1 := make([]byte, 4)
	s2 := make([]byte, 4)
	s3 := make([]byte, 8)
	binary.BigEndian.PutUint32(s1, p.dataIndex)
	binary.BigEndian.PutUint32(s2, p.SuccessRewards)
	binary.BigEndian.PutUint64(s3, p.NodeNumber)
	buf.Write(s1)
	buf.Write(s2)
	buf.Write(s3)
	totalpowval := p.TotalMiningPower.Bytes()
	if len(totalpowval) > 128 {
		panic("TotalMiningPower is too big, bytes lenght overflow 128 bytes.")
	}
	buf.Write(bytes.Repeat([]byte{0}, 128-len(totalpowval)))
	buf.Write(totalpowval)
	// 分别统计值
	p.MiningPowerStatistics.Range(func(key interface{}, val interface{}) bool {
		addrstr := key.(string)
		valuebts := val.(*big.Int).Bytes()
		//client.Close()
		addr, err := fields.CheckReadableAddress(addrstr)
		if err == nil && len(valuebts) < 32 {
			buf.Write(*addr)
			buf.Write(bytes.Repeat([]byte{0}, 32-len(valuebts)))
			buf.Write(valuebts)
		}
		return true
	})
	// 返回
	return buf.Bytes()
}

type NewCreateBlock struct {
	Block    block.Block
	Coinbase *transactions.Transaction_0_Coinbase
}

type MiningPool struct {
	flushFileName string

	currentPoolPeriodStateData *PoolPeriodStateData
	currentNewCreateBlock      *NewCreateBlock // 正在挖掘的区块
	prevSuccessBlockHash       []byte

	allActiveConns sync.Map // 全部客户端连接

	autoincrementMiningCoinbaseStuffNum uint64 // 自增的挖矿序号
	prevPowerStatisticsBlockHeadMeta    []byte // 89 上一个区块的区块头，用来统计算力

	// 数据通信
	NewCreateBlockCh    chan NewCreateBlock // 新挖掘的区块到来
	DiscoveryNewBlockCh chan block.Block    // 发现新的区块
	CalcSuccessBlockCh  chan x16rs.MiningSuccess

	// 配置
	addressMaxNum uint64 // 统计地址数量上限
	markword      string // 区块播报方

}

var (
	globalInstanceMiningPoolMutex sync.Mutex
	globalInstanceMiningPool      *MiningPool = nil
)

func GetGlobalInstanceMiningPool() *MiningPool {
	globalInstanceMiningPoolMutex.Lock()
	defer globalInstanceMiningPoolMutex.Unlock()
	if globalInstanceMiningPool == nil {
		lg := config.GetGlobalInstanceLogger()
		globalInstanceMiningPool = NewMiningPool(lg)
	}
	return globalInstanceMiningPool
}

func NewMiningPool(logger sys_log.Logger) *MiningPool {

	fnp := config.Config.MiningPool.StatisticsDir
	if strings.Compare(fnp, "") == 0 {
		panic("config.Config.MiningPool.StatisticsDir must be set.")
	}

	return &MiningPool{
		flushFileName:                    fnp,
		currentPoolPeriodStateData:       nil,
		NewCreateBlockCh:                 make(chan NewCreateBlock, 8),
		DiscoveryNewBlockCh:              make(chan block.Block, 8),
		CalcSuccessBlockCh:               make(chan x16rs.MiningSuccess, 8),
		prevPowerStatisticsBlockHeadMeta: nil,
		prevSuccessBlockHash:             nil,

		addressMaxNum: config.Config.MiningPool.AddressMax,
		markword:      config.Config.MiningPool.Markword,
	}

}

// 启动
func (mp *MiningPool) Start() error {
	// 读取保存的数据文件
	for {
		time.Sleep(time.Second)
		nhei := GetGlobalInstanceHacashMiner().State.CurrentHeight()
		if nhei > 0 {
			mp.readStoreData(nhei)
			break
		}
	}
	// 接口监听
	go mp.startApiListen()
	// 区块创建
	go mp.createBlockLoop()
	// 定时保存数据到磁盘
	go mp.storeDataLoop()

	return nil
}

// 读取保存的数据文件
func (mp *MiningPool) readStoreData(blkheight uint64) {
	newdataidx := blkheight / (288 * 7)
	flushFileName := mp.flushFileName + "/miningpooldata.dat." + strconv.FormatUint(newdataidx, 10)
	fmt.Println("--- --- load miner pool statistics file: " + flushFileName)
	mp.currentPoolPeriodStateData = ReadStoreDataByFileName(flushFileName)
}

func ReadStoreDataByFileName(flushFileName string) *PoolPeriodStateData {

	if file.IsExist(flushFileName) {
		var data PoolPeriodStateData
		data.flushFileName = flushFileName
		file, _ := os.Open(flushFileName)
		info, _ := file.Stat()
		fbytes := make([]byte, info.Size())
		file.Read(fbytes)
		data.Parse(fbytes, 0)
		return &data
	} else {
		return nil
	}
}

//
func (mp *MiningPool) storeDataLoop() {
	for {
		time.Sleep(time.Second * 13)
		// time.Sleep( time.Minute * 5 )
		mp.StoreCountDataToDisk()
	}
}

// 保存至磁盘
func (mp *MiningPool) StoreCountDataToDisk() error {
	data := mp.currentPoolPeriodStateData
	if data == nil {
		return nil
	}
	if !file.IsExist(data.flushFileName) {
		file.CreatePath(path.Dir(data.flushFileName))
	}
	filebts := data.Serialize()
	file, fe := os.OpenFile(data.flushFileName, os.O_RDWR|os.O_CREATE, 0777)
	if fe != nil {
		panic("file '" + data.flushFileName + "'")
	}
	file.WriteAt(filebts, 0)
	file.Close()
	//fmt.Println( "flush file '" + data.flushFileName + "'")

	return nil
}

//
func (mp *MiningPool) createBlockLoop() {
	for {
		ncb := <-mp.NewCreateBlockCh
		nhei := ncb.Block.GetHeight()
		// 记录上一个区块头
		if mp.currentNewCreateBlock != nil {
			if mp.currentNewCreateBlock.Block.GetHeight() == nhei {
				continue // 重复的区块，下次再说
			}
			// 用来统计算力
			mp.prevPowerStatisticsBlockHeadMeta = blocks.CalculateBlockHashBaseStuff(mp.currentNewCreateBlock.Block)
		}
		// 记录挖出区块总奖励
		if mp.currentPoolPeriodStateData != nil &&
			bytes.Compare(mp.prevSuccessBlockHash, ncb.Block.GetPrevHash()) == 0 {
			mp.currentPoolPeriodStateData.SuccessRewards += uint32(coin.BlockCoinBaseRewardNumber(nhei - 1))
		}
		// 重置统计数据
		newdataidx := nhei / (288 * 7)
		if mp.currentPoolPeriodStateData == nil || uint32(newdataidx) > mp.currentPoolPeriodStateData.dataIndex {
			// 储存旧的
			if mp.currentPoolPeriodStateData != nil {
				mp.StoreCountDataToDisk()
			}
			// 创建新的
			sd := &PoolPeriodStateData{
				flushFileName:         mp.flushFileName + "/miningpooldata.dat." + strconv.FormatUint(newdataidx, 10),
				dataIndex:             uint32(newdataidx),
				SuccessRewards:        0,
				NodeNumber:            0,
				TotalMiningPower:      big.NewInt(0),
				MiningPowerStatistics: sync.Map{},
			}
			mp.currentPoolPeriodStateData = sd // 重设
		}
		// 当前区块
		mp.currentNewCreateBlock = &ncb
		// 重置数据
		mp.autoincrementMiningCoinbaseStuffNum = 0
		// 给所有节点发送重启挖矿的消息
		mp.sendMiningStuffToAllNode(&ncb)
	}
}

// 获得成功的区块
func (mp *MiningPool) gotSuccessBlock(client *Client, success x16rs.MiningSuccess) {
	// 检查区块难度值是否满足要求
	cncb := mp.currentNewCreateBlock
	if cncb == nil {
		return // 无
	}
	mp.fillCoinbaseMsg(cncb.Coinbase, success.MiningIndex)
	cncb.Block.SetMrklRoot(blocks.CalculateMrklRoot(cncb.Block.GetTransactions()))
	cncb.Block.SetNonce(binary.BigEndian.Uint32(success.Nonce))
	hx := cncb.Block.HashFresh()
	curdiff := difficulty.HashToBig(cncb.Block.GetHeight(), hx)
	targetDifficultyHash := difficulty.Uint32ToBig(cncb.Block.GetHeight(), cncb.Block.GetDifficulty())
	//fmt.Println(curdiff, targetDifficulty)
	//fmt.Println(hx)
	//fmt.Println(difficulty.Uint32ToHash(cncb.Block.GetHeight(), cncb.Block.GetDifficulty()))
	if curdiff.Cmp(targetDifficultyHash) == -1 {
		log.Println("mining pool find a valid nonce for block", "height", cncb.Block.GetHeight())
		// OK !!!!!!!!!!!!!!!
		//goto MINING_SUCCESS
		// 发现满足要求的区块
		mp.currentNewCreateBlock = nil // 暂停挖矿
		mp.prevSuccessBlockHash = hx
		// 加入
		mp.CalcSuccessBlockCh <- success
	} else {
		log.Println("hx not ok", cncb.Block.GetHeight(), hex.EncodeToString(hx))
	}

	// 增加算力统计
	addMinerPowerValue(hx, mp.currentPoolPeriodStateData, client)
}

//////////////////////////////////

func (mp *MiningPool) fillCoinbaseMsg(coin *transactions.Transaction_0_Coinbase, miningNum uint64) {

	s1 := make([]byte, 8)
	binary.BigEndian.PutUint64(s1, miningNum)
	blkmsg := []byte("        ")
	copy(blkmsg, mp.markword) // []byte("pool.HCX")
	blkmsg = append(blkmsg, s1...)
	coin.Message = fields.TrimString16(blkmsg)
}
