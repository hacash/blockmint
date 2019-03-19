package hashtreedb

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/blockmint/sys/file"
	"os"
	"path"
	"strconv"
	"sync"
)

// 单个文件大小至少支持 256^4×5×8 MenuWide=8 时约 80GB
type HashTreeDB struct {
	HashSize   uint32 // 哈希大小 16,32,64,128,256
	KeyReverse bool   // key值倒序

	MaxValueSize uint32 // 最大数据尺寸大小 + hash32

	MenuWide uint8 // 单层索引宽度数（不可超过256）

	FilePartitionLevel uint32 // 文件分区层级 0为不分区

	FileAbsPath string // 文件的储存路径
	FileName    string // 保存文件的名称
	FileSuffix  string // 保存文件后缀名 .idx

	DeleteMark bool // 删除也会保存key标记
	//gc *GarbageCollectionDB
	OpenGc       bool                            // 是否开启gc
	gcPool       map[string]*GarbageCollectionDB // gc管理器
	MaxNumGCPool int

	// fileLock
	FileLock map[string]*sync.Mutex
}

// 创建DataBase
func NewHashTreeDB(FileAbsPath string, MaxValueSize uint32, HashSize uint32) *HashTreeDB {

	menuWide := (MaxValueSize+HashSize)/IndexItemSize + 1 // 最优空间
	if (MaxValueSize+HashSize)%IndexItemSize == 0 {
		menuWide -= 1 // 刚好合适
	}
	if menuWide > 255 {
		panic("menuWide cannot over 255")
	}

	return &HashTreeDB{
		HashSize:           HashSize,
		KeyReverse:         false,
		MenuWide:           uint8(menuWide),
		FilePartitionLevel: 0,
		FileName:           "INDEX",
		FileSuffix:         ".idx",
		MaxNumGCPool:       64,
		OpenGc:             true,
		gcPool:             make(map[string]*GarbageCollectionDB),

		FileAbsPath:  FileAbsPath,
		MaxValueSize: MaxValueSize,

		FileLock: make(map[string]*sync.Mutex),
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
	//
	keyhash := hash
	if this.KeyReverse {
		keyhash = ReverseHashOrder(hash) // 倒序
	}
	filename := this.getPartFileName(keyhash)
	// 文件操作锁
	lock, has := this.FileLock[filename]
	if !has {
		lock = new(sync.Mutex)
		this.FileLock[filename] = lock
	}
	//fmt.Println("LOCK FILE - "+filename)
	lock.Lock()
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
	return NewQueryInstance(this, hash, this.SpreadTheKey(keyhash), this.GetQueryHashKey(hash), curfile, &filename), nil
}

// 建立数据操作
func (this *HashTreeDB) ReadBytesByPosition(keyprefix []byte, ptrnum uint32) ([]byte, error) {
	segsize := this.getSegmentSize()
	return this.ReadBytesByPositionWithLength(keyprefix, ptrnum, int64(segsize))
}

// 建立数据操作
func (this *HashTreeDB) ReadBytesByPositionWithLength(keyprefix []byte, ptrnum uint32, length int64) ([]byte, error) {

	//fmt.Println(keyprefix)

	filename := this.getPartFileName(keyprefix)
	//fmt.Println(filename)
	//fmt.Println(path.Dir(filename))
	//file.CreatePath(path.Dir(filename))
	// 打开相应文件
	curfile, fe := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	if fe != nil {
		return nil, fe
	}
	// 读取内容
	segsize := this.getSegmentSize()
	var valuebody = make([]byte, length)
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
	return IndexItemSize * uint32(this.MenuWide)
}

// 展开key
func (this *HashTreeDB) SpreadTheKey(key []byte) []byte {
	var queryKey bytes.Buffer
	for _, v := range key {
		num := uint8(v)
		count := uint8(0)
		for {
			if num/this.MenuWide == 0 {
				queryKey.Write([]byte{count, num})
				break
			} else {
				count++
				if count+1 == this.MenuWide {
					queryKey.Write([]byte{count})
					count = 0
				}
				//queryKey.Write([]byte{this.db.MenuWide-1})
				num -= this.MenuWide
			}
		}
	}
	bts := queryKey.Bytes()
	return bts // ReverseHashOrder(bts) // 倒序
}

// Segment Size
func (this *HashTreeDB) GetQueryHashKey(hash []byte) []byte {
	hashkey := this.SpreadTheKey(hash)
	return hashkey[this.FilePartitionLevel:]
}

// 获取打开的文件名
func (this *HashTreeDB) getPartFileName(hash []byte) string {
	return this.getPartFileNameEx(hash, this.FileSuffix)
}

// 获取打开的文件名
func (this *HashTreeDB) getPartFileNameEx(hash []byte, ffix string) string {
	hash = this.SpreadTheKey(hash)

	var partPath = "" // 路径分区
	var partNum = ""  // 文件编号

	if this.FilePartitionLevel > 0 {

		var lv = int(this.FilePartitionLevel) - 1
		var bsm = int(this.MenuWide)

		partNum = strconv.Itoa(int(hash[lv]) % bsm)

		var mergedir = 1
		var mw = uint32(this.MenuWide)
		var mwd = uint32(mw)
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

// 遍历拷贝
// 只能是单文件数据库
func (this *HashTreeDB) TraversalCopy(get *HashTreeDB) error {
	if get.FilePartitionLevel > 0 {
		panic("unsupported operations for TraversalCopy: FilePartitionLevel must be 0")
	}
	filename := get.getPartFileName([]byte{})
	file.CreatePath(path.Dir(filename))
	curfile, fe := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
	if fe != nil {
		panic("unsupported operations for TraversalCopy: file '" + filename + "' must be existence")
	}
	// do copy
	recursTraversalCopy(curfile, 0, get.getSegmentSize(), get, this.doCallTraversalCopy)
	return nil
}

func (this *HashTreeDB) doCallTraversalCopy(ty uint8, itembytes []byte, get *HashTreeDB) {
	if ty == 0 {
		return // do nothing
	}
	if ty != 2 && ty != 3 {
		return // do nothing
	}
	if len(itembytes) != int(get.getSegmentSize()) {
		return // do nothing
	}
	key := itembytes[0:get.HashSize]
	query, e := this.CreateQuery(key)
	if e != nil {
		return // do nothing
	}
	defer query.Close()
	if ty == 2 {
		// copy
		//fmt.Println(get.HashSize)
		//fmt.Println(len(itembytes) - int(get.HashSize))
		//fmt.Println("copy save " + hex.EncodeToString(key) + "(" +address.NewAddressReadableFromAddress(key) + ") => " + hex.EncodeToString(itembytes[get.HashSize:]))
		query.Save(itembytes[get.HashSize:])
	} else if ty == 3 {
		// delete
		query.Save([]byte{}) // save empty = delete mark
	}
}

func recursTraversalCopy(file *os.File, fileseek int64, segmentSize uint32, get *HashTreeDB, docall func(uint8, []byte, *HashTreeDB)) {
	segment := make([]byte, segmentSize)
	rdlen, e := file.ReadAt(segment, fileseek)
	if e != nil || uint32(rdlen) != segmentSize {
		return // end
	}
	// down
	for i := 0; i < int(segmentSize/5); i++ {
		start := i * 5
		//fmt.Println(segment)
		//fmt.Println(len(segment))
		//fmt.Println(start)
		//fmt.Println(start+1+4)
		ty := segment[start]
		fileseek = int64(binary.BigEndian.Uint32(segment[start+1:start+1+4])) * int64(segmentSize)
		if ty == 1 {
			recursTraversalCopy(file, fileseek, segmentSize, get, docall)
			continue
		} else {
			itembytes := []byte{}
			if fileseek > 0 {
				itembytes = make([]byte, segmentSize)
				rdlen, e := file.ReadAt(itembytes, fileseek)
				if e != nil && rdlen != int(segmentSize) {
					itembytes = []byte{}
				}
			}
			//fmt.Println("segmentSize :: ", segmentSize)
			//fmt.Println("segmentSize :: ", hex.EncodeToString(itembytes))
			docall(ty, itembytes[:], get)
		}
	}

}
