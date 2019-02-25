package state

import "github.com/hacash/blockmint/block/fields"

// chain state 操作

type ChainStateOperation interface {
	// query
	Balance(fields.Address) fields.Amount

	// operate
	BalanceSet(fields.Address, fields.Amount)
	BalanceDel(fields.Address)
}
