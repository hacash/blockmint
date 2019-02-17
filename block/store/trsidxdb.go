package store

// 交易索引 database

type TrsIdxOneItem struct {
	BlockHeadInfoPtrNumber uint32        // 区块头信息指针位置
	location               BlockLocation // 在区块文件中的位置

}

//////////////////////////////////////////////////////

type TrsIdxDB struct {
}
