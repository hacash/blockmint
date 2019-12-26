package block

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/chain/state"
	"github.com/hacash/blockmint/core/account"
	"math/big"
	"testing"
)

func Test_5(t *testing.T) {

	/*

		hashone7, _ := hex.DecodeString("12a1633cafcc01ebfb6d78e39f687a1f0995c62fc95f51ead10a02ee0be551b5dc")
		fmt.Println( len(hashone7) )


		amt1 :=  fields.NewAmountSmall(1, 248).GetValue()
		fmt.Println( amt1.String() )
		fmt.Println( amt1.Bytes() )
		fmt.Println( len(amt1.Bytes()) )

		var amt2 big.Int
		amt2.SetString("12345678901234567890123456789012345678901234567890", 10)
		fmt.Println( amt2.String() )
		fmt.Println( amt2.Bytes() )
		fmt.Println( len(amt2.Bytes()) )

	*/

	/*

		bgstr := "12a1633cafcc01ebfb6d78e39f687a1f0995c62fc95f51ead10a02ee0be551b5dce551b5dce551b5d1b5d12a1633cafcc01ebfb6d78e39f687a1f0995c62fc95f51ead10a02ee0be551b5dce551b5dce551b5d1b5d12a1633cafcc01ebfb6d78e39f687a1f0995c62fc95f51ead10a02ee0be551b5dce551b5dce551b5d1b5d"
		bbb, ok := new(big.Int).SetString(bgstr,16)
		if !ok {
			panic("big too big")
		}
		fmt.Println( bbb.String() )
		fmt.Println( len(bbb.String()) )

		hashone9, _ := hex.DecodeString(bgstr)
		fmt.Println( hashone9 )
		fmt.Println( len(hashone9) )

		//bgamt := fields.NewAmount()

	*/

	/*

		numstr := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
		fmt.Println(numstr)
		fmt.Println( len(numstr) )
		bignum, _ := new(big.Int).SetString(numstr, 10)
		amt := fields.NewAmountByBigInt(bignum)

		fmt.Println( amt.GetValue().String() )

		fmt.Println( amt.Unit )
		fmt.Println( amt.Dist )
		fmt.Println( amt.Numeral )
		fmt.Println( len(amt.Numeral) )


		amt2 := fields.NewAmountSmall(1, 0)

		amt3 := amt.Sub(amt2)

		fmt.Println( amt3.GetValue().String() )

		amt4 := amt3.EllipsisDecimalFor23SizeStore()

		fmt.Println( amt4.GetValue().String() )


	*/

	hashone8, _ := hex.DecodeString("12a1633cafcc01ebfb6d78e39f687a1f0995c62fc95f51ead10a02ee0be551b5dce551b5dce551b5d1b5d")
	fmt.Println(len(hashone8))
	amt3 := fields.NewAmount(0, hashone8)
	amt4, _ := amt3.EllipsisDecimalFor23SizeStore()
	longnum3 := new(big.Int).SetBytes(amt3.Numeral).String()
	longnum4 := new(big.Int).SetBytes(amt4.Numeral).String()
	fmt.Println(longnum3)
	fmt.Println(longnum4)
	fmt.Println(len(longnum3), len(longnum4))
	fmt.Println(amt3.GetValue().String())
	fmt.Println(amt4.GetValue().String())
	fmt.Println(amt3.Numeral)
	fmt.Println(amt4.Numeral)
	fmt.Println(len(amt3.Numeral), len(amt4.Numeral))
	fmt.Println(amt3.Dist, amt4.Dist)
	fmt.Println(amt3.Unit, amt4.Unit)

	fmt.Println(len("fc916f213a3d7f1369313d5fa30f6168f9446a2d"))

}

func Test_loop_nice(t *testing.T) {

	fmt.Println([]byte("ㄜ"))

}

func Test_7(t *testing.T) {

	one := fields.NewAmountNumSmallCoin(1)
	base := fields.NewAmountNumSmallCoin(1)

	var e error
	for i := 2; i < 1000000000; i++ {
		base, e = base.Add(one)
		if e != nil {
			panic(e)
		}
		fmt.Println(i, base.ToFinString())
	}

}

func Test_8(t *testing.T) {

	acc1 := account.CreateAccountByPassword("121")
	acc2 := account.CreateAccountByPassword("122")
	acc3 := account.CreateAccountByPassword("123")
	acc4 := account.CreateAccountByPassword("124")

	oneamt := fields.NewAmountNumSmallCoin(1)
	baseamt := fields.NewAmountNumSmallCoin(0)

	basestate := state.NewTempChainState(nil)
	defer basestate.Destroy() // 清理垃圾

	for i := uint64(0); ; i++ {
		tmpstate := state.NewTempChainState(basestate)
		bbamt1 := tmpstate.Balance(acc1.Address)
		bbamt2 := tmpstate.Balance(acc2.Address)
		bbamt3 := tmpstate.Balance(acc3.Address)
		bbamt4 := tmpstate.Balance(acc4.Address)
		if !bbamt1.Equal(baseamt) ||
			!bbamt2.Equal(baseamt) ||
			!bbamt3.Equal(baseamt) ||
			!bbamt4.Equal(baseamt) {
			fmt.Println(baseamt.ToFinString() + " but get " + bbamt1.ToFinString())
			panic("")
		}
		if i%1000 == 0 {
			fmt.Println(i, baseamt.ToFinString())
		}
		baseamt, _ = baseamt.Add(oneamt)
		tmpstate.BalanceSet(acc1.Address, *baseamt)
		tmpstate.BalanceSet(acc2.Address, *baseamt)
		tmpstate.BalanceSet(acc3.Address, *baseamt)
		tmpstate.BalanceSet(acc4.Address, *baseamt)
		basestate.TraversalCopy(tmpstate)
		tmpstate.Destroy() // 清理垃圾
	}

}
