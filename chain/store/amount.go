package store

import "math/big"

type StoreData struct {
	LockHeight uint32 // 锁定区块高度

	Amount big.Int // 可用余额

}
