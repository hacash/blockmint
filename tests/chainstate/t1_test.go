package chainstate

import (
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	store2 "github.com/hacash/blockmint/chain/store"
	"testing"
)

func Test_state1(t *testing.T) {

	var testdir = "/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/tests/chainstate/datas"
	//os.Remove(testdir)

	var db1 store2.ChainStateBalanceDB
	db1.Init(testdir)

	address := []byte("address0000000address")

	db1.SaveAmountByClearCreate(address, fields.NewAmountSmall(1, 248))

	myamt, _ := db1.Read(address)

	fmt.Println(myamt.Amount.ToAccountingString())

}

func Test_state2(t *testing.T) {

	var testdir = "/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/tests/chainstate/datas"
	//os.Remove(testdir)

	var db1 store2.ChainStateBalanceDB
	db1.Init(testdir)

	address1 := []byte("add1ess000000address1")
	address2 := []byte("add2ess000000address2")
	address3 := []byte("add3ess000000address3")
	address4 := []byte("add4ess000000address4")
	address5 := []byte("add5ess000000address5")

	db1.SaveAmountByClearCreate(address1, fields.NewAmountSmall(1, 248))
	db1.SaveAmountByClearCreate(address2, fields.NewAmountSmall(2, 248))
	db1.SaveAmountByClearCreate(address3, fields.NewAmountSmall(3, 248))
	db1.SaveAmountByClearCreate(address4, fields.NewAmountSmall(4, 248))
	db1.SaveAmountByClearCreate(address5, fields.NewAmountSmall(5, 248))

	myamt1, _ := db1.Read(address1)
	fmt.Println(myamt1.Amount.ToAccountingString())
	myamt2, _ := db1.Read(address2)
	fmt.Println(myamt2.Amount.ToAccountingString())
	myamt3, _ := db1.Read(address3)
	fmt.Println(myamt3.Amount.ToAccountingString())
	myamt4, _ := db1.Read(address4)
	fmt.Println(myamt4.Amount.ToAccountingString())
	myamt5, _ := db1.Read(address5)
	fmt.Println(myamt5.Amount.ToAccountingString())

	// DELETE
	//db1.Remove(address1)
	//db1.Remove(address2)
	//db1.Remove(address3)
	//db1.Remove(address4)
	//db1.Remove(address5)

}
