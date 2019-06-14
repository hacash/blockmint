package difficulty

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var (
	version_v2_height = uint64(288 * 160)
)


// 封装的版本   对外接口

func CalculateNextTarget(
	currentBits uint32,
	currentHeight uint64,
	prevTimestamp uint64,
	lastTimestamp uint64,
	eachblocktime uint64,
	changeblocknum uint64,
	printInfo *string,
) ([]byte, *big.Int, uint32) {
	// 使用新版
	if currentHeight >= version_v2_height {
		return DifficultyCalculateNextTarget_v2(currentBits, currentHeight, prevTimestamp, lastTimestamp, eachblocktime, changeblocknum, printInfo)
	}
	// 最旧版
	b1, u1 := CalculateNextTargetDifficulty(currentBits, currentHeight, prevTimestamp, lastTimestamp, eachblocktime, changeblocknum, printInfo)
	return BigToHash256_v1(b1), b1, u1
}


func Uint32ToBig(currentHeight uint64, diff_num uint32) *big.Int {
	// 使用新版
	if currentHeight >= version_v2_height {
		return DifficultyUint32ToBig(diff_num)
	}
	// 最旧版
	return Uint32ToBig_v1(diff_num)
}


func Uint32ToHash(currentHeight uint64, diff_num uint32) []byte {
	// 使用新版
	if currentHeight >= version_v2_height {
		return DifficultyUint32ToHash(diff_num)
	}
	// 最旧版
	return Uint32ToHash256_v1(diff_num)
}

func HashToBig(currentHeight uint64, hash []byte) *big.Int {
	// 使用新版
	if currentHeight >= version_v2_height {
		return DifficultyHashToBig(hash)
	}
	// 最旧版
	return HashToBig_v1(hash)
}

func HashToUint32(currentHeight uint64, hash []byte) uint32 {
	// 使用新版
	if currentHeight >= version_v2_height {
		return DifficultyHashToUint32(hash)
	}
	// 最旧版
	return Hash256ToUint32_v1(hash)
}

func BigToHash(currentHeight uint64, bignum *big.Int) []byte {
	// 使用新版
	if currentHeight >= version_v2_height {
		return DifficultyBigToHash(bignum)
	}
	// 最旧版
	return BigToHash256_v1(bignum)
}





//////////////////////////////////////////////////////////////////////////////////////////










// 计算下一阶段区块难度
func DifficultyCalculateNextTarget_v2(
	currentBits uint32,
	currentHeight uint64,
	prevTimestamp uint64,
	lastTimestamp uint64,
	eachblocktime uint64,
	changeblocknum uint64,
	printInfo *string,
) ([]byte, *big.Int, uint32) {


	powTargetTimespan := time.Second * time.Duration(eachblocktime*changeblocknum) // 一分钟一个快
	// 如果新区块height不是 288 的整数倍，则不需要更新，仍然是最后一个区块的 bits
	if currentHeight%changeblocknum != 0 {
		currentBig := DifficultyUint32ToBig(currentBits)
		return DifficultyBigToHash(currentBig), currentBig, currentBits
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

	lastTarget := DifficultyUint32ToBig(currentBits)
	// 计算公式： target = lastTarget * actualTime / expectTime
	newTarget := lastTarget.Mul(lastTarget, big.NewInt(int64(actualTimespan.Seconds())))
	newTarget = newTarget.Div(newTarget, big.NewInt(int64(powTargetTimespan.Seconds())))

	nextBits := DifficultyBigToUint32(newTarget)
	nextHash := DifficultyBigToHash(newTarget)

	// 打印数据
	if printInfo != nil {
		actual_t, target_t := uint64(actualTimespan.Seconds()), uint64(powTargetTimespan.Seconds())
		nhs := strings.TrimRight( hex.EncodeToString(nextHash), "0")
		printStr := fmt.Sprintf("==== new ==== difficulty calculate next target at height %d ==== %ds/%ds ==== %ds/%ds ==== %d -> %d ==== "+nhs+" ====",
			currentHeight,
			actual_t/changeblocknum,
			target_t/changeblocknum,
			actual_t,
			target_t,
			currentBits,
			nextBits)
		*printInfo = printStr
	}

	return nextHash, newTarget, nextBits
}


func DifficultyUint32ToBig(diff_num uint32) *big.Int {
	hashbyte := DifficultyUint32ToHash(diff_num)
	return DifficultyHashToBig(hashbyte)
}

func DifficultyHashToBig(hashbyte []byte) *big.Int {
	cur_big := new(big.Int).SetBytes(bytes.TrimLeft(hashbyte, string([]byte{0})))
	return cur_big
}

func DifficultyUint32ToHash(diff_num uint32) []byte {

	diff_byte := make([]byte, 4)
	binary.BigEndian.PutUint32(diff_byte, diff_num)

	// 还原
	originally_bits_1 := bytes.Repeat([]byte{0}, 255 - int(diff_byte[0]))
	//fmt.Println("originally_bits_1:", len(originally_bits_1), originally_bits_1)
	originally_bits_2 := BytesToBits([]byte{diff_byte[1],diff_byte[2],diff_byte[3]})
	//fmt.Println("originally_bits_2:", len(originally_bits_2), originally_bits_2)
	originally_yushu :=  256 - len(originally_bits_1) - len(originally_bits_2)
	originally_bits_3 := []byte{}
	if originally_yushu > 0 {
		originally_bits_3 = bytes.Repeat([]byte{0}, originally_yushu)
	}
	originally_bits_bufs := bytes.NewBuffer(originally_bits_1)
	originally_bits_bufs.Write(originally_bits_2)
	originally_bits_bufs.Write(originally_bits_3)
	originally_bits := originally_bits_bufs.Bytes()
	//fmt.Println("originally_bits:", len(originally_bits), originally_bits)
	originally_byte := BitsToBytes(originally_bits)[0:32]
	//fmt.Println("originally_byte:", len(originally_byte), originally_byte)
	return originally_byte
}

func DifficultyBigToHash(diff_big *big.Int) []byte {
	bigbytes := diff_big.Bytes()
	if len(bigbytes) > 32 {
		fmt.Println(len(bigbytes), bigbytes)
		panic("bigbytes length not more than 32.")
	}
	buf := bytes.NewBuffer(bytes.Repeat([]byte{0}, 32-len(bigbytes)))
	buf.Write(bigbytes)
	return buf.Bytes()
}

func DifficultyBigToUint32(diff_big *big.Int) uint32 {
	bighash := DifficultyBigToHash(diff_big)
	return DifficultyHashToUint32(bighash)
}

func DifficultyHashToUint32(hash_byte []byte) uint32 {

	//hash_byte, _ := hex.DecodeString(hash)
	//
	//fmt.Println("\n--------------", hash, "-------------")
	//fmt.Println("           byte:", len(hash_byte), hash_byte)

	// HASH256 转 UINT32
	//fmt.Println(hash_byte)
	hash_bits := BytesToBits(hash_byte)
	//fmt.Println(len(hash_bits), hash_bits)
	headzero := 0
	for _, v := range hash_bits {
		if v!=0{
			break
		}else{
			headzero++
		}
	}
	hash_bits = append(hash_bits, bytes.Repeat([]byte{1}, 3*8 + 12)...)
	//fmt.Println(len(hash_bits), hash_bits)
	//fmt.Println(headzero, headzero+3*8)
	hash_bits_2 := BitsToBytes(hash_bits[headzero: headzero+3*8])
	//fmt.Println(len(hash_bits_2), hash_bits_2)
	//
	diff_byte := make([]byte, 4)
	diff_byte[0] = 255 - uint8(headzero)
	diff_byte[1] = hash_bits_2[0]
	diff_byte[2] = hash_bits_2[1]
	diff_byte[3] = hash_bits_2[2]

	diff_number := binary.BigEndian.Uint32(diff_byte)
	//fmt.Println("diff_number:", diff_number)

	return diff_number
}



