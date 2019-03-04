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

var (
	genesisBlock block.Block = nil
)

/**
 * 创世区块
 */
func GetGenesisBlock() block.Block {
	if genesisBlock != nil {
		return genesisBlock
	}
	genesis := blocks.NewEmptyBlock_v1(nil)
	loc, _ := time.LoadLocation("Asia/Chongqing")
	//fmt.Println(time.Now().In(loc))
	ttt := time.Date(2019, time.February, 4, 11, 25, 0, 0, loc).Unix()
	//fmt.Println( ttt )
	genesis.Timestamp = fields.VarInt5(ttt)
	genesis.Nonce = fields.VarInt4(160117829)
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
	hash := genesis.HashFresh()
	check_hash := "000000077790ba2fcdeaef4a4299d9b667135bac577ce204dee8388f1b97f7e6"
	check, _ := hex.DecodeString(check_hash)
	if 0 != bytes.Compare(hash, check) {
		panic("Genesis Block Hash Error: need " + check_hash + ", but give " + hex.EncodeToString(hash))
	}
	genesisBlock = genesis
	return genesis
}
