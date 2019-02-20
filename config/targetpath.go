package config

import (
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
