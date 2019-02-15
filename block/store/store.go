package store

import (
	"github.com/hacash/blockmint/config"
	"github.com/hacash/blockmint/types/block"
	"github.com/syndtr/goleveldb/leveldb"
)

type BlocksDataStore struct {
	db *leveldb.DB
}

var blockStoreInstance *BlocksDataStore = nil

func GetBlocksDataStoreInstance() *BlocksDataStore {
	if blockStoreInstance != nil {
		return blockStoreInstance
	}
	blockStoreInstance = new(BlocksDataStore)
	blockStoreInstance.Init(config.GetCnfPathBlocks())
	return blockStoreInstance
}

//////////////

func (sto *BlocksDataStore) Init(dir string, idxdir string) error {

	db, err := leveldb.OpenFile(idxdir, nil)
	if err != nil {
		return err
	}
	sto.db = db
	return nil
}

func (sto *BlocksDataStore) Save(hash []byte, height uint64, body *[]byte) error {

	return nil
}

func (sto *BlocksDataStore) ReadHead(hash []byte) block.Block {

	return nil
}
func (sto *BlocksDataStore) ReadHeadByHeight(height uint64) []block.Block {

	return nil
}

func (sto *BlocksDataStore) ReadBody(hash []byte) block.Block {

	return nil
}

func (sto *BlocksDataStore) ReadBodyByHeight(height uint64) []block.Block {

	return nil
}
