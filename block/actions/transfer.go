package actions

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/store"
)

type Action_1_SimpleTransfer struct {
	Address fields.Address
	Amount  fields.Amount
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

func (*Action_1_SimpleTransfer) ChangeChainState(*store.ChainStateDB) {
}

/*************************************************************************/
