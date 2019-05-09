package actions

import (
	"bytes"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/state"
)

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

// 增加余额
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

// 扣除余额
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
