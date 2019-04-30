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
)

var Config = struct {
	Datadir string `default:"~/.hacash"` // 数据目录

	Loglevel string `default:"News"`

	Miner struct {
		Forcestart  string   `default:"false"` // 启动时强制开始挖矿
		Minfeeratio string   `default:"1Y"`    // 接受的最小手续费比例
		Rewards     []string // 矿工奖励地址
		Supervene   uint64   // 启动多线程挖矿，指定线程数量（数量必须小于200）
		Markword    string   // 矿工寄语/标识，例如 hacash.org （不超过 15位）
		// 慎重参数
		Backtoheight  uint64 // state 数据状态退回到指定区块高低
		Stepsleepnano string `default:"1KK"` // 矿工单次计算后休眠时间 纳秒  1秒=1000*1000*1000纳秒
	}

	P2p struct {
		Port struct {
			Node string `default:"3337"`
			Rpc  string `default:"3338"`
		}
		Myname     string   `default:""`
		Maxpeernum uint64   // 连接节点数量上线
		Bootnodes  []string // 起始节点
	}

	DiamondMiner struct {
		PrevBlockHash string   `default:""`
		Rewards       []string //
		Supervene     uint64   // 启动多线程挖矿，指定线程数量（数量必须小于200）
	}


	GpuMiner struct {
		Address string   `default:""` // 矿工地址
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
			cnffile = fn // 尝试打开配置文件
			f.Close()
		}
	}
	fmt.Printf("load config file \"%s\"\n", cnffile)
	// 加载配置
	configor.Load(&Config, cnffile)
	//fmt.Printf("config: %#v\n\n", Config)
	// handle
	Config.Miner.Minfeeratio = strnumdeal(Config.Miner.Minfeeratio)
	Config.Miner.Stepsleepnano = strnumdeal(Config.Miner.Stepsleepnano)
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
