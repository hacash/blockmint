package state

import (
	"github.com/hacash/blockmint/block/fields"
)

////////////////////  query  /////////////////////

func (this *ChainState) Balance(addr fields.Address) fields.Amount {
	blc, _ := this.balanceDB.Read(addr)
	if blc != nil {
		return blc.Amount
	}
	if this.base == nil {
		return *fields.NewEmptyAmount()
	}
	// 递归查询
	return this.base.Balance(addr)
}

////////////////////  operate  /////////////////////

func (this *ChainState) BalanceDel(addr fields.Address) {
	this.balanceDB.Remove(addr)
}

func (this *ChainState) BalanceSet(addr fields.Address, amt fields.Amount) {
	//fmt.Println("BalanceSet address " + hex.EncodeToString(addr))
	blc, e := this.balanceDB.Read(addr)
	if e != nil || blc == nil || blc.Locitem == nil {
		this.balanceDB.SaveAmountByClearCreate(addr, amt)
		return // new insert
	}
	//fmt.Println("BalanceSet address ", blc.LockHeight)
	blc.Amount = amt
	this.balanceDB.Save(addr, blc)
}
