package config

import (
	"fmt"
	"github.com/jinzhu/configor"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (

	// min fee
	MinimalFeePurity = uint64(0)

	// data dir
	DirDataBlock          = "blocks/"
	DirDataChainState     = "chainstate/"
	DirDataMinerState     = "minerstate/"
	DirDataTemporaryState = "tempstate/" // 临时状态，最好是内存文件系统
	DirDataNodes          = "nodes/"

	// consensus rule, prohibit change !
	MaximumBlockSize            = int64(1024 * 1024 * 2) // 区块最大尺寸值byte  =  2MB
	ChangeDifficultyBlockNumber = uint64(10)
	EachBlockTakesTime          = uint64(60) // 秒

)

var Config = struct {
	Datadir string `default:"~/.hacash"` // 数据目录

	Miner struct {
		Forcestart  string   `default:"false"` // 启动时强制开始挖矿
		Minfeeratio string   `default:"1Y"`    // 接受的最小手续费比例
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
	// 随机数种子
	rand.Seed(time.Now().Unix())
	// 读取配置文件路径
	cnffile := "hacash.config.yml"
	if len(os.Args) >= 2 {
		fn := os.Args[1]
		f, e := os.Open(fn)
		if e == nil {
			fmt.Printf("load config file `%s`\n", fn)
			cnffile = fn // 尝试打开配置文件
			f.Close()
		}
	}
	// 加载配置
	configor.Load(&Config, cnffile)
	//fmt.Printf("config: %#v\n\n", Config)
	// handle
	Config.Miner.Minfeeratio = strnumdeal(Config.Miner.Minfeeratio)
	// deal
	MainnetBootnodes = append(MainnetBootnodes, Config.P2p.Bootnodes...)
	if Config.P2p.Myname == "" {
		Config.P2p.Myname = "hacash_node_" + strconv.FormatUint(rand.Uint64(), 10)
	}
	feep, e1 := strconv.ParseUint(Config.Miner.Minfeeratio, 10, 0)
	if e1 != nil {
		panic("Config.Miner.Minfeeratio value format error")
	}
	MinimalFeePurity = feep
}

func strnumdeal(in string) string {
	in = strings.Replace(in, "H", "00", -1)       // 百
	in = strings.Replace(in, "K", "000", -1)      // 千
	in = strings.Replace(in, "W", "0000", -1)     // 万
	in = strings.Replace(in, "M", "000000", -1)   // 百万
	in = strings.Replace(in, "Y", "00000000", -1) // 亿
	return in
}
