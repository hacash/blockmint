package block

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
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
	amt4 := amt3.EllipsisDecimalFor23SizeStore()
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

func Test_6(t *testing.T) {

	addr, err := fields.CheckReadableAddress("1DJcykHUKFjMJbJVcmCynQqviDrm373988")
	fmt.Println(err)
	fmt.Println(addr)

}
