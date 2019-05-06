package miner

type Miner interface {
	GetPrevDiamondHash() []byte // 获取当前基于的钻石区块hash
	SetPrevDiamondHash([]byte)  // 设置钻石区块hash

}
