package state

import (
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
)

func (this *ChainState) Diamond(diamond fields.Bytes6) interface{} {
	item, _ := this.diamondDB.Read(diamond)
	if item != nil {
		return item
	}
	if this.base == nil {
		return nil
	}
	// 递归查询
	return this.base.Diamond(diamond)
}

func (this *ChainState) DiamondSet(diamond fields.Bytes6, store interface{}) {
	item := store.(*db.DiamondStoreItemData)
	this.diamondDB.Save(diamond, item)
}

func (this *ChainState) DiamondBelong(diamond fields.Bytes6, address fields.Address) {
	item, _ := this.diamondDB.Read(diamond)
	if item != nil {
		return
	}
	item.Address = address
	this.diamondDB.Save(diamond, item)
}

func (this *ChainState) DiamondDel(diamond fields.Bytes6) {
	this.diamondDB.Delete(diamond)
}
