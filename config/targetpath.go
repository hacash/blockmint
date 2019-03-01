package config

import (
	"math/rand"
	"os"
	"path"
	"strings"
)

func dealHomeDirBase(dir string) string {
	if strings.HasPrefix(dir, "~/") {
		return strings.Replace(dir, "~", os.Getenv("HOME"), 1)
	} else {
		return dir
	}
}

func GetCnfPathBlocks() string {
	base := dealHomeDirBase(Config.Datadir)
	blocks := path.Join(base, DirDataBlock)
	return blocks
}

func GetCnfPathChainState() string {
	base := dealHomeDirBase(Config.Datadir)
	states := path.Join(base, DirDataChainState)
	return states
}

func GetCnfPathMinerState() string {
	base := dealHomeDirBase(Config.Datadir)
	states := path.Join(base, DirDataMinerState)
	return states
}
func GetCnfPathTemporaryState() string {
	base := dealHomeDirBase(Config.Datadir)
	states := path.Join(base, DirDataTemporaryState)
	return states
}
func GetCnfPathNodes() string {
	base := dealHomeDirBase(Config.Datadir)
	nodes := path.Join(base, DirDataNodes)
	return nodes
}

func GetRandomMinerRewardAddress() string {
	length := len(Config.Miner.Rewards)
	if length == 0 {
		panic("Miner Reward Address must be give at lest one !")
	}
	idx := rand.Intn(length)
	//fmt.Println(idx)
	return MinerRewardAddress[idx]
}
