package file

import "os"

/*/////////////////////////////////////*/

func IsExist(path string) bool {
	_, err := os.Stat(path) //os.Stat 获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func CreatePath(filePath string) error {
	if !IsExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}
