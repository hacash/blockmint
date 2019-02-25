package coin

import (
	"bytes"
	"encoding/hex"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/types/block"
	"time"
)

/**
 * 创世区块
 */
func GetGenesisBlock() block.Block {
	genesis := blocks.NewEmptyBlock_v1(nil)
	loc, _ := time.LoadLocation("Asia/Chongqing")
	//fmt.Println(time.Now().In(loc))
	ttt := time.Date(2019, time.February, 4, 11, 25, 0, 0, loc).Unix()
	//fmt.Println( ttt )
	genesis.Timestamp = fields.VarInt5(ttt)
	// coinbase
	addrreadble := "1271438866CSDpJUqrnchoJAiGGBFSQhjd"
	addr, _ := fields.CheckReadableAddress(addrreadble)
	coinbase := transactions.NewTransaction_0_Coinbase()
	coinbase.Address = *addr
	coinbase.Reward = *(BlockCoinBaseReward(uint64(0)))
	coinbase.Message = "hardertodobetter"
	genesis.TransactionCount = 1
	genesis.Transactions = make([]block.Transaction, 1)
	genesis.Transactions[0] = coinbase
	root := blocks.CalculateMrklRoot(genesis.GetTransactions())
	//fmt.Println( hex.EncodeToString(root) )
	genesis.SetMrklRoot(root)
	// check data
	hash := genesis.Hash()
	//fmt.Println( hex.EncodeToString(hash) )
	check_str := "7c256d39d8be6aa35587b687116198da0bad0d2f9d2fa030fd8f1afa080a05b3"
	check, _ := hex.DecodeString(check_str)
	if 0 != bytes.Compare(hash, check) {
		panic("Genesis Block Data Error: need " + check_str + " but give " + hex.EncodeToString(hash))
	}
	return genesis
}
