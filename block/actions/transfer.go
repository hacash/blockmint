package actions

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

//////////////////////////////////////////////////////////

func DoSimpleTransferFromChainState(state state.ChainStateOperation, addr1 fields.Address, addr2 fields.Address, amt fields.Amount) error {

	//fmt.Println("addr1:", base58check.Encode(addr1), "addr2:", base58check.Encode(addr2), "amt:", amt.ToFinString())

	if bytes.Compare(addr1, addr2) == 0 {
		return nil // 自己转给自己
	}
	amt1 := state.Balance(addr1)
	//fmt.Println("amt1: " + amt1.ToFinString())
	if amt1.LessThan(&amt) {
		return fmt.Errorf("balance not enough")
	}
	amt2 := state.Balance(addr2)
	//fmt.Println("amt2: " + amt2.ToFinString())
	// add
	amtsub, e1 := amt1.Sub(&amt)
	if e1 != nil {
		//fmt.Println("e1: ", e1)
		return e1
	}
	amtadd, e2 := amt2.Add(&amt)
	if e2 != nil {
		//fmt.Println("e2: ", e2)
		return e2
	}
	//fmt.Println("EllipsisDecimalFor23SizeStore: ")
	amtsub = amtsub.EllipsisDecimalFor23SizeStore()
	amtadd = amtadd.EllipsisDecimalFor23SizeStore()
	/*if amtsub1 != amtsub || amtadd1 != amtadd {
		return fmt.Errorf("amount can not to store")
	}*/
	if amtsub.IsEmpty() {
		state.BalanceDel(addr1) // 归零
	} else {
		//fmt.Println("amtsub: " + amtsub.ToFinString())
		state.BalanceSet(addr1, *amtsub)
	}
	state.BalanceSet(addr2, *amtadd)
	return nil
}

func DoAddBalanceFromChainState(state state.ChainStateOperation, addr fields.Address, amt fields.Amount) error {
	baseamt := state.Balance(addr)
	//fmt.Println( "baseamt: ", baseamt.ToFinString() )
	amtnew, e1 := baseamt.Add(&amt)
	if e1 != nil {
		return e1
	}
	amtsave := *amtnew.EllipsisDecimalFor23SizeStore()
	//addrrr, _ := base58check.Encode(addr)
	//fmt.Println( "DoAddBalanceFromChainState: ++++++++++ ", addrrr, amtsave.ToFinString() )
	state.BalanceSet(addr, amtsave)
	return nil
}

func DoSubBalanceFromChainState(state state.ChainStateOperation, addr fields.Address, amt fields.Amount) error {
	baseamt := state.Balance(addr)
	//fmt.Println("baseamt: " + baseamt.ToFinString())
	if baseamt.LessThan(&amt) {
		return fmt.Errorf("balance not enough")
	}
	//fmt.Println("amt fee: " + amt.ToFinString())
	amtnew, e1 := baseamt.Sub(&amt)
	if e1 != nil {
		return e1
	}
	amtnew1 := *amtnew.EllipsisDecimalFor23SizeStore()
	//fmt.Println("amtnew1: " + amtnew1.ToFinString())
	state.BalanceSet(addr, amtnew1)
	return nil
}

/*************************************************************************/
