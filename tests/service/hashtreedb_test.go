package service

import (
	"encoding/hex"
	"github.com/hacash/blockmint/service/hashtreedb"
	"os"
	"testing"
)

func Test_d1(t *testing.T) {

	testdir := "/media/yangjie/500GB/Hacash/src/github.com/hacash/blockmint/tests/service/datas"
	os.Remove(testdir)

	db1 := hashtreedb.NewHashTreeDB(testdir, 80)

	db1.FilePartitionLevel = 7
	db1.MenuWide = 25

	hash1, _ := hex.DecodeString("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	db1.CreateQuery(hash1)

}
