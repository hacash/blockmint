package coin

import (
	"bytes"
	"encoding/hex"
	"github.com/hacash/blockmint/block/blocks"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/block/transactions"
	"github.com/hacash/blockmint/types/block"
	"strings"
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
	/*for i:=0; i<1000; i++ {
		hash := genesis.HashFresh()
		fmt.Printf("%d %s \n", i, hex.EncodeToString(hash) )
		time.Sleep(time.Duration(100) * time.Millisecond)
	}*/
	hash := genesis.HashFresh()
	bodybytes, e := genesis.Serialize()
	if e != nil {
		panic(e)
	}
	bdbts := hex.EncodeToString(bodybytes)
	check_bdbts := "010000000000005c57b08c0000000000000000000000000000000000000000000000000000000000000000ad557702fc70afaf70a855e7b8a4400159643cb5a7fc8a89ba2bce6f818a9b01000000010000000000000000000000000c1aaa4e6007cc58cfb932052ac0ec25ca356183f80101686172646572746f646f62657474657200"
	//fmt.Println( bdbts )
	//fmt.Println( check_bdbts )
	if 0 != strings.Compare(bdbts, check_bdbts) {
		panic("Genesis Block Data Error: need " + check_bdbts + ", but give " + bdbts)
	}
	check_hash := "57cef097f9a7cc0c45bcac6325b5b6e58199c8197763734cac6664e8d2b8e63e"
	check, _ := hex.DecodeString(check_hash)
	if 0 != bytes.Compare(hash, check) {
		panic("Genesis Block Hash Error: need " + check_hash + ", but give " + hex.EncodeToString(hash))
	}
	return genesis
}
