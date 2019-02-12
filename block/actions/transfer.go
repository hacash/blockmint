package actions

import (
	"bytes"
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
	//var kindByte = make([]byte, 2)
	//binary.PutUvarint(kindByte, uint64( elm.Kind() ))
	var addrBytes, _ = elm.Address.Serialize()
	var amtBytes, _ = elm.Amount.Serialize()
	var buffer bytes.Buffer
	//buffer.Write(kindByte)
	buffer.Write(addrBytes)
	buffer.Write(amtBytes)
	return buffer.Bytes(), nil
}

func (elm *Action_1_SimpleTransfer) Parse(but *[]byte, seek int) (int, error) {
	var moveseek, _ = elm.Address.Parse(but, seek)
	var moveseek2, _ = elm.Amount.Parse(but, moveseek)
	return moveseek2, nil
}

func (*Action_1_SimpleTransfer) SignatureRequestAddress() *[]string {
	return nil
}

func (*Action_1_SimpleTransfer) ChangeChainState(*store.ChainStateDB) {
}

/*************************************************************************/
