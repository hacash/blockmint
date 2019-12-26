package pool

import (
	"github.com/hacash/blockmint/block/fields"
	"math/big"
	"net"
	"time"
)

// 客户端连接
type Client struct {
	// 标识
	Id uint64
	// tcp 连接
	Conn net.Conn
	// 奖励地址
	RewordAddress *fields.Address
	// 当前的挖矿标号
	//CurrentMiningMarkNumber uint64
	// 上次活跃的时间戳
	ActiveTimestamp *time.Time
	// 挖矿地址统计
	Worker *PowWorker

	// 用于统计算力的 区块头
	MiningBlockStuffCurrent []byte // 当前正在挖掘的区块头
	MiningBlockStuffPrev    []byte // 上一个挖掘的区块头，用于统计算力

}

/////////////////////////////////////////////////////////

//
type PowWorker struct {
	// 奖励地址
	RewordAddress *fields.Address
	// 持久化统计信息
	StatisticsData *AddressStatisticsStoreItem
	// 实时算力统计
	RealtimePower           *big.Int // 上一个挖掘周期的实时算力
	RealtimeWorkSubmitCount uint64   // 挖掘周期工作次数（线程数量统计）提交
	ClientCount             uint32   // tcp连接数量

}
