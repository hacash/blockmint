package coin

import (
	"github.com/hacash/blockmint/block/fields"
)

/**
 * 货币发行算法： 前 66 年总量 2200 万枚
 */
func BlockCoinBaseRewardNumber(blockHeight uint64) uint8 {
	num := uint8(1)
	part1 := []uint8{1, 1, 2, 3, 5, 8}
	part2 := []uint8{8, 5, 3, 2, 1, 1}
	part3 := uint8(1)
	tbn1 := uint64(10000 * 10)
	tbn2 := uint64(10000 * 100)
	spx1 := uint64(len(part1)) * tbn1
	spx2 := uint64(len(part2))*tbn2 + spx1
	if blockHeight <= spx1 {
		base := blockHeight
		num = part1[base/tbn1]
	} else if blockHeight <= spx2 {
		base := blockHeight - spx1
		num = part2[base/tbn2]
	} else {
		num = part3
	}
	return num
}

func BlockCoinBaseReward(blockHeight uint64) *fields.Amount {
	num := BlockCoinBaseRewardNumber(blockHeight)
	return fields.NewAmountNumSmallCoin(num)
}
