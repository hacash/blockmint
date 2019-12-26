package state

import (
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/miner"
)

// chain state 操作

type ChainStateOperation interface {

	// status

	Block() interface{}   // block.Block
	SetBlock(interface{}) // block.Block
	Miner() miner.Miner
	SetMiner(miner.Miner)

	// state

	GetPrevDiamondHash() (uint32, []byte) // 获取当前基于的钻石区块hash
	SetPrevDiamondHash(uint32, []byte)    // 设置钻石区块hash

	// query

	Balance(fields.Address) fields.Amount // 查询账户余额
	Channel(fields.Bytes16) interface{}   // 查询交易通道
	Diamond(fields.Bytes6) interface{}    // 查询钻石所属

	// operate

	BalanceSet(fields.Address, fields.Amount) // 余额设定
	BalanceDel(fields.Address)                // 余额删除

	ChannelCreate(fields.Bytes16, interface{}) // 开启通道
	ChannelDelete(fields.Bytes16)              // 删除通道

	DiamondSet(fields.Bytes6, interface{})       // 设置钻石
	DiamondBelong(fields.Bytes6, fields.Address) // 更改所属
	DiamondDel(fields.Bytes6)                    // 移除钻石

}
