package state

import "github.com/hacash/blockmint/block/fields"

func (this *ChainState) Diamond(diamond fields.Bytes6) fields.Address {
	addr, _ := this.diamondDB.Read(diamond)
	if addr != nil {
		return addr
	}
	if this.base == nil {
		return nil
	}
	// 递归查询
	return this.base.Diamond(diamond)
}

func (this *ChainState) DiamondSet(diamond fields.Bytes6, addr fields.Address) {
	this.diamondDB.SetBelong(diamond, addr)
}

func (this *ChainState) DiamondDel(diamond fields.Bytes6) {
	this.diamondDB.Delete(diamond)
}
