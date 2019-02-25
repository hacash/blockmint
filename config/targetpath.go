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
	base := dealHomeDirBase(DirBase)
	blocks := path.Join(base, DirDataBlock)
	return blocks
}

func GetCnfPathChainState() string {
	base := dealHomeDirBase(DirBase)
	states := path.Join(base, DirDataChainState)
	return states
}

func GetCnfPathMinerState() string {
	base := dealHomeDirBase(DirBase)
	states := path.Join(base, DirDataMinerState)
	return states
}
func GetCnfPathTemporaryState() string {
	base := dealHomeDirBase(DirBase)
	states := path.Join(base, DirDataTemporaryState)
	return states
}

func GetRandomMinerRewardAddress() string {
	length := len(MinerRewardAddress)
	idx := rand.Intn(length)
	return MinerRewardAddress[idx]
}
