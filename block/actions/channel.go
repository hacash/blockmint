package actions

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/state"
)

/**
 * 支付通道交易类型
 */

// 开启支付通道
type Action_2_OpenPaymentChannel struct {
	ChannelId    fields.Bytes16 // 通道id
	Amount       fields.Amount  // 锁定金额
	LeftAddress  fields.Address // 账户1
	RightAddress fields.Address // 账户2

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
	return elm.ChannelId.Size() + elm.Amount.Size() + (elm.LeftAddress.Size() * 2)
}

func (elm *Action_2_OpenPaymentChannel) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var idBytes, _ = elm.ChannelId.Serialize()
	var amtBytes, _ = elm.Amount.Serialize()
	var addr1Bytes, _ = elm.LeftAddress.Serialize()
	var addr2Bytes, _ = elm.RightAddress.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(idBytes)
	buffer.Write(amtBytes)
	buffer.Write(addr1Bytes)
	buffer.Write(addr2Bytes)
	return buffer.Bytes(), nil
}

func (elm *Action_2_OpenPaymentChannel) Parse(buf []byte, seek uint32) (uint32, error) {
	var moveseek1, _ = elm.ChannelId.Parse(buf, seek)
	var moveseek2, _ = elm.Amount.Parse(buf, moveseek1)
	var moveseek3, _ = elm.LeftAddress.Parse(buf, moveseek2)
	var moveseek4, _ = elm.RightAddress.Parse(buf, moveseek3)
	return moveseek4, nil
}

func (elm *Action_2_OpenPaymentChannel) RequestSignAddrs() [][]byte {
	reqs := make([][]byte, 2)
	reqs[0] = elm.LeftAddress
	reqs[1] = elm.RightAddress
	return reqs
}

func (act *Action_2_OpenPaymentChannel) ChangeChainState(state state.ChainStateOperation) error {
	return nil
}

func (act *Action_2_OpenPaymentChannel) RecoverChainState(state state.ChainStateOperation) error {
	return nil
}

/////////////////////////////////////////////////////////////////
