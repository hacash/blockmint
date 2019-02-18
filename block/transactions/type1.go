package transactions

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/sys/err"
	typesblock "github.com/hacash/blockmint/types/block"

	"golang.org/x/crypto/sha3"
)

type Transaction_1_Simple struct {
	Timestamp fields.VarInt5
	Address   fields.Address
	Fee       fields.Amount

	ActionCount fields.VarInt1
	Actions     []typesblock.Action

	// cache data
	hash []byte
}

func (trs *Transaction_1_Simple) Type() uint8 {
	return 1
}

func (trs *Transaction_1_Simple) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	b1, _ := trs.Timestamp.Serialize()
	b2, _ := trs.Address.Serialize()
	b3, _ := trs.Fee.Serialize()
	b4, _ := trs.ActionCount.Serialize()
	buffer.Write([]byte{trs.Type()}) // type
	buffer.Write(b1)
	buffer.Write(b2)
	buffer.Write(b3)
	buffer.Write(b4)
	for i := 0; i < len(trs.Actions); i++ {
		var bi, _ = trs.Actions[i].Serialize()
		buffer.Write(bi)
	}
	return buffer.Bytes(), nil
}

func (trs *Transaction_1_Simple) Parse(buf []byte, seek uint32) (uint32, error) {
	m1, _ := trs.Timestamp.Parse(buf, seek)
	m2, _ := trs.Address.Parse(buf, m1)
	m3, _ := trs.Fee.Parse(buf, m2)
	m4, _ := trs.ActionCount.Parse(buf, m3)
	iseek := m4
	for i := 0; i < int(trs.ActionCount); i++ {
		var act, sk, err = ParseAction(buf, iseek)
		trs.Actions = append(trs.Actions, act)
		iseek = sk
		if err != nil {
			break
		}
	}
	return iseek, nil
}

func (trs *Transaction_1_Simple) Size() uint32 {
	totalsize := 1 + trs.Timestamp.Size() + trs.Address.Size() + trs.Fee.Size() + trs.ActionCount.Size()
	for i := 0; i < int(trs.ActionCount); i++ {
		totalsize += trs.Actions[i].Size()
	}
	return totalsize
}

// 交易唯一哈希值
func (trs *Transaction_1_Simple) Hash() []byte {
	if trs.hash == nil {
		stuff, _ := trs.Serialize()
		digest := sha3.Sum256(stuff)
		trs.hash = digest[:]
	}
	return trs.hash
}

/* *********************************************************** */

func NewActionByKind(kind uint16) (typesblock.Action, error) {
	switch kind {
	////////////////////   ACTIONS   ////////////////////
	case 1:
		return new(actions.Action_1_SimpleTransfer), nil
		////////////////////    END      ////////////////////
	}
	return nil, err.New("Cannot find Action kind of " + string(kind))
}

func ParseAction(buf []byte, seek uint32) (typesblock.Action, uint32, error) {
	var kind = binary.BigEndian.Uint16(buf[seek : seek+2])
	var act, _ = NewActionByKind(kind)
	var mv, err = act.Parse(buf, seek+2)
	return act, mv, err
}
