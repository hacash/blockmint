package txpool

import "github.com/hacash/blockmint/types/block"

type txOne struct {
	feePer int64 // 手续费比值， 用于排序 fee/txsize
	tx     block.Transaction
}

type MemTxPool struct {
}
