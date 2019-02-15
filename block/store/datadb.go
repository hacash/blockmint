package store

var (
	maxPartFileSize = 200 // MB

)

type BlockDataDB struct {
	filepath string
}

func (db *BlockDataDB) InitForTest(filepath string) {

}

func (db *BlockDataDB) SaveBlockByte(filepath string) {

}
