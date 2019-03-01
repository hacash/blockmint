package config

import (
	"github.com/jinzhu/configor"
	"math/rand"
	"strconv"
	"strings"
)

var Config = struct {
	Datadir string `default:"~/.hacash"` // 数据目录

	Miner struct {
		Minfeeratio string   `default:"1y"` // 接受的最小手续费比例
		Rewards     []string // 矿工奖励地址
	}

	P2p struct {
		Port struct {
			Node string `default:"3337"`
			Rpc  string `default:"3338"`
		}
		Myname    string   `default:""`
		Bootnodes []string // 起始节点
	}
}{}

func LoadConfigFile() {

	configor.Load(&Config, "hacash.config.yml")
	//fmt.Printf("config: %#v\n\n", Config)
	// deal
	MainnetBootnodes = append(MainnetBootnodes, Config.P2p.Bootnodes...)
	if Config.P2p.Myname == "" {
		Config.P2p.Myname = "hacash_node_" + strconv.FormatUint(rand.Uint64(), 10)
	}
	// handle
	Config.Miner.Minfeeratio = strnumdeal(Config.Miner.Minfeeratio)
}

func strnumdeal(in string) string {
	in = strings.Replace(in, "H", "00", -1)       // 百
	in = strings.Replace(in, "K", "000", -1)      // 千
	in = strings.Replace(in, "W", "0000", -1)     // 万
	in = strings.Replace(in, "M", "000000", -1)   // 百万
	in = strings.Replace(in, "Y", "00000000", -1) // 亿
	return in
}
