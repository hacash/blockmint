package difficulty

import (
	"fmt"
	"math/big"
	"time"
)

var (
	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh256 = new(big.Int).Lsh(bigOne, 256)

	//
	LowestCompact = uint32(521000000) // 508000000  521000000
)

// HashToBig converts a chainhash.Hash into a big.Int that can be used to
// perform math comparisons.
func HashToBig(hash *[]byte) *big.Int {
	// A Hash is in big-endian
	buf := *hash
	//blen := len(buf)
	//for i := 0; i < blen/2; i++ {
	//	buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	//}

	return new(big.Int).SetBytes(buf[:])
}

// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number.  The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number.  See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	// No need to do any work if it's zero.
	if n.Sign() == 0 {
		return 0
	}

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes.  So, shift the number right or left
	// accordingly.  This is equivalent to:
	// mantissa = mantissa / 256^(exponent-3)
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		// Use a copy to avoid modifying the caller's original number.
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	// When the mantissa already has the sign bit set, the number is too
	// large to fit into the available 23-bits, so divide the number by 256
	// and increment the exponent accordingly.
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	// Pack the exponent, sign bit, and mantissa into an unsigned 32-bit
	// int and return it.
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

// This compact form is only used in bitcoin to encode unsigned 256-bit numbers
// which represent difficulty targets, thus there really is not a need for a
// sign bit, but it is implemented here to stay consistent with bitcoind.
func CompactToBig(compact uint32) *big.Int {
	// Extract the mantissa, sign bit, and exponent.
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	// Since the base for the exponent is 256, the exponent can be treated
	// as the number of bytes to represent the full 256-bit number.  So,
	// treat the exponent as the number of bytes and shift the mantissa
	// right or left accordingly.  This is equivalent to:
	// N = mantissa * 256^(exponent-3)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	// Make it negative if the sign bit is set.
	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

// result adds 1 to the denominator and multiplies the numerator by 2^256.
func CalcWork(bits uint32) *big.Int {
	// Return a work value of zero if the passed difficulty bits represent
	// a negative number. Note this should not happen in practice with valid
	// blocks, but an invalid block could trigger it.
	difficultyNum := CompactToBig(bits)
	if difficultyNum.Sign() <= 0 {
		return big.NewInt(0)
	}

	// (1 << 256) / (difficultyNum + 1)
	denominator := new(big.Int).Add(difficultyNum, bigOne)
	return new(big.Int).Div(oneLsh256, denominator)
}

func CalcNextRequiredDifficulty(prevCompact uint32, totalSecond uint32) uint32 {

	fmt.Println(totalSecond)
	fmt.Println(prevCompact)

	targetCompact := prevCompact * totalSecond / 2016 * 5 * 60

	fmt.Println(targetCompact)

	/*




		work := CalcWork(LowestCompact)

		a := 1.0
		b := float64(2016*5*60 / totalSecond)

		a1 := a / b
		fmt.Println(a1)
		fmt.Println(work.Uint64())

		targetWork := float64(work.Uint64()) / a1

		fmt.Println(targetWork)


	*/

	return LowestCompact
}

var (
	bigOneValue = big.NewInt(1)
	// 最大难度：00000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffff，2^224，0x1d00ffff
	MaxPowLimit = new(big.Int).Sub(new(big.Int).Lsh(bigOneValue, 224), bigOneValue)
	//powTargetTimespan =  // time.Hour * 24 * 7 // 一周
)

// 计算下一阶段区块难度
func CalculateNextWorkTarget(currentBits uint32, currentHeight uint64, prevTimestamp uint64, lastTimestamp uint64, eachblocktime uint64, changeblocknum uint64, printInfo *string) uint32 {

	powTargetTimespan := time.Second * time.Duration(eachblocktime*changeblocknum) // 一分钟一个快
	// 如果新区块height不是 288 的整数倍，则不需要更新，仍然是最后一个区块的 bits
	if currentHeight%changeblocknum != 0 {
		return currentBits
	}
	prev2016blockTimestamp := time.Unix(int64(prevTimestamp), 0)
	lastBlockTimestamp := time.Unix(int64(lastTimestamp), 0)
	// 计算 288 个区块出块时间
	actualTimespan := lastBlockTimestamp.Sub(prev2016blockTimestamp)
	if actualTimespan < powTargetTimespan/8 {
		// 如果小于1/8天，则按1/8天计算
		actualTimespan = powTargetTimespan / 8
	} else if actualTimespan > powTargetTimespan*8 {
		// 如果超过8天，则按8天计算
		actualTimespan = powTargetTimespan * 8
	}

	lastTarget := CompactToBig(currentBits)
	// 计算公式： target = lastTarget * actualTime / expectTime
	newTarget := lastTarget.Mul(lastTarget, big.NewInt(int64(actualTimespan.Seconds())))
	newTarget = newTarget.Div(newTarget, big.NewInt(int64(powTargetTimespan.Seconds())))
	//超过最多难度，则重置
	//if newTarget.Cmp(MaxPowLimit) > 0 {
	//	newTarget.Set(MaxPowLimit)
	//}
	nextBits := BigToCompact(newTarget)

	if printInfo != nil {
		actual_t, target_t := uint64(actualTimespan.Seconds()), uint64(powTargetTimespan.Seconds())
		printStr := fmt.Sprintf("calculate next work target difficulty at height %d == %ds/%ds == %ds/%ds == %d->%d ==\n",
			currentHeight,
			actual_t/changeblocknum,
			target_t/changeblocknum,
			actual_t,
			target_t,
			currentBits,
			nextBits)
		*printInfo = printStr
	}

	return nextBits
}
