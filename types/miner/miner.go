package miner

type Miner interface {
	GetPrevDiamondHash() (uint32, []byte) // 获取当前基于的钻石区块hash
	SetPrevDiamondHash(uint32, []byte)    // 设置钻石区块hash

}
