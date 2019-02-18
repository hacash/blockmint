package transactions

import (
	"bytes"
	//"fmt"
	"github.com/hacash/blockmint/block/fields"

	"golang.org/x/crypto/sha3"
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

func (trs *Transaction_0_Coinbase) Parse(buf []byte, seek uint32) (uint32, error) {

	m1, _ := trs.Address.Parse(buf, seek)
	m2, _ := trs.Reward.Parse(buf, m1)
	m3, _ := trs.Message.Parse(buf, m2)

	return m3, nil
}

func (trs *Transaction_0_Coinbase) Size() uint32 {
	return trs.Address.Size() + trs.Reward.Size() + trs.Message.Size()
}

// 交易唯一哈希值
func (trs *Transaction_0_Coinbase) Hash() []byte {
	stuff, _ := trs.Serialize()
	digest := sha3.Sum256(stuff)
	return digest[:]
}
