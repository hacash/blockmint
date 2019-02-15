package config

import (
	"github.com/hacash/blockmint/sys/file"
	"os"
	"path"
	"strings"
)

func dealHomeDirBase(dir string) string {
	if strings.HasPrefix(dir, "~/") {
		return strings.Replace(dir, "~/", os.Getenv("HOME"), 1)
	} else {
		return dir
	}
}

func GetCnfPathBlocks() string {
	base := dealHomeDirBase(DirBase)
	blocks := path.Join(base, DirDataBlock)
	file.CreatePath(blocks)
	return blocks
}
