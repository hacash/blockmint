package hashtreedb

import (
	"github.com/hacash/blockmint/sys/file"
	"os"
	"path"
)

// 不安全的迭代器
type UnsafeIteration struct {
	sigmentSize uint32
}

// 迭代器
func (this *HashTreeDB) CreateUnsafeIteration() *UnsafeIteration {

	if this.FilePartitionLevel > 0 {
		panic("unsupported operations for TraversalCopy: FilePartitionLevel must be 0")
	}
	filename := this.getPartFileName([]byte{})
	file.CreatePath(path.Dir(filename))
	_, fe := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
	if fe != nil {
		panic("unsupported operations for TraversalCopy: file '" + filename + "' must be existence")
	}

	return nil
}
