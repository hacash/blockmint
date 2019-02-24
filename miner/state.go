package miner

type MinerState struct {
	PrevBlockHash []byte // 上一个区块的哈希

	TargetDifficulty uint32 // 挖矿目标难度值

}
