package miner

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/blockmint/miner/difficulty"
	"os"
	"path"
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
		data1, _ := hex.DecodeString(hsxstrary[i])
		bignum := difficulty.HashToBig(&data1)
		num32 := difficulty.BigToCompact(bignum)
		fmt.Println(hsxstrary[i]+" difficulty value is", num32)
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
