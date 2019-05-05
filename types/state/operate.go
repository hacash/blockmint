package state

import "github.com/hacash/blockmint/block/fields"

// chain state 操作

type ChainStateOperation interface {

	// query

	Balance(fields.Address) fields.Amount // 查询账户余额

	Channel(fields.Bytes16) // 查询交易通道

	// operate

	BalanceSet(fields.Address, fields.Amount) // 余额设定
	BalanceDel(fields.Address)                // 余额删除

	ChannelCreate(fields.Bytes16, fields.Address, fields.Address, fields.Amount) // 开启通道
	ChannelDelete(fields.Bytes16)                                                // 删除通道
}
