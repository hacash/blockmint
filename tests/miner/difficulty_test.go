package miner

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/miner/difficulty"
	"math"
	"math/big"
	"os"
	"path"
	"strconv"
	"testing"
)

func Test_d1(t *testing.T) {

	hsxstrary := []string{
		"1b8008d8ed4df1b800b258ed908d8ed4dffd87b87105e0b6e11a5e7f465b741d", // 538673160
		"000908d8ed4df1b800b258ed908d8ed4dffd87b87105e0b6e11a5e7f465b741d", // 520685784
		"00000008ed9081b800b258ed908d8ed4dffd87b87105e0b6e11a5e7f465b741d", // 487124368
		"0000000008d8e1b800b258ed908d8ed4dffd87b87105e0b6e11a5e7f465b741d", // 470341857
		"0000000000000000000258ed908d8ed4dffd87b87105e0b6e11a5e7f465b741d", // 386029805
		"00000000000000000000000000008ed4dffd87b87105e0b6e11a5e7f465b741d", // 318803668
		"00000000000000000000000000000000000000b87105e0b6e11a5e7f465b741d", // 234928241
		"0000000000000000000000000000000000000000000000b6e11a5e7f465b741d", // 167818977
		"0000000000000000000000000000000000000000000000000000007f465b741d", // 92227163
		"000000000000000000000000000000000000000000000000000000000000041d", // 33824000
		"0000000000000000000000000000000000000000000000000000000000000001", // 16842752
	}

	for i := 0; i < len(hsxstrary); i++ {
		hash1, _ := hex.DecodeString(hsxstrary[i])
		//bignum := difficulty.HashToBig(&hash1)
		//num32 := difficulty.BigToCompact(bignum)
		//fmt.Println(hsxstrary[i]+" difficulty value is", num32)

		val := Hash256ToUint32(hash1)
		fmt.Println(hsxstrary[i]+" difficulty value is", val)


	}

}

func Test_d2(t *testing.T) {

	//difficulty.CalcNextRequiredDifficulty(difficulty.LowestCompact, 2016*5*59)

	//fmt.Println(difficulty.MaxPowLimit)
	//fmt.Println(difficulty.BigToCompact(difficulty.MaxPowLimit))

	compact := uint32(538673160)

	fmt.Println(compact)

	target := difficulty.CalculateNextWorkTarget(
		compact,
		2016,
		1549704600, /*2019/2/9 17:30:00*/
		1550050200, /*2019/2/13 17:30:00*/
		15,
		10,
		nil,
	)

	fmt.Println(target)

}

func Test_d3(t *testing.T) {

	tar := path.Join(os.Getenv("HOME"), "~/.hacash", "blocks")

	fmt.Println(tar)

}



func Test_t4(t *testing.T) {

	bits, _ := strconv.ParseUint("172e6117", 16, 0)

	hash, _ := hex.DecodeString("00000000000000000028197fec7637dd3ff7a0e2b1efa7917772b8c94a37acb2")
	hash2, _ := hex.DecodeString("00000078ee442ce4177cf835217b6e9e4f55990467f5e1575868385c64d20c43")

	hash3, _ := hex.DecodeString("000000009b7262315dbf071787ad3656097b892abffd1f95a1a022f896f533fc")
	bits3, _ := strconv.ParseUint("1d00ffff", 16, 0)


	cccbits := difficulty.BigToCompact( difficulty.HashToBig(&hash) )
	cccbits2 := difficulty.BigToCompact( difficulty.HashToBig(&hash2) )
	cccbits3 := difficulty.BigToCompact( difficulty.HashToBig(&hash3) )


	fmt.Println(bits)
	fmt.Println(cccbits)
	fmt.Println(cccbits2)

	fmt.Println(cccbits3)
	fmt.Println(bits3)


	fmt.Println("----------")


	hhhh := difficulty.BigToHash( difficulty.CompactToBig(483787182) )
	bbbb := make([]byte, 32)
	copy(bbbb[32-len(hhhh):], hhhh)

	fmt.Println("00000000000000000028197fec7637dd3ff7a0e2b1efa7917772b8c94a37acb2")
	fmt.Println(hex.EncodeToString(bbbb))



}


func Test_t5(t *testing.T) {

	//hash2, _ := hex.DecodeString("00000000ee442ce4177cf835217b6e9e4f55990467f5e1575868385c64d20c43")
	//cccbits2 := difficulty.BigToCompact( difficulty.HashToBig(&hash2) )
	//fmt.Println(cccbits2)

	//maxlagestr := "FFFFFF0000000000000000000000000000000000000000000000000000000000"
	//maxlagehash, _ := hex.DecodeString(maxlagestr)
	//diffval := Hash256ToUint32(maxlagehash)
	//fmt.Println(diffval)



	basestr := "FFFFFF" // 0000000000000000000000000000000000000000000000000000000000
	basehash, _ := hex.DecodeString(basestr)
	maxbig := new(big.Int).SetBytes(basehash)
	stepbig := new(big.Int).SetUint64(256)
	fmt.Println( maxbig.String() )

	for i:=0; i<1000; i++ {

		maxbig = maxbig.Add(maxbig, stepbig)

		fmt.Println( maxbig.String() )



	}



	step := 1000*1
	basebig := new(big.Int).SetUint64(256*256*256)
	total := int64(256)*256*256*256 - 1
	total2 := int64(3892314111)

	fmt.Println(step, total, total2, stepbig.String(), basebig.String())

	for i:=total; i>0; i-=256*256*256 {
		break
		step --
		if step == 0 {
			break
		}
		//
		//fmt.Println( hex.EncodeToString(reverse(basebig.Bytes())), basebig.String() )
		//basebig = basebig.Add(basebig, stepbig)



		targethash := Uint32ToHash(uint32(i))
		hexbts := hex.EncodeToString( targethash )
		fmt.Printf("%-64s %-10d %s \n", hexbts, i, new(big.Int).SetBytes(reverse(targethash)).String())




	}







	//for i:=uint64(0); i<uint64(math.Pow(8, 63)); i+=99 {
	//	bts := new(big.Int).SetUint64(i).Bytes()
	//	fmt.Println(bts)
	//}




}






func Test_t6(t *testing.T) {

	maxbighash, _ := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")

	maxbig := new(big.Int).SetBytes(maxbighash)
	spcbig := new(big.Int).SetUint64(256*256*256*256)
	stepbig := maxbig.Div(maxbig, spcbig)
	//stepbig := new(big.Int).SetUint64(919)

	fmt.Println(stepbig.Bytes())


	for i:=0; i<1000; i++ {

		//maxbig = maxbig.Add(maxbig, stepbig)
		//bigbts := maxbig.Bytes()
		//fmt.Println( bigbts, hex.EncodeToString(bigbts), maxbig.String(), Hash256ToUint32(bigbts) )



	}

}




func Test_t7(t *testing.T) {
	spcbig := new(big.Int).SetUint64(256*256*256*256)
	fmt.Println(spcbig.String())

	for i:=0; i<256; i++ {

		base := math.Pow(1.1, float64(i))

		fmt.Printf("%3d %11f \n", i, base)

	}


}




func Test_t8(t *testing.T) {

	total := int64(256)*256*256*256 - 1
	step := 1000

	//zero := new(big.Int).SetUint64(0)
	//prevbignum := zero

	for i:=total; i>0; i-=19999999 {


		hash := Uint32ToHash256(uint32(i))

		bignum := new(big.Int).SetBytes(hash)
		//stepnum := new(big.Int).Div(prevbignum, bignum)

		fmt.Printf( "%s %11d %s %s \n", hex.EncodeToString(hash), i, bignum, "" )
		//prevbignum = bignum

		if step--; step==0 {
			break
		}
	}
}


func Test_t9(t *testing.T) {

	var oldbig *big.Int = nil
	var oldnum = uint32(0)

	for i:=0; i<200; i++ {
		fmt.Println("-----------------------------------")
		for k:=256*256*256-1; k>0; k-=10000*200 {
			rnumbt := make([]byte, 4)
			binary.BigEndian.PutUint32(rnumbt, uint32(k))
			onebits := BytesToBits(rnumbt[1:])
			//fmt.Println(rnumbt[1:], onebits)
			bits := bytes.Repeat([]byte{0}, i)
			bits = append(bits, 1)
			onebits = append(onebits, bytes.Repeat([]byte{1}, 7)...)
			bits = append(bits, onebits...)
			//fmt.Println(bits)
			bytes := BitsToBytes(bits)
			bytes32 := make([]byte, 32)
			copy(bytes32, bytes)
			//fmt.Println(hex.EncodeToString(bytes))
			bignum := new(big.Int).SetBytes(bytes32)
			numval := Hash256ToUint32(bytes32)
			bei := ""
			val := uint32(0)
			if oldbig != nil {
				bei = new(big.Int).Div(oldbig, new(big.Int).Div(bignum, new(big.Int).SetUint64(10000))).String()
				val = oldnum - numval
				//bei = new(big.Int).Sub(oldbig, bignum).String()
			}
			if false {
				fmt.Println(oldbig)
			}
			if true {
				fmt.Printf( "%s %12d %10d %s %s  \n", hex.EncodeToString(bytes32), numval, val, bignum, bei)
			}
			oldbig = bignum
			oldnum = numval
		}
	}



}


func Test_t10(t *testing.T) {

	maxbighash, _  := hex.DecodeString("FFFFFF0000000000000000000000000000000000000000000000000000000000")
	maxbig := new(big.Int).SetBytes(maxbighash)

	for i:=0; i<255; i++ {
		bytes := maxbig.Bytes()
		bytes32 := make([]byte, 32)
		start := 32-len(bytes)
		copy(bytes32[start:], bytes)
		fmt.Printf("%3d %32s %d\n", start, hex.EncodeToString(bytes32), Hash256ToUint32(bytes32))

		maxbig = maxbig.Div(maxbig, new(big.Int).SetUint64(2))
	}


}



func Test_t11(t *testing.T) {

	var oldbig *big.Int = nil

	step := 1000
	// int64(256)*256*256*256-1
	for i:=int64(3892314111); i>0; i-=256*256*251 {
		if step--; step==0 {
			break
		}
		hash := Uint32ToHash256(uint32(i))
		bignum := new(big.Int).SetBytes(hash)
		bei := ""
		if oldbig != nil {
			cu := new(big.Int).Div(bignum, new(big.Int).SetUint64(100))
			bei = new(big.Int).Div(oldbig, cu).String()
		}

		fmt.Printf("%12d %s %6s %s \n", i, hex.EncodeToString(hash), bei, bignum)

		oldbig = bignum
	}



}


func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}



//
func Uint32ToHash256(number uint32) []byte {
	resbytes := Uint32ToHash(number)
	results := make([]byte, 32)
	copy(results, resbytes)
	return results
}

//
func Uint32ToHash(number uint32) []byte {
	numbts := make([]byte, 4)
	binary.BigEndian.PutUint32(numbts, number)
	//fmt.Println(numbts)
	headzero := 255 - numbts[0]
	bitary := bytes.Repeat([]byte{0}, int(headzero) )
	bitary = append(bitary, 1)
	bitary = append(bitary, BytesToBits(numbts[1:])... )
	resbytes := BitsToBytes(bitary)
	//fmt.Println(bitary)
	return resbytes
}

//
func Hash256ToUint32(hash []byte) uint32 {
	if len(hash) != 32 {
		panic(fmt.Sprintf("hash length not be %d", len(hash)))
	}
	headzero := uint8(0)
	bits := BytesToBits(hash)
	for k, v := range bits {
		if v != 0 {
			headzero = uint8(k)
			break
		}
	}
	rightcut := uint8(255) - 8*3
	if headzero > rightcut {
		headzero = rightcut
	}
	valbits := bits[headzero:headzero+8*3]
	valbytes := BitsToBytes(valbits)
	results := make([]byte, 0, 4)
	results = append(results, 255 - headzero)
	results = append(results, valbytes...)
	return binary.BigEndian.Uint32(results)
}

// 256进制变2进制
func BitsToBytes(bits []byte) []byte {
	retults := make([]byte, 0, len(bits)/8)
	for i:=0; i<len(bits)/8; i++ {
		retults = append(retults, BitsToByte(bits[i*8:i*8+8]))
	}
	return retults
}


// 256进制变2进制
func BytesToBits(stuff []byte) []byte {
	results := make([]byte, 0, 32*8)
	for _, v := range stuff {
		results = append(results, ByteToBits(v)...)
	}
	return results
}



// 256进制变2进制
func BitsToByte(bits []byte) byte {
	b := uint8(0)
	b += uint8(1) * bits[7]
	b += uint8(2) * bits[6]
	b += uint8(4) * bits[5]
	b += uint8(8) * bits[4]
	b += uint8(16) * bits[3]
	b += uint8(32) * bits[2]
	b += uint8(64) * bits[1]
	b += uint8(128) * bits[0]
	return b
}



func ByteToBits(b byte) []byte {
	return []byte{
		(byte) ((b >> 7) & 0x1),
		(byte) ((b >> 6) & 0x1),
		(byte) ((b >> 5) & 0x1),
		(byte) ((b >> 4) & 0x1),
		(byte) ((b >> 3) & 0x1),
		(byte) ((b >> 2) & 0x1),
		(byte) ((b >> 1) & 0x1),
		(byte) ((b >> 0) & 0x1),
	}
}