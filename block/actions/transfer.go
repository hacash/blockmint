package actions

import (
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/types/store"
)

type Action_1_SimpleTransfer struct {
	Address fields.Address
	Amount  fields.Amount
}

func (*Action_1_SimpleTransfer) Kind() uint16 {
	return 1
}

func (elm *Action_1_SimpleTransfer) Serialize() ([]byte, error) {
	return nil, nil
}
func (elm *Action_1_SimpleTransfer) Parse(but *[]byte, seek int) (int, error) {
	var moveseek, _ = elm.Address.Parse(but, seek)
	var moveseek2, _ = elm.Address.Parse(but, moveseek)
	return moveseek2, nil
}

func (*Action_1_SimpleTransfer) SignatureRequestAddress() *[]string {
	return nil
}

func (*Action_1_SimpleTransfer) ChangeChainState(*store.ChainStateDB) {
}

/*************************************************************************/
