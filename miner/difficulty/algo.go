package difficulty

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"
)


// 计算下一阶段区块难度
func CalculateNextTargetDifficulty(
	currentBits uint32,
	currentHeight uint64,
	prevTimestamp uint64,
	lastTimestamp uint64,
	eachblocktime uint64,
	changeblocknum uint64,
	printInfo *string,
	) (*big.Int, uint32) {

	powTargetTimespan := time.Second * time.Duration(eachblocktime*changeblocknum) // 一分钟一个快
	// 如果新区块height不是 288 的整数倍，则不需要更新，仍然是最后一个区块的 bits
	if currentHeight%changeblocknum != 0 {
		return Uint32ToBig(currentBits), currentBits
	}
	prev2016blockTimestamp := time.Unix(int64(prevTimestamp), 0)
	lastBlockTimestamp := time.Unix(int64(lastTimestamp), 0)
	// 计算 288 个区块出块时间
	actualTimespan := lastBlockTimestamp.Sub(prev2016blockTimestamp)
	if actualTimespan < powTargetTimespan/4 {
		// 如果小于1/4天，则按1/4天计算
		actualTimespan = powTargetTimespan / 4
	} else if actualTimespan > powTargetTimespan*4 {
		// 如果超过4天，则按4天计算
		actualTimespan = powTargetTimespan * 4
	}

	lastTarget := Uint32ToBig(currentBits)
	// 计算公式： target = lastTarget * actualTime / expectTime
	newTarget := lastTarget.Mul(lastTarget, big.NewInt(int64(actualTimespan.Seconds())))
	newTarget = newTarget.Div(newTarget, big.NewInt(int64(powTargetTimespan.Seconds())))

	nextBits := BigToUint32(newTarget)

	if printInfo != nil {
		actual_t, target_t := uint64(actualTimespan.Seconds()), uint64(powTargetTimespan.Seconds())
		printStr := fmt.Sprintf("==== calculate next target difficulty at height %d ==== %ds/%ds ==== %ds/%ds ==== %d -> %d ====",
			currentHeight,
			actual_t/changeblocknum,
			target_t/changeblocknum,
			actual_t,
			target_t,
			currentBits,
			nextBits)
		*printInfo = printStr
	}

	return newTarget, nextBits
}

func Uint32ToBig(number uint32) *big.Int {
	resbytes := Uint32ToHash256(number)
	return new(big.Int).SetBytes(resbytes)
}


func HashToBig(hash []byte) *big.Int {
	return new(big.Int).SetBytes(hash)
}

func Uint32ToHash256(number uint32) []byte {
	resbytes := Uint32ToHash(number)
	results := bytes.Repeat([]byte{0}, 32)
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

func BigToHash256(bignum *big.Int) []byte {
	bigbytes := bignum.Bytes()
	bytes32 := bytes.Repeat([]byte{0}, 32)
	start := 32-len(bigbytes)
	if start < 0 {
		start = 0
	}
	copy(bytes32[start:], bigbytes)
	return bytes32
}

func BigToUint32(bignum *big.Int) uint32 {

	bytes32 := BigToHash256(bignum)
	return Hash256ToUint32( bytes32 )
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








