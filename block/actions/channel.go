package actions

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/state"
)

/**
 * 支付通道交易类型
 */

// 开启支付通道
type Action_2_OpenPaymentChannel struct {
	ChannelId    fields.Bytes16 // 通道id
	LeftAddress  fields.Address // 账户1
	LeftAmount   fields.Amount  // 锁定金额
	RightAddress fields.Address // 账户2
	RightAmount  fields.Amount  // 锁定金额

	// 数据指针
	// 所属交易
	// trs block.Transaction
}

func (elm *Action_2_OpenPaymentChannel) Kind() uint16 {
	return 2
}

func (elm *Action_2_OpenPaymentChannel) SetBelongTrs(t block.Transaction) {

}

func (elm *Action_2_OpenPaymentChannel) Size() uint32 {
	return elm.ChannelId.Size() + ((elm.LeftAddress.Size() + elm.LeftAddress.Size()) * 2)
}

func (elm *Action_2_OpenPaymentChannel) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var idBytes, _ = elm.ChannelId.Serialize()
	var addr1Bytes, _ = elm.LeftAddress.Serialize()
	var amt1Bytes, _ = elm.LeftAmount.Serialize()
	var addr2Bytes, _ = elm.RightAddress.Serialize()
	var amt2Bytes, _ = elm.RightAmount.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(idBytes)
	buffer.Write(addr1Bytes)
	buffer.Write(amt1Bytes)
	buffer.Write(addr2Bytes)
	buffer.Write(amt2Bytes)
	return buffer.Bytes(), nil
}

func (elm *Action_2_OpenPaymentChannel) Parse(buf []byte, seek uint32) (uint32, error) {
	seek, _ = elm.ChannelId.Parse(buf, seek)
	seek, _ = elm.LeftAddress.Parse(buf, seek)
	seek, _ = elm.LeftAmount.Parse(buf, seek)
	seek, _ = elm.RightAddress.Parse(buf, seek)
	seek, _ = elm.RightAmount.Parse(buf, seek)
	return seek, nil
}

func (elm *Action_2_OpenPaymentChannel) RequestSignAddrs() [][]byte {
	reqs := make([][]byte, 2)
	reqs[0] = elm.LeftAddress
	reqs[1] = elm.RightAddress
	return reqs
}

func (act *Action_2_OpenPaymentChannel) ChangeChainState(state state.ChainStateOperation) error {
	// 检查金额储存的位数
	labt, _ := act.LeftAmount.Serialize()
	rabt, _ := act.RightAmount.Serialize()
	if len(labt) > 6 || len(rabt) > 6 {
		return fmt.Errorf("Payment Channel create error: left or right Amount bytes too long")
	}
	// 检查余额是否充足
	amt1 := state.Balance(act.LeftAddress)
	if amt1.LessThan(&act.LeftAmount) {
		return fmt.Errorf("Address %s Balance is not enough.", act.LeftAddress.ToReadable())
	}
	amt2 := state.Balance(act.RightAddress)
	if amt2.LessThan(&act.RightAmount) {
		return fmt.Errorf("Address %s Balance is not enough.", act.RightAddress.ToReadable())
	}
	// 查询通道是否存在
	sto := state.Channel(act.ChannelId)
	if sto != nil {
		return fmt.Errorf("Payment Channel Id <%s> already exist.", hex.EncodeToString(act.ChannelId))
	}
	curblk := state.Block().(block.Block)
	// 创建 channel
	var storeItem db.ChannelStoreItemData
	storeItem.BelongHeight = fields.VarInt5(curblk.GetHeight())
	storeItem.LockBlock = fields.VarInt2(uint16(5000))
	storeItem.LeftAddress = act.LeftAddress
	storeItem.LeftAmount = act.LeftAmount
	storeItem.RightAddress = act.RightAddress
	storeItem.RightAmount = act.RightAmount
	// 扣除余额
	DoSubBalanceFromChainState(state, act.LeftAddress, act.LeftAmount)
	DoSubBalanceFromChainState(state, act.RightAddress, act.RightAmount)
	// 储存通道
	state.ChannelCreate(act.ChannelId, &storeItem)
	//
	return nil
}

func (act *Action_2_OpenPaymentChannel) RecoverChainState(state state.ChainStateOperation) error {
	// 删除通道
	state.ChannelDelete(act.ChannelId)
	// 恢复余额
	DoAddBalanceFromChainState(state, act.LeftAddress, act.LeftAmount)
	DoAddBalanceFromChainState(state, act.RightAddress, act.RightAmount)
	return nil
}

/////////////////////////////////////////////////////////////////
