package pool

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/hacash/blockmint/block/fields"
	"github.com/hacash/blockmint/sys/file"
	"github.com/hacash/blockmint/types/block"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"sync"
)

type TransferRecordData struct {
	Latest uint64 // 最新加入的交易序号
	Submit uint64 // 已经提交的交易序号

	TxLatestId uint64 // 实际交易的id
	TxConfirm  uint64 // 已经确认的交易序号

	PrevSendHeight uint64 // 上次打币的区块
}

type TransferData struct {
	Id uint64 //转账ID

	TxId            uint64         // 实际交易的id
	CreateTimestamp uint64         // 转账生成时间
	SubmitTimestamp uint64         // 转账提交时间
	Amount          uint64         // 转账数量 单位:240
	Address         fields.Address // 转账地址
}

type Store struct {
	DirBase string
	db      *leveldb.DB

	// 毒蜥读写锁
	workerLock         sync.Mutex
	transferRecordLock sync.Mutex
}

func NewStore(dirbase string) *Store {
	dbdir := dirbase + "/leveldb_minerpool"
	file.CreatePath(dbdir)
	db, e := leveldb.OpenFile(dbdir, nil)
	if e != nil {
		panic(e)
	}
	return &Store{
		DirBase: dirbase,
		db:      db,
	}
}

func (s *Store) SaveTransactionBody(id uint64, trs block.Transaction) error {
	key := []byte(fmt.Sprintf("txbody:%d", id))
	bts, e1 := trs.Serialize()
	if e1 != nil {
		return e1
	}
	return s.db.Put(key, bts, nil)
}
func (s *Store) ReadTransactionBody(id uint64) []byte {
	key := []byte(fmt.Sprintf("txbody:%d", id))
	body, e := s.db.Get(key, nil)
	if e != nil {
		return nil
	}
	return body
}

////////////////////////////////

// 转账交易
func (s *Store) ReadTransfer(id uint64) *TransferData {
	// 读取
	key := []byte(fmt.Sprintf("transfer:%d", id))
	valbts, notfind := s.db.Get(key, nil)
	if notfind != nil {
		return nil
	}
	var td = &TransferData{Id: id}
	td.TxId = binary.BigEndian.Uint64(valbts[0:8])
	td.CreateTimestamp = binary.BigEndian.Uint64(valbts[8:16])
	td.SubmitTimestamp = binary.BigEndian.Uint64(valbts[16:24])
	td.Amount = binary.BigEndian.Uint64(valbts[24:32])
	td.Address = fields.Address(valbts[32:53])
	return td
}
func (s *Store) SaveTransfer(td *TransferData) error {
	if td == nil {
		return fmt.Errorf("TransferData is nil")
	}
	// 保存
	key := []byte(fmt.Sprintf("transfer:%d", td.Id))
	var savebts = make([]byte, 4*8+21)
	binary.BigEndian.PutUint64(savebts[0:8], td.TxId)
	binary.BigEndian.PutUint64(savebts[8:16], td.CreateTimestamp)
	binary.BigEndian.PutUint64(savebts[16:24], td.SubmitTimestamp)
	binary.BigEndian.PutUint64(savebts[24:32], td.Amount)
	copy(savebts[32:53], td.Address)
	return s.db.Put(key, savebts, nil)
}

//////////////////////////////

// 转账记录
func (s *Store) ReadTransferRecord(lock bool) *TransferRecordData {
	// 读写锁
	if lock {
		s.transferRecordLock.Lock()
	} else {
		s.transferRecordLock.Lock()
		defer s.transferRecordLock.Unlock()
	}
	// 读取
	var rec = &TransferRecordData{0, 0, 0, 0, 0}
	key := []byte("transfer_record")
	valbts, notfind := s.db.Get(key, nil)
	if notfind != nil {
		return rec
	}
	rec.Latest = binary.BigEndian.Uint64(valbts[0:8])
	rec.Submit = binary.BigEndian.Uint64(valbts[8:16])
	rec.TxLatestId = binary.BigEndian.Uint64(valbts[16:24])
	rec.TxConfirm = binary.BigEndian.Uint64(valbts[24:32])
	rec.PrevSendHeight = binary.BigEndian.Uint64(valbts[32:40])
	return rec
}
func (s *Store) SaveTransferRecord(rec *TransferRecordData, lock bool) error {
	// 读写锁
	// 读写锁
	if lock {
		s.transferRecordLock.Unlock()
	} else {
		s.transferRecordLock.Lock()
		defer s.transferRecordLock.Unlock()
	}
	// 保存
	key := []byte("transfer_record")
	var savebts = make([]byte, 5*8)
	binary.BigEndian.PutUint64(savebts[0:8], rec.Latest)
	binary.BigEndian.PutUint64(savebts[8:16], rec.Submit)
	binary.BigEndian.PutUint64(savebts[16:24], rec.TxLatestId)
	binary.BigEndian.PutUint64(savebts[24:32], rec.TxConfirm)
	binary.BigEndian.PutUint64(savebts[32:40], rec.PrevSendHeight)
	return s.db.Put(key, savebts, nil)
}

///////////////////////////////////////

// 读取
func (s *Store) ReadWorker(addr *fields.Address) *PowWorker {
	// 读写锁
	s.workerLock.Lock()
	defer s.workerLock.Unlock()
	// 读取
	key := []byte("worker:" + string(*addr))
	val, notfind := s.db.Get(key, nil)
	if notfind != nil {
		return nil
	}
	var data AddressStatisticsStoreItem
	seek, e1 := data.Parse(val, 0)
	if e1 != nil {
		return nil
	}
	// fmt.Println("ReadWorker", seek, val, val[seek:], big.NewInt(0).SetBytes(val[seek:]).String())
	return &PowWorker{
		StatisticsData:          &data,
		RewordAddress:           addr,
		RealtimePower:           big.NewInt(0).SetBytes(val[seek:]), // 历史保存的
		RealtimeWorkSubmitCount: 0,
		ClientCount:             0,
	}
}

// 储存
func (s *Store) SaveWorker(wk *PowWorker) error {
	// 读写锁
	s.workerLock.Lock()
	defer s.workerLock.Unlock()
	// 保存
	key := []byte("worker:" + string(*wk.RewordAddress))
	val, _ := wk.StatisticsData.Serialize()
	buf := bytes.NewBuffer(val)
	// fmt.Println("SaveWorker", wk.RealtimePower.Bytes())
	buf.Write(wk.RealtimePower.Bytes()) // 保存挖矿记录
	return s.db.Put(key, buf.Bytes(), nil)
}
