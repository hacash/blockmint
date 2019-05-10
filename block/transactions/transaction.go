package transactions

import (
	"encoding/binary"
	"github.com/hacash/blockmint/block/actions"
	"github.com/hacash/blockmint/sys/err"
	"github.com/hacash/blockmint/types/block"
)

/* *********************************************************** */

func NewActionByKind(kind uint16) (block.Action, error) {
	////////////////////   ACTIONS   ////////////////////
	switch kind {
	case 1:
		return new(actions.Action_1_SimpleTransfer), nil
	case 2:
		return new(actions.Action_2_OpenPaymentChannel), nil
	case 3:
		return new(actions.Action_3_ClosePaymentChannel), nil
	case 4:
		return new(actions.Action_4_DiamondCreate), nil
	case 5:
		return new(actions.Action_5_DiamondTransfer), nil
	}
	////////////////////    END      ////////////////////
	return nil, err.New("Cannot find Action kind of " + string(kind))
}

func ParseAction(buf []byte, seek uint32) (block.Action, uint32, error) {
	var kind = binary.BigEndian.Uint16(buf[seek : seek+2])
	var act, e1 = NewActionByKind(kind)
	if e1 != nil {
		return nil, 0, e1
	}
	var mv, err = act.Parse(buf, seek+2)
	return act, mv, err
}
