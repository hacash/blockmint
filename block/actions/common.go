package actions

import (
	"bytes"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/state"
	"math"
	"math/big"
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

// 2500个区块万分之一的复利计算
func DoAppendCompoundInterest1Of10000By2500Height(amt1 *fields.Amount, amt2 *fields.Amount, insnum uint64) (*fields.Amount, *fields.Amount) {
	if insnum == 0 {
		panic("insnum cannot be 0.")
	}
	if len(amt1.Numeral) > 4 || len(amt2.Numeral) > 4 {
		panic("amount numeral bytes too long.")
	}

	amts := []*fields.Amount{amt1, amt2}
	coinnums := make([]*fields.Amount, 2)

	for i := 0; i < 2; i++ {
		//fmt.Println("----------")
		// amt
		coinnum := new(big.Int).SetBytes(amts[i].Numeral).Uint64()
		//fmt.Println(coinnum)
		mulnum := math.Pow(1.0001, float64(insnum)) * float64(coinnum) * float64(100000000)
		//fmt.Println(mulnum)
		mulnumint := int64(mulnum)
		//fmt.Println(mulnumint)
		newunit := int(amts[i].Unit) - 8
		if newunit < 0 {
			coinnums[i] = amts[i] // 数额极小， 余额不变
			continue
		}
		for {
			if newunit<255 && mulnumint%10 == 0 {
				mulnumint /= 10
				newunit++
			} else {
				break
			}
		}
		newNumeral := big.NewInt(int64(mulnumint)).Bytes()
		//fmt.Println(newNumeral)
		if newunit > 0 && newunit <= 255 {
			newamt := fields.NewAmount(uint8(newunit), newNumeral)
			coinnums[i] = newamt // 正常情况
		}else{
			coinnums[i] = amts[i] // 计算错误， 余额不变
		}
	}

	//fmt.Println("insnum: ", insnum)
	//fmt.Println(amts[0].ToFinString(), " => ", coinnums[0].ToFinString())
	//fmt.Println(amts[1].ToFinString(), " => ", coinnums[1].ToFinString())

	return coinnums[0], coinnums[1]

}
