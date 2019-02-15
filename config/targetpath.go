package config

import (
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

func GetCnfPathBlocks() (string, string) {
	base := dealHomeDirBase(DirBase)
	blocks := path.Join(base, DirDataBlock)
	indexs := path.Join(blocks, "indexs/")
	createPath(indexs)
	return blocks, indexs
}

/*/////////////////////////////////////*/

func isExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func createPath(filePath string) error {
	if !isExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}
