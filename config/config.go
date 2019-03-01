package config

import (
	"fmt"
	"github.com/jinzhu/configor"
	"strings"
)

var Config = struct {
	Datadir string `default:"~/.hacash"` // 数据目录

	Miner struct {
		Minfeeratio string   `default:"1ww"` // 接受的最小手续费比例
		Rewards     []string // 矿工奖励地址
	}
}{}

func LoadConfigFile() {

	configor.Load(&Config, "hacash.config.yml")
	fmt.Printf("config: %#v", Config)
	// handle
	Config.Miner.Minfeeratio = strnumdeal(Config.Miner.Minfeeratio)
}

func strnumdeal(in string) string {
	in = strings.Replace(in, "k", "000", -1)
	in = strings.Replace(in, "w", "0000", -1)
	in = strings.Replace(in, "m", "000000", -1)
	return in
}
