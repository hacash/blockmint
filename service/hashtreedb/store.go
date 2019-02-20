package hashtreedb

import (
	"github.com/hacash/blockmint/sys/file"
	"os"
	"path"
	"strconv"
)

// 单个文件大小至少支持 256^4×5×8 MenuWide=8 时约 80GB
type HashTreeDB struct {
	HashSize   uint32 // 哈希大小 16,32,64,128,256
	KeyReverse bool   // key值倒序

	MaxValueSize uint32 // 最大数据尺寸大小 + hash32

	MenuWide uint32 // 单层索引宽度数（不可超过256）

	FilePartitionLevel uint32 // 文件分区层级 0为不分区

	FileAbsPath string // 文件的储存路径
	FileName    string // 保存文件的名称
	FileSuffix  string // 保存文件后缀名 .idx

	//gc *GarbageCollectionDB
	gcPool       map[string]*GarbageCollectionDB // gc管理器
	MaxNumGCPool int
}

// 创建DataBase
func NewHashTreeDB(FileAbsPath string, MaxValueSize uint32, HashSize uint32) *HashTreeDB {

	menuWide := (MaxValueSize+HashSize)/IndexItemSize + 1 // 最优空间

	return &HashTreeDB{
		HashSize:           HashSize,
		KeyReverse:         false,
		MenuWide:           menuWide,
		FilePartitionLevel: 0,
		FileName:           "INDEX",
		FileSuffix:         ".idx",
		MaxNumGCPool:       64,
		gcPool:             make(map[string]*GarbageCollectionDB),

		FileAbsPath:  FileAbsPath,
		MaxValueSize: MaxValueSize,
	}
}

// 建立数据操作
func (this *HashTreeDB) GetGcService(keyhash []byte) (*GarbageCollectionDB, error) {
	gcfile := this.getPartFileNameEx(keyhash, ".gc")
	gc, got := this.gcPool[gcfile]
	if got {
		return gc, nil
	}
	if len(this.gcPool) >= this.MaxNumGCPool {
		// remove one
		for k := range this.gcPool {
			this.gcPool[k].Close()
			delete(this.gcPool, k)
			break
		}
	}
	// create
	gc, e := NewGarbageCollectionDB(gcfile)
	if e != nil {
		panic("NewGarbageCollectionDB error ！")
	}
	this.gcPool[gcfile] = gc
	return gc, nil
}

// 建立数据操作
func (this *HashTreeDB) CreateQuery(hash []byte) (*QueryInstance, error) {
	keyhash := hash
	if this.KeyReverse {
		keyhash = ReverseHashOrder(hash) // 倒序
	}
	filename := this.getPartFileName(keyhash)
	//fmt.Println(hash)
	//fmt.Println(keyhash)
	//fmt.Println(filename)
	//fmt.Println(path.Dir(filename))
	file.CreatePath(path.Dir(filename))
	// 打开相应文件
	curfile, fe := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	if fe != nil {
		return nil, fe
	}
	// 返回
	return NewQueryInstance(this, hash, keyhash, this.GetQueryHashKey(hash), curfile), nil
}

// 建立数据操作
func (this *HashTreeDB) ReadBytesByPosition(keyprefix []byte, ptrnum uint32) ([]byte, error) {

	filename := this.getPartFileName(keyprefix)
	//fmt.Println(filename)
	//fmt.Println(path.Dir(filename))
	file.CreatePath(path.Dir(filename))
	// 打开相应文件
	curfile, fe := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	if fe != nil {
		return nil, fe
	}
	// 读取内容
	segsize := this.getSegmentSize()
	var valuebody = make([]byte, segsize)
	_, e1 := curfile.ReadAt(valuebody, int64(segsize)*int64(ptrnum))
	if e1 != nil {
		return nil, e1
	}
	curfile.Close()
	// ok
	return valuebody, nil
}

// Segment Size
func (this *HashTreeDB) getSegmentSize() uint32 {
	return IndexItemSize * this.MenuWide
}

// Segment Size
func (this *HashTreeDB) GetQueryHashKey(hash []byte) []byte {
	hashkey := hash
	if this.KeyReverse {
		hashkey = ReverseHashOrder(hash) // 倒序
	}
	return hashkey[this.FilePartitionLevel:]
}

// 获取打开的文件名
func (this *HashTreeDB) getPartFileName(hash []byte) string {
	return this.getPartFileNameEx(hash, this.FileSuffix)
}

// 获取打开的文件名
func (this *HashTreeDB) getPartFileNameEx(hash []byte, ffix string) string {

	var partPath = "" // 路径分区
	var partNum = ""  // 文件编号

	if this.FilePartitionLevel > 0 {

		var lv = int(this.FilePartitionLevel) - 1
		var bsm = int(this.MenuWide)

		partNum = strconv.Itoa(int(hash[lv]) % bsm)

		var mergedir = 1
		var mw = this.MenuWide
		var mwd = mw
		var filenummax = uint32(20000)
		for true {
			mwd = mwd * mw
			if mwd > filenummax {
				break
			}
			mergedir += 1
		}

		var loop = int(lv)/mergedir + 1
		for i := 0; i < loop; i++ {
			if len(partPath) > 0 {
				partPath += "/"
			}
			var seg = ""
			for k := 0; k < mergedir; k++ {
				a := i*mergedir + k
				if a >= lv {
					break
				}
				if len(seg) > 0 {
					seg += "-"
				}
				seg += strconv.Itoa(int(hash[a]) % bsm)
			}
			partPath += seg
		}

	}

	return path.Join(this.FileAbsPath, partPath, this.FileName+partNum+ffix)
}