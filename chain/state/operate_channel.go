package state

import (
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state/db"
)

func (this *ChainState) Channel(cid fields.Bytes16) interface{} {
	sto, err := this.channelDB.Read(cid)
	if err != nil || sto == nil {
		return nil
	} else {
		return sto
	}
}

func (this *ChainState) ChannelCreate(cid fields.Bytes16, store interface{}) {
	item := store.(*db.ChannelStoreItemData)
	this.channelDB.Save(cid, item)
}

func (this *ChainState) ChannelDelete(cid fields.Bytes16) {
	this.channelDB.Delete(cid)
}
