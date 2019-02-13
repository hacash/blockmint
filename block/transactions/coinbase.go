package transactions

import (
	"bytes"
	//"fmt"
	"github.com/hacash/blockmint/block/fields"
)

type Transaction_0_Coinbase struct {
	Address fields.Address
	Reward  fields.Amount
	Message fields.TrimString16

	// nonce fields.VarInt8

}

func (trs *Transaction_0_Coinbase) Type() uint8 {
	return 0
}

func (trs *Transaction_0_Coinbase) Serialize() ([]byte, error) {

	var buffer bytes.Buffer
	b1, _ := trs.Address.Serialize()
	b2, _ := trs.Reward.Serialize()
	b3, _ := trs.Message.Serialize()
	// fmt.Println("trs.Message=", trs.Message)
	buffer.Write([]byte{trs.Type()}) // type
	buffer.Write(b1)
	buffer.Write(b2)
	buffer.Write(b3)
	return buffer.Bytes(), nil
}

func (trs *Transaction_0_Coinbase) Parse(buf *[]byte, seek int) (int, error) {

	m1, _ := trs.Address.Parse(buf, seek)
	m2, _ := trs.Reward.Parse(buf, m1)
	m3, _ := trs.Message.Parse(buf, m2)

	return m3, nil
}
