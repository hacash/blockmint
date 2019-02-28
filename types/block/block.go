package block

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

	// HASH
	Hash() []byte
	HashFresh() []byte

	GetTransactions() []Transaction
	AddTransaction(Transaction)

	GetHeight() uint64
	GetDifficulty() uint32
	GetPrevHash() []byte

	SetMrklRoot([]byte)
	SetNonce(uint32)
}

type SerializeTransactionsIterator interface {
	Init(uint32)
	FinishOneTrs(uint32, Transaction, []byte)
}
