package actions

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/state"
)

/**
 * 钻石交易类型
 */

// 挖出钻石
type Action_3_DiamondCreate struct {
	Diamond fields.Bytes6  // 钻石字面量 WTYUIAHXVMEKBSZN
	Address fields.Address // 所属账户
	Nonce   fields.VarInt8 // 随机数

	// 数据指针
	// 所属交易
	// trs block.Transaction
}

func (elm *Action_3_DiamondCreate) Kind() uint16 {
	return 3
}

func (elm *Action_3_DiamondCreate) SetBelongTrs(t block.Transaction) {

}

func (elm *Action_3_DiamondCreate) Size() uint32 {
	return elm.Diamond.Size() + elm.Address.Size() + elm.Nonce.Size()
}

func (elm *Action_3_DiamondCreate) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var diamondBytes, _ = elm.Diamond.Serialize()
	var addrBytes, _ = elm.Address.Serialize()
	var nonceBytes, _ = elm.Nonce.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(diamondBytes)
	buffer.Write(addrBytes)
	buffer.Write(nonceBytes)
	return buffer.Bytes(), nil
}

func (elm *Action_3_DiamondCreate) Parse(buf []byte, seek uint32) (uint32, error) {
	var moveseek1, _ = elm.Diamond.Parse(buf, seek)
	var moveseek2, _ = elm.Address.Parse(buf, moveseek1)
	var moveseek3, _ = elm.Nonce.Parse(buf, moveseek2)
	return moveseek3, nil
}

func (elm *Action_3_DiamondCreate) RequestSignAddrs() [][]byte {
	return make([][]byte, 0) // 无需签名
}

func (act *Action_3_DiamondCreate) ChangeChainState(state state.ChainStateOperation) error {
	return nil
}

func (act *Action_3_DiamondCreate) RecoverChainState(state state.ChainStateOperation) error {
	return nil
}

///////////////////////////////////////////////////////////////

// 转移钻石
type Action_4_DiamondTransfer struct {
	Diamond fields.Bytes6  // 钻石字面量 WTYUIAHXVMEKBSZN
	Address fields.Address // 收钻方账户

	// 数据指针
	// 所属交易
	trs block.Transaction
}

func (elm *Action_4_DiamondTransfer) Kind() uint16 {
	return 3
}

func (elm *Action_4_DiamondTransfer) SetBelongTrs(t block.Transaction) {
	elm.trs = t
}

func (elm *Action_4_DiamondTransfer) Size() uint32 {
	return elm.Diamond.Size() + elm.Address.Size()
}

func (elm *Action_4_DiamondTransfer) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var diamondBytes, _ = elm.Diamond.Serialize()
	var addrBytes, _ = elm.Address.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(diamondBytes)
	buffer.Write(addrBytes)
	return buffer.Bytes(), nil
}

func (elm *Action_4_DiamondTransfer) Parse(buf []byte, seek uint32) (uint32, error) {
	var moveseek1, _ = elm.Diamond.Parse(buf, seek)
	var moveseek2, _ = elm.Address.Parse(buf, moveseek1)
	return moveseek2, nil
}

func (elm *Action_4_DiamondTransfer) RequestSignAddrs() [][]byte {
	return make([][]byte, 0) // 无需签名
}

func (act *Action_4_DiamondTransfer) ChangeChainState(state state.ChainStateOperation) error {
	return nil
}

func (act *Action_4_DiamondTransfer) RecoverChainState(state state.ChainStateOperation) error {
	return nil
}
