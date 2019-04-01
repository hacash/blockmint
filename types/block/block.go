package block

import . "github.com/hacash/blockmint/types/state"

type Block interface {

	// 序列化 与 反序列化
	Serialize() ([]byte, error)
	Parse([]byte, uint32) (uint32, error)
	Size() uint32

	SerializeHead() ([]byte, error)
	ParseHead([]byte, uint32) (uint32, error)

	SerializeBody() ([]byte, error)
	ParseBody([]byte, uint32) (uint32, error)

	SerializeMeta() ([]byte, error)
	ParseMeta([]byte, uint32) (uint32, error)

	SerializeTransactions(SerializeTransactionsIterator) ([]byte, error)
	ParseTransactions([]byte, uint32) (uint32, error)

	ParseExcludeTransactions([]byte, uint32) (uint32, error)

	// 修改 / 恢复 状态数据库
	ChangeChainState(ChainStateOperation) error
	RecoverChainState(ChainStateOperation) error

	// HASH
	Hash() []byte
	HashFresh() []byte
	Fresh() // 刷新所有缓存数据

	GetTransactions() []Transaction
	AddTransaction(Transaction)

	GetHeight() uint64
	GetDifficulty() uint32
	GetNonce() uint32
	GetPrevHash() []byte
	GetTimestamp() uint64
	GetTransactionCount() uint32
	GetMrklRoot() []byte

	SetMrklRoot([]byte)
	SetNonce(uint32)

	// 验证需要的签名
	VerifyNeedSigns() (bool, error)
}

type SerializeTransactionsIterator interface {
	Init(uint32)
	FinishOneTrs(uint32, Transaction, []byte)
}
