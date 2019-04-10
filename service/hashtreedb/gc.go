package hashtreedb

import (
	"encoding/binary"
	"github.com/hacash/blockmint/sys/err"
	"github.com/hacash/blockmint/sys/file"
	"os"
	"path"
)

type GarbageCollectionDB struct {
	filepath string
	file     *os.File

	content []uint32
}

func openGcFile(filepath string) (*os.File, error) {
	return os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0777) // |os.O_TRUNC =清空
}

func NewGarbageCollectionDB(filepath string) (*GarbageCollectionDB, error) {
	file.CreatePath(path.Dir(filepath))
	curfile, fe := openGcFile(filepath)
	if fe != nil {
		return nil, fe
	}
	that := &GarbageCollectionDB{
		filepath: filepath,
		file:     curfile,
	}
	defer func() {
		that.file.Close()
		that.file = nil
	}()
	// 读取文件
	seek := int64(0)
	fstat, _ := curfile.Stat()
	fizs := fstat.Size()
	if fizs == 0 {
		return that, nil
	}
	that.content = make([]uint32, 0)
	for {
		sglen := int64(4 * 1024)
		if fizs < seek+sglen {
			sglen = fizs - seek
		}
		rdseg := make([]byte, sglen)
		rdlen, e1 := curfile.ReadAt(rdseg, seek)
		//fmt.Println(rdlen)
		//fmt.Println(rdseg)
		if e1 != nil {
			return nil, e1
		}
		if rdlen%4 != 0 {
			return nil, err.New("file size error")
		}
		rdseg = rdseg[0:rdlen]
		addctn := make([]uint32, rdlen/4)
		for i := 0; i < rdlen/4; i++ {
			//fmt.Println(rdseg[i*4:i*4+4])
			addctn[i] = binary.BigEndian.Uint32(rdseg[i*4 : i*4+4])
		}
		that.content = append(that.content, addctn...)
		if int64(rdlen) < fizs {
			break
		}
		seek += sglen
	}
	// ok
	return that, nil
}

// 收集一个空间
func (this *GarbageCollectionDB) Collect(ptrnum uint32) error {
	curlen := len(this.content)
	valueadd := make([]byte, 4)
	binary.BigEndian.PutUint32(valueadd, ptrnum)
	var e error
	this.file, e = openGcFile(this.filepath)
	if e != nil {
		return e
	}
	defer func() {
		this.file.Close()
		this.file = nil
	}()
	_, e2 := this.file.WriteAt(valueadd, int64(curlen)*4)
	if e2 != nil {
		return e2
	}
	this.content = append(this.content, ptrnum)
	return nil
}

// 取回一个空间
func (this *GarbageCollectionDB) Release() (uint32, bool, error) {
	//fmt.Println(this)
	//fmt.Println(this.content)
	curlen := len(this.content)
	if curlen == 0 {
		return 0, false, err.New("no space")
	}
	rt := this.content[curlen-1]
	var e error
	this.file, e = openGcFile(this.filepath)
	if e != nil {
		return 0, false, e
	}
	defer func() {
		this.file.Close()
		this.file = nil
	}()
	e = this.file.Truncate(int64(curlen-1) * 4) // 截断文件
	if e != nil {
		return 0, false, e
	}
	// ok
	if curlen-1 > 0 {
		this.content = this.content[0 : curlen-1]
	} else {
		this.content = make([]uint32, 0)
	}
	return rt, true, nil
}

// 取回一个空间
func (this *GarbageCollectionDB) Close() error {
	if this.file != nil {
		this.file.Close()
	}
	return nil
}
