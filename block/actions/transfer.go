package actions

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/block"
	"github.com/hacash/blockmint/types/state"
)

type Action_1_SimpleTransfer struct {
	Address fields.Address
	Amount  fields.Amount

	// 数据指针
	// 所属交易
	trs block.Transaction
}

func NewAction_1_SimpleTransfer(addr fields.Address, amt fields.Amount) *Action_1_SimpleTransfer {
	return &Action_1_SimpleTransfer{
		Address: addr,
		Amount:  amt,
	}
}

func (elm *Action_1_SimpleTransfer) Kind() uint16 {
	return 1
}

func (elm *Action_1_SimpleTransfer) Serialize() ([]byte, error) {
	var kindByte = make([]byte, 2)
	binary.BigEndian.PutUint16(kindByte, elm.Kind())
	var addrBytes, _ = elm.Address.Serialize()
	var amtBytes, _ = elm.Amount.Serialize()
	var buffer bytes.Buffer
	buffer.Write(kindByte)
	buffer.Write(addrBytes)
	buffer.Write(amtBytes)
	return buffer.Bytes(), nil
}

func (elm *Action_1_SimpleTransfer) Parse(buf []byte, seek uint32) (uint32, error) {
	var moveseek, _ = elm.Address.Parse(buf, seek)
	var moveseek2, _ = elm.Amount.Parse(buf, moveseek)
	return moveseek2, nil
}

func (elm *Action_1_SimpleTransfer) Size() uint32 {
	return 2 + elm.Address.Size() + elm.Amount.Size()
}

///////////////////////////////////////////////////////////////////////////

func (*Action_1_SimpleTransfer) RequestSignAddrs() [][]byte {
	return make([][]byte, 0) // 无需签名
}

func (act *Action_1_SimpleTransfer) ChangeChainState(state state.ChainStateOperation) error {
	if act.trs == nil {
		panic("Action belong to transaction not be nil !")
	}
	//fmt.Println("address - - - - - - - "+hex.EncodeToString( act.Address))
	//fmt.Println( act.Address[3] )

	//fmt.Println("addr2 - - - - - - - "+hex.EncodeToString( act.Address ))
	//addr2 := make([]byte, 21)
	//copy(addr2, act.Address)
	//fmt.Println("addr2 - - - - - - - "+hex.EncodeToString( addr2 ))

	// 转移
	return DoSimpleTransferFromChainState(state, act.trs.GetAddress(), act.Address, act.Amount)
}

func (act *Action_1_SimpleTransfer) RecoverChainState(state state.ChainStateOperation) error {
	if act.trs == nil {
		panic("Action belong to transaction not be nil !")
	}
	// 回退
	return DoSimpleTransferFromChainState(state, act.Address, act.trs.GetAddress(), act.Amount)
}

// 设置所属 trs
func (act *Action_1_SimpleTransfer) SetBelongTrs(trs block.Transaction) {
	act.trs = trs
}
