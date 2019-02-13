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

func (elm *Action_1_SimpleTransfer) Parse(buf *[]byte, seek int) (int, error) {
	var moveseek, _ = elm.Address.Parse(buf, seek)
	var moveseek2, _ = elm.Amount.Parse(buf, moveseek)
	return moveseek2, nil
}

func (*Action_1_SimpleTransfer) SignatureRequestAddress() []string {
	return nil
}

func (*Action_1_SimpleTransfer) ChangeChainState(*store.ChainStateDB) {
}

/*************************************************************************/
